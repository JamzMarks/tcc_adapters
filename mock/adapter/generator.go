package adapter

import (
	"math"
	"math/rand"
)

type RandomGenerator struct {
	r *rand.Rand
}

func NewRandomGenerator(seed int64) *RandomGenerator {
	if seed != 0 {
		return &RandomGenerator{r: rand.New(rand.NewSource(seed))}
	}
	return &RandomGenerator{r: rand.New(rand.NewSource(rand.Int63()))}
}

func (rg *RandomGenerator) ComputeNewValue(last *float64, deltaMax float64) float64 {
	base := Round5(rg.r.Float64())
	if last == nil {
		return base
	}

	delta := (rg.r.Float64()*2 - 1) * deltaMax
	newVal := *last + delta
	if newVal < 0 {
		newVal = 0
	}
	if newVal > 1 {
		newVal = 1
	}
	return Round5(newVal)
}

func Round5(f float64) float64 {
	p := math.Pow10(5)
	return math.Round(f*p) / p
}
