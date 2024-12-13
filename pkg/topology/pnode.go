package topology

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/ibm/chic-sched/pkg/util"
)

// PNode : a node in a physical tree topology
type PNode struct {
	// extends Node
	Node
	// the level of this node in the tree (leaves are at level 0)
	level int
	// resource capacity
	capacity *util.Allocation
	// resource allocated
	allocated *util.Allocation

	// number of group instances that can fit
	numFit int
	// number of group instances that are claimed
	numClaimed int

	// weight for guiding placement decision
	weight int
	// Track totalWeight when multiple weights added
	totalWeight int
	// Counter for how many weights have been added
	weightCount int
}

// NewPNode : create a new physical node with zero capacity and allocated resources
//   - returns nil if bad parameters
func NewPNode(node *Node, level int, numResources int) *PNode {
	if node == nil || level < 0 || numResources < 0 {
		return nil
	}
	capacity, _ := util.NewAllocation(numResources)
	allocated, _ := util.NewAllocation(numResources)
	return &PNode{
		Node:       *node,
		level:      level,
		capacity:   capacity,
		allocated:  allocated,
		numFit:     0,
		numClaimed: 0,
		weight:     util.DefaultWeight,
	}
}

// GetWeight : get the weight of the node
func (pNode *PNode) GetWeight() int {
	return pNode.weight
}

// SetWeight : set the weight of the node
func (pNode *PNode) SetWeight(weight int) {
	pNode.weight = weight
}

// AddWeight adds weight and calculates the average of the node
func (pNode *PNode) AddWeight(weight int) {
	// Update the total weight and the count of weights added
	pNode.totalWeight += weight
	pNode.weightCount++

	// Calculate the average weight and round up to the next highest integer
	pNode.weight = int(math.Ceil(float64(pNode.totalWeight) / float64(pNode.weightCount)))
}

// GetLevel : get the level of the node
func (pNode *PNode) GetLevel() int {
	return pNode.level
}

// SetLevel : set the level of the node
func (pNode *PNode) SetLevel(level int) {
	pNode.level = level
}

// SetLevelSubtree : set the level of all nodes in subtree rooted at this node
func (pNode *PNode) SetLevelSubtree(level int) {
	pNode.SetLevel(level)
	for _, node := range pNode.children {
		pChild := (*PNode)(unsafe.Pointer(node))
		pChild.SetLevelSubtree(level - 1)
	}
}

// GetCapacity : get resource capacity
func (pNode *PNode) GetCapacity() *util.Allocation {
	return pNode.capacity
}

// GetAllocated : get resource allocated
func (pNode *PNode) GetAllocated() *util.Allocation {
	return pNode.allocated
}

// GetAvailable : get resource available
func (pNode *PNode) GetAvailable() *util.Allocation {
	available := pNode.capacity.Clone()
	available.Subtract(pNode.allocated)
	return available
}

// GetNumFit : get number can fit
func (pNode *PNode) GetNumFit() int {
	return pNode.numFit
}

// SetNumFit : set number can fit
func (pNode *PNode) SetNumFit(numFit int) {
	pNode.numFit = numFit
}

// IncNumFit : increment number can fit
func (pNode *PNode) IncNumFit(deltaNumFit int) {
	pNode.numFit += deltaNumFit
}

// ResetNumFit : set number can fit to zero in subtree
func (pNode *PNode) ResetNumFit() {
	pNode.numFit = 0
	for _, node := range pNode.children {
		pChild := (*PNode)(unsafe.Pointer(node))
		pChild.ResetNumFit()
	}
}

// GetNumClaimed : get the number claimed
func (pNode *PNode) GetNumClaimed() int {
	return pNode.numClaimed
}

// SetNumClaimed : set the number claimed
func (pNode *PNode) SetNumClaimed(numClaimed int) {
	pNode.numClaimed = numClaimed
}

// ResetNumClaimed : set number claimed to zero in subtree
func (pNode *PNode) ResetNumClaimed() {
	pNode.numClaimed = 0
	for _, node := range pNode.children {
		pChild := (*PNode)(unsafe.Pointer(node))
		pChild.ResetNumClaimed()
	}
}

// ResetResources : reset resource capacity and allocated fields in subtree
func (pNode *PNode) ResetResources() {
	pNode.capacity.SetZero()
	pNode.allocated.SetZero()
	for _, node := range pNode.children {
		pChild := (*PNode)(unsafe.Pointer(node))
		pChild.ResetResources()
	}
}

// GetNumResources : get number of resources
func (pNode *PNode) GetNumResources() int {
	if cap := pNode.capacity; cap != nil {
		return cap.GetSize()
	}
	return 0
}

// Compare : comparator function between this and other node,
// based on number of demands to fit on available resources on the node
//   - two options: increasing or decreasing
//   - return {-1, 0, or +1} if this compared to other is {before, same, or after}
func (pNode *PNode) Compare(oNode *PNode, isIncreasing bool) int {
	numFitThis := pNode.GetNumFit()
	numFitOther := oNode.GetNumFit()

	if numFitThis == numFitOther {
		return 0
	}
	below := numFitThis < numFitOther
	return util.BoolValue(util.Xor(below, isIncreasing))
}

// CompareClaimed : same as Compare() except order on number claimed as primary
func (pNode *PNode) CompareClaimed(oNode *PNode, isIncreasing bool) int {
	numClaimedThis := pNode.GetNumClaimed()
	numClaimedOther := oNode.GetNumClaimed()

	if numClaimedThis == numClaimedOther {
		return pNode.Compare(oNode, isIncreasing)
	}
	below := numClaimedThis < numClaimedOther
	return util.BoolValue(below)
}

// CompareByWeight : Compares nodes by weight
func (pNode *PNode) CompareByWeight(oNode *PNode, isIncreasing bool) int {
	if pNode.weight == oNode.weight {
		return 0
	}
	below := pNode.weight < oNode.weight
	return util.BoolValue(util.Xor(below, isIncreasing))
}

// CompareFitThenWeights : similar to Compare() but consider CompareByWeight() if equal
func (pNode *PNode) CompareFitThenWeights(oNode *PNode, isIncreasing bool) int {
	numFitThis := pNode.GetNumFit()
	numFitOther := oNode.GetNumFit()

	if numFitThis == numFitOther {
		return pNode.CompareByWeight(oNode, isIncreasing)
	}
	below := numFitThis < numFitOther
	return util.BoolValue(util.Xor(below, isIncreasing))
}

// CompareClaimedThenWeights : similar to CompareClaimed() but consider CompareByWeight() if equal
func (pNode *PNode) CompareClaimedThenWeights(oNode *PNode, isIncreasing bool) int {
	numClaimedThis := pNode.GetNumClaimed()
	numClaimedOther := oNode.GetNumClaimed()

	if numClaimedThis == numClaimedOther {
		return pNode.CompareByWeight(oNode, isIncreasing)
	}
	below := numClaimedThis < numClaimedOther
	return util.BoolValue(below)
}

// CompareByClaimedWeightProduct : Compares by the product of weight and claimed resources.
func (pNode *PNode) CompareByClaimedWeightProduct(oNode *PNode, isIncreasing bool) int {
	numClaimedThis := pNode.weight * pNode.GetNumClaimed()
	numClaimedOther := oNode.weight * oNode.GetNumClaimed()

	if numClaimedThis == numClaimedOther {
		return pNode.CompareByWeight(oNode, isIncreasing)
	}
	below := numClaimedThis < numClaimedOther
	return util.BoolValue(below)
}

// CompareByFitWeightProduct : Compares by the product of weight and number of demands that fit.
func (pNode *PNode) CompareByFitWeightProduct(oNode *PNode, isIncreasing bool) int {
	numFitThis := pNode.weight * pNode.GetNumFit()
	numFitOther := oNode.weight * oNode.GetNumFit()

	if numFitThis == numFitOther {
		return pNode.Compare(oNode, isIncreasing)
	}
	below := numFitThis < numFitOther
	return util.BoolValue(util.Xor(below, isIncreasing))
}

// CompareByMinWeightedAvailability : Compares nodes by the minimum weighted available resource.
func (pNode *PNode) CompareByMinWeightedAvailability(oNode *PNode, isIncreasing bool) int {
	availableThis := pNode.weight * pNode.GetAvailable().Minimum()
	availableOther := oNode.weight * oNode.GetAvailable().Minimum()

	if availableThis == availableOther {
		return pNode.CompareByWeight(oNode, isIncreasing)
	}
	below := availableThis < availableOther
	return util.BoolValue(util.Xor(below, isIncreasing))
}

// String : a print out of the physical node
func (pNode *PNode) String() string {
	return fmt.Sprintf("pNode: ID=%s; weight=%v; level=%d; cap=%v; alloc=%v; numClaimed=%d",
		pNode.GetID(), pNode.weight, pNode.level, pNode.capacity, pNode.allocated, pNode.numClaimed)
}
