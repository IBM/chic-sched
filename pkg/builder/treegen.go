package builder

import (
	"github.com/ibm/chic-sched/pkg/system"
	"github.com/ibm/chic-sched/pkg/topology"
	"github.com/ibm/chic-sched/pkg/util"
	"strconv"
	"unsafe"
)

// TreeGen : a physical tree and system generator
type TreeGen struct {
	// the pTree
	pTree *topology.PTree
	// list of PEs
	pes []*system.PE
}

// NewTreeGen : create a new physical tree generator
func NewTreeGen() *TreeGen {
	return &TreeGen{
		pTree: nil,
		pes:   make([]*system.PE, 0),
	}
}

// GetPTree : get the physical tree
func (tg *TreeGen) GetPTree() *topology.PTree {
	return tg.pTree
}

// GetPEs : get the PEs at the leaves of the physical tree
func (tg *TreeGen) GetPEs() []*system.PE {
	return tg.pes
}

// CreateUniformTree : create a uniform physical tree given degree at each level,
// e.g. degree=[2,3] results into:
// root -> ( node1 -> ( server0 server1 server2 ) node2 -> ( server3 server4 server5 ) )
// where all servers are homogeneous with the given resource capacity per server.
//   - returns nil if bad parameters
func (tg *TreeGen) CreateUniformTree(degree []int, capacity []int) *topology.PTree {
	td := createTreeData(degree)
	if td == nil {
		return nil
	}
	tg.pes = td.makeHomogeneousNodes(capacity)
	if tg.pes == nil {
		return nil
	}
	tg.pTree = td.connectNodes()
	return tg.pTree
}

// CreateUniformGroupedTree : create a uniform physical tree given degree at each level,
// e.g. degree=[2,3] results into:
// root -> ( node1 -> ( server0 server1 server2 ) node2 -> ( server3 server4 server5 ) )
// where groups of homogeneous servers are defined at some level in tree.
// For example if groupLevel=1 and groupCapacity=[[16 64] [32 256]] then
// servers under node1 have resource capacities [16 64] and
// servers under node2 have resource capacities [32 256].
// The first dimension of groupCapacity is the total number of nodes at level groupLevel,
// and the second dimension is the number of resources.
//   - returns nil if bad parameters
func (tg *TreeGen) CreateUniformGroupedTree(degree []int, groupLevel int, groupCapacity [][]int) *topology.PTree {
	td := createTreeData(degree)
	if td == nil {
		return nil
	}
	tg.pes = td.makeGroupedNodes(groupLevel, groupCapacity)
	if tg.pes == nil {
		return nil
	}
	tg.pTree = td.connectNodes()
	return tg.pTree
}

// treedata : data related to tree (vectors start from root down to leaves)
type treedata struct {
	// degree at each level (uniform tree)
	degree []int
	// height of tree
	height int

	// number of levels in the tree
	numLevels int
	// total number of nodes per level
	numPerLevel []int
	// cumulative number of nodes from root down to a level
	cumNum []int

	// total number of leaf nodes
	numLeaves int
	// total number of internal (non-leaf) nodes
	numInternalNodes int
	// total number of nodes in the tree
	numNodes int

	// an array of references to PNode objects
	// corresponding to the nodes in the tree
	allNodes []*topology.PNode
}

// createTreeData : calculate tree data given the tree degree
func createTreeData(degree []int) *treedata {
	if len(degree) == 0 {
		return nil
	}

	height := len(degree)
	numLevels := height + 1
	numPerLevel := make([]int, numLevels)
	cumNum := make([]int, numLevels)

	// count nodes
	numPerLevel[0] = 1
	cumNum[0] = 0
	for l := 1; l <= height; l++ {
		numPerLevel[l] = numPerLevel[l-1] * degree[l-1]
		cumNum[l] = cumNum[l-1] + numPerLevel[l-1]
	}
	numLeaves := numPerLevel[height]
	numInternalNodes := cumNum[height]
	numNodes := numInternalNodes + numLeaves

	return &treedata{
		degree: degree,
		height: height,

		numLevels:   numLevels,
		numPerLevel: numPerLevel,
		cumNum:      cumNum,

		numLeaves:        numLeaves,
		numInternalNodes: numInternalNodes,
		numNodes:         numNodes,
	}
}

// makeHomogeneousNodes : create all nodes in the tree with homogeneous servers and return leaf PEs
func (td *treedata) makeHomogeneousNodes(capacity []int) (pes []*system.PE) {
	numResources := len(capacity)
	if numResources == 0 {
		return nil
	}
	allNodes := make([]*topology.PNode, td.numNodes)
	nodeIdx := 0
	serverCapacity, _ := util.NewAllocationCopy(capacity)
	pes = make([]*system.PE, td.numLeaves)
	for l := 0; l < td.numLevels; l++ {
		for a := 0; a < td.numPerLevel[l]; a++ {
			name := nameAtLevel(l, a, td.height)
			var node *topology.PNode

			if l == td.numLevels-1 {
				// create PE leaf node
				pe := system.NewPE(name, serverCapacity)
				pes[a] = pe
				node = topology.NewPNode(topology.NewNode((*system.Entity)(unsafe.Pointer(pe))),
					0, numResources)
			} else {
				// create internal pNode
				node = topology.NewPNode(topology.NewNode(&system.Entity{ID: name}),
					0, numResources)
			}

			allNodes[nodeIdx] = node
			nodeIdx++
		}
	}
	td.allNodes = allNodes
	return pes
}

// makeGroupedNodes : create all nodes in the tree and return grouped leaf PEs
func (td *treedata) makeGroupedNodes(groupLevel int, groupCapacity [][]int) (pes []*system.PE) {
	numGroups := len(groupCapacity)
	numResources := len(groupCapacity[0])
	if groupLevel < 0 || groupLevel > td.height || numGroups != td.numPerLevel[groupLevel] || numResources == 0 {
		return nil
	}

	allNodes := make([]*topology.PNode, td.numNodes)
	nodeIdx := 0
	pes = make([]*system.PE, td.numLeaves)

	// create internal pNodes
	for l := 0; l < td.height; l++ {
		for a := 0; a < td.numPerLevel[l]; a++ {
			name := nameAtLevel(l, a, td.height)
			node := topology.NewPNode(topology.NewNode(&system.Entity{ID: name}),
				0, numResources)
			allNodes[nodeIdx] = node
			nodeIdx++
		}
	}

	// create grouped leaf pNodes
	/* group size should be divisible given uniform tree */
	groupSize := td.numLeaves / numGroups
	peIdx := 0
	for g := 0; g < numGroups; g++ {
		serverCapacity, _ := util.NewAllocationCopy(groupCapacity[g])
		for a := 0; a < groupSize; a++ {
			name := nameAtLevel(td.height, peIdx, td.height)
			pe := system.NewPE(name, serverCapacity)
			node := topology.NewPNode(topology.NewNode((*system.Entity)(unsafe.Pointer(pe))),
				0, numResources)
			pes[peIdx] = pe
			peIdx++
			allNodes[nodeIdx] = node
			nodeIdx++
		}
	}

	td.allNodes = allNodes
	return pes
}

// connectNodes : connect nodes, set level values, and return the PTree
func (td *treedata) connectNodes() (pTree *topology.PTree) {
	if td.allNodes == nil {
		return nil
	}
	root := td.allNodes[0]
	root.SetLevel(td.height)
	for l := 0; l < td.height; l++ {
		for a := 0; a < td.numPerLevel[l]; a++ {
			indexParent := td.cumNum[l] + a
			parent := td.allNodes[indexParent]
			heightParent := parent.GetLevel()
			for b := 0; b < td.degree[l]; b++ {
				indexChild := td.cumNum[l+1] + (a*td.degree[l] + b)
				child := td.allNodes[indexChild]
				parent.AddChild(&child.Node)
				child.SetLevel(heightParent - 1)
			}
		}
	}

	// create and initialize tree
	tree := topology.NewTree(&root.Node)
	pTree = topology.NewPTree(tree)
	pTree.PercolateResources()
	return pTree
}

// nameAtLevel : id for a node at a given level and index within a level
func nameAtLevel(level int, index int, height int) string {
	if level == 0 {
		return util.DefaultRootName
	}
	var key string
	l := height - level
	if l >= 0 && l < len(util.DefaultLevelNames) {
		key = util.DefaultLevelNames[l]
	} else {
		key = util.DefaultLevelName + strconv.FormatInt(int64(l), 10)
	}
	return key + "-" + strconv.FormatInt(int64(index), 10)
}
