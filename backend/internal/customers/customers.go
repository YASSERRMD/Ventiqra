// Package customers models per-product customer dynamics: acquisition, churn,
// satisfaction drift, and monthly active users. All formulas are pure and
// deterministic for a given (seed, day) so a reloaded game evolves identically.
package customers

import "math/rand/v2"

// Product is the per-product customer state advanced by the model.
type Product struct {
	Total        int
	MAU          int
	Churned      int
	Satisfaction int // 0..100
}

const (
	// SatisfactionBaseline is the equilibrium satisfaction drifts toward.
	SatisfactionBaseline = 70
	customerStreamSalt   = 0x4B1C_7A55_D00D_F00D
)

// Acquisition returns the deterministic new customers gained in one day. Growth
// scales with satisfaction: happier customers drive more word-of-mouth.
func Acquisition(satisfaction int, r *rand.Rand) int {
	sat := clamp(satisfaction, 0, 100)
	base := float64(sat) / 10.0 // satisfaction 80 → ~8/day baseline
	jitter := r.Float64() * 3
	n := int(base + jitter)
	if n < 0 {
		return 0
	}
	return n
}

// Churn returns the deterministic customers lost in one day. The churn rate
// rises sharply as satisfaction falls: unhappy users leave faster.
func Churn(total, satisfaction int, r *rand.Rand) int {
	if total <= 0 {
		return 0
	}
	sat := clamp(satisfaction, 0, 100)
	// Daily churn rate up to ~3% at satisfaction 0, near 0 at 100.
	rate := (100.0 - float64(sat)) / 100.0 * 0.03
	lost := int(float64(total)*rate + r.Float64())
	if lost < 0 {
		return 0
	}
	if lost > total {
		lost = total
	}
	return lost
}

// MauRatio returns the fraction of total customers who are active in a month,
// as a function of satisfaction. Higher satisfaction → higher activity.
func MauRatio(satisfaction int) float64 {
	sat := clamp(satisfaction, 0, 100)
	return float64(sat)/100.0*0.85 + 0.1 // satisfaction 80 → 0.78
}

// SatisfactionDrift nudges satisfaction toward the baseline with small noise.
func SatisfactionDrift(satisfaction int, r *rand.Rand) int {
	sat := satisfaction
	switch {
	case sat < SatisfactionBaseline:
		sat++
	case sat > SatisfactionBaseline:
		sat--
	}
	noise := int(r.Float64()*3) - 1 // -1..+1
	sat += noise
	return clamp(sat, 0, 100)
}

// Advance applies one simulated day of acquisition, churn, satisfaction drift,
// and MAU recompute to a product's customer state, deterministically for the
// given (seed, day) round.
func Advance(p Product, seed int64, day int) Product {
	r := rand.New(rand.NewPCG(uint64(seed), uint64(day)^customerStreamSalt))
	acq := Acquisition(p.Satisfaction, r)
	lost := Churn(p.Total, p.Satisfaction, r)

	total := p.Total + acq - lost
	if total < 0 {
		total = 0
	}
	churned := p.Churned + lost
	sat := SatisfactionDrift(p.Satisfaction, r)
	mau := int(float64(total) * MauRatio(sat))

	return Product{Total: total, MAU: mau, Churned: churned, Satisfaction: sat}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
