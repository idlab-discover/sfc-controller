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
)

// Scheduler instance structure
type Scheduler struct {
	clientset  *kubernetes.Clientset
	podQueue   chan *k8sApi.Pod
	nodeLister listersv1.NodeLister
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

	return Scheduler{
		clientset:  clientset,
		podQueue:   podQueue,
		nodeLister: initInformers(clientset, podQueue, quit),
	}
}

// Create Node and Pod informers
func initInformers(clientset *kubernetes.Clientset, podQueue chan *k8sApi.Pod, quit chan struct{}) listersv1.NodeLister {
	factory := informers.NewSharedInformerFactory(clientset, 0)

	nodeInformer := factory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*k8sApi.Node)
			if !ok {
				log.Println("This is not a node")
				return
			}
			log.Printf("New Node Added to Store: %s", node.GetName())
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
			if pod.Spec.SchedulerName == schedulerName { //if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == schedulerName {
				podQueue <- pod
				log.Printf("New Pod Added to Store: %s", pod.Name)
			}
		},
	})

	factory.Start(quit)
	return nodeInformer.Lister()
}

func main() {
	log.Printf("SFC-controller v0.0.3 Starting...\n")

	// init router
	svr := &http.Server{Addr: ":" + port}
	svr.Handler = http.HandlerFunc(handler)

	//Create Scheduler
	podQueue := make(chan *k8sApi.Pod, 300)
	defer close(podQueue)

	quit := make(chan struct{})
	defer close(quit)

	scheduler = NewScheduler(podQueue, quit)

	// start http server
	log.Printf("Started on port %v\n", port)
	if err := svr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	// Create Channel
	ch := make(chan bool)
	<-ch
}
