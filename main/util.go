package main

import (
	"fmt"
	k8sApi "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"log"
	"math"
	"strconv"
)

// logNodes prints a line for every candidate node.
func logNodes(nodes *k8sApi.NodeList) {
	log.Printf("---------------New Scheduling request------------\n")
	for _, n := range nodes.Items {
		log.Printf("Received node: %s \n", n.Name)
	}
}

// convert Bandwidth String to Float
func stringtoFloatBandwidth(minBandwidth string) float64 {
	bandwidth, err := strconv.ParseFloat(minBandwidth, 64)
	if err == nil {
		return bandwidth
	}
	return 0.250 // Default Value: 250 Kbit/s
}

// convert String to Int
func stringtoInt(value string) int {
	newValue, err := strconv.Atoi(value)
	if err == nil {
		return newValue
	}
	return 1 // Default Value: 1
}

// getDesiredFromLabels parses the LabelValue from a pod's label
func getDesiredFromLabels(pod *k8sApi.Pod, label string) string {
	labelValue, exists := pod.Labels[label]
	if exists {
		labelValue = string(pod.Labels[label])
		return labelValue
	}
	return "Any"
}

// add service Hash
func addService(key string, node k8sApi.Node) {
	serviceHash[key] = node.Name
	log.Printf("Service Hash Added: Key: %v  - Value: %v \n", key, serviceHash[key])
}

// GetMinDelay for the specified Location
func getMinDelay(nodes *k8sApi.NodeList, targetLocation string) float64 {
	minDelay := math.MaxFloat64
	for _, node := range nodes.Items {
		delay, _ := graphLatency.getPath(node.Name, targetLocation)
		//fmt.Printf("Delay value for %v received for Node %v : %v \n", targetLocation, node.Name, float64(delay))
		minDelay = math.Min(minDelay, float64(delay))
	}
	return minDelay
}

// getKey
func getKey(id int, appName string, nsh string, chainPos int, totalChainServ int) string {
	return strconv.Itoa(id) + "-" + appName + "-" + nsh + "-" + strconv.Itoa(chainPos) + "-" + strconv.Itoa(totalChainServ)
}

//getValue
func getValue(shortPathCost map[string]float64, key string) float64 {
	return shortPathCost[key]
}

// GetBandwidthValue parses the bandwidth from a node's label or returns
// the max float value if the label doesn't exist.
func getBandwidthValue(node *k8sApi.Node, avBandwidth string) float64 {
	nodeBandwidth, exists := node.Labels[avBandwidth]
	if exists {
		nodeBandwidth, err := strconv.ParseFloat(nodeBandwidth, 64)
		if err == nil {
			return nodeBandwidth
		}
	}
	return math.MaxFloat64
}

func updateBandwidthLabel(label string, kubeClient kubernetes.Interface, nodes listersv1.NodeLister, candidateNode *k8sApi.Node) error {

	// New: Using Node Informers - Faster!
	nodeLabels := candidateNode.GetLabels()
	prevLabel := nodeLabels["avBandwidth"]
	nodeLabels["avBandwidth"] = label

	node, err := nodes.Get(candidateNode.Name)
	if err != nil {
		return fmt.Errorf("node could not be found")
	}

	node.SetLabels(nodeLabels)

	log.Printf("Node %v Updating Bandwidth Label: previous: %v / avBandwidth = %v \n", node.Name, prevLabel, label)

	if _, err = kubeClient.CoreV1().Nodes().Update(node); err != nil {
		return fmt.Errorf("failed to update Label")
	}
	return nil
}
