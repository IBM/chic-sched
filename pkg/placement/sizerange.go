package placement

import (
	"fmt"
	"github.com/ibm/chic-sched/pkg/util"
)

// SizeRange : desired range of number of members to place in a node
type SizeRange struct {
	// minimum value
	Min int
	// desired value
	Desired int
	// max value
	Max int
	// factor
	Factor int

	// reference to the placement group
	pg *PGroup
}

// CreateSizeRange : create a new size range to place in a node
//   - level: the level of the node
//   - numToPlace: number of members to place on node and its sibling nodes
//   - numNodes: number of sibling nodes (including the node)
//   - numPartitionsPlaced: number of partitions placed so far
//   - returns nil if bad parameters, or no placement possible
func CreateSizeRange(pg *PGroup, level int, numToPlace int, numNodes int,
	numPartitionsPlaced int) *SizeRange {

	if pg == nil || level < 0 || numToPlace <= 0 || numNodes <= 0 || numPartitionsPlaced < 0 {
		return nil
	}

	// get level constraint of placement group (default if unspecified)
	lc := pg.GetLevelConstraint(level)
	isHard := lc.IsHard()
	affinity := lc.Affinity()
	factor := 1

	// create size range based on constraints
	var min, desired, max int
	if isHard {
		// HARD constraint
		if affinity == util.Pack {
			min = numToPlace
			desired = numToPlace
			max = numToPlace
		} else {
			min = 1
			desired = 1
			max = 1
		}
	} else {
		// SOFT constraint

		// calculate remaining partitions (if applicable)
		numPartitionsLeft := -1
		if numPartitions, isSet := lc.GetNumPartitions(); isSet {
			numPartitionsLeft = numPartitions - numPartitionsPlaced
			if numPartitionsLeft <= 0 || numPartitionsLeft > numNodes {
				return nil
			}
		}

		// get partition range (if applicable)
		minRange, maxRange, okRange := lc.GetRange()
		minToPlace := 1
		if okRange {
			minToPlace = minRange
		}
		minToLeave := 0
		if numPartitionsLeft > 1 {
			minToLeave = (numPartitionsLeft - 1)
			if okRange {
				minToLeave *= minRange
			}
		}
		if numToPlace < minToPlace+minToLeave {
			return nil
		}

		// determine size range according to constraints
		if affinity == util.Pack {
			// PACK
			desired = numToPlace - minToLeave
			if okRange {
				desired = util.Min(desired, maxRange)
				min = minRange
				max = maxRange
			} else {
				min = 1
				max = desired
			}
		} else {
			// SPREAD
			divisor := numNodes
			if numPartitionsLeft > 0 {
				divisor = numPartitionsLeft
			}
			desired = util.CeilDivide(numToPlace, divisor)
			desired = util.Max(desired, 1)
			if okRange {
				desired = util.Min(util.Max(desired, minRange), maxRange)
				min = minRange
				max = maxRange
			} else {
				min = 1
				max = numToPlace
			}
		}

		// handle factor
		var hasFactor bool
		if factor, hasFactor = lc.GetFactor(); hasFactor {
			var ok bool
			if min, ok = util.AboveMultiple(min, factor); !ok {
				return nil
			}
			if max, ok = util.BelowMultiple(max, factor); !ok {
				return nil
			}
			if desired, ok = util.BelowMultiple(desired, factor); !ok {
				return nil
			}
			if min > desired || desired > max {
				return nil
			}
		}
	}

	return &SizeRange{
		Min:     min,
		Desired: desired,
		Max:     max,
		Factor:  factor,
		pg:      pg,
	}
}

// NumberToPlace : best choice of number of members to place given the number that fits
func (sr *SizeRange) NumberToPlace(numFit int) int {
	if numFit < sr.Min {
		return 0
	}
	if sr.Factor > 1 {
		var valid bool
		if numFit, valid = util.BelowMultiple(numFit, sr.Factor); !valid {
			return 0
		}
	}
	return util.Min(util.Min(numFit, sr.Max), sr.Desired)
}

// NumberInRange : check if given number falls within the range
func (sr *SizeRange) NumberInRange(num int) bool {
	return num >= sr.Min && num <= sr.Max && num%sr.Factor == 0
}

// String : a print out of the size range
func (sr *SizeRange) String() string {
	return fmt.Sprintf("[%d,%d,%d]", sr.Min, sr.Desired, sr.Max)
}
