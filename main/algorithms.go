package main

import (
	"fmt"
	k8sApi "k8s.io/api/core/v1"
	"math"
	"math/rand"
	"time"
)

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
		_, delay, _ := graphLatency.Path(node.Name, targetLocation)
		if minDelay == float64(delay) {
			nodeBand := getBandwidthValue(&node, "avBandwidth")
			if podMinBandwith <= nodeBand {
				fmt.Printf("Selected Node: %v \n", node.Name)
				return node
			} else {
				fmt.Printf("Remove candidate Node %v , available Bandwidth not enough: %v (Mi)\n", node.Name, nodeBand)
				copyItems = append(nodes.Items[:i], nodes.Items[i+1:]...)
			}
		}
	}

	if len(copyItems) == 0 {
		return k8sApi.Node{}
	} else {
		nodes.Items = copyItems
		newDelay := getMinDelay(nodes, targetLocation)
		fmt.Printf("---------------------Recursive Iteration ------------------\n")
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
			fmt.Printf("Node: %v \n", node.Name)
			podList.start()
			for podList.current != nil {
				// calculate each shortest path
				_, cost, _ := graphLatency.Path(node.Name, podList.current.nodeAllocated)
				fmt.Printf("Current Cost: %v \n", cost)
				previousValue := getValue(delayCost, node.Name)
				fmt.Printf("Previous Cost: %v \n", previousValue)
				delayCost[node.Name] = previousValue + float64(cost)
				fmt.Printf("Updated Cost: %v \n", delayCost[node.Name])
				podList.next()
			}

			minCost = math.Min(minCost, float64(delayCost[node.Name]))

			if prevCost > minCost {
				prevCost = minCost
				selectedNode = node
				fmt.Printf("Updated min Node (Delay Cost): %v \n", node.Name)
			}
		} else {
			fmt.Printf("Node %v av bandwidth not enough!\n", node.Name)
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
		fmt.Printf("Node: %v - Cost: %v \n", node.Name, linkCost[node.Name])

		if prevCost < linkCost[node.Name] && linkCost[node.Name] <= 1.0 {
			prevCost = linkCost[node.Name]
			selectedNode = node
			fmt.Printf("Updated Max Node (Link Cost): %v \n", node.Name)
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
