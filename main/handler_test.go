package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	k8sApi "k8s.io/kubernetes/pkg/api"
	k8sSchedulerApi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
)

// TestHandler tests the handler function.
func TestHandler(t *testing.T) {
	// New test server
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	// Input to send for test and expected value
	expected := "node1"
	args := &k8sSchedulerApi.ExtenderArgs{
		Pod: k8sApi.Pod{},
		Nodes: newNodeList(
			newNode("node1", "50.5"),
			newNode("node2", "70.5"),
			newNode("node3", "80.5"),
		),
	}

	// convert to json
	b, err := json.Marshal(args)
	if err != nil {
		t.Errorf("Error when trying to convert args to bytes: %v", err)
		return
	}
	// send request to fake server
	res, err := http.Post(srv.URL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		t.Errorf("Error when making post request: %v", err)
		return
	}

	// decode result from fake server
	dec := json.NewDecoder(res.Body)
	received := &k8sSchedulerApi.ExtenderFilterResult{}
	err = dec.Decode(received)
	if err != nil {
		t.Errorf("Error when trying to convert result to ExtenderFilterResult: %v", err)
		return
	}

	// handle result
	if len(received.Nodes.Items) != 1 {
		t.Errorf("Expected one node but received %v", len(received.Nodes.Items))
		return
	}
	if received.Nodes.Items[0].Name != expected {
		t.Errorf("Expected node %v to be scheduled but %v was chosen", expected, received.Nodes.Items[0].Name)
	}
}
