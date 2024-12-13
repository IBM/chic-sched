package system

import (
	"fmt"
	"github.com/ibm/chic-sched/pkg/util"
	"sort"
)

// PE : Physical Entity
type PE struct {
	// extends Entity
	Entity
	// resource capacity
	capacity *util.Allocation
	// resource allocated
	allocated *util.Allocation
	// hosted LEs
	hosted map[string]*LE
	// weight for guiding placement decision
	weight int
}

// NewPE : create a new PE
//   - makes a copy of resource capacity
//   - assumes zero resource allocated
//   - returns nil if bad parameters
func NewPE(id string, capacity *util.Allocation) *PE {
	if len(id) == 0 || capacity == nil || capacity.GetSize() == 0 {
		return nil
	}
	allocated, _ := util.NewAllocation(capacity.GetSize())
	return &PE{
		Entity:    Entity{ID: id},
		capacity:  capacity.Clone(),
		allocated: allocated,
		hosted:    make(map[string]*LE),
		weight:    util.DefaultWeight,
	}
}

// GetWeight : get the weight of this PE
func (pe *PE) GetWeight() int {
	return pe.weight
}

// SetWeight : set the weight of this PE
func (pe *PE) SetWeight(weight int) {
	pe.weight = weight
}

// GetCapacity : get resource capacity
func (pe *PE) GetCapacity() *util.Allocation {
	return pe.capacity
}

// GetAllocated : get resource allocated
func (pe *PE) GetAllocated() *util.Allocation {
	return pe.allocated
}

// SetAllocated : set resource allocated
//   - assuming length of allocated same as length of capacity
func (pe *PE) SetAllocated(allocated *util.Allocation) {
	if pe.capacity.SameSize(allocated) {
		pe.allocated = allocated
	}
}

// AddAllocated : add resource allocated
//   - assuming length of allocated same as length of capacity
func (pe *PE) AddAllocated(allocated *util.Allocation) {
	if pe.capacity.SameSize(allocated) {
		pe.allocated.Add(allocated)
	}
}

// PlaceLE : place an LE on this PE
//   - assuming allowed resource overflow
func (pe *PE) PlaceLE(le *LE) bool {
	if le == nil || le.demand == nil || !le.demand.SameSize(pe.capacity) {
		return false
	}
	leID := le.GetID()
	if _, exists := pe.hosted[leID]; exists {
		return false
	}
	pe.allocated.Add(le.demand)
	pe.hosted[leID] = le
	le.SetHost(pe)
	pe.hosted[leID] = le
	return true
}

// UnPlaceLE : unplace an LE from this PE
func (pe *PE) UnPlaceLE(le *LE) bool {
	if le == nil || le.demand == nil || !le.demand.SameSize(pe.capacity) {
		return false
	}
	leID := le.GetID()
	if _, exists := pe.hosted[leID]; !exists {
		return false
	}
	pe.allocated.Subtract(le.demand)
	delete(pe.hosted, leID)
	le.SetHost(nil)
	return true
}

// GetHostedIDs : a sorted list of IDs of all LEs hosted by this PE
func (pe *PE) GetHostedIDs() []string {
	ids := make([]string, len(pe.hosted))
	i := 0
	for leID := range pe.hosted {
		ids[i] = leID
		i++
	}
	sort.Strings(ids)
	return ids
}

// UNDONE:

// TODO: add host add/remove LE

// String : a print out of the PE
func (pe *PE) String() string {
	return fmt.Sprintf("PE: ID=%s; weight=%v; cap=%v; alloc=%v; hosted=%v", pe.GetID(), pe.weight, pe.capacity, pe.allocated,
		pe.GetHostedIDs())
}
