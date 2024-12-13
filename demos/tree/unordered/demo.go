package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.ibm.com/chic/chic-sched/pkg/system"
	"github.ibm.com/chic/chic-sched/pkg/topology"
)

// Demo operating on a tree

func main() {

	// misc parameters
	lineLength := 64

	// create a tree
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Create a sample tree")
	fmt.Println(strings.Repeat("=", lineLength))

	// create nodes
	nodeA := topology.NewNode(&system.Entity{ID: "A"})
	nodeB := topology.NewNode(&system.Entity{ID: "B"})
	nodeC := topology.NewNode(&system.Entity{ID: "C"})
	nodeD := topology.NewNode(&system.Entity{ID: "D"})
	nodeE := topology.NewNode(&system.Entity{ID: "E"})
	nodeF := topology.NewNode(&system.Entity{ID: "F"})
	nodeG := topology.NewNode(&system.Entity{ID: "G"})
	nodeH := topology.NewNode(&system.Entity{ID: "H"})
	nodeI := topology.NewNode(&system.Entity{ID: "I"})
	nodeJ := topology.NewNode(&system.Entity{ID: "J"})

	// connect nodes
	//A -> ( B -> ( D E ) C -> ( F ) )
	nodeA.AddChild(nodeD)
	nodeA.AddChild(nodeB)
	nodeA.AddChild(nodeC)

	nodeD.AddChild(nodeH)
	nodeD.AddChild(nodeJ)

	nodeB.AddChild(nodeE)
	nodeB.AddChild(nodeF)

	nodeC.AddChild(nodeI)
	nodeC.AddChild(nodeG)

	// create a tree
	tree := topology.NewTree(nodeA)
	fmt.Println("tree: ", tree)
	fmt.Println("height =", tree.GetHeight())
	fmt.Println("nodes =", tree.GetNodeIDs())
	fmt.Println("leaves =", tree.GetLeafIDs())
	fmt.Println()

	mapTree := tree.ToMap()
	jsonStr, err := json.Marshal(mapTree)
	if err != nil {
		fmt.Printf("mapTree error: %s", err.Error())
	} else {
		fmt.Println("mapTree =", string(jsonStr))
	}
	fmt.Println()

	var leafIDs []string
	leafIDs = []string{"E"}
	fmt.Printf("subtree %v: %v \n", leafIDs, tree.CopyByLeafIDs(leafIDs))
	leafIDs = []string{"F"}
	fmt.Printf("subtree %v: %v \n", leafIDs, tree.CopyByLeafIDs(leafIDs))
	leafIDs = []string{"E", "F"}
	fmt.Printf("subtree %v: %v \n", leafIDs, tree.CopyByLeafIDs(leafIDs))
	leafIDs = []string{"D", "E", "F"}
	fmt.Printf("subtree %v: %v \n", leafIDs, tree.CopyByLeafIDs(leafIDs))
	leafIDs = []string{}
	fmt.Printf("subtree %v: %v \n", leafIDs, tree.CopyByLeafIDs(leafIDs))
	fmt.Println()

	leavesDepth := nodeA.GetLeavesDepth()
	fmt.Printf("leavesDepth(%s): %v \n", nodeA.GetID(), leavesDepth)
	leavesDepth = nodeB.GetLeavesDepth()
	fmt.Printf("leavesDepth(%s): %v \n", nodeB.GetID(), leavesDepth)
	leavesDepth = nodeC.GetLeavesDepth()
	fmt.Printf("leavesDepth(%s): %v \n", nodeC.GetID(), leavesDepth)
	leavesDepth = nodeE.GetLeavesDepth()
	fmt.Printf("leavesDepth(%s): %v \n", nodeE.GetID(), leavesDepth)
	fmt.Println()

	leavesDistance := tree.GetLeavesDistanceFrom(nodeE.GetID())
	fmt.Printf("leavesDistance(%s): %v \n", nodeE.GetID(), leavesDistance)
	leavesDistance = tree.GetLeavesDistanceFrom(nodeD.GetID())
	fmt.Printf("leavesDistance(%s): %v \n", nodeD.GetID(), leavesDistance)
	leavesDistance = tree.GetLeavesDistanceFrom(nodeF.GetID())
	fmt.Printf("leavesDistance(%s): %v \n", nodeF.GetID(), leavesDistance)
	leavesDistance = tree.GetLeavesDistanceFrom(nodeB.GetID())
	fmt.Printf("leavesDistance(%s): %v \n", nodeB.GetID(), leavesDistance)
	fmt.Println()
}
