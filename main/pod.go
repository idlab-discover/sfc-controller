package main

import "fmt"

type pod struct {
	key string
	nodeAllocated string
	next *pod
}

type podList struct {
	name       string
	head       *pod
	current	   *pod
}

func createPodList(name string) *podList {
	return &podList{
		name: name,
	}
}

func (p *podList) addPod(key, nodeAllocated string) error {
	s := &pod{
		key:   key,
		nodeAllocated: nodeAllocated,
	}
	if p.head == nil {
		p.head = s
	} else {
		currentNode := p.head
		for currentNode.next != nil {
			currentNode = currentNode.next
		}
		currentNode.next = s
	}
	return nil
}

func (p *podList) showAllPods() error {
	currentNode := p.head
	if currentNode == nil {
		fmt.Println("PodList is empty.")
		return nil
	}
	fmt.Printf("%+v\n", *currentNode)
	for currentNode.next != nil {
		currentNode = currentNode.next
		fmt.Printf("%+v\n", *currentNode)
	}

	return nil
}

func (p *podList) start() *pod {
	p.current = p.head
	return p.current
}

func (p *podList) next() *pod {
	p.current = p.current.next
	return p.current
}

// IsEmpty returns true if the list is empty
func (p *podList) isEmpty() bool {
	if p.head == nil {
		return true
	}
	return false
}

// Size returns the linked list size
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