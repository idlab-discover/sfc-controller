package main

import (
	"encoding/json"
	k8sApi "k8s.io/api/core/v1"
	k8sSchedulerApi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"
	"log"
	"net/http"
	"time"
)

func checkBody(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
}

// handler receives a request from the kubernetes scheduler.
func handler(w http.ResponseWriter, r *http.Request) {
	// check request body.
	checkBody(w, r)

	// Decode request
	start := time.Now()
	dec := json.NewDecoder(r.Body)
	received := &k8sSchedulerApi.ExtenderArgs{}
	err := dec.Decode(received)
	if err != nil {
		panic(err)
		return
	}

	logNodes(received.Nodes)

	//verify nodes available bandwidth if Pods were already allocated
	if !allocatedPods.isEmpty() {
		log.Printf("Pods were already allocated! Updating available bandwidth...")
		watchScheduledPods(scheduler)
	}

	// select the node to schedule on.
	nodes, err := selectNode(received.Nodes, received.Pod, scheduler)
	if err != nil {
		panic(err)
	}

	// return the result.
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_ = enc.Encode(&k8sSchedulerApi.ExtenderFilterResult{
		Nodes: &k8sApi.NodeList{
			Items: nodes,
		},
	})

	log.Printf("Choose Node %v for Pod %v \n", nodes[0].Name, received.Pod.Name)
	log.Printf("Response Time: %v \n", time.Since(start))
	log.Printf("---------------------------------------------------------\n")
	return
}
