package main

import (
	"fmt"
	k8sApi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"math"
	"strconv"
	"strings"
)

var (
	// Service Hash IDs
	id = 0

	// Service Hash map
	serviceHash = make(map[string]string)
)

// initial infrastructure Graph
var graphLatency = Graph{
	"docker-desktop": {"Bruges": 5.0},
	"work1.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": {"Bruges": 3.0},
	"work2.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": {"Bruges": 3.0},
	"work3.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": {"Bruges": 5.0},
	"Bruges": {"work1.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 3.0, "work2.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 3.0, "work3.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 5.0, "Ghent": 15.0},
	"work4.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": {"Ghent": 3.0},
	"work5.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": {"Ghent": 5.0},
	"work6.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": {"Ghent": 3.0},
	"Ghent": {"work4.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 3.0, "work5.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 5.0, "work6.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "Bruges": 15.0, "Brussels": 25.0},
	"work7.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": {"Leuven": 3.0},
	"work8.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": {"Leuven": 3.0},
	"work9.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": {"Leuven": 5.0},
	"Brussels": {"work13.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 1.0, "work14.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 1.0, "master0.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": 1.0, "Leuven": 25.0, "Ghent": 25.0},
	"work13.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be":  {"Brussels": 1.0},
	"work14.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be":  {"Brussels": 1.0},
	"master0.kbcluster1.wall2-ilabt-iminds-be.wall2.ilabt.iminds.be": {"Brussels": 1.0},
	"Leuven": {"work7.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "work8.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "work9.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 5.0, "Brussels": 25.0, "Antwerp": 15.0},
	"work10.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": {"Antwerp": 3.0},
	"work11.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": {"Antwerp": 3.0},
	"work12.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": {"Antwerp": 3.0},
	"Antwerp": {"work10.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "work11.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "work12.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be": 3.0, "Leuven": 15.0},
}

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
	log.Printf("Service Hash Added: Key %v  - Value %v \n", key, serviceHash[key])
}

// selectNode
func selectNode(nodes *k8sApi.NodeList, pod *k8sApi.Pod, scheduler Scheduler) ([]k8sApi.Node, error) {

	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("no nodes were provided")
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
	var appList []string

	// find next and previous services in the service chain
	if chainPos == 1 {
		nextApp = getDesiredFromLabels(pod, "nextService")
		appList = []string{nextApp}
	} else if chainPos == totalChainServ {
		prevApp = getDesiredFromLabels(pod, "prevService")
		appList = []string{prevApp}
	} else {
		prevApp = getDesiredFromLabels(pod, "prevService")
		nextApp = getDesiredFromLabels(pod, "nextService")
		appList = []string{prevApp, nextApp}
	}

	//fmt.Printf("Pod Network Service Header: %v \n", nsh)
	//fmt.Printf("Pod Chain Position: %v \n", chainPos)
	//fmt.Printf("Pod Total Chain Services: %v \n", totalChainServ)
	//fmt.Printf("Pod Desired location: %v \n", targetLocation)
	//

	log.Printf("Pod Name: %v \n", pod.Name)
	log.Printf("Pod Desired location: %v \n", targetLocation)
	log.Printf("Pod Desired bandwidth: %v (Mi)\n", podMinBandwith)
	log.Printf("Scheduling Policy: %v \n", policy)
	log.Printf("prevApp: %v \n", prevApp)
	log.Printf("nextApp: %v \n", nextApp)
	log.Printf("Service Chain: %v \n", appList)

	if policy == "Location" { // If Location Policy enabled

		log.Printf("--------------------------------------------------------------\n")
		log.Printf("---------------------Location Policy Selected ------------------\n")
		log.Printf("Target Location: %v \n", targetLocation)

		minDelay := getMinDelay(nodes, targetLocation)
		node := locationSelection(nodes, minDelay, targetLocation, podMinBandwith)

		if node.GetName() == "" { // No suitable node found
			return nil, fmt.Errorf("no suitable node for target Location with enough bandwidth")
		} else {
			// add pod to Service Hash
			id++
			addService(getKey(id, appName, nsh, chainPos, totalChainServ), node)

			// update Link bandwidth
			nodeBand := getBandwidthValue(&node, "avBandwidth")
			value := nodeBand - podMinBandwith

			label := strconv.FormatFloat(value, 'f', 2, 64)

			err := updateBandwidthLabel(label, scheduler.clientset, &node, "kubernetes.io/hostname") // &node, "kubernetes.io/hostname")
			if err != nil {
				log.Printf("encountered error when updating label: %v", err)
			}

			//updateNodeBandwidth(value, node)
			return []k8sApi.Node{node}, nil
		}

	} else if policy == "Latency" { // If Latency Policy enabled
		log.Printf("---------------------------------------------------------------\n")
		log.Printf("---------------------Latency Policy Selected ------------------\n")

		// find services belonging to this service chain and put them in a Linked List
		podList := createPodList(nsh)

		for i := 1; i <= id; i++ {
			for j := 1; j <= totalChainServ; j++ {
				for _, app := range appList {
					if j != chainPos {
						key := getKey(i, app, nsh, j, totalChainServ)
						//fmt.Printf("Key: %v \n", key)
						allocatedNode, ok := serviceHash[key]
						if ok {
							log.Printf("Key found! Allocated on Node: %v \n", allocatedNode)
							err := podList.addPod(key, allocatedNode)
							if err != nil {
								log.Printf("encountered error when adding Pod to the List: %v", err)
							}
						} //else {
						//fmt.Printf("Key not found! \n")
						//}
					}
				}
			}
		}

		// calculate shortest path for each filtered node
		// node with min short path is selected
		if !podList.isEmpty() {
			log.Printf("Pod List is not empty! \n")
			log.Printf("Calculate Delay Cost (Short Paths) and find Best Node! \n")
			nodeDelay, _ := calculateShortPath(nodes, podList, podMinBandwith)

			if nodeDelay.GetName() != "" {
				// Return Node Delay
				log.Printf("Node Delay selected! \n")

				// add pod to Service Hash
				id++
				addService(getKey(id, appName, nsh, chainPos, totalChainServ), nodeDelay)

				// update Link bandwidth
				nodeBand := getBandwidthValue(&nodeDelay, "avBandwidth")
				value := nodeBand - podMinBandwith

				label := strconv.FormatFloat(value, 'f', 2, 64)

				err := updateBandwidthLabel(label, scheduler.clientset, &nodeDelay, "kubernetes.io/hostname")
				if err != nil {
					log.Printf("encountered error when updating label: %v", err)
				}

				//updateNodeBandwidth(value, nodeDelay)
				return []k8sApi.Node{nodeDelay}, nil
			}
		} else {
			log.Printf("Pod List is empty! \n")
			log.Printf("Target Location: %v \n", targetLocation)

			if targetLocation != "Any" { // Location Selection -> Location Policy
				log.Printf("As if Location Policy was selected!! \n")
				minDelay := getMinDelay(nodes, targetLocation)
				node := locationSelection(nodes, minDelay, targetLocation, podMinBandwith)

				if node.Name == "" { // No suitable Node found
					return nil, fmt.Errorf("no suitable node for target Location with enough bandwidth")
				} else {
					// add pod to Service Hash
					id++
					addService(getKey(id, appName, nsh, chainPos, totalChainServ), node)

					// update Link bandwidth
					nodeBand := getBandwidthValue(&node, "avBandwidth")
					value := nodeBand - podMinBandwith

					label := strconv.FormatFloat(value, 'f', 2, 64)

					err := updateBandwidthLabel(label, scheduler.clientset, &node, "kubernetes.io/hostname")
					if err != nil {
						log.Printf("encountered error when updating label: %v", err)
					}

					//updateNodeBandwidth(value, node)
					return []k8sApi.Node{node}, nil
				}
			}
		}
	}
	// Link MAX Cost Selection
	fmt.Printf("---------------------------------------------------------------\n")
	fmt.Printf("---------------------MAX Link Cost Selection-------------------\n")
	//fmt.Printf("Calculate Max Link Cost!! Higher amount of bandwidth used! \n")
	nodeMaxLink, _ := calculateMaxLinkCost(nodes, podMinBandwith)

	if nodeMaxLink.GetName() != "" {
		log.Printf("Node Max Link selected! \n")

		// add pod to Service Hash
		id++
		addService(getKey(id, appName, nsh, chainPos, totalChainServ), nodeMaxLink)

		// update Link bandwidth
		nodeBand := getBandwidthValue(&nodeMaxLink, "avBandwidth")
		value := nodeBand - podMinBandwith

		label := strconv.FormatFloat(value, 'f', 2, 64)

		err := updateBandwidthLabel(label, scheduler.clientset, &nodeMaxLink, "kubernetes.io/hostname")
		if err != nil {
			log.Printf("encountered error when updating label: %v", err)
		}

		//updateNodeBandwidth(value, nodeMaxLink)
		return []k8sApi.Node{nodeMaxLink}, nil
	}

	log.Printf("---------------------------------------------------------------\n")
	log.Printf("----------------Last Resource: Random Selection ---------------\n")

	pick := randomSelection(nodes)
	// add pod to Service Hash
	id++
	addService(getKey(id, appName, nsh, chainPos, totalChainServ), pick)

	// update Link bandwidth
	nodeBand := getBandwidthValue(&pick, "avBandwidth")
	value := nodeBand - podMinBandwith

	if value < 0 {
		value = 0.0
	}

	label := strconv.FormatFloat(value, 'f', 2, 64)

	err := updateBandwidthLabel(label, scheduler.clientset, &pick, "kubernetes.io/hostname")
	if err != nil {
		log.Printf("encountered error when updating label: %v", err)
	}

	//updateNodeBandwidth(value, pick)
	return []k8sApi.Node{pick}, nil
}

// GetMinDelay for the specified Location
func getMinDelay(nodes *k8sApi.NodeList, targetLocation string) float64 {
	minDelay := math.MaxFloat64
	for _, node := range nodes.Items {
		_, delay, _ := graphLatency.Path(node.Name, targetLocation)
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

func updateBandwidthLabel(label string, kubeClient kubernetes.Interface, candidateNode *k8sApi.Node, hostnameLabel string) error {

	nodeLabels := candidateNode.GetLabels()
	nodeLabels["avBandwidth"] = label

	log.Printf("Updating Bandwidth Label: avBandwidth = %v \n", label)

	k8sNodeList, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("no nodes were provided")
	}

	for _, node := range k8sNodeList.Items {
		if node.Labels[hostnameLabel] == candidateNode.Labels[hostnameLabel] {
			node.SetLabels(nodeLabels)
			if _, err = kubeClient.CoreV1().Nodes().Update(&node); err != nil {
				return fmt.Errorf("failed to update Label")
			}
		}
	}
	return nil
}

/*
// Return keys of the given map
func getAllKeys(serviceHash map[string]string) (keys []string) {
	for k := range serviceHash {
		keys = append(keys, k)
	}
	return keys
}

// GetMinRTT finds the node with min RTT for the target Location
func getMinRTT(nodes *k8sApi.NodeList, targetLocation string) float64 {
	minRTT := math.MaxFloat64
	for _, node := range nodes.Items {
		rtt := getRTTValue(&node, targetLocation)
		//fmt.Printf("RTT value for %v received for Node %v : %v \n", targetLocation, node.Name, rtt)
		minRTT = math.Min(minRTT, rtt)
	}
	return minRTT
}

func getDeviceType(node *k8sApi.Node) string {
	deviceType, exists := node.Labels["deviceType"]
	if exists {
		return deviceType
	}
	return ""
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
*/
