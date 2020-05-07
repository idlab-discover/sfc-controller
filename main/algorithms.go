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

	// extract information from Pod - label values
	appName := getDesiredFromLabels(pod, "app")
	targetLocation := getDesiredFromLabels(pod, "targetLocation")
	minBandwidth := getDesiredFromLabels(pod, "minBandwidth")
	chainPosString := getDesiredFromLabels(pod, "chainPosition")
	nsh := getDesiredFromLabels(pod, "networkServiceHeader")
	totalChain := getDesiredFromLabels(pod, "totalChainServ")
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

			// update Link bandwidth
			nodeBand := getBandwidthValue(&node, "avBandwidth")
			value := nodeBand - podMinBandwith

			label := strconv.FormatFloat(value, 'f', 2, 64)

			err := updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &node)
			if err != nil {
				log.Printf("encountered error when updating Bandiwdth label: %v", err)
			}

			err = allocatedPods.addPod(pod.Name, pod.Namespace, key, podMinBandwith, node.Name)
			if err != nil {
				log.Printf("encountered error when adding Pod to the List: %v", err)
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
						allocatedNode, ok := serviceHash[key]
						if ok {
							log.Printf("Key found! Allocated on Node: %v \n", allocatedNode)
							err := podList.addPod("pod", "dijkstra", key, 0.0, allocatedNode)
							if err != nil {
								log.Printf("encountered error when adding Pod to the List: %v", err)
							}
						}
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

				// update Link bandwidth
				nodeBand := getBandwidthValue(&nodeDelay, "avBandwidth")
				value := nodeBand - podMinBandwith

				label := strconv.FormatFloat(value, 'f', 2, 64)

				err := updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &nodeDelay)
				if err != nil {
					log.Printf("encountered error when updating label: %v", err)
				}

				err = allocatedPods.addPod(pod.Name, pod.Namespace, key, podMinBandwith, nodeDelay.Name)
				if err != nil {
					log.Printf("encountered error when adding Pod to the List: %v", err)
				}

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

					// update Link bandwidth
					nodeBand := getBandwidthValue(&node, "avBandwidth")
					value := nodeBand - podMinBandwith

					label := strconv.FormatFloat(value, 'f', 2, 64)

					err := updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &node)
					if err != nil {
						log.Printf("encountered error when updating bandwidth label: %v", err)
					}

					err = allocatedPods.addPod(pod.Name, pod.Namespace, key, podMinBandwith, node.Name)
					if err != nil {
						log.Printf("encountered error when adding Pod to the List: %v", err)
					}

					return []k8sApi.Node{node}, nil
				}
			}
		}
	}
	// Link MAX Cost Selection
	log.Printf("---------------------------------------------------------------\n")
	log.Printf("-------------------- MAX Link Cost Selection ------------------\n")

	nodeMaxLink, _ := calculateMaxLinkCost(nodes, podMinBandwith)

	if nodeMaxLink.GetName() != "" {
		log.Printf("Node Max Link selected! \n")

		// add pod to Service Hash
		id++
		key := getKey(id, appName, nsh, chainPos, totalChainServ)
		addService(key, nodeMaxLink)

		// update Link bandwidth
		nodeBand := getBandwidthValue(&nodeMaxLink, "avBandwidth")
		value := nodeBand - podMinBandwith

		label := strconv.FormatFloat(value, 'f', 2, 64)

		err := updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &nodeMaxLink)
		if err != nil {
			log.Printf("encountered error when updating bandwidth label: %v", err)
		}

		err = allocatedPods.addPod(pod.Name, pod.Namespace, key, podMinBandwith, nodeMaxLink.Name)
		if err != nil {
			log.Printf("encountered error when adding Pod to the List: %v", err)
		}

		return []k8sApi.Node{nodeMaxLink}, nil
	}

	log.Printf("---------------------------------------------------------------\n")
	log.Printf("---------------- Last Solution: Random Selection --------------\n")

	pick := randomSelection(nodes)
	// add pod to Service Hash
	id++
	key := getKey(id, appName, nsh, chainPos, totalChainServ)
	addService(key, pick)

	// update Link bandwidth
	nodeBand := getBandwidthValue(&pick, "avBandwidth")
	value := nodeBand - podMinBandwith

	if value < 0 {
		value = 0.0
	}

	label := strconv.FormatFloat(value, 'f', 2, 64)

	err := updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, &pick)
	if err != nil {
		log.Printf("encountered error when updating bandwidth label: %v", err)
	}

	err = allocatedPods.addPod(pod.Name, pod.Namespace, key, podMinBandwith, pick.Name)
	if err != nil {
		log.Printf("encountered error when adding Pod to the List: %v", err)
	}

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
			//log.Printf("Node: %v \n", node.Name)
			podList.start()
			for podList.current != nil {
				// calculate each shortest path
				cost, _ := graphLatency.getPath(node.Name, podList.current.nodeAllocated)
				//log.Printf("Current Cost: %v \n", cost)
				previousValue := getValue(delayCost, node.Name)
				//log.Printf("Previous Cost: %v \n", previousValue)
				delayCost[node.Name] = previousValue + float64(cost)
				//log.Printf("Updated Cost: %v \n", delayCost[node.Name])
				podList.next()
			}

			minCost = math.Min(minCost, float64(delayCost[node.Name]))

			if prevCost > minCost {
				prevCost = minCost
				selectedNode = node
				log.Printf("Updated min Node (Delay Cost): %v / %v \n", minCost, node.Name)
			}
		} else {
			//log.Printf("Node %v av bandwidth not enough!\n", node.Name)
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
		//log.Printf("Node: %v - Cost: %v \n", node.Name, linkCost[node.Name])

		if prevCost < linkCost[node.Name] && linkCost[node.Name] <= 1.0 {
			prevCost = linkCost[node.Name]
			selectedNode = node
			log.Printf("Updated Max Node (Link Cost): %v / %v \n", prevCost, node.Name)
		}
	}
	return selectedNode, linkCost
}

// Verify scheduled Pods: Otherwise free bandwidth
func watchScheduledPods(scheduler Scheduler) {

	log.Printf("--------------- Watching Pods ------------\n")

	// Update allocatedPods List
	log.Printf("Initial List: ")
	err := allocatedPods.showAllPods()
	if err != nil {
		log.Printf("encountered error when printing pods: %v", err)
		return
	}

	// Check if allocated Pods are still deployed. Otherwise free bandwidth on the correspondent node
	nodes := scheduler.nodeLister
	pods := scheduler.podLister

	// Start Checking
	allocatedPods.start()
	for allocatedPods.current != nil {

		//log.Printf("Found a pod to check: %v / %v", allocatedPods.current.namespace, allocatedPods.current.name)

		// Using Pod Lister instead of API! Faster Access!
		podScheduled, err := pods.Pods(allocatedPods.current.namespace).Get(allocatedPods.current.name)

		if err != nil {
			// Pod is not deployed anymore!
			log.Printf("Check failed: Pod %s is not deployed anymore", allocatedPods.current.name)

			// Remove hash key / update node bandwidth
			nodeName := allocatedPods.current.nodeAllocated
			key := allocatedPods.current.key

			delete(serviceHash, key)

			log.Printf("Service Hash removed...")

			// Get node from node lister
			node, err := nodes.Get(nodeName)
			if err != nil {
				log.Printf("cannot find node %v", err.Error())
				return
			}

			// Get Node current Bandwidth
			nodeBand := getBandwidthValue(node, "avBandwidth")

			// Get Pod Min Bandwidth Requirement from List
			podMinBandwith := allocatedPods.current.minBandwidth

			// Update current avBandwidth
			newValue := nodeBand + podMinBandwith
			label := strconv.FormatFloat(newValue, 'f', 2, 64)

			err = updateBandwidthLabel(label, scheduler.clientset, scheduler.nodeLister, node)
			if err != nil {
				log.Printf("encountered error when updating label after pod verification: %v", err)
				return
			}

			log.Printf("Bandwidth updated...")

			//remove from list of allocated pods
			log.Printf("Remove Pod from List...")
			allocatedPods = allocatedPods.removePod(allocatedPods.current.name)

		} else {
			// Check confirmed
			log.Printf("Check confirmed: Pod still alocated %s", podScheduled.Name)
		}
		allocatedPods.next()
	}

	log.Printf("Updated List: ")
	err = allocatedPods.showAllPods()
	if err != nil {
		log.Printf("encountered error when printing pods: %v", err)
		return
	}
	return
}
