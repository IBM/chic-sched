package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/ibm/chic-sched/demos"
	"github.com/ibm/chic-sched/pkg/builder"
	"github.com/ibm/chic-sched/pkg/placement"
	"github.com/ibm/chic-sched/pkg/system"
	"github.com/ibm/chic-sched/pkg/util"
)

// Large system demo
//   - create a pTree and PEs given tree degrees at various levels
//   - generate random allocation with given distribution
//   - generate a group with level constraints
//   - place group, generate lTree
//   - partial claim of group resources
//   - alter system allocation state
//   - place partially placed group given new system state
func main() {

	// system parameters
	degree := []int{2, 4, 8}
	capacity := []int{16}

	// loading parameters
	loadFactor := 0.4
	alpha := 0.250
	beta := 0.125
	cov := 1.5
	avg := demos.ComputeAverageLoad(loadFactor, alpha, beta)

	// group parameters
	groupSize := 16
	demand := []int{2}

	// partial group parameters
	var fractionClaimed float64 = 0.5

	// constraint parameters
	// constraint at rack level
	level := 1
	affinity := util.Spread
	isHard := false
	lc1 := placement.NewLevelConstraint("lc-1", level, util.Affinity(affinity), isHard)

	//constraint at server level
	level = 0
	affinity = util.Pack
	isHard = false
	lc0 := placement.NewLevelConstraint("lc-0", level, util.Affinity(affinity), isHard)

	// misc parameters
	lineLength := 64
	isPrint := false

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

	// place some random allocation on the servers
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Place some load: ")
	fmt.Printf("(loadFactor=%f; alpha=%f; beta=%f; \n", loadFactor, alpha, beta)
	fmt.Printf("avg=%f; cov=%f)", avg, cov)
	fmt.Println()
	fmt.Println(strings.Repeat("=", lineLength))

	allocateLoad(pes, capacity, loadFactor, cov, avg, alpha, beta, false)

	// place random weights
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())
	demos.PlaceWeights(pes)
	for i := 0; i < len(pes); i++ {
		fmt.Println(pes[i])
	}
	fmt.Println()

	pTree.PercolateResources()
	fmt.Print(pTree)

	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Servers:")
	fmt.Println(strings.Repeat("=", lineLength))
	for i := 0; i < len(pes); i++ {
		fmt.Println(pes[i])
	}
	fmt.Println()

	// create level constraint
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Create level constraint:")
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println(lc1)
	fmt.Println(lc0)
	fmt.Println()

	defaultPolicy := true
	ByWeightPolicy := true
	ByWeightProductPolicy := true
	ByFitWeightProductPolicy := true
	ByMinWeightedAvailabilityPolicy := true

	if defaultPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Default Policy (no weights)")
		fmt.Println("Create placement group:")
		groupDemand, _ := util.NewAllocationCopy(demand)
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)
		if isPrint {
			fmt.Println(pg)
			fmt.Println()
		}

		// place group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Place group result: (logical tree)")
			fmt.Println(strings.Repeat("=", lineLength))
		}
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroup(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// claim partial group placement
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Placement group after partial claim:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		if isPrint {
			fmt.Print(pg)
		}

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers after partial claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		for i := 0; i < len(pes); i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// alter system allocation state
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("System after altering allocation state:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		loadFactor = 0.8
		alpha = 0
		beta = 0
		avg = demos.ComputeAverageLoad(loadFactor, alpha, beta)
		cov = 0.5

		allocateLoad(pes, capacity, loadFactor, cov, avg, alpha, beta, true)
		pTree.PercolateResources()
		if isPrint {
			for i := 0; i < len(pes); i++ {
				fmt.Println(pes[i])
			}
			fmt.Println()
		}

		// place partial group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Partial place group result:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		_, err = p.PlacePartialGroup(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}
	}
	if ByWeightPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByWeight Policy")
		fmt.Println("Create placement group:")
		groupDemand, _ := util.NewAllocationCopy(demand)
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)
		if isPrint {
			fmt.Println(pg)
			fmt.Println()
		}

		// place group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Place group result: (logical tree)")
			fmt.Println(strings.Repeat("=", lineLength))
		}
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroupByWeight(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// claim partial group placement
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Placement group after partial claim:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		if isPrint {
			fmt.Print(pg)
		}

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers after partial claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		for i := 0; i < len(pes); i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// alter system allocation state
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("System after altering allocation state:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		loadFactor = 0.8
		alpha = 0
		beta = 0
		avg = demos.ComputeAverageLoad(loadFactor, alpha, beta)
		cov = 0.5

		allocateLoad(pes, capacity, loadFactor, cov, avg, alpha, beta, true)
		pTree.PercolateResources()
		if isPrint {
			for i := 0; i < len(pes); i++ {
				fmt.Println(pes[i])
			}
			fmt.Println()
		}

		// place partial group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Partial place group result:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		_, err = p.PlacePartialGroupByWeight(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}
	}
	if ByWeightProductPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByWeightProduct Policy")
		fmt.Println("Create placement group:")
		groupDemand, _ := util.NewAllocationCopy(demand)
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)
		if isPrint {
			fmt.Println(pg)
			fmt.Println()
		}

		// place group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Place group result: (logical tree)")
			fmt.Println(strings.Repeat("=", lineLength))
		}
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroupByWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// claim partial group placement
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Placement group after partial claim:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		if isPrint {
			fmt.Print(pg)
		}

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers after partial claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		for i := 0; i < len(pes); i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// alter system allocation state
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("System after altering allocation state:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		loadFactor = 0.8
		alpha = 0
		beta = 0
		avg = demos.ComputeAverageLoad(loadFactor, alpha, beta)
		cov = 0.5

		allocateLoad(pes, capacity, loadFactor, cov, avg, alpha, beta, true)
		pTree.PercolateResources()
		if isPrint {
			for i := 0; i < len(pes); i++ {
				fmt.Println(pes[i])
			}
			fmt.Println()
		}

		// place partial group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Partial place group result:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		_, err = p.PlacePartialGroupByWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}
	}
	if ByFitWeightProductPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByFitWeightProduct Policy")
		fmt.Println("Create placement group:")
		groupDemand, _ := util.NewAllocationCopy(demand)
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)
		if isPrint {
			fmt.Println(pg)
			fmt.Println()
		}

		// place group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Place group result: (logical tree)")
			fmt.Println(strings.Repeat("=", lineLength))
		}
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroupByFitWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// claim partial group placement
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Placement group after partial claim:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		if isPrint {
			fmt.Print(pg)
		}

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers after partial claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		for i := 0; i < len(pes); i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// alter system allocation state
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("System after altering allocation state:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		loadFactor = 0.8
		alpha = 0
		beta = 0
		avg = demos.ComputeAverageLoad(loadFactor, alpha, beta)
		cov = 0.5

		allocateLoad(pes, capacity, loadFactor, cov, avg, alpha, beta, true)
		pTree.PercolateResources()
		if isPrint {
			for i := 0; i < len(pes); i++ {
				fmt.Println(pes[i])
			}
			fmt.Println()
		}

		// place partial group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Partial place group result:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		_, err = p.PlacePartialGroupByFitWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}
	}
	if ByMinWeightedAvailabilityPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByMinWeightedAvailability Policy")
		fmt.Println("Create placement group:")
		groupDemand, _ := util.NewAllocationCopy(demand)
		pg := placement.NewPGroup("pg0", groupSize, groupDemand)
		pg.AddLevelConstraint(lc1)
		pg.AddLevelConstraint(lc0)
		if isPrint {
			fmt.Println(pg)
			fmt.Println()
		}

		// place group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Place group result: (logical tree)")
			fmt.Println(strings.Repeat("=", lineLength))
		}
		p := placement.NewPlacer(pTree)
		_, err := p.PlaceGroupByMinWeightedAvailability(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}

		// claim partial group placement
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Placement group after partial claim:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		if isPrint {
			fmt.Print(pg)
		}

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers after partial claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		for i := 0; i < len(pes); i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		// alter system allocation state
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("System after altering allocation state:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		loadFactor = 0.8
		alpha = 0
		beta = 0
		avg = demos.ComputeAverageLoad(loadFactor, alpha, beta)
		cov = 0.5

		allocateLoad(pes, capacity, loadFactor, cov, avg, alpha, beta, true)
		pTree.PercolateResources()
		if isPrint {
			for i := 0; i < len(pes); i++ {
				fmt.Println(pes[i])
			}
			fmt.Println()
		}

		// place partial group
		if isPrint {
			fmt.Println(strings.Repeat("=", lineLength))
			fmt.Println("Partial place group result:")
			fmt.Println(strings.Repeat("=", lineLength))
		}

		_, err = p.PlacePartialGroupByMinWeightedAvailability(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isPrint {
			fmt.Print(pg)
		}
	}
}

// allocateLoad : allocate some random load on the servers
func allocateLoad(servers []*system.PE, capacity []int, loadFactor, cov, avg, alpha, beta float64, noChangeToClaimed bool) {
	numServers := len(servers)
	numResources := len(capacity)
	for i := 0; i < numServers; i++ {
		if noChangeToClaimed && len(servers[i].GetHostedIDs()) > 0 {
			continue
		}
		alloc := servers[i].GetAllocated().GetValue()

		var y float64
		x := rand.Float64()
		if x < alpha {
			y = 1
		} else if x < alpha+beta {
			y = 0
		} else {
			y = avg * (rand.NormFloat64()*cov + 1)
		}

		for k := 0; k < numResources; k++ {
			z := int(math.Round(y * float64(capacity[k])))
			z = util.Max(util.Min(z, capacity[k]), 0)
			alloc[k] = z
		}
	}
}
