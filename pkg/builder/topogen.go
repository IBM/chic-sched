package builder

import (
	"encoding/json"
	"fmt"
	"github.com/ibm/chic-sched/pkg/system"
	"github.com/ibm/chic-sched/pkg/topology"
	"github.com/ibm/chic-sched/pkg/util"
	"unsafe"
)

// CreateTopologyTreeFromJson : create a PTree fron JSON string
func CreateTopologyTreeFromJson(topologyTreeString string) (pTree *topology.PTree, err error) {
	var topologyTree util.TopologyTree
	err = json.Unmarshal([]byte(topologyTreeString), &topologyTree)
	if err != nil {
		return nil, fmt.Errorf("error parsing tree: %s", err.Error())
	}

	// process kind field
	fmt.Println("kind=" + topologyTree.Kind)
	if topologyTree.Kind != util.DefaultTreeKind {
		return nil, fmt.Errorf("invalid kind: %s", topologyTree.Kind)
	}

	// process tree name field
	treeName := topologyTree.MetaData.Name
	fmt.Println("treeName=" + treeName)

	// process topology levels
	resourceNames := make([]string, len(topologyTree.Spec.ResourceNames))
	copy(resourceNames, topologyTree.Spec.ResourceNames)
	numResources := len(resourceNames)
	fmt.Println("numResources=", numResources)
	fmt.Println("resourceNames=", resourceNames)

	levelNames := make([]string, len(topologyTree.Spec.LevelNames))
	copy(levelNames, topologyTree.Spec.LevelNames)
	fmt.Println("numLevels=", len(levelNames))
	fmt.Println("levelNames=", levelNames)

	// make PTree
	root := topology.NewPNode(topology.NewNode(&system.Entity{ID: "root"}), 0, numResources)
	MakeSubtreeFromSpec(root, topologyTree.Spec.Tree)
	pTree = topology.NewPTree(topology.NewTree((*topology.Node)(unsafe.Pointer(root))))
	pTree.SetNodeLevels()
	fmt.Println("pTree = ", pTree)
	return pTree, nil
}

// MakeSubtreeFromSpec : make a substree rooted at a node given tree spec level hierarchy,
// without setting level values
func MakeSubtreeFromSpec(pNode *topology.PNode, spec util.TreeSpec) {
	numResources := pNode.GetNumResources()
	for childName, childSpec := range spec.Level {
		child := topology.NewPNode(topology.NewNode(&system.Entity{ID: childName}), 0, numResources)
		pNode.AddChild((*topology.Node)(unsafe.Pointer(child)))
		MakeSubtreeFromSpec(child, childSpec)
	}
}

// CreateFlatTopology : create a flat PTree
func CreateFlatTopology(leafIDs []string, numResources int) (pTree *topology.PTree) {
	root := topology.NewPNode(topology.NewNode(&system.Entity{ID: "root"}), 0, numResources)
	root.SetLevel(1)
	for _, id := range leafIDs {
		leaf := topology.NewPNode(topology.NewNode(&system.Entity{ID: id}), 0, numResources)
		leaf.SetLevel(0)
		root.AddChild((*topology.Node)(unsafe.Pointer(leaf)))
	}
	pTree = topology.NewPTree(topology.NewTree((*topology.Node)(unsafe.Pointer(root))))
	return pTree
}
