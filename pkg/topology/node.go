package topology

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ibm/chic-sched/pkg/system"
)

// Node : a basic node in a tree
type Node struct {
	// the entity that this node represents
	Entity *system.Entity
	// value associated with the node
	value int
	// the parent of this node in the tree
	parent *Node
	// set of children of this node in the tree (node ID -> node)
	children map[string]*Node
}

// NewNode : create a node
func NewNode(entity *system.Entity) *Node {
	if entity == nil || len(entity.ID) == 0 {
		return nil
	}
	return &Node{
		Entity:   entity,
		value:    0,
		parent:   nil,
		children: make(map[string]*Node),
	}
}

// GetID : the unique ID of this node
func (n *Node) GetID() string {
	return n.Entity.GetID()
}

// GetValue : the value of this node
func (n *Node) GetValue() int {
	return n.value
}

// SetValue : set the value of this node
func (n *Node) SetValue(value int) {
	n.value = value
}

// GetParent : the parent of this node
func (n *Node) GetParent() *Node {
	return n.parent
}

// setParent : set the parent of this node
func (n *Node) setParent(parent *Node) {
	n.parent = parent
}

// GetChildren : the children of this node
func (n *Node) GetChildren() []*Node {
	children := make([]*Node, 0, len(n.children))
	for _, c := range n.children {
		children = append(children, c)
	}
	return children
}

// IsRoot : is this node the root of the tree
func (n *Node) IsRoot() bool {
	return n.parent == nil
}

// IsLeaf : is this node a leaf in the tree
func (n *Node) IsLeaf() bool {
	return len(n.children) == 0
}

// HasLeaf : does the subtree from this node has a given node as a leaf
func (n *Node) HasLeaf(leafID string) bool {
	for _, leaf := range n.GetLeaves() {
		if leaf.GetID() == leafID {
			return true
		}
	}
	return false
}

// AddChild : add a child to this node;
// return false if child already exists
func (n *Node) AddChild(child *Node) bool {
	if child != nil {
		cid := child.GetID()
		if _, exists := n.children[cid]; !exists {
			n.children[cid] = child
			child.setParent(n)
			return true
		}
	}
	return false
}

// RemoveChild : remove a child from this node
func (n *Node) RemoveChild(child *Node) bool {
	if child != nil {
		cid := child.GetID()
		if _, exists := n.children[cid]; exists {
			delete(n.children, cid)
			child.setParent(nil)
			return true
		}
	}
	return false
}

// RemoveChildren : remove all children from this node
func (n *Node) RemoveChildren() {
	for _, c := range n.children {
		c.setParent(nil)
	}
	n.children = make(map[string]*Node)
}

// GetNumChildren : the number of children of this node
func (n *Node) GetNumChildren() int {
	return len(n.children)
}

// GetHeight : the height of this node in the tree
func (n *Node) GetHeight() int {
	h := 0
	for _, c := range n.children {
		ch := c.GetHeight() + 1
		if ch > h {
			h = ch
		}
	}
	return h
}

// GetLeaves : the leaf nodes in the subtree consisting of this node as a root
func (n *Node) GetLeaves() []*Node {
	list := make([]*Node, 0)
	if n.IsLeaf() {
		list = append(list, n)
	} else {
		for _, c := range n.children {
			list = append(list, c.GetLeaves()...)
		}
	}
	return list
}

// GetPathToRoot : the path from this node to the root of the tree;
// node (root) is first (last) in list
func (n *Node) GetPathToRoot() []*Node {
	path := make([]*Node, 0)
	node := n
	for !node.IsRoot() {
		path = append(path, node)
		node = node.GetParent()
	}
	path = append(path, node)
	return path
}

// ToMap : create a map of the subtree rooted at this node
func (n *Node) ToMap() TreeMap {
	m := make(map[string]TreeMap)
	for _, c := range n.GetChildren() {
		m[c.GetID()] = c.ToMap()
	}
	return m
}

// NodeValue : name and value pair
type NodeValue struct {
	Name  string
	Value int
}

func (nodeValue *NodeValue) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "(%s,%d)", nodeValue.Name, nodeValue.Value)
	return b.String()
}

// GetLeavesDepth : get a list of leaves and their depth from this node
func (node *Node) GetLeavesDepth() (nodeValue []*NodeValue) {
	if node.IsLeaf() {
		nodeValue = []*NodeValue{{Name: node.GetID(), Value: 0}}
	} else {
		nodeValue = make([]*NodeValue, 0)
		for _, child := range node.children {
			nvc := child.GetLeavesDepth()
			for _, nv := range nvc {
				nv.Value++
			}
			nodeValue = append(nodeValue, nvc...)
		}
	}
	return nodeValue
}

// String : a print out of the node;
// e.g. A -> ( C -> ( D ) B )
func (n *Node) String() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%s ", n.GetID()))
	if !n.IsLeaf() {
		b.WriteString(" -> ( ")
		// order children by name
		ids := make([]string, 0, len(n.children))
		for id := range n.children {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			fmt.Fprintf(&b, "%s ", n.children[id])
		}
		b.WriteString(")")
	}
	return b.String()
}
