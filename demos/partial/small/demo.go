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
	"github.com/ibm/chic-sched/pkg/util"
)

// Small system demo
//   - create a pTree and PEs given tree degrees at various levels
//   - generate random allocation using normal distribution
//   - generate a group with level constraints
//   - place group, generate lTree
//   - partial claim of group resources
//   - alter system allocation state
//   - place partially placed group given new system state
//   - deallocate resources
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

	// partial group parameters
	var fractionClaimed float64 = 0.5

	// system allocation state change parameters
	var multiplierMean float64 = 0.25
	var multiplierStdev float64 = 2.0

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
	fmt.Print("Place some load and weights: ")
	fmt.Printf("(avg=%f; cov=%f)", loadFactor, cov)
	fmt.Println()
	fmt.Println(strings.Repeat("=", lineLength))

	demos.PlaceBackgroungLoad(pes, loadFactor, 0, 0, cov)

	// place random weights
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())
	demos.PlaceWeights(pes)
	for i := 0; i < numServers; i++ {
		fmt.Println(pes[i])
	}
	fmt.Println()

	pTree.PercolateResources()
	fmt.Print(pTree)

	// create level constraint
	fmt.Println(strings.Repeat("=", lineLength))
	fmt.Println("Create level constraint:")
	fmt.Println(strings.Repeat("=", lineLength))
	lc0 := placement.NewLevelConstraint("lc0", level, util.Affinity(affinity), isHard)
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
		fmt.Println("Default Strategy (no weights)")
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

		// claim partial group placement
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers:")
		fmt.Println(strings.Repeat("=", lineLength))
		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pTree)

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Logical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pg.GetLTree())

		// change system state
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after some resource allocation changes:")
		fmt.Println(strings.Repeat("=", lineLength))

		numResources := len(capacity)
		mean := make([]float64, numResources)
		std := make([]float64, numResources)
		for k := 0; k < numResources; k++ {
			mean[k] = float64(capacity[k]) * loadFactor
			std[k] = cov * mean[k]
		}

		for i := 0; i < numServers; i++ {
			if len(pes[i].GetHostedIDs()) > 0 {
				continue
			}
			alloc := pes[i].GetAllocated().GetValue()
			for k := 0; k < numResources; k++ {
				z := int(math.Round(rand.NormFloat64()*multiplierStdev*std[k] + multiplierMean*mean[k]))
				z = util.Max(util.Min(z, capacity[k]), 0)
				alloc[k] = z
			}
		}
		pTree.PercolateResources()
		fmt.Print(pTree)

		// place partial group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Partial place group result: (logical tree)")
		fmt.Println(strings.Repeat("=", lineLength))
		_, err = p.PlacePartialGroup(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// unplace group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after group de-allocation:")
		fmt.Println(strings.Repeat("=", lineLength))
		pg.UnClaimAll(pTree)
		fmt.Print(pTree)
	}
	if ByWeightPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByWeight Strategy")
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
		_, err := p.PlaceGroupByWeight(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// claim partial group placement
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers:")
		fmt.Println(strings.Repeat("=", lineLength))
		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pTree)

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Logical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pg.GetLTree())

		// change system state
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after some resource allocation changes:")
		fmt.Println(strings.Repeat("=", lineLength))

		numResources := len(capacity)
		mean := make([]float64, numResources)
		std := make([]float64, numResources)
		for k := 0; k < numResources; k++ {
			mean[k] = float64(capacity[k]) * loadFactor
			std[k] = cov * mean[k]
		}

		for i := 0; i < numServers; i++ {
			if len(pes[i].GetHostedIDs()) > 0 {
				continue
			}
			alloc := pes[i].GetAllocated().GetValue()
			for k := 0; k < numResources; k++ {
				z := int(math.Round(rand.NormFloat64()*multiplierStdev*std[k] + multiplierMean*mean[k]))
				z = util.Max(util.Min(z, capacity[k]), 0)
				alloc[k] = z
			}
		}
		pTree.PercolateResources()
		fmt.Print(pTree)

		// place partial group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Partial place group result: (logical tree)")
		fmt.Println(strings.Repeat("=", lineLength))
		_, err = p.PlacePartialGroupByWeight(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// unplace group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after group de-allocation:")
		fmt.Println(strings.Repeat("=", lineLength))
		pg.UnClaimAll(pTree)
		fmt.Print(pTree)
	}
	if ByWeightProductPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByWeightProduct Strategy")
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
		_, err := p.PlaceGroupByWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// claim partial group placement
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers:")
		fmt.Println(strings.Repeat("=", lineLength))
		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pTree)

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Logical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pg.GetLTree())

		// change system state
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after some resource allocation changes:")
		fmt.Println(strings.Repeat("=", lineLength))

		numResources := len(capacity)
		mean := make([]float64, numResources)
		std := make([]float64, numResources)
		for k := 0; k < numResources; k++ {
			mean[k] = float64(capacity[k]) * loadFactor
			std[k] = cov * mean[k]
		}

		for i := 0; i < numServers; i++ {
			if len(pes[i].GetHostedIDs()) > 0 {
				continue
			}
			alloc := pes[i].GetAllocated().GetValue()
			for k := 0; k < numResources; k++ {
				z := int(math.Round(rand.NormFloat64()*multiplierStdev*std[k] + multiplierMean*mean[k]))
				z = util.Max(util.Min(z, capacity[k]), 0)
				alloc[k] = z
			}
		}
		pTree.PercolateResources()
		fmt.Print(pTree)

		// place partial group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Partial place group result: (logical tree)")
		fmt.Println(strings.Repeat("=", lineLength))
		_, err = p.PlacePartialGroupByWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// unplace group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after group de-allocation:")
		fmt.Println(strings.Repeat("=", lineLength))
		pg.UnClaimAll(pTree)
		fmt.Print(pTree)
	}
	if ByFitWeightProductPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByFitWeightProduct Strategy")
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
		_, err := p.PlaceGroupByFitWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// claim partial group placement
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers:")
		fmt.Println(strings.Repeat("=", lineLength))
		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pTree)

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Logical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pg.GetLTree())

		// change system state
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after some resource allocation changes:")
		fmt.Println(strings.Repeat("=", lineLength))

		numResources := len(capacity)
		mean := make([]float64, numResources)
		std := make([]float64, numResources)
		for k := 0; k < numResources; k++ {
			mean[k] = float64(capacity[k]) * loadFactor
			std[k] = cov * mean[k]
		}

		for i := 0; i < numServers; i++ {
			if len(pes[i].GetHostedIDs()) > 0 {
				continue
			}
			alloc := pes[i].GetAllocated().GetValue()
			for k := 0; k < numResources; k++ {
				z := int(math.Round(rand.NormFloat64()*multiplierStdev*std[k] + multiplierMean*mean[k]))
				z = util.Max(util.Min(z, capacity[k]), 0)
				alloc[k] = z
			}
		}
		pTree.PercolateResources()
		fmt.Print(pTree)

		// place partial group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Partial place group result: (logical tree)")
		fmt.Println(strings.Repeat("=", lineLength))
		_, err = p.PlacePartialGroupByFitWeightProduct(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// unplace group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after group de-allocation:")
		fmt.Println(strings.Repeat("=", lineLength))
		pg.UnClaimAll(pTree)
		fmt.Print(pTree)
	}
	if ByMinWeightedAvailabilityPolicy {
		// create placement group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("ByMinWeightedAvailability Strategy")
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
		_, err := p.PlaceGroupByMinWeightedAvailability(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// claim partial group placement
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Allocation on servers:")
		fmt.Println(strings.Repeat("=", lineLength))
		numClaimed := int(math.Ceil(fractionClaimed * float64(groupSize)))
		pg.Claim(numClaimed, pTree)
		for i := 0; i < numServers; i++ {
			fmt.Println(pes[i])
		}
		fmt.Println()

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pTree)

		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Logical tree after partial group claim:")
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Print(pg.GetLTree())

		// change system state
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after some resource allocation changes:")
		fmt.Println(strings.Repeat("=", lineLength))

		numResources := len(capacity)
		mean := make([]float64, numResources)
		std := make([]float64, numResources)
		for k := 0; k < numResources; k++ {
			mean[k] = float64(capacity[k]) * loadFactor
			std[k] = cov * mean[k]
		}

		for i := 0; i < numServers; i++ {
			if len(pes[i].GetHostedIDs()) > 0 {
				continue
			}
			alloc := pes[i].GetAllocated().GetValue()
			for k := 0; k < numResources; k++ {
				z := int(math.Round(rand.NormFloat64()*multiplierStdev*std[k] + multiplierMean*mean[k]))
				z = util.Max(util.Min(z, capacity[k]), 0)
				alloc[k] = z
			}
		}
		pTree.PercolateResources()
		fmt.Print(pTree)

		// place partial group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Partial place group result: (logical tree)")
		fmt.Println(strings.Repeat("=", lineLength))
		_, err = p.PlacePartialGroupByMinWeightedAvailability(pg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(pg)

		// unplace group
		fmt.Println(strings.Repeat("=", lineLength))
		fmt.Println("Physical tree after group de-allocation:")
		fmt.Println(strings.Repeat("=", lineLength))
		pg.UnClaimAll(pTree)
		fmt.Print(pTree)
	}
}
