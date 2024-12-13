package main

import (
	"fmt"
	"github.com/ibm/chic-sched/demos"
	"github.com/ibm/chic-sched/pkg/builder"
	"github.com/ibm/chic-sched/pkg/placement"
	"github.com/ibm/chic-sched/pkg/util"
	"strings"
)

// Large system demo
//   - create a pTree and PEs given tree degrees at various levels
//   - generate random allocation with given distribution
//   - generate a group with level constraints
//   - place group, generate lTree
func main() {

	// system parameters
	degree := []int{3, 8, 20}
	capacity := []int{16}

	// loading parameters
	loadFactor := 0.4
	alpha := 0.250
	beta := 0.125
	cov := 1.5
	avg := demos.ComputeAverageLoad(loadFactor, alpha, beta)

	// group parameters
	groupSize := 120
	demand := []int{2}

	// constraint parameters
	// constraint at rack level
	level := 1
	affinity := util.Pack
	isHard := false
	lc1 := placement.NewLevelConstraint("lc-1", level, util.Affinity(affinity), isHard)

	//constraint at server level
	level = 0
	affinity = util.Spread
	isHard = false
	lc0 := placement.NewLevelConstraint("lc-0", level, util.Affinity(affinity), isHard)

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
	//fmt.Print(pTree)
	pes := tg.GetPEs()
	numServers := len(pes)

	// place some random allocation on the servers
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Place some load: ")
	fmt.Printf("(loadFactor=%f; alpha=%f; beta=%f; \n", loadFactor, alpha, beta)
	fmt.Printf("avg=%f; cov=%f)", avg, cov)
	fmt.Println()
	fmt.Println(strings.Repeat("=", lineLength))

	demos.PlaceBackgroungLoad(pes, loadFactor, alpha, beta, cov)
	pTree.PercolateResources()
	//fmt.Print(pTree)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Servers:")
	fmt.Println(strings.Repeat("=", lineLength))
	for i := 0; i < numServers; i++ {
		fmt.Println(pes[i])
	}

	// create level constraint
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Create level constraint:")
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println(lc1)
	fmt.Println(lc0)
	fmt.Println()

	// create placement group
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Create placement group:")
	fmt.Println(strings.Repeat("=", lineLength))
	groupDemand, _ := util.NewAllocationCopy(demand)
	pg := placement.NewPGroup("pg0", groupSize, groupDemand)
	pg.AddLevelConstraint(lc1)
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
	//fmt.Print(pTree)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Allocation on servers:")
	fmt.Println(strings.Repeat("=", lineLength))
	for i := 0; i < numServers; i++ {
		fmt.Println(pes[i])
	}
}
