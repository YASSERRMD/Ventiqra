// Package launch models product launch readiness and the initial customer
// boost a launch produces. The formulas are pure and deterministic so launch
// outcomes are reproducible and testable.
package launch

import "math/rand/v2"

// MinReadiness is the readiness score required to launch a product.
const MinReadiness = 40.0

// Inputs describes the state used to compute launch readiness.
type Inputs struct {
	DevProgress float64 // 0..100
	AvgSkill    float64 // 0..100 (team average; 0 when no team)
	TeamSize    int
}

// Readiness computes a 0..100 launch readiness score from development progress,
// average team skill, and team size. Development maturity dominates, with skill
// and a modest team-size bonus rounding it out.
func Readiness(in Inputs) float64 {
	dev := clampFloat(in.DevProgress, 0, 100)
	skill := clampFloat(in.AvgSkill, 0, 100)
	size := in.TeamSize
	if size < 0 {
		size = 0
	}
	if size > 10 {
		size = 10
	}
	score := 0.6*dev + 0.3*skill + 0.1*float64(size)*10
	return clampFloat(score, 0, 100)
}

// CanLaunch reports whether a readiness score clears the launch threshold.
func CanLaunch(readiness float64) bool {
	return readiness >= MinReadiness
}

// InitialCustomers returns the deterministic customer count a launch produces.
// It scales with readiness and adds reproducible jitter derived from seed so the
// same product/seed always launches with the same starting base.
func InitialCustomers(readiness float64, seed int64) int {
	r := rand.New(rand.NewPCG(uint64(seed), 0xA17C_51ED_2701_BEEF))
	base := readiness * 5 // readiness 80 → ~400 customers
	jitter := r.Float64() * 50
	n := int(base + jitter)
	if n < 0 {
		return 0
	}
	return n
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
