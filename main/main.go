package main

import (
	k8sApi "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"net/http"
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

	//Infrastructure Locations: change locations according to your infrastructure!
	locations = [5]string{"sw-Bruges", "sw-Antwerp", "sw-Ghent", "sw-Brussels", "sw-Leuven"}

	// Graph Latency - For Dijkstra Short Path Calculation
	graphLatency = newGraph()

	// Linked List: allocated Pods
	allocatedPods = createPodList("scheduledPods")
)

// Scheduler instance structure
type Scheduler struct {
	clientset *kubernetes.Clientset
	//podQueue   chan *k8sApi.Pod
	nodeLister listersv1.NodeLister
	podLister  listersv1.PodLister
}

// Create the Scheduler Instance
func NewScheduler(quit chan struct{}) Scheduler {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	nodeLister, podLister := initInformers(clientset, quit)

	return Scheduler{
		clientset:  clientset,
		nodeLister: nodeLister,
		podLister:  podLister,
	}
}

// Create Informers
func initInformers(clientset *kubernetes.Clientset, quit chan struct{}) (listersv1.NodeLister, listersv1.PodLister) {
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
				log.Printf("New Pod Added to Store: %s", pod.Name)
			}
		},
	})

	factory.Start(quit)
	return nodeInformer.Lister(), podInformer.Lister()
}

func main() {
	// configure Logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// The SFC controller is starting
	log.Printf("SFC-controller v0.0.4 Starting...\n")

	//Add infrastructure Edges to Graph: change weights and locations according to your infrastructure!
	graphLatency.addEdge("sw-Bruges", "sw-Ghent", 15)
	graphLatency.addEdge("sw-Antwerp", "sw-Leuven", 15)
	graphLatency.addEdge("sw-Ghent", "sw-Brussels", 25)
	graphLatency.addEdge("sw-Brussels", "sw-Leuven", 25)

	//Create Scheduler
	quit := make(chan struct{})
	defer close(quit)

	scheduler = NewScheduler(quit)

	// start http server
	svr := &http.Server{Addr: ":" + port}
	svr.Handler = http.HandlerFunc(handler)
	log.Printf("Extender Call started on port %v\n", port)

	if err := svr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
