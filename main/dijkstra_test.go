package main

import (
	"fmt"
	"math"
	"testing"
)

/*
func TestEmptyGraph(t *testing.T) {
	g := make(Graph)

	_, _, err := g.Path("a", "z")
	if err == nil {
		t.Error("Error nil; want error message")
	}
}
*/
/*
func TestGraphErrors(t *testing.T) {
	g := Graph{
		"a": {"b": 20, "c": 80},
		"b": {"a": 20, "c": 20},
		"c": {"a": 80, "b": 20},
	}

	_, _, err := g.Path("a", "z")
	if err == nil {
		t.Error("err = nil; want not in graph error")
	}

	_, _, err = g.Path("z", "c")
	if err == nil {
		t.Error("err = nil; want not in graph error")
	}
}


func TestPath1(t *testing.T) {
	g := Graph{
		"a": {"b": 20, "c": 80},
		"b": {"a": 20, "c": 20},
		"c": {"a": 80, "b": 20},
	}

	// The shortest path is correct
	path, cost, err := g.Path("a", "c")
	if err != nil {
		t.Errorf("err = %v; want nil", err)
	}

	expectedPath := []string{"a", "b", "c"}

	if len(path) != len(expectedPath) {
		t.Errorf("path = %v; want %v", path, expectedPath)
	}
	for i, key := range path {
		if key != expectedPath[i] {
			t.Errorf("path = %v; want %v", path, expectedPath)
		}
	}

	expectedCost := 40
	if cost != expectedCost {
		t.Errorf("cost = %v; want %v", cost, expectedCost)
	}
}

func TestPath2(t *testing.T) {
	g := Graph{
		"a": {"b": 7, "c": 9, "f": 14},
		"b": {"c": 10, "d": 15},
		"c": {"d": 11, "f": 2},
		"d": {"e": 6},
		"e": {"f": 9},
	}

	// The shortest path is correct
	path, _, err := g.Path("a", "e")
	if err != nil {
		t.Errorf("err = %v; want nil", err)
	}

	expectedPath := []string{"a", "c", "d", "e"}

	if len(path) != len(expectedPath) {
		t.Errorf("path = %v; want %v", path, expectedPath)
	}
	for i, key := range path {
		if key != expectedPath[i] {
			t.Errorf("path = %v; want %v", path, expectedPath)
		}
	}
}

func BenchmarkPath(b *testing.B) {
	g := Graph{
		"a": {"b": 20, "c": 80},
		"b": {"a": 20, "c": 20},
		"c": {"a": 80, "b": 20},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g.Path("a", "c")
	}
}

func ExampleGraph_Path() {
	g := Graph{
		"a": {"b": 20, "c": 80},
		"b": {"a": 20, "c": 20},
		"c": {"a": 80, "b": 20},
	}

	path, cost, _ := g.Path("a", "c") // skipping error handling

	fmt.Printf("path: %v, cost: %v", path, cost)
	// Output: path: [a b c], cost: 40
}
*/

func TestNOMS2020v2(t *testing.T) {
	g := Graph{
		"1": 	{"sw1": 3.0},
		"2": 	{"sw1": 3.0},
		"3": 	{"sw1": 5.0},
		"sw1": 	{"1": 3.0, "2": 3.0, "3": 5.0,"sw2": 15.0},
		"4":	{"sw2": 3.0},
		"5": 	{"sw2": 5.0},
		"6": 	{"sw2": 3.0},
		"sw2": 	{"4": 3.0, "5": 5.0, "6": 3.0, "sw1": 15.0 ,"sw3": 25.0},
		"7": 	{"sw4": 3},
		"8": 	{"sw4": 3},
		"9": 	{"sw4": 5},
		"sw3": 	{"13": 1.0, "14": 1.0, "m": 1.0, "sw4": 25.0,"sw2": 25.0},
		"13": 	{"sw3": 1.0},
		"14": 	{"sw3": 1.0},
		"m": 	{"sw3": 1.0},
		"sw4": 	{"7": 3.0, "8": 3.0, "9": 5.0, "sw3": 25.0, "sw5": 15.0},
		"10": 	{"sw5": 3.0},
		"11": 	{"sw5": 3.0},
		"12": 	{"sw5": 3.0},
		"sw5": 	{"10": 3.0, "11": 3.0, "12": 3.0, "sw4": 15.0},
	}

	//path, cost, _ := g.Path("1", "2") // skipping error handling

	//fmt.Printf("Test path: %v, cost: %v", path, cost)

	// Waste Management Use Case
	replicasAPI := 3
	replicasWasteDB := 4
	replicasRP := 4
	replicasServer := 3

	podsAPI := [3]string{"3","5","7"}
	podsWasteDB := [4]string{"4","10","11","14"}
	podsRP := [4]string{"6","13","14", "m"}
	podsServer := [3]string{"13","14","m"}

	// Surveillance Camera Use Case
	replicasFD := 4
	replicasFM := 2
	replicasCamDB := 2
	replicasDash := 4

	podsFD := [4]string{"1", "2", "12","m"}
	podsFM := [2]string{"13","m"}
	podsCamDB := [2]string{"7","8"}
	podsDash := [4]string{"8","9","10","11"}

	latency := 0.0
	number := 0
	totalCost := 0.0
	max := 0.0
	min := math.MaxFloat64
	var numberCost [144]float64

	for i := 0; i < replicasAPI; i++ {
		for j := 0; j < replicasWasteDB; j++ {
			for k := 0; k < replicasRP; k++ {
				for l := 0; l < replicasServer; l++ {
						number = number + 1
						_, cost1, _ := g.Path(podsAPI[i], podsWasteDB[j])
						//fmt.Printf("cost1: %v ms \n", float64(cost1))
						_, cost2, _ := g.Path(podsWasteDB[j], podsRP[k])
						//fmt.Printf("cost2: %v ms \n", float64(cost2))
						_, cost3, _ := g.Path(podsRP[k], podsServer[l])
						//fmt.Printf("cost3: %v ms \n", float64(cost3))
						totalCost = float64(cost1 + cost2 + cost3)
						numberCost[number-1] = totalCost
						//fmt.Printf("Total Cost: %v ms \n", float64(totalCost))
						latency = latency + totalCost
						//fmt.Printf("Updated Latency: %v ms \n", float64(latency))
						if max < totalCost{
							max = totalCost
						}
						if min > totalCost{
							min = totalCost
						}
				}
			}
		}
	}

	fmt.Printf("Waste Management Use Case: -----------\n")
	fmt.Printf("latency cost: %v ms \n", float64(latency))
	fmt.Printf("Number of Runs: %v \n", number)
	fmt.Printf("av Latency: %v ms \n", latency/float64(number))
	fmt.Printf("max Latency: %v ms \n", max)
	fmt.Printf("min Latency: %v ms \n", min)

	std := 0.0
	avg := latency/float64(number)
	for i := 0; i < number; i++ {
		std = std + math.Pow(float64(numberCost[i] - avg), 2)
	}

	std = math.Sqrt(std/float64(number))
	fmt.Printf("Standard Deviation: %v ms \n", std)

	latency = 0.0
	number = 0
	totalCost = 0.0
	max = 0.0
	min = math.MaxFloat64
	var numberCost2 [64]float64

	for i := 0; i < replicasFD; i++ {
		for j := 0; j < replicasFM; j++ {
			for k := 0; k < replicasCamDB; k++ {
				for l := 0; l < replicasDash; l++ {
					number = number + 1
					_, cost1, _ := g.Path(podsFD[i], podsFM[j])
					//fmt.Printf("cost1: %v ms \n", float64(cost1))
					_, cost2, _ := g.Path(podsFM[j], podsCamDB[k])
					//fmt.Printf("cost2: %v ms \n", float64(cost2))
					_, cost3, _ := g.Path(podsCamDB[k], podsDash[l])
					//fmt.Printf("cost3: %v ms \n", float64(cost3))
					totalCost = float64(cost1 + cost2 + cost3)
					//fmt.Printf("Total Cost: %v ms \n", float64(totalCost))
					numberCost2[number-1] = totalCost
					latency = latency + totalCost
					//fmt.Printf("Updated Latency: %v ms \n", float64(latency))

					if max < totalCost{
						max = totalCost
					}
					if min > totalCost{
						min = totalCost
					}

				}
			}
		}
	}

	fmt.Printf("Surveillance Camera Use Case: -----------\n")
	fmt.Printf("latency cost: %v ms \n", float64(latency))
	fmt.Printf("Number of Runs: %v \n", number)
	fmt.Printf("av Latency: %v ms \n", latency/float64(number))
	fmt.Printf("max Latency: %v ms \n", max)
	fmt.Printf("min Latency: %v ms \n", min)

	std = 0.0
	avg = latency/float64(number)
	for i := 0; i < number; i++ {
		std = std + math.Pow(float64(numberCost2[i] - avg), 2)
	}

	std = math.Sqrt(std/float64(number))
	fmt.Printf("Standard Deviation: %v ms \n", std)

	// path, cost, _ := g.Path("5", "6") // skipping error handling
	// Output: path: [a b c], cost: 40
}