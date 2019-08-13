package main

import (
	"fmt"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/rest"
	k8sApi "k8s.io/kubernetes/pkg/api"
	"math"
	"strconv"
	"strings"

	//schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
	//utilfeature "k8s.io/apiserver/pkg/util/feature"
	//"k8s.io/kubernetes/pkg/features"
	//schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

var id = 0
var serviceHash = make(map[string]string)

var graphLatency = Graph{
	"work1.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 	{"Bruges": 3.0},
	"work2.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 	{"Bruges": 3.0},
	"work3.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 	{"Bruges": 5.0},
	"Bruges": 	{"work1.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 3.0, "work2.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 3.0, "work3.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 5.0,"Ghent": 15.0},
	"work4.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be":		{"Ghent": 3.0},
	"work5.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Ghent": 5.0},
	"work6.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Ghent": 3.0},
	"Ghent": 	{"work4.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 3.0, "work5.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 5.0, "work6.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "Bruges": 15.0 ,"Brussels": 25.0},
	"work7.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Leuven": 3.0},
	"work8.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Leuven": 3.0},
	"work9.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Leuven": 5.0},
	"Brussels": 	{"work13.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 1.0, "work14.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 1.0, "master0.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 1.0, "Leuven": 25.0,"Ghent": 25.0},
	"work13.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Brussels": 1.0},
	"work14.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Brussels": 1.0},
	"master0.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 	{"Brussels": 1.0},
	"Leuven": 	{"work7.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "work8.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "work9.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 5.0, "Brussels": 25.0, "Antwerp": 15.0},
	"work10.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Antwerp": 3.0},
	"work11.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Antwerp": 3.0},
	"work12.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 	{"Antwerp": 3.0},
	"Antwerp": 	{"work10.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "work11.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "work12.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "Leuven": 15.0},
}

// logNodes prints a line for every node.
func logNodes(nodes *k8sApi.NodeList) {
	fmt.Printf("---------------New Scheduling request------------\n")
	for _, n := range nodes.Items {
		fmt.Printf("Received node: %v \n", n.Name)
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

func addService(key string, node k8sApi.Node){
	serviceHash[key] = node.Name
	fmt.Printf("Service Hash Added: Key %v  - Value %v \n", key, serviceHash[key])
}
// selectNode
func selectNode(nodes *k8sApi.NodeList, pod *k8sApi.Pod) ([]k8sApi.Node, error) {

	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("No nodes were provided")
	}

	// extract information from Pod Template file - label values
	appName := getDesiredFromLabels(pod, "app")
	targetLocation := getDesiredFromLabels(pod, "targetLocation")
	minBandwidth := getDesiredFromLabels(pod, "minBandwidth")
	chainPosString := getDesiredFromLabels(pod, "chainPosition")
	nsh := getDesiredFromLabels(pod, "networkServiceHeader")
	totalChain := getDesiredFromLabels(pod, "totalChainServ")
	//deviceType := getDesiredFromLabels(pod, "deviceType")
	policy := getDesiredFromLabels(pod, "policy")

	minBandwidth = strings.TrimRight(minBandwidth, "Mi")
	chainPosString = strings.TrimRight(chainPosString, "pos")
	totalChain = strings.TrimRight(totalChain, "serv")

	podMinBandwith := stringtoFloatBandwidth(minBandwidth)
	chainPos := stringtoInt(chainPosString)
	totalChainServ := stringtoInt(totalChain)

	nextApp := ""
	prevApp := ""
	appList := []string{}

	if chainPos == 1{
		nextApp = getDesiredFromLabels(pod, "nextService")
		appList= []string{nextApp}
	} else if chainPos == totalChainServ{
		prevApp = getDesiredFromLabels(pod, "prevService")
		appList = []string{prevApp}
	} else{
		prevApp = getDesiredFromLabels(pod, "prevService")
		nextApp = getDesiredFromLabels(pod, "nextService")
		appList = []string{prevApp, nextApp}
	}

	//fmt.Printf("Pod Network Service Header: %v \n", nsh)
	//fmt.Printf("Pod Chain Position: %v \n", chainPos)
	//fmt.Printf("Pod Total Chain Services: %v \n", totalChainServ)
	//fmt.Printf("Pod Desired location: %v \n", targetLocation)
	//

	fmt.Printf("Pod Name: %v \n", pod.Name)
	fmt.Printf("Pod Desired location: %v \n", targetLocation)
	fmt.Printf("Pod Desired bandwidth: %v (Mi)\n", podMinBandwith)
	//fmt.Printf("Pod Desired Device Type: %v \n", deviceType)
	fmt.Printf("Policy: %v \n", policy)
	fmt.Printf("prevApp: %v \n", prevApp)
	fmt.Printf("nextApp: %v \n", nextApp)
	fmt.Printf("appList: %v \n", appList)

	 if policy == "Location" { // Location Policy

		fmt.Printf("--------------------------------------------------------------\n")
		fmt.Printf("---------------------Location Policy Selected ------------------\n")
		fmt.Printf("Target Location: %v \n", targetLocation)

		minDelay := getMinDelay(nodes, targetLocation)
		node := locationSelection(nodes, minDelay, targetLocation, podMinBandwith)

		if node.GetName() == "" {
			return nil, fmt.Errorf("No suitable node for target Location with enough bandwidth!")
		} else {
			// add pod to Service Hash
			id++
			addService(getKey(id, appName, nsh, chainPos, totalChainServ), node)

			// update Link bandwidth
			value := getNodeBandwidth(node) - podMinBandwith
			updateNodeBandwidth(value,node)
			return []k8sApi.Node{node}, nil
		}

	} else if policy == "Latency" { // Latency Policy
		fmt.Printf("---------------------------------------------------------------\n")
		fmt.Printf("---------------------Latency Policy Selected ------------------\n")

		// find services belonging to this service chain and put them in a Linked List
		podList := createPodList(nsh)

		for i := 1; i <= id; i++ {
			for j := 1; j <= totalChainServ; j++ {
				for _, app := range appList {
					if j != chainPos {
						key := getKey(i, app, nsh, j, totalChainServ)
						fmt.Printf("Key: %v \n", key)
						allocatedNode, ok := serviceHash[key]
						if ok {
							fmt.Printf("Key found! Allocated on Node: %v \n", allocatedNode)
							podList.addPod(key, allocatedNode)
						} else {
							fmt.Printf("Key not found! \n")
						}
					}
				}
			}
		}

		// calculate shortest path for each filtered node
		// node with min short path is selected
		if !podList.isEmpty() {
			fmt.Printf("Pod List is not empty! \n")
			fmt.Printf("Calculate Delay Cost (Short Paths) and find Best Node! \n")
			nodeDelay, _ := calculateShortPath(nodes, podList, podMinBandwith)

			if nodeDelay.GetName() != "" {
				// Return Node Delay
				fmt.Printf("Node Delay selected! \n")

				// add pod to Service Hash
				id++
				addService(getKey(id, appName, nsh, chainPos, totalChainServ), nodeDelay)

				// update Link bandwidth
				value := getNodeBandwidth(nodeDelay) - podMinBandwith
				updateNodeBandwidth(value,nodeDelay)
				return []k8sApi.Node{nodeDelay}, nil
			}
		}else {
			fmt.Printf("Pod List is empty! \n")
			fmt.Printf("Target Location: %v \n", targetLocation)

			if targetLocation != "Any"{ // Location Selection -> Location Policy
				fmt.Printf("As if Location Policy was selected!! \n")
				minDelay := getMinDelay(nodes, targetLocation)
				node := locationSelection(nodes, minDelay, targetLocation, podMinBandwith)

				if node.Name == "" {
					return nil, fmt.Errorf("No suitable node for target Location with enough bandwidth!")
				} else {
					// add pod to Service Hash
					id++
					addService(getKey(id, appName, nsh, chainPos, totalChainServ), node)

					// update Link bandwidth
					value := getNodeBandwidth(node) - podMinBandwith
					updateNodeBandwidth(value,node)

					return []k8sApi.Node{node}, nil
				}
			}
		}
	}
	// Device Type & Link MAX Cost Selection
	fmt.Printf("Calculate Max Link Cost!! Higher amount of bandwidth used! \n")
	nodeMaxLink, _ := calculateMaxLinkCost(nodes, podMinBandwith)

	if nodeMaxLink.GetName() != "" {
			fmt.Printf("Node Max Link selected! \n")

			// add pod to Service Hash
			id++
			addService(getKey(id, appName, nsh, chainPos, totalChainServ), nodeMaxLink)

			// update Link bandwidth
			value := getNodeBandwidth(nodeMaxLink) - podMinBandwith
			updateNodeBandwidth(value, nodeMaxLink)
			return []k8sApi.Node{nodeMaxLink}, nil
	}

	fmt.Printf("--------------------------------------------------------------\n")
	fmt.Printf("---------------------Random Selection-------------------------\n")
	pick := randomSelection(nodes)
	// add pod to Service Hash
	id++
	addService(getKey(id, appName, nsh, chainPos, totalChainServ), pick)

	// update Link bandwidth
	value := getNodeBandwidth(pick) - podMinBandwith
	updateNodeBandwidth(value,pick)
	return []k8sApi.Node{pick}, nil
}

// GetMinRTT for the specified Location
// the max float value if the label doesn't exist.
func getMinDelay(nodes *k8sApi.NodeList, targetLocation string) float64 {
	minDelay := math.MaxFloat64
	for _, node := range nodes.Items {
		_, delay, _ := graphLatency.Path(node.Name, targetLocation)
		//fmt.Printf("Delay value for %v received for Node %v : %v \n", targetLocation, node.Name, float64(delay))
		minDelay = math.Min(minDelay, float64(delay))
	}
	return minDelay
}

// Return keys of the given map
func getAllKeys(serviceHash map[string]string) (keys []string) {
	for k := range serviceHash {
		keys = append(keys, k)
	}
	return keys
}

// getKey
func getKey(id int, appName string, nsh string, chainPos int, totalChainServ int) string {
	return strconv.Itoa(id)+"-"+appName+"-"+nsh+"-"+ strconv.Itoa(chainPos)+"-"+strconv.Itoa(totalChainServ)
}

//getValue
func getValue(shortPathCost map[string]float64, key string) float64 {
	return shortPathCost[key]
}

func getDeviceType(node *k8sApi.Node) string {
	deviceType, exists := node.Labels["deviceType"]
	if exists {
		return deviceType
	}
	return ""
}

/*
func getMinRTT(nodes *k8sApi.NodeList, targetLocation string) float64 {
	minRTT := math.MaxFloat64
	for _, node := range nodes.Items {
		rtt := getRTTValue(&node, targetLocation)
		//fmt.Printf("RTT value for %v received for Node %v : %v \n", targetLocation, node.Name, rtt)
		minRTT = math.Min(minRTT, rtt)
	}
	return minRTT
}

// GetRTTValue parses the RTT from a node's label or returns
// the max float value if the label doesn't exist.
func getRTTValue(node *k8sApi.Node, rttLocation string) float64 {
	rtt, exists := node.Labels[rttLocation]
	if exists {
		rtt, err := strconv.ParseFloat(rtt, 64)
		if err == nil {
			return rtt
		}
	}
	return math.MaxFloat64
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

func updateBandwidthLabel(nodeLabels map[string]string, kubeClient kubernetes.Interface, candidateNode *k8sApi.Node, hostnameLabel string) error {
	k8sNodeList, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("No nodes were provided")
	}

	for _, node := range k8sNodeList.Items {
		if node.Labels[hostnameLabel] == candidateNode.Labels[hostnameLabel] {
			node.SetLabels(nodeLabels)
			if _, err = kubeClient.CoreV1().Nodes().Update(&node); err != nil {
				fmt.Printf("Failed to update Label")
				return fmt.Errorf("Failed to update Label")
			}
		}
	}
	return nil
}
*/
