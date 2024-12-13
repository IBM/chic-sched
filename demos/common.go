package demos

import (
	"github.com/ibm/chic-sched/pkg/system"
	"github.com/ibm/chic-sched/pkg/util"
	"math"
	"math/rand"
)

// PlaceBackgroungLoad : place random allocation on PEs
//   - loadFactor: target load
//   - alpha: probability of full server
//   - beta: probability of idle server
//   - cov: coefficient of variation
func PlaceBackgroungLoad(pes []*system.PE, loadFactor float64, alpha float64, beta float64, cov float64) (avg float64) {
	avg = ComputeAverageLoad(loadFactor, alpha, beta)
	for i := 0; i < len(pes); i++ {
		alloc := pes[i].GetAllocated().GetValue()

		var y float64
		x := rand.Float64()
		if x < alpha {
			y = 1
		} else if x < alpha+beta {
			y = 0
		} else {
			y = avg * (rand.NormFloat64()*cov + 1)
		}
		y = math.Max(math.Min(y, 1), 0)

		capacity := pes[i].GetCapacity().GetValue()
		numResources := len(capacity)
		for k := 0; k < numResources; k++ {
			z := int(math.Round(y * float64(capacity[k])))
			z = util.Max(util.Min(z, capacity[k]), 0)
			alloc[k] = z
		}
	}
	return avg
}

func PlaceWeights(pes []*system.PE) {
	for i := 0; i < len(pes); i++ {
		pes[i].SetWeight(rand.Intn(util.MaxWeight) + util.MinWeight)
	}
}

// ComputeAverageLoad : average utilization (non zero and non one)
func ComputeAverageLoad(loadFactor, alpha, beta float64) (avg float64) {
	avg = 0.0
	if alpha+beta < 1 {
		avg = (loadFactor - alpha) / (1 - (alpha + beta))
	}
	avg = math.Min(math.Max(avg, 0), 1)
	return avg
}
