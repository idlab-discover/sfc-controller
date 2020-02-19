package main

import (
	k8sApi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// The port on which the SFC controller listens for HTTP traffic.
	port = "8100"

	// The scheduler name
	schedulerName = "sfc-controller"
)

var (
	// Scheduler instance
	scheduler Scheduler

	// Service Hash IDs
	id = 0

	// Service Hash map
	serviceHash = make(map[string]string)

	//Infrastructure Locations:
	locations = [5]string{"sw-Bruges", "sw-Antwerp", "sw-Ghent", "sw-Brussels", "sw-Leuven"}

	// Graph Latency - For Dijkstra Short Path Calculation
	graphLatency = newGraph()
)

// Scheduler instance structure
type Scheduler struct {
	clientset  *kubernetes.Clientset
	podQueue   chan *k8sApi.Pod
	nodeLister listersv1.NodeLister
	podLister  listersv1.PodLister
}

// Create the Scheduler Instance
func NewScheduler(podQueue chan *k8sApi.Pod, quit chan struct{}) Scheduler {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	nodeLister, podLister := initInformers(clientset, podQueue, quit)

	return Scheduler{
		clientset:  clientset,
		podQueue:   podQueue,
		nodeLister: nodeLister,
		podLister:  podLister,
	}
}

// Create Informers
func initInformers(clientset *kubernetes.Clientset, podQueue chan *k8sApi.Pod, quit chan struct{}) (listersv1.NodeLister, listersv1.PodLister) {
	factory := informers.NewSharedInformerFactory(clientset, 0)

	nodeInformer := factory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*k8sApi.Node)
			if !ok {
				log.Println("This is not a node")
				return
			}

			// Listen to all nodes in the infrastructure
			log.Printf("New Node Added to Store: %s", node.GetName())

			// Create Graph Entry
			log.Printf("Add Node in Graph for Dijkstra Calculation...\n")
			for _, location := range locations {
				label := node.Labels[location]
				if label != "" {
					log.Printf("Location found! Add Edge in Graph...\n")
					value := stringtoInt(label)
					graphLatency.addEdge(node.GetName(), location, value)
				}
			}
			// print Graph Latency
			for _, location := range locations {
				cost, path := graphLatency.getPath(node.GetName(), location)
				if cost != 0 {
					log.Printf("Graph Entry added: %v | %v \n", cost, path)
				}
			}

		},
	})

	podInformer := factory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*k8sApi.Pod)
			if !ok {
				log.Println("This is not a pod")
				return
			} // Listen to all pods related with sfc-controller
			if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == schedulerName {
				podQueue <- pod
				log.Printf("New Pod Added to Store: %s", pod.Name)
			}
		},
	})

	factory.Start(quit)
	return nodeInformer.Lister(), podInformer.Lister()
}

func main() {
	log.Printf("SFC-controller v0.0.3 Starting...\n")

	// configure Logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//Add infrastructure Edges to Graph
	graphLatency.addEdge("sw-Bruges", "sw-Ghent", 15)
	graphLatency.addEdge("sw-Antwerp", "sw-Leuven", 15)
	graphLatency.addEdge("sw-Ghent", "sw-Brussels", 25)
	graphLatency.addEdge("sw-Brussels", "sw-Leuven", 25)

	//Create Scheduler
	podQueue := make(chan *k8sApi.Pod, 300)
	defer close(podQueue)

	quit := make(chan struct{})
	defer close(quit)

	scheduler = NewScheduler(podQueue, quit)

	// start http server
	svr := &http.Server{Addr: ":" + port}
	svr.Handler = http.HandlerFunc(handler)
	log.Printf("Extender Call started on port %v\n", port)

	if err := svr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	// WatchPods Routine
	// log.Printf("Start WatchingPods Routine...\n")
	// scheduler.Run(quit)

	// Create Channel: Live forever
	//ch := make(chan bool)
	//<-ch
}

func (scheduler *Scheduler) Run(quit chan struct{}) {
	wait.Until(scheduler.watchScheduledPods, 30*time.Second, quit)
}

func (scheduler *Scheduler) watchScheduledPods() {

	// Check if allocated Pods are still deployed. Otherwise free bandwidth on the correspondent node
	nodes := scheduler.nodeLister
	podScheduled := <-scheduler.podQueue

	log.Printf("---------------Watching Pods------------\n")
	log.Printf("Found a pod to check: %v / %v", podScheduled.Namespace, podScheduled.Name)

	pod, err := scheduler.clientset.CoreV1().Pods(podScheduled.Namespace).Get(podScheduled.Name, metav1.GetOptions{})
	if err != nil {
		// Pod is not deployed anymore!
		log.Printf("Check failed: Pod %s is not deployed anymore", podScheduled.Name)

		// Remove hash key / update node bandwidth
		appName := getDesiredFromLabels(podScheduled, "app")
		chainPosString := getDesiredFromLabels(podScheduled, "chainPosition")
		nsh := getDesiredFromLabels(podScheduled, "networkServiceHeader")
		totalChain := getDesiredFromLabels(podScheduled, "totalChainServ")
		chainPosString = strings.TrimRight(chainPosString, "pos")
		totalChain = strings.TrimRight(totalChain, "serv")

		chainPos := stringtoInt(chainPosString)
		totalChainServ := stringtoInt(totalChain)

		nodeName := ""

		//Find the correct Service Hash key and remove pod from Service Hash
		for i := 1; i <= id; i++ {
			key := getKey(i, appName, nsh, chainPos, totalChainServ)
			allocatedNode, ok := serviceHash[key]
			if ok {
				log.Printf("Pod found! Allocated on Node: %v \n", allocatedNode)
				delete(serviceHash, key)
				log.Printf("Service Hash removed...")
				nodeName = allocatedNode
			}
		}

		// Get node where the pod was allocated
		node, err := nodes.Get(nodeName)
		if err != nil {
			log.Printf("cannot find node %v", err.Error())
			return
		}

		// Get Node current Bandwidth
		nodeBand := getBandwidthValue(node, "avBandwidth")

		// Get Pod Min Bandwidth Requirement
		minBandwidth := getDesiredFromLabels(podScheduled, "minBandwidth")
		minBandwidth = strings.TrimRight(minBandwidth, "Mi")
		podMinBandwith := stringtoFloatBandwidth(minBandwidth)

		// Update current avBandwidth
		newValue := nodeBand + podMinBandwith
		label := strconv.FormatFloat(newValue, 'f', 2, 64)

		err = updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, node)
		if err != nil {
			log.Printf("encountered error when updating label after pod verification: %v", err)
			return
		}

		log.Printf("Node %v bandwidth updated. Pod %s is not deployed anymore", node.Name, podScheduled.Name)
		return
	}

	// Check confirmed
	log.Printf("Check confirmed: Pod still alocated %s", pod.Name)

	// Send Pod back to the channel
	scheduler.podQueue <- pod

	return
}
