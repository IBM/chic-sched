package placement

import (
	"fmt"
	"sort"
	"unsafe"

	"github.com/ibm/chic-sched/pkg/topology"
	"github.com/ibm/chic-sched/pkg/util"
)

// Placer : placer of a placement group
type Placer struct {
	// physical tree
	pTree *topology.PTree

	// placement group
	pg *PGroup
	// keep track of number remaining to place
	numRemaining int
	// keep track of number of claimed remaining
	numClaimedRemaining int
}

// NewPlacer : create a new placer
//   - returns nil if bad parameters
func NewPlacer(pTree *topology.PTree) *Placer {
	if pTree == nil {
		return nil
	}
	return &Placer{
		pTree:               pTree,
		pg:                  nil,
		numRemaining:        0,
		numClaimedRemaining: 0,
	}
}

// PlaceInit : initialize group placement
func (p *Placer) PlaceInit(pg *PGroup) (*topology.PNode, error) {
	if pg == nil {
		return nil, fmt.Errorf("PGroup is nil")
	}
	pRoot := (*topology.PNode)(unsafe.Pointer(p.pTree.GetRoot()))
	if pRoot == nil {
		return nil, fmt.Errorf("pRoot is nil, empty pTree")
	}
	p.pg = pg
	p.numRemaining = pg.GetSize()
	if p.numRemaining == 0 {
		return pRoot, fmt.Errorf("empty group")
	}
	demand := p.pg.GetDemand()
	if demand == nil {
		return pRoot, fmt.Errorf("group demand is nil")
	}
	// calculate number of members that can fit on all nodes
	// of the physical tree
	p.pTree.PercolateNumFit(demand)
	return pRoot, nil
}

// PlaceCleanup : cleanup after group placement
func (p *Placer) PlaceCleanup() {
	if p.pTree != nil {
		p.pTree.ResetNumFit()
	}
}

// PlaceGroup : place a group
func (p *Placer) PlaceGroup(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	lRoot := p.placeAtNode(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	pg.SetLTree(lTree)
	return lTree, nil
}

// PlaceGroupByWeight : place a group with Weights Compare
func (p *Placer) PlaceGroupByWeight(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	lRoot := p.placeAtNodeByWeight(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	pg.SetLTree(lTree)
	return lTree, nil
}

// PlaceGroupByWeightProduct : place a group ByWeightProduct Compare
func (p *Placer) PlaceGroupByWeightProduct(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	lRoot := p.placeAtNodeByWeightProduct(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	pg.SetLTree(lTree)
	return lTree, nil
}

// PlaceGroupByFitWeightProduct : place a group with ByFitWeightProduct Compare
func (p *Placer) PlaceGroupByFitWeightProduct(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	lRoot := p.placeAtNodeByFitWeightProduct(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	pg.SetLTree(lTree)
	return lTree, nil
}

// PlaceGroupByMinWeightedAvailability : place a group with MinWeightedAvailability Compare
func (p *Placer) PlaceGroupByMinWeightedAvailability(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	lRoot := p.placeAtNodeByMinWeightedAvailability(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	pg.SetLTree(lTree)
	return lTree, nil
}

// placeAtNode : recursive function to place subgroup on a subtree rooted at a given pNode
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
//   - numPartitionsPlaced: number of partitions already placed (if applicable)
func (p *Placer) placeAtNode(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)
	// calculate range of number to place on node given constraint
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), numToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	if numDesired == 0 {
		return lNode
	}

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		p.numRemaining -= numPlaced
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && numDesired >= numPartitions*minRange {
			p.sortNodes(children, false)
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				node := p.placeAtNode(pChild, numNodes, numDesired, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					numPartitionsUsed++
				}
			}
		}
	}
	if numPlaced == 0 || sr.NumberInRange(numPlaced) {
		lNode.SetCount(numPlaced)
	} else {
		// placement failed size range
		fmt.Printf("==> Failed placement at node %s: canPlace=%d; sizeRange=%v \n", pNode.GetID(), numPlaced, sr)
		lNode.RemoveChildren()
		p.numRemaining += numPlaced
		lNode.SetCount(0)
	}
	return lNode
}

// placeAtNodeByWeight : recursive function to place subgroup on a subtree rooted at a given pNode with weight consideration
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
//   - numPartitionsPlaced: number of partitions already placed (if applicable)
func (p *Placer) placeAtNodeByWeight(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)
	// calculate range of number to place on node given constraint
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), numToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	if numDesired == 0 {
		return lNode
	}

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		p.numRemaining -= numPlaced
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && numDesired >= numPartitions*minRange {
			p.sortNodesByWeight(children, false)
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				node := p.placeAtNodeByWeight(pChild, numNodes, numDesired, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					numPartitionsUsed++
				}
			}
		}
	}
	if numPlaced == 0 || sr.NumberInRange(numPlaced) {
		lNode.SetCount(numPlaced)
	} else {
		// placement failed size range
		fmt.Printf("==> Failed placement at node %s: canPlace=%d; sizeRange=%v \n", pNode.GetID(), numPlaced, sr)
		lNode.RemoveChildren()
		p.numRemaining += numPlaced
		lNode.SetCount(0)
	}
	return lNode
}

// placeAtNodeByWeightProduct : recursive function to place subgroup on a subtree rooted at a given pNode with weight consideration
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
//   - numPartitionsPlaced: number of partitions already placed (if applicable)
func (p *Placer) placeAtNodeByWeightProduct(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)
	// calculate range of number to place on node given constraint
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), numToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	if numDesired == 0 {
		return lNode
	}

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		p.numRemaining -= numPlaced
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && numDesired >= numPartitions*minRange {
			p.sortNodesByWeightProduct(children, false)
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				node := p.placeAtNodeByWeightProduct(pChild, numNodes, numDesired, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					numPartitionsUsed++
				}
			}
		}
	}
	if numPlaced == 0 || sr.NumberInRange(numPlaced) {
		lNode.SetCount(numPlaced)
	} else {
		// placement failed size range
		fmt.Printf("==> Failed placement at node %s: canPlace=%d; sizeRange=%v \n", pNode.GetID(), numPlaced, sr)
		lNode.RemoveChildren()
		p.numRemaining += numPlaced
		lNode.SetCount(0)
	}
	return lNode
}

// placeAtNodeByFitWeightProduct : recursive function to place subgroup on a subtree rooted at a given pNode with weight consideration
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
//   - numPartitionsPlaced: number of partitions already placed (if applicable)
func (p *Placer) placeAtNodeByFitWeightProduct(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)
	// calculate range of number to place on node given constraint
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), numToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	if numDesired == 0 {
		return lNode
	}

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		p.numRemaining -= numPlaced
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && numDesired >= numPartitions*minRange {
			p.sortNodesByFitWeightProduct(children, false)
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				node := p.placeAtNodeByFitWeightProduct(pChild, numNodes, numDesired, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					numPartitionsUsed++
				}
			}
		}
	}
	if numPlaced == 0 || sr.NumberInRange(numPlaced) {
		lNode.SetCount(numPlaced)
	} else {
		// placement failed size range
		fmt.Printf("==> Failed placement at node %s: canPlace=%d; sizeRange=%v \n", pNode.GetID(), numPlaced, sr)
		lNode.RemoveChildren()
		p.numRemaining += numPlaced
		lNode.SetCount(0)
	}
	return lNode
}

// placeAtNodeByMinWeightedAvailability : recursive function to place subgroup on a subtree rooted at a given pNode with weight consideration
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
//   - numPartitionsPlaced: number of partitions already placed (if applicable)
func (p *Placer) placeAtNodeByMinWeightedAvailability(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)
	// calculate range of number to place on node given constraint
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), numToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	if numDesired == 0 {
		return lNode
	}

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		p.numRemaining -= numPlaced
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && numDesired >= numPartitions*minRange {
			p.sortNodesByMinWeightedAvailability(children, false)
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				node := p.placeAtNodeByMinWeightedAvailability(pChild, numNodes, numDesired, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					numPartitionsUsed++
				}
			}
		}
	}
	if numPlaced == 0 || sr.NumberInRange(numPlaced) {
		lNode.SetCount(numPlaced)
	} else {
		// placement failed size range
		fmt.Printf("==> Failed placement at node %s: canPlace=%d; sizeRange=%v \n", pNode.GetID(), numPlaced, sr)
		lNode.RemoveChildren()
		p.numRemaining += numPlaced
		lNode.SetCount(0)
	}
	return lNode
}

// PlacePartialGroup : place a group with some members already placed (claimed resources)
func (p *Placer) PlacePartialGroup(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	// lTree of partial placement
	partialLTree := pg.GetLTree()
	if partialLTree == nil {
		return nil, fmt.Errorf("partial placement LTree is nil")
	}
	// account for members with claimed resources
	p.pTree.ResetNumClaimed()
	partialLTree.SetPhysicalClaimed()
	p.numRemaining -= pRoot.GetNumClaimed()
	if p.numRemaining == 0 {
		// nothing to do
		return partialLTree, nil
	}
	if p.numRemaining < 0 {
		return nil, fmt.Errorf("number claimed larger than group size")
	}
	p.pTree.MergeClaimedToFit()

	// place recursively
	p.numClaimedRemaining = pRoot.GetNumClaimed()
	lRoot := p.placePartialGroupAtNode(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	if p.numClaimedRemaining > 0 {
		fmt.Println("Not all claimed instances were re-placed.")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	lTree.PercolateClaimed()
	pg.SetLTree(lTree)
	return lTree, nil
}

// PlacePartialGroupByWeight : place a group with some members already placed (claimed resources) with weight consideration
func (p *Placer) PlacePartialGroupByWeight(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	// lTree of partial placement
	partialLTree := pg.GetLTree()
	if partialLTree == nil {
		return nil, fmt.Errorf("partial placement LTree is nil")
	}
	// account for members with claimed resources
	p.pTree.ResetNumClaimed()
	partialLTree.SetPhysicalClaimed()
	p.numRemaining -= pRoot.GetNumClaimed()
	if p.numRemaining == 0 {
		// nothing to do
		return partialLTree, nil
	}
	if p.numRemaining < 0 {
		return nil, fmt.Errorf("number claimed larger than group size")
	}
	p.pTree.MergeClaimedToFit()

	// place recursively
	p.numClaimedRemaining = pRoot.GetNumClaimed()
	lRoot := p.placePartialGroupAtNodeByWeight(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	if p.numClaimedRemaining > 0 {
		fmt.Println("Not all claimed instances were re-placed.")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	lTree.PercolateClaimed()
	pg.SetLTree(lTree)
	return lTree, nil
}

// PlacePartialGroupByWeightProduct : place a group with some members already placed (claimed resources) with weights in calculation
func (p *Placer) PlacePartialGroupByWeightProduct(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	// lTree of partial placement
	partialLTree := pg.GetLTree()
	if partialLTree == nil {
		return nil, fmt.Errorf("partial placement LTree is nil")
	}
	// account for members with claimed resources
	p.pTree.ResetNumClaimed()
	partialLTree.SetPhysicalClaimed()
	p.numRemaining -= pRoot.GetNumClaimed()
	if p.numRemaining == 0 {
		// nothing to do
		return partialLTree, nil
	}
	if p.numRemaining < 0 {
		return nil, fmt.Errorf("number claimed larger than group size")
	}
	p.pTree.MergeClaimedToFit()

	// place recursively
	p.numClaimedRemaining = pRoot.GetNumClaimed()
	lRoot := p.placePartialGroupAtNodeByWeightProduct(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	if p.numClaimedRemaining > 0 {
		fmt.Println("Not all claimed instances were re-placed.")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	lTree.PercolateClaimed()
	pg.SetLTree(lTree)
	return lTree, nil
}

// PlacePartialGroupByFitWeightProduct : place a group with some members already placed (claimed resources) with weights in calculation
func (p *Placer) PlacePartialGroupByFitWeightProduct(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	// lTree of partial placement
	partialLTree := pg.GetLTree()
	if partialLTree == nil {
		return nil, fmt.Errorf("partial placement LTree is nil")
	}
	// account for members with claimed resources
	p.pTree.ResetNumClaimed()
	partialLTree.SetPhysicalClaimed()
	p.numRemaining -= pRoot.GetNumClaimed()
	if p.numRemaining == 0 {
		// nothing to do
		return partialLTree, nil
	}
	if p.numRemaining < 0 {
		return nil, fmt.Errorf("number claimed larger than group size")
	}
	p.pTree.MergeClaimedToFit()

	// place recursively
	p.numClaimedRemaining = pRoot.GetNumClaimed()
	lRoot := p.placePartialGroupAtNodeByFitWeightProduct(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	if p.numClaimedRemaining > 0 {
		fmt.Println("Not all claimed instances were re-placed.")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	lTree.PercolateClaimed()
	pg.SetLTree(lTree)
	return lTree, nil
}

// PlacePartialGroupByMinWeightedAvailability : place a group with some members already placed (claimed resources) with weights in calculation
func (p *Placer) PlacePartialGroupByMinWeightedAvailability(pg *PGroup) (*topology.LTree, error) {
	defer p.PlaceCleanup()
	pRoot, err := p.PlaceInit(pg)
	if err != nil {
		return nil, err
	}
	// lTree of partial placement
	partialLTree := pg.GetLTree()
	if partialLTree == nil {
		return nil, fmt.Errorf("partial placement LTree is nil")
	}
	// account for members with claimed resources
	p.pTree.ResetNumClaimed()
	partialLTree.SetPhysicalClaimed()
	p.numRemaining -= pRoot.GetNumClaimed()
	if p.numRemaining == 0 {
		// nothing to do
		return partialLTree, nil
	}
	if p.numRemaining < 0 {
		return nil, fmt.Errorf("number claimed larger than group size")
	}
	p.pTree.MergeClaimedToFit()

	// place recursively
	p.numClaimedRemaining = pRoot.GetNumClaimed()
	lRoot := p.placePartialGroupAtNodeByMinWeightedAvailability(pRoot, 1, p.numRemaining, 0)
	if lRoot == nil {
		return nil, fmt.Errorf("lRoot is nil, failed placement")
	}
	if p.numClaimedRemaining > 0 {
		fmt.Println("Not all claimed instances were re-placed.")
	}
	tree := topology.NewTree(&lRoot.Node)
	lTree := topology.NewLTree(tree)
	lTree.PercolateClaimed()
	pg.SetLTree(lTree)
	return lTree, nil
}

// placePartialGroupAtNode : recursive function to place subgroup of a partial group
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
func (p *Placer) placePartialGroupAtNode(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)

	// calculate range of number to place on node given constraint
	totalNumToPlace := numToPlace + pNode.GetNumClaimed()
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), totalNumToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	numDesired = util.Max(numDesired, pNode.GetNumClaimed())

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		numClaimedAndPlaced := util.Min(numPlaced, pNode.GetNumClaimed())
		p.numRemaining -= (numPlaced - numClaimedAndPlaced)
		p.numClaimedRemaining -= numClaimedAndPlaced
		lNode.SetClaimed(numClaimedAndPlaced)
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && totalNumToPlace >= numPartitions*minRange {
			p.sortNodes(children, true)
			claimedRemaining := pNode.GetNumClaimed()
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				numUnclaimedToPlace := util.Max(numDesired-claimedRemaining, 0)
				node := p.placePartialGroupAtNode(pChild, numNodes, numUnclaimedToPlace, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					claimedRemaining -= util.Min(node.GetCount(), pChild.GetNumClaimed())
					numPartitionsUsed++
				}
			}
		}
	}
	lNode.SetCount(numPlaced)
	return lNode
}

// placePartialGroupAtNodeByWeight : recursive function to place subgroup of a partial group
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
func (p *Placer) placePartialGroupAtNodeByWeight(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)

	// calculate range of number to place on node given constraint
	totalNumToPlace := numToPlace + pNode.GetNumClaimed()
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), totalNumToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	numDesired = util.Max(numDesired, pNode.GetNumClaimed())

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		numClaimedAndPlaced := util.Min(numPlaced, pNode.GetNumClaimed())
		p.numRemaining -= (numPlaced - numClaimedAndPlaced)
		p.numClaimedRemaining -= numClaimedAndPlaced
		lNode.SetClaimed(numClaimedAndPlaced)
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && totalNumToPlace >= numPartitions*minRange {
			p.sortNodesByWeight(children, true)
			claimedRemaining := pNode.GetNumClaimed()
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				numUnclaimedToPlace := util.Max(numDesired-claimedRemaining, 0)
				node := p.placePartialGroupAtNodeByWeight(pChild, numNodes, numUnclaimedToPlace, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					claimedRemaining -= util.Min(node.GetCount(), pChild.GetNumClaimed())
					numPartitionsUsed++
				}
			}
		}
	}
	lNode.SetCount(numPlaced)
	return lNode
}

// placePartialGroupAtNodeByWeightProduct : recursive function to place subgroup of a partial group
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
func (p *Placer) placePartialGroupAtNodeByWeightProduct(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)

	// calculate range of number to place on node given constraint
	totalNumToPlace := numToPlace + pNode.GetNumClaimed()
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), totalNumToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	numDesired = util.Max(numDesired, pNode.GetNumClaimed())

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		numClaimedAndPlaced := util.Min(numPlaced, pNode.GetNumClaimed())
		p.numRemaining -= (numPlaced - numClaimedAndPlaced)
		p.numClaimedRemaining -= numClaimedAndPlaced
		lNode.SetClaimed(numClaimedAndPlaced)
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && totalNumToPlace >= numPartitions*minRange {
			p.sortNodesByWeightProduct(children, true)
			claimedRemaining := pNode.GetNumClaimed()
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				numUnclaimedToPlace := util.Max(numDesired-claimedRemaining, 0)
				node := p.placePartialGroupAtNodeByWeightProduct(pChild, numNodes, numUnclaimedToPlace, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					claimedRemaining -= util.Min(node.GetCount(), pChild.GetNumClaimed())
					numPartitionsUsed++
				}
			}
		}
	}
	lNode.SetCount(numPlaced)
	return lNode
}

// placePartialGroupAtNodeByFitWeightProduct : recursive function to place subgroup of a partial group
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
func (p *Placer) placePartialGroupAtNodeByFitWeightProduct(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)

	// calculate range of number to place on node given constraint
	totalNumToPlace := numToPlace + pNode.GetNumClaimed()
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), totalNumToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	numDesired = util.Max(numDesired, pNode.GetNumClaimed())

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		numClaimedAndPlaced := util.Min(numPlaced, pNode.GetNumClaimed())
		p.numRemaining -= (numPlaced - numClaimedAndPlaced)
		p.numClaimedRemaining -= numClaimedAndPlaced
		lNode.SetClaimed(numClaimedAndPlaced)
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && totalNumToPlace >= numPartitions*minRange {
			p.sortNodesByFitWeightProduct(children, true)
			claimedRemaining := pNode.GetNumClaimed()
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				numUnclaimedToPlace := util.Max(numDesired-claimedRemaining, 0)
				node := p.placePartialGroupAtNodeByFitWeightProduct(pChild, numNodes, numUnclaimedToPlace, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					claimedRemaining -= util.Min(node.GetCount(), pChild.GetNumClaimed())
					numPartitionsUsed++
				}
			}
		}
	}
	lNode.SetCount(numPlaced)
	return lNode
}

// placePartialGroupAtNodeByMinWeightedAvailability : recursive function to place subgroup of a partial group
//   - numNodes: number of sibling pNodes targeted for placement (including this pNode)
//   - numToPlace: number of members available for placement at this node
func (p *Placer) placePartialGroupAtNodeByMinWeightedAvailability(pNode *topology.PNode, numNodes int, numToPlace int,
	numPartitionsPlaced int) *topology.LNode {

	// create lNode corresponding to the pNode
	lNode := topology.NewLNode(pNode, 0)

	// calculate range of number to place on node given constraint
	totalNumToPlace := numToPlace + pNode.GetNumClaimed()
	sr := CreateSizeRange(p.pg, pNode.GetLevel(), totalNumToPlace, numNodes, numPartitionsPlaced)
	if sr == nil {
		return lNode
	}
	// select number in range based on node availability
	numDesired := sr.NumberToPlace(pNode.GetNumFit())
	numDesired = util.Max(numDesired, pNode.GetNumClaimed())

	// visit the subtree rooted at pNode
	numPlaced := 0
	if pNode.GetLevel() == 0 {
		// leaf node, place desired number
		numPlaced = numDesired
		numClaimedAndPlaced := util.Min(numPlaced, pNode.GetNumClaimed())
		p.numRemaining -= (numPlaced - numClaimedAndPlaced)
		p.numClaimedRemaining -= numClaimedAndPlaced
		lNode.SetClaimed(numClaimedAndPlaced)
	} else {
		// process children of pNode
		children := pNode.GetChildren()
		numChildren := len(children)

		// check number of partitions and range
		lc := p.pg.GetLevelConstraint(pNode.GetLevel() - 1)
		numPartitions, _ := lc.GetNumPartitions()
		minRange, _, okRange := lc.GetRange()
		if !okRange {
			minRange = 1
		}

		if numPartitions <= numChildren && totalNumToPlace >= numPartitions*minRange {
			p.sortNodesByMinWeightedAvailability(children, true)
			claimedRemaining := pNode.GetNumClaimed()
			startFrom := 0
			numNodes := numChildren
			if numPartitions > 0 {
				numNodes = numPartitions
				lcp := p.pg.GetLevelConstraint(pNode.GetLevel())
				if lcp.affinity == util.Spread {
					startFrom = numChildren - numNodes
				}
			}
			numPartitionsUsed := 0
			for i := startFrom; i < numChildren; i++ {
				if numDesired <= 0 {
					break
				}
				child := children[i]
				pChild := (*topology.PNode)(unsafe.Pointer(child))
				numUnclaimedToPlace := util.Max(numDesired-claimedRemaining, 0)
				node := p.placePartialGroupAtNodeByMinWeightedAvailability(pChild, numNodes, numUnclaimedToPlace, numPartitionsUsed)
				numNodes--
				if node.GetCount() > 0 {
					lNode.AddChild(&node.Node)
					numPlaced += node.GetCount()
					numDesired -= node.GetCount()
					claimedRemaining -= util.Min(node.GetCount(), pChild.GetNumClaimed())
					numPartitionsUsed++
				}
			}
		}
	}
	lNode.SetCount(numPlaced)
	return lNode
}

// sortNodes : sort a set of nodes based on constraint (assuming numFit already calculated)
func (p *Placer) sortNodes(nodes []*topology.Node, isPartialPlacement bool) {
	sort.Slice(nodes, func(i, j int) bool {
		pNodei := (*topology.PNode)(unsafe.Pointer(nodes[i]))
		pNodej := (*topology.PNode)(unsafe.Pointer(nodes[j]))
		lc := p.pg.GetLevelConstraint(pNodei.GetLevel())
		isIncreasing := lc.Affinity() == util.Spread
		if isPartialPlacement {
			return pNodei.CompareClaimed(pNodej, isIncreasing) < 0
		}
		return pNodei.Compare(pNodej, isIncreasing) < 0
	})
}

// sortNodesWeights : sort a set of nodes based on constraint and weights (assuming numFit already calculated)
func (p *Placer) sortNodesByWeight(nodes []*topology.Node, isPartialPlacement bool) {
	sort.Slice(nodes, func(i, j int) bool {
		pNodei := (*topology.PNode)(unsafe.Pointer(nodes[i]))
		pNodej := (*topology.PNode)(unsafe.Pointer(nodes[j]))
		lc := p.pg.GetLevelConstraint(pNodei.GetLevel())
		isIncreasing := lc.Affinity() == util.Spread
		if isPartialPlacement {
			return pNodei.CompareClaimedThenWeights(pNodej, isIncreasing) < 0
		}
		return pNodei.CompareFitThenWeights(pNodej, isIncreasing) < 0
	})
}

// sortNodesByWeightProduct : sort a set of nodes based on constraint and weight calculation (assuming numFit already calculated)
func (p *Placer) sortNodesByWeightProduct(nodes []*topology.Node, isPartialPlacement bool) {
	sort.Slice(nodes, func(i, j int) bool {
		pNodei := (*topology.PNode)(unsafe.Pointer(nodes[i]))
		pNodej := (*topology.PNode)(unsafe.Pointer(nodes[j]))
		lc := p.pg.GetLevelConstraint(pNodei.GetLevel())
		isIncreasing := lc.Affinity() == util.Spread
		if isPartialPlacement {
			return pNodei.CompareByClaimedWeightProduct(pNodej, isIncreasing) < 0
		}
		return pNodei.CompareFitThenWeights(pNodej, isIncreasing) < 0
	})
}

// sortNodesFitWeights : sort a set of nodes based on constraint and weight calculation (assuming numFit already calculated)
func (p *Placer) sortNodesByFitWeightProduct(nodes []*topology.Node, isPartialPlacement bool) {
	sort.Slice(nodes, func(i, j int) bool {
		pNodei := (*topology.PNode)(unsafe.Pointer(nodes[i]))
		pNodej := (*topology.PNode)(unsafe.Pointer(nodes[j]))
		lc := p.pg.GetLevelConstraint(pNodei.GetLevel())
		isIncreasing := lc.Affinity() == util.Spread
		if isPartialPlacement {
			return pNodei.CompareByFitWeightProduct(pNodej, isIncreasing) < 0
		}
		return pNodei.CompareFitThenWeights(pNodej, isIncreasing) < 0
	})
}

// sortNodesMinWeightedAvailability : sort a set of nodes based on constraint and weight calculation (assuming numFit already calculated)
func (p *Placer) sortNodesByMinWeightedAvailability(nodes []*topology.Node, isPartialPlacement bool) {
	sort.Slice(nodes, func(i, j int) bool {
		pNodei := (*topology.PNode)(unsafe.Pointer(nodes[i]))
		pNodej := (*topology.PNode)(unsafe.Pointer(nodes[j]))
		lc := p.pg.GetLevelConstraint(pNodei.GetLevel())
		isIncreasing := lc.Affinity() == util.Spread
		if isPartialPlacement {
			return pNodei.CompareByMinWeightedAvailability(pNodej, isIncreasing) < 0
		}
		return pNodei.CompareFitThenWeights(pNodej, isIncreasing) < 0
	})
}
