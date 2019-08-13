package main

import (
	"fmt"
	"testing"
)

func TestServiceHash(t *testing.T) {
	key := getKey(1, "birch-api", "anomalyDetection-v1", 1, 3)

	fmt.Printf("key: %v \n", key)

	var serviceHash = make(map[string]string)

	serviceHash[key] = "work8.kbcluster2.wall2-ilabt-iminds-be.wall1.ilabt.iminds.be"

	fmt.Printf("Map key: %v \n", serviceHash[key])

	fmt.Printf("String %v to int %v \n", "3",stringtoInt("3"))
}

