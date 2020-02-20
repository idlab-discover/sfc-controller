package main

import (
	"log"
)

//pod structure
type pod struct {
	name          string
	namespace     string
	key           string
	minBandwidth  float64
	nodeAllocated string
	next          *pod
}

//List structure
type podList struct {
	name    string
	head    *pod
	current *pod
}

// create the list
func createPodList(name string) *podList {
	return &podList{
		name: name,
	}
}

// addPod() adds an element to the list
func (p *podList) addPod(name string, namespace string, key string, minBandwidth float64, nodeAllocated string) error {
	s := &pod{
		name:          name,
		namespace:     namespace,
		minBandwidth:  minBandwidth,
		key:           key,
		nodeAllocated: nodeAllocated,
	}
	if p.head == nil {
		p.head = s
	} else {
		currentPod := p.head
		for currentPod.next != nil {
			currentPod = currentPod.next
		}
		currentPod.next = s
	}
	return nil
}

// removePod() removes a particular element from the list
func (p *podList) removePod(name string) *podList {

	// List is empty, cannot remove Pod
	if p.isEmpty() {
		log.Printf("list is empty")
	}

	// auxiliary variable
	temp := p.head

	// If head holds the Pod to be deleted
	if temp != nil && temp.name == name {
		p.head = temp.next // Changed head
		return p
	}

	// Search for the Pod to be deleted, keep track of the
	// previous Pod as we need to change temp.next
	for temp != nil {
		if temp.next.name == name {
			temp.next = temp.next.next
			return p
		}
		temp = temp.next
	}
	return p
}

// showAllPods() prints all elements on the list
func (p *podList) showAllPods() error {
	currentPod := p.head
	if currentPod == nil {
		log.Printf("PodList is empty.")
		return nil
	}
	log.Printf("%v \n", currentPod.name)
	for currentPod.next != nil {
		currentPod = currentPod.next
		log.Printf("%v \n", currentPod.name)
	}

	return nil
}

//start() returns the first/head element
func (p *podList) start() *pod {
	p.current = p.head
	return p.current
}

//next() returns the next element on the list
func (p *podList) next() *pod {
	p.current = p.current.next
	return p.current
}

// IsEmpty() returns true if the list is empty
func (p *podList) isEmpty() bool {
	if p.head == nil {
		return true
	}
	return false
}

// getSize() returns the linked list size
func (p *podList) getSize() int {
	size := 1
	last := p.head
	for {
		if last == nil || last.next == nil {
			break
		}
		last = last.next
		size++
	}
	return size
}
