package main

import (
	"flag"
	"fmt"
	"github.com/ibm/chic-sched/demos"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"k8s.io/klog/v2"

	"github.com/ibm/chic-sched/pkg/placement"
	"github.com/ibm/chic-sched/pkg/system"
	"github.com/ibm/chic-sched/pkg/topology"
	"github.com/ibm/chic-sched/pkg/util"
)

// Small system demo
//   - create a pTree and PEs given tree degrees at various levels
//   - generate random allocation using normal distribution
//   - generate a group with level constraints
//   - place group, generate lTree
//   - deallocate resources
func main() {
	klog.InitFlags(nil)
	flag.Set("v", "4")
	flag.Set("skip_headers", "true")
	klog.SetOutput(os.Stdout)
	flag.Parse()
	defer klog.Flush()

	isPrint := false

	// parameters
	numResources := 2
	maxLevel := 2
	numServers := 5
	serverCapacity := []int{16, 256}
	groupSize := 4
	demand := []int{4, 32}

	// misc parameters
	lineLength := 64

	groupDemand, _ := util.NewAllocationCopy(demand)

	// create non-server nodes in the physical tree
	root := topology.NewPNode(topology.NewNode(&system.Entity{ID: "root"}), maxLevel, numResources)
	rack0 := topology.NewPNode(topology.NewNode(&system.Entity{ID: "rack-0"}), maxLevel-1, numResources)
	rack1 := topology.NewPNode(topology.NewNode(&system.Entity{ID: "rack-1"}), maxLevel-1, numResources)

	// create servers
	capacity, _ := util.NewAllocationCopy(serverCapacity)
	allocated, _ := util.NewAllocationCopy(demand)

	fmt.Println("Servers:")
	pes := make([]*system.PE, numServers)
	servers := make([]*topology.PNode, numServers)
	for i := 0; i < numServers; i++ {
		namei := "server" + strconv.FormatInt(int64(i), 10)
		pei := system.NewPE(namei, capacity)
		pes[i] = pei
		fmt.Println(pei)
		pei.SetAllocated(allocated.Clone())
		nodei := topology.NewNode((*system.Entity)(unsafe.Pointer(pei)))
		servers[i] = topology.NewPNode(nodei, 0, numResources)
	}
	fmt.Println()

	fmt.Println("Place some load:")
	pes[0].GetAllocated().Scale(3)
	pes[1].GetAllocated().Scale(2)
	pes[2].GetAllocated().Scale(1)
	pes[3].GetAllocated().Scale(2)
	pes[4].GetAllocated().Scale(3)
	for i := 0; i < numServers; i++ {
		fmt.Println(pes[i])
	}
	fmt.Println()

	// place random weights
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())
	fmt.Print("Place some weights: ")
	fmt.Println()
	demos.PlaceWeights(pes)
	for i := 0; i < numServers; i++ {
		fmt.Println(pes[i])
	}
	fmt.Println()

	// build topology
	fmt.Println("Build physical topology:")
	root.AddChild(&rack0.Node)
	root.AddChild(&rack1.Node)

	rack0.AddChild(&servers[0].Node)
	rack0.AddChild(&servers[1].Node)
	rack0.AddChild(&servers[2].Node)

	rack1.AddChild(&servers[3].Node)
	rack1.AddChild(&servers[4].Node)

	tree := topology.NewTree(&root.Node)
	pTree := topology.NewPTree(tree)
	pTree.PercolateResources()
	fmt.Print(pTree)

	// create level constraint
	fmt.Println("Create level constraints:")

	// constraint at rack level
	level := 1
	affinity := util.Pack
	isHard := false
	lc1 := placement.NewLevelConstraint("lc-1", level, util.Affinity(affinity), isHard)
	fmt.Println(lc1)

	//constraint at server level
	level = 0
	affinity = util.Spread
	isHard = false
	lc0 := placement.NewLevelConstraint("lc-0", level, util.Affinity(affinity), isHard)
	fmt.Println(lc0)
	fmt.Println()

	defaultPolicy := true
	ByWeightPolicy := true
	ByWeightProductPolicy := true
	ByFitWeightProductPolicy := true
	ByMinWeightedAvailability := true

	if defaultPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Default Strategy (no weights)")
		fmt.Println("Create placement group:")
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)

		// place group
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroup(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// allocate resources
		if isPrint {
			fmt.Println("Allocate logical tree:")
		}
		pg.ClaimAll(pTree)

		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}

		fmt.Println("Servers:")
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// de-allocate resources
		if isPrint {
			fmt.Println("DeAllocate logical tree:")
		}
		pg.UnClaimAll(pTree)
		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}
	}
	if ByWeightPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByWeight Strategy")
		fmt.Println("Create placement group:")
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)

		// place group
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroupByWeight(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// allocate resources
		if isPrint {
			fmt.Println("Allocate logical tree:")
		}
		pg.ClaimAll(pTree)

		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}

		fmt.Println("Servers:")
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// de-allocate resources
		if isPrint {
			fmt.Println("DeAllocate logical tree:")
		}
		pg.UnClaimAll(pTree)
		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}
	}
	if ByWeightProductPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByWeightProduct Strategy")
		fmt.Println("Create placement group:")
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)

		// place group
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroupByWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// allocate resources
		if isPrint {
			fmt.Println("Allocate logical tree:")
		}
		pg.ClaimAll(pTree)

		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}

		fmt.Println("Servers:")
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// de-allocate resources
		if isPrint {
			fmt.Println("DeAllocate logical tree:")
		}
		pg.UnClaimAll(pTree)
		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}
	}
	if ByFitWeightProductPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByFitWeightProduct Strategy")
		fmt.Println("Create placement group:")
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)

		// place group
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroupByFitWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// allocate resources
		if isPrint {
			fmt.Println("Allocate logical tree:")
		}
		pg.ClaimAll(pTree)

		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}

		fmt.Println("Servers:")
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// de-allocate resources
		if isPrint {
			fmt.Println("DeAllocate logical tree:")
		}
		pg.UnClaimAll(pTree)
		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}
		fmt.Println(strings.Repeat("=", lineLength))
	}
	if ByMinWeightedAvailability {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByMinWeightedAvailability Strategy")
		fmt.Println("Create placement group:")
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)

		// place group
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroupByMinWeightedAvailability(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// allocate resources
		if isPrint {
			fmt.Println("Allocate logical tree:")
		}
		pg.ClaimAll(pTree)

		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}

		fmt.Println("Servers:")
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// de-allocate resources
		if isPrint {
			fmt.Println("DeAllocate logical tree:")
		}
		pg.UnClaimAll(pTree)
		if isPrint {
			fmt.Print(pTree)
			fmt.Print(pg)
		}
		fmt.Println(strings.Repeat("=", lineLength))
	}
}
