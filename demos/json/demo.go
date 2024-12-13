package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.ibm.com/chic/chic-sched/pkg/builder"
)

func main() {
	fName := "../../samples/testTree.json"

	// create topology tree from json file
	jsonTree, err := os.ReadFile(fName)
	if err != nil {
		fmt.Printf("error reading quota tree file: %s", fName)
		return
	}
	pTree, err := builder.CreateTopologyTreeFromJson(string(jsonTree))
	if err != nil {
		fmt.Println("error creating tree ", err.Error())
		return
	}

	// create subtree with a random subset of leaf nodes
	numSampled := 4
	leafIDs := pTree.GetLeafIDs()
	numLeaves := len(leafIDs)
	selectedLeavesMap := make(map[string]bool)
	for i := 0; i < numSampled; i++ {
		index := rand.Intn(numLeaves)
		id := leafIDs[index]
		if !selectedLeavesMap[id] {
			selectedLeavesMap[id] = true
		}
	}
	j := 0
	selectedLeaves := make([]string, len(selectedLeavesMap))
	for k := range selectedLeavesMap {
		selectedLeaves[j] = k
		j++
	}
	fmt.Println("selectedLeaves: ", selectedLeaves)

	//selectedLeaves := []string{"node-3", "node-3", "node-5"}
	selectPTree := pTree.CopyByLeafIDs(selectedLeaves)
	fmt.Println("selectPTree: ", selectPTree)

	// create a flat topology
	numResources := 1
	flatPTree := builder.CreateFlatTopology(leafIDs, numResources)
	fmt.Println("flatPTree: ", flatPTree)
}
