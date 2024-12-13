package topology

import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/ibm/chic-sched/pkg/system"
	"github.com/ibm/chic-sched/pkg/util"
)

// PTree : a physical tree topology
//   - all nodes are of type PNode
//   - leaf nodes point to type PE objects
type PTree struct {
	// extends Tree
	Tree
}

// NewPTree : create a new physical tree
//   - returns nil if bad parameters
func NewPTree(tree *Tree) *PTree {
	if tree == nil {
		return nil
	}
	return &PTree{
		Tree: *tree,
	}
}

// GetPEs : get a map of all PEs (leaf nodes)
func (pTree *PTree) GetPEs() map[string]*system.PE {
	pLeaves := pTree.GetLeaves()
	serverMap := make(map[string]*system.PE)
	for _, node := range pLeaves {
		pe := (*system.PE)(unsafe.Pointer(node.Entity))
		serverMap[pe.GetID()] = pe
	}
	return serverMap
}

// SetPEs : set PEs as leaf nodes
func (pTree *PTree) SetPEs(pes map[string]*system.PE) {
	leavesMap := pTree.GetLeavesMap()
	for id, node := range leavesMap {
		if pe, exists := pes[id]; exists {
			node.Entity = (*system.Entity)(unsafe.Pointer(pe))
		}
	}
}

// ResetNumFit : set number can fit to zero for all nodes in the tree
func (pTree *PTree) ResetNumFit() {
	if pTree.root != nil {
		pRoot := (*PNode)(unsafe.Pointer(pTree.root))
		pRoot.ResetNumFit()
	}
}

// ResetNumClaimed : set number claimed to zero for all nodes in the tree
func (pTree *PTree) ResetNumClaimed() {
	if pTree.root != nil {
		pRoot := (*PNode)(unsafe.Pointer(pTree.root))
		pRoot.ResetNumClaimed()
	}
}

// ResetResources : reset resource capacity and allocated fields for all nodes in the tree
func (pTree *PTree) ResetResources() {
	if pTree.root != nil {
		pRoot := (*PNode)(unsafe.Pointer(pTree.root))
		pRoot.ResetResources()
	}
}

// PercolateResources : set allocation and capacity of all nodes, from the leaves up to the root
func (pTree *PTree) PercolateResources() {
	pTree.ResetResources()
	leaves := pTree.GetLeaves()
	for _, leaf := range leaves {
		pe := (*system.PE)(unsafe.Pointer(leaf.Entity))
		allocated := pe.GetAllocated()
		capacity := pe.GetCapacity()
		weight := pe.GetWeight()
		path := leaf.GetPathToRoot()
		for i, node := range path {
			pNode := (*PNode)(unsafe.Pointer(node))

			if i == 0 {
				pNode.allocated = allocated.Clone()
				pNode.capacity = capacity.Clone()
				pNode.AddWeight(weight)
			} else {
				pNode.allocated.Add(allocated)
				pNode.capacity.Add(capacity)
				pNode.AddWeight(weight)
			}
		}
	}
}

// PercolateNumFit : set number that can fit given demand for all nodes, from the leaves up to the root
func (pTree *PTree) PercolateNumFit(demand *util.Allocation) {
	pTree.ResetNumFit()
	leaves := pTree.GetLeaves()
	for _, leaf := range leaves {
		pe := (*system.PE)(unsafe.Pointer(leaf.Entity))
		numFit := demand.NumberToFit(pe.GetAllocated(), pe.GetCapacity())
		path := leaf.GetPathToRoot()
		for i, node := range path {
			pNode := (*PNode)(unsafe.Pointer(node))
			if i == 0 {
				pNode.SetNumFit(numFit)
			} else {
				pNode.IncNumFit(numFit)
			}
		}
	}
}

// MergeClaimedToFit : add number claimed to number can fit in all nodes
func (pTree *PTree) MergeClaimedToFit() {
	for _, node := range pTree.GetNodeListBFS() {
		pNode := (*PNode)(unsafe.Pointer(node))
		pNode.IncNumFit(pNode.GetNumClaimed())
	}
}

// SetNodeLevels : set levels in all nodes in the tree
func (pTree *PTree) SetNodeLevels() {
	if pTree.root != nil {
		pRoot := (*PNode)(unsafe.Pointer(pTree.root))
		pRoot.SetLevelSubtree(pTree.GetHeight())
	}
}

// CopyByLeafIDs : create a copy of pTree, only with specified subset of leaves, may return nil
//   - Node ID and value are copied, but not parent and childern links
//   - PNode level is copied, but not capacity, allocated, and other data
//   - PE leaves are copied by reference
func (pTree *PTree) CopyByLeafIDs(leafIDs []string) *PTree {
	pRoot := (*PNode)(unsafe.Pointer(pTree.root))
	if pRoot == nil {
		return nil
	}

	// initialize
	numResources := pRoot.GetNumResources()

	leavesMap := pTree.GetLeavesMap()
	var pRootCopy *PNode
	allNodes := make(map[string]*PNode)
	var prevNode *PNode

	// go over leaf IDs
	for _, id := range leafIDs {
		leafNode := leavesMap[id]
		// skip non-existing leaf IDs
		if leafNode == nil {
			continue
		}
		// visit (and copy non-copied) nodes from root to leaf node
		path := leafNode.GetPathToRoot()
		for i := len(path) - 1; i >= 0; i-- {
			curNode := path[i]
			id := curNode.GetID()
			if curNodeCopy, exists := allNodes[id]; exists {
				prevNode = (*PNode)(unsafe.Pointer(curNodeCopy))
				continue
			}
			// handle if leaf PE node
			var node *Node
			if i == 0 {
				node = NewNode((*system.Entity)(unsafe.Pointer(curNode.Entity)))
			} else {
				node = NewNode(&system.Entity{ID: id})
			}
			// create PNode copy
			curNodeCopy := NewPNode(node, 0, numResources)
			curNodeCopy.SetValue(curNode.GetValue())
			curNodeCopy.SetLevel((*PNode)(unsafe.Pointer(curNode)).GetLevel())
			curNodeCopy.SetWeight((*PNode)(unsafe.Pointer(curNode)).GetWeight())
			allNodes[id] = curNodeCopy
			// check if we are at the root node, otherwise link to parent
			if prevNode == nil {
				pRootCopy = curNodeCopy
			} else {
				prevNode.AddChild((*Node)(unsafe.Pointer(curNodeCopy)))
			}
			prevNode = curNodeCopy
		}
	}
	return NewPTree(NewTree((*Node)(unsafe.Pointer(pRootCopy))))
}

// String : a print out of the physical tree
func (pTree *PTree) String() string {
	var b bytes.Buffer
	b.WriteString("pTree:\n")
	fmt.Fprintf(&b, "%s", &pTree.Tree)
	b.WriteString("\n")

	b.WriteString("pNodes:\n")
	nodes := pTree.GetNodeListBFS()
	for _, node := range nodes {
		pNode := (*PNode)(unsafe.Pointer(node))
		fmt.Fprintf(&b, "%s\n", pNode)
	}
	b.WriteString("\n")
	return b.String()
}
