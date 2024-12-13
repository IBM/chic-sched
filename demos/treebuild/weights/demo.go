package main

import (
	"fmt"
	"github.com/ibm/chic-sched/demos"
	"github.com/ibm/chic-sched/pkg/builder"
	"github.com/ibm/chic-sched/pkg/placement"
	"github.com/ibm/chic-sched/pkg/util"
	"math/rand"
	"strings"
	"time"
)

// Demo creating a pTree and PEs given tree degrees at various levels
// and placing a group

func main() {

	// system parameters
	degree := []int{2, 3}
	capacity := []int{16, 128}

	// loading parameters
	loadFactor := 0.5
	cov := 0.5

	// group parameters
	groupSize := 4
	demand := []int{4, 32}

	// constraint parameters
	level := 1
	affinity := util.Pack
	isHard := false

	// misc parameters
	lineLength := 64

	// create physical tree and servers
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Print("Create physical tree topology: ")
	fmt.Printf("(degree=%v; capacity=%v)", degree, capacity)
	fmt.Println()
	fmt.Println(strings.Repeat("=", lineLength))

	tg := builder.NewTreeGen()
	pTree := tg.CreateUniformTree(degree, capacity)
	fmt.Print(pTree)
	pes := tg.GetPEs()
	numServers := len(pes)

	// place some random allocation on the servers
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Print("Place some load: ")
	fmt.Printf("(avg=%f; cov=%f)", loadFactor, cov)
	fmt.Println()
	fmt.Println(strings.Repeat("=", lineLength))

	demos.PlaceBackgroungLoad(pes, loadFactor, 0, 0, cov)
	pTree.PercolateResources()
	fmt.Print(pTree)

	// place random weights
	fmt.Println(strings.Repeat("=", lineLength))
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())
	fmt.Print("Place some weights: ")
	fmt.Println()
	fmt.Println(strings.Repeat("=", lineLength))
	demos.PlaceWeights(pes)
	pTree.PercolateResources()
	fmt.Print(pTree)

	// create level constraint
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Create level constraint:")
	fmt.Println(strings.Repeat("=", lineLength))
	lc0 := placement.NewLevelConstraint("lc0", level, util.Affinity(affinity), isHard)
	fmt.Println(lc0)
	fmt.Println()

	// create placement group
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Create placement group:")
	fmt.Println(strings.Repeat("=", lineLength))
	groupDemand, _ := util.NewAllocationCopy(demand)
	pg := placement.NewPGroup("pg0", groupSize, groupDemand)
	pg.AddLevelConstraint(lc0)
	fmt.Println(pg)
	fmt.Println()

	// place group
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Place group result: (logical tree)")
	fmt.Println(strings.Repeat("=", lineLength))
	p := placement.NewPlacer(pTree)
	_, err := p.PlaceGroup(pg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(pg)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Physical tree after group allocation:")
	fmt.Println(strings.Repeat("=", lineLength))
	pg.ClaimAll(pTree)
	fmt.Print(pTree)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Allocation on servers:")
	fmt.Println(strings.Repeat("=", lineLength))
	for i := 0; i < numServers; i++ {
		fmt.Println(pes[i])
	}
	fmt.Println()

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Logical tree after group allocation:")
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Print(pg.GetLTree())

	// unplace group
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Physical tree after group de-allocation:")
	fmt.Println(strings.Repeat("=", lineLength))
	pg.UnClaimAll(pTree)
	fmt.Print(pTree)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("logical tree after group de-allocation:")
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Print(pg.GetLTree())
	fmt.Println()

	// create placement group
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Create placement group (weights):")
	fmt.Println(strings.Repeat("=", lineLength))
	groupDemandWeights, _ := util.NewAllocationCopy(demand)
	pgWeights := placement.NewPGroup("pg0", groupSize, groupDemandWeights)
	pgWeights.AddLevelConstraint(lc0)
	fmt.Println(pgWeights)
	fmt.Println()

	// place group with weights
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Place group result (ByWeight): (logical tree)")
	fmt.Println(strings.Repeat("=", lineLength))
	pWeights := placement.NewPlacer(pTree)
	_, errWeights := pWeights.PlaceGroupByWeight(pgWeights)
	if errWeights != nil {
		fmt.Println(errWeights)
		return
	}
	fmt.Print(pgWeights)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Physical tree after group allocation (weights):")
	fmt.Println(strings.Repeat("=", lineLength))
	pgWeights.ClaimAll(pTree)
	fmt.Print(pTree)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Allocation on servers (weights):")
	fmt.Println(strings.Repeat("=", lineLength))
	for i := 0; i < numServers; i++ {
		fmt.Println(pes[i])
	}
	fmt.Println()

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Logical tree after group allocation (weights):")
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Print(pgWeights.GetLTree())

	// unplace group
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Physical tree after group de-allocation (weights):")
	fmt.Println(strings.Repeat("=", lineLength))
	pgWeights.UnClaimAll(pTree)
	fmt.Print(pTree)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("logical tree after group de-allocation (weights):")
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Print(pgWeights.GetLTree())
	fmt.Println()
}
