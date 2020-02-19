package main

import (
	"fmt"
	k8sApi "k8s.io/api/core/v1"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

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
			key := getKey(id, appName, nsh, chainPos, totalChainServ)
			addService(key, node)

			//add Service key to Pod Label
			err := addKeyPodLabel(key, scheduler.clientset, scheduler.podLister, pod)
			if err != nil {
				log.Printf("encountered error when updating Service Key label: %v", err)
			}

			// update Link bandwidth
			nodeBand := getBandwidthValue(&node, "avBandwidth")
			value := nodeBand - podMinBandwith

			label := strconv.FormatFloat(value, 'f', 2, 64)

			err = updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &node)
			if err != nil {
				log.Printf("encountered error when updating Bandiwdth label: %v", err)
			}

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
				key := getKey(id, appName, nsh, chainPos, totalChainServ)
				addService(key, nodeDelay)

				//add Service key to Pod Label
				err := addKeyPodLabel(key, scheduler.clientset, scheduler.podLister, pod)
				if err != nil {
					log.Printf("encountered error when updating Service Key label: %v", err)
				}

				// update Link bandwidth
				nodeBand := getBandwidthValue(&nodeDelay, "avBandwidth")
				value := nodeBand - podMinBandwith

				label := strconv.FormatFloat(value, 'f', 2, 64)

				err = updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &nodeDelay)
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
					key := getKey(id, appName, nsh, chainPos, totalChainServ)
					addService(key, node)

					//add Service key to Pod Label
					err := addKeyPodLabel(key, scheduler.clientset, scheduler.podLister, pod)
					if err != nil {
						log.Printf("encountered error when updating service key label: %v", err)
					}

					// update Link bandwidth
					nodeBand := getBandwidthValue(&node, "avBandwidth")
					value := nodeBand - podMinBandwith

					label := strconv.FormatFloat(value, 'f', 2, 64)

					err = updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &node)
					if err != nil {
						log.Printf("encountered error when updating bandwidth label: %v", err)
					}

					//updateNodeBandwidth(value, node)
					return []k8sApi.Node{node}, nil
				}
			}
		}
	}
	// Link MAX Cost Selection
	log.Printf("---------------------------------------------------------------\n")
	log.Printf("---------------------MAX Link Cost Selection-------------------\n")
	//fmt.Printf("Calculate Max Link Cost!! Higher amount of bandwidth used! \n")
	nodeMaxLink, _ := calculateMaxLinkCost(nodes, podMinBandwith)

	if nodeMaxLink.GetName() != "" {
		log.Printf("Node Max Link selected! \n")

		// add pod to Service Hash
		id++
		key := getKey(id, appName, nsh, chainPos, totalChainServ)
		addService(key, nodeMaxLink)

		//add Service key to Pod Label
		err := addKeyPodLabel(key, scheduler.clientset, scheduler.podLister, pod)
		if err != nil {
			log.Printf("encountered error when updating Service Key label: %v", err)
		}

		// update Link bandwidth
		nodeBand := getBandwidthValue(&nodeMaxLink, "avBandwidth")
		value := nodeBand - podMinBandwith

		label := strconv.FormatFloat(value, 'f', 2, 64)

		err = updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &nodeMaxLink)
		if err != nil {
			log.Printf("encountered error when updating bandwidth label: %v", err)
		}

		//updateNodeBandwidth(value, nodeMaxLink)
		return []k8sApi.Node{nodeMaxLink}, nil
	}

	log.Printf("---------------------------------------------------------------\n")
	log.Printf("----------------Last Resource: Random Selection ---------------\n")

	pick := randomSelection(nodes)
	// add pod to Service Hash
	id++
	key := getKey(id, appName, nsh, chainPos, totalChainServ)
	addService(key, pick)

	//add Service key to Pod Label
	err := addKeyPodLabel(key, scheduler.clientset, scheduler.podLister, pod)
	if err != nil {
		log.Printf("encountered error when updating Service Key label: %v", err)
	}

	// update Link bandwidth
	nodeBand := getBandwidthValue(&pick, "avBandwidth")
	value := nodeBand - podMinBandwith

	if value < 0 {
		value = 0.0
	}

	label := strconv.FormatFloat(value, 'f', 2, 64)

	err = updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &pick)
	if err != nil {
		log.Printf("encountered error when updating bandwidth label: %v", err)
	}

	//updateNodeBandwidth(value, pick)
	return []k8sApi.Node{pick}, nil
}

// Random Selection: pick a node randomly
func randomSelection(nodes *k8sApi.NodeList) k8sApi.Node {
	// Random Pick between the filtered Nodes
	rand.Seed(time.Now().Unix())
	numNodes := len(nodes.Items)
	pick := nodes.Items[rand.Int()%numNodes]
	return pick
}

// locationSelection: Select Node for the pod deployment based on the specified Location OR
// the max float value if the label doesn't exist.
func locationSelection(nodes *k8sApi.NodeList, minDelay float64, targetLocation string, podMinBandwith float64) k8sApi.Node {
	copyItems := nodes.Items

	for i, node := range nodes.Items {
		delay, _ := graphLatency.getPath(node.Name, targetLocation)
		if minDelay == float64(delay) {
			nodeBand := getBandwidthValue(&node, "avBandwidth")
			if podMinBandwith <= nodeBand {
				log.Printf("Selected Node: %v \n", node.Name)
				return node
			} else {
				log.Printf("Remove candidate Node %s , available Bandwidth not enough: %v (Mi)\n", node.Name, nodeBand)
				copyItems = append(nodes.Items[:i], nodes.Items[i+1:]...)
			}
		}
	}

	if len(copyItems) == 0 {
		return k8sApi.Node{}
	} else {
		nodes.Items = copyItems
		newDelay := getMinDelay(nodes, targetLocation)
		log.Printf("---------------------Recursive Iteration ------------------\n")
		return locationSelection(nodes, newDelay, targetLocation, podMinBandwith)
	}
}

//calculate short path - Delay Cost
func calculateShortPath(nodes *k8sApi.NodeList, podList *podList, podMinBandwidth float64) (k8sApi.Node, map[string]float64) {
	delayCost := make(map[string]float64)
	minCost := math.MaxFloat64
	prevCost := minCost
	selectedNode := k8sApi.Node{}

	for _, node := range nodes.Items {
		nodeBand := getBandwidthValue(&node, "avBandwidth")
		if podMinBandwidth <= nodeBand {
			log.Printf("Node: %v \n", node.Name)
			podList.start()
			for podList.current != nil {
				// calculate each shortest path
				cost, _ := graphLatency.getPath(node.Name, podList.current.nodeAllocated)
				log.Printf("Current Cost: %v \n", cost)
				previousValue := getValue(delayCost, node.Name)
				log.Printf("Previous Cost: %v \n", previousValue)
				delayCost[node.Name] = previousValue + float64(cost)
				log.Printf("Updated Cost: %v \n", delayCost[node.Name])
				podList.next()
			}

			minCost = math.Min(minCost, float64(delayCost[node.Name]))

			if prevCost > minCost {
				prevCost = minCost
				selectedNode = node
				log.Printf("Updated min Node (Delay Cost): %v \n", node.Name)
			}
		} else {
			log.Printf("Node %v av bandwidth not enough!\n", node.Name)
			delayCost[node.Name] = 100000.0
		}
	}
	return selectedNode, delayCost
}

//calculate Max Link Cost
func calculateMaxLinkCost(nodes *k8sApi.NodeList, minBandwidth float64) (k8sApi.Node, map[string]float64) {

	selectedNode := k8sApi.Node{}
	prevCost := 0.0
	linkCost := make(map[string]float64)

	for _, node := range nodes.Items {

		linkCost[node.Name] = minBandwidth / getBandwidthValue(&node, "avBandwidth")
		log.Printf("Node: %v - Cost: %v \n", node.Name, linkCost[node.Name])

		if prevCost < linkCost[node.Name] && linkCost[node.Name] <= 1.0 {
			prevCost = linkCost[node.Name]
			selectedNode = node
			log.Printf("Updated Max Node (Link Cost): %v \n", node.Name)
		}
	}
	return selectedNode, linkCost
}

/*
// min = 0.0
// max = 10.0
func normalize(value float64) float64{
	normalized := (value - 0.0) / (10.0 - 0.0)
	return normalized
}

//calculate Location Cost
func calculateLocationCost(nodes *k8sApi.NodeList, targetLocation string) (k8sApi.Node, map[string]float64) {

	selectedNode := k8sApi.Node{}
	minCost := math.MaxFloat64
	prevCost := minCost
	locationCost := make(map[string]float64)

	for _, node := range nodes.Items {

		fmt.Printf("Calculate Location Cost for Node: %v \n", node.Name)

		_, cost, _ := graphLatency.Path(node.Name, targetLocation)
		locationCost[node.Name] = float64(cost)

		fmt.Printf("Cost: %v \n", locationCost[node.Name])

		minCost = math.Min(minCost, locationCost[node.Name])

		if prevCost > minCost {
			prevCost = minCost
			selectedNode = node
			fmt.Printf("Updated min Node (Location Cost): %v \n", node.Name)
		}
	}
	return selectedNode, locationCost
}

//calculate Min Link Cost
func calculateMinLinkCost(nodes *k8sApi.NodeList, minBandwidth float64) (k8sApi.Node, map[string]float64) {

	selectedNode := k8sApi.Node{}
	minCost := math.MaxFloat64
	prevCost := minCost
	linkCost := make(map[string]float64)

	for _, node := range nodes.Items {

		fmt.Printf("Calculate Link Cost for Node: %v \n", node.Name)

		linkCost[node.Name] = minBandwidth / getNodeBandwidth(node)

		fmt.Printf("Cost: %v \n", linkCost[node.Name])

		minCost = math.Min(minCost, linkCost[node.Name])

		if prevCost > minCost {
			prevCost = minCost
			selectedNode = node
			fmt.Printf("Updated min Node (Link Cost): %v \n", node.Name)
		}
	}
	return selectedNode, linkCost
}

//calculate balanced score - Resource Cost
func calculateResourceCost(nodes *k8sApi.NodeList) (k8sApi.Node, map[string]int){

	selectedNode := k8sApi.Node{}
	minCost := math.MaxFloat64
	prevCost := minCost
	resourceCost := make(map[string]int)

	for _, node := range nodes.Items {

		fmt.Printf("Calculate Resource Cost for Node: %v \n", node.Name)

		scoreBalance := balancedResourceScorer(&node)
		scoreLeast := leastResourceScorer(&node)

		fmt.Printf("ScoreBalance: %v \n", scoreBalance)
		fmt.Printf("scoreLeast: %v \n", scoreLeast)

		resourceCost[node.Name] = int(float64(scoreBalance) * 0.5 + float64(scoreLeast) * 0.5)
		minCost = math.Min(minCost, float64(resourceCost[node.Name]))

		if prevCost > minCost {
			prevCost = minCost
			selectedNode = node
			fmt.Printf("Updated min Node (Resource Cost): %v \n", node.Name)
		}
	}
	return selectedNode, resourceCost
}

func fractionOfCapacity(requested, capacity int64) float64 {
	if capacity == 0 {
		return 1
	}
	return float64(requested) / float64(capacity)
}

func balancedResourceScorer(node *k8sApi.Node) int64 {

	requestedCPU, _ := node.Status.Capacity.Cpu().AsInt64()
	allocatedCPU, _  := node.Status.Allocatable.Cpu().AsInt64()
	requestedMEM, _ := node.Status.Capacity.Memory().AsInt64()
	allocatedMEM, _  := node.Status.Allocatable.Memory().AsInt64()

	cpuFraction := fractionOfCapacity(requestedCPU,allocatedCPU)
	memoryFraction := fractionOfCapacity(requestedMEM, allocatedMEM)

	fmt.Printf("Node: %v \n", node.Name)
	fmt.Printf("cpu requested: %v \n", node.Status.Capacity.Cpu())
	fmt.Printf("mem requested: %v \n", node.Status.Capacity.Memory())
	fmt.Printf("cpu available: %v \n", node.Status.Allocatable.Cpu())
	fmt.Printf("mem available: %v \n", node.Status.Allocatable.Memory())

	fmt.Printf("cpu fraction: %v \n", cpuFraction)
	fmt.Printf("mem fraction: %v \n", memoryFraction)

	if cpuFraction >= 1 || memoryFraction >= 1 {
		// if requested >= capacity, the corresponding host should never be preferred.
		return 0
	}
	// Upper and lower boundary of difference between cpuFraction and memoryFraction are -1 and 1
	// respectively. Multiplying the absolute value of the difference by 10 scales the value to
	// 0-10 with 0 representing well balanced allocation and 10 poorly balanced. Subtracting it from
	// 10 leads to the score which also scales from 0 to 10 while 10 representing well balanced.
	diff := math.Abs(cpuFraction - memoryFraction)
	return int64((1 - diff) * float64(10))
}

func leastResourceScorer(node *k8sApi.Node) int64 {

	requestedCPU, _ := node.Status.Capacity.Cpu().AsInt64()
	allocatedCPU, _  := node.Status.Allocatable.Cpu().AsInt64()
	requestedMEM, _ := node.Status.Capacity.Memory().AsInt64()
	allocatedMEM, _  := node.Status.Allocatable.Memory().AsInt64()

	return (leastRequestedScore(requestedCPU, allocatedCPU) +
		leastRequestedScore(requestedMEM, allocatedMEM)) / 2
}

func leastRequestedScore(requested, capacity int64) int64 {
	if capacity == 0 {
		return 0
	}
	if requested > capacity {
		return 0
	}

	return ((capacity - requested) * int64(10)) / capacity
}
*/
