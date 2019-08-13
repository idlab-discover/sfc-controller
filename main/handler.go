package main

import (
	"encoding/json"
	"fmt"
	k8sApi "k8s.io/kubernetes/pkg/api"
	k8sSchedulerApi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
	"net/http"
	"time"
)

// handler handles a request from the kubernetes scheduler.
func handler(w http.ResponseWriter, r *http.Request) {
	// decode request body.
	start := time.Now()
	dec := json.NewDecoder(r.Body)
	received := &k8sSchedulerApi.ExtenderArgs{}
	err := dec.Decode(received)
	if err != nil {
		fmt.Printf("Error when trying to decode response body to struct: %v\n", err)
		return
	}

	logNodes(&received.Nodes)

	// select the node to schedule on.
	nodes, err := selectNode(&received.Nodes, &received.Pod)
	if err != nil {
		fmt.Printf("Encountered error when selecting node: %v", err)
	}

	// return the result.
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	enc.Encode(&k8sSchedulerApi.ExtenderFilterResult{
		Nodes: k8sApi.NodeList{
			Items: nodes,
		},
	})
	fmt.Printf("Choose Node %v for Pod %v\n", nodes[0].Name, received.Pod.Name)
	fmt.Printf("Response Time: %v\n", time.Since(start))
	fmt.Printf("---------------------------------------------------------\n")
	return
}

//func homeDir() string {
//	if h := os.Getenv("HOME"); h != "" {
//		return h
//	}
//	return os.Getenv("USERPROFILE") // windows
//}
