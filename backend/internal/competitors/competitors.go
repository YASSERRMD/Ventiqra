// Package competitors models rival companies: deterministic generation, strength
// growth, periodic launches, and the aggregate market pressure they exert on the
// player's acquisition. All formulas are pure.
package competitors

import "math/rand/v2"

// Competitor is a single rival.
type Competitor struct {
	Name          string
	Strength      int     // 0..100
	MarketShare   float64 // 0..1
	LastLaunchDay int
}

var rivalNames = []string{
	"Nimbus Labs", "Orbital Systems", "Cobalt Inc", "Helix Corp",
	"Vector AI", "Quanta Tech", "Forge Dynamics", "Polaris Co",
}

// competitorCount is the number of rivals generated per company.
const competitorCount = 3

// compStreamSalt separates the competitor RNG stream.
const compStreamSalt uint64 = 7845129630854702912

// Generate returns the deterministic initial rival set for a company.
func Generate(seed int64) []Competitor {
	r := rand.New(rand.NewPCG(uint64(seed)^compStreamSalt, uint64(seed)))
	out := make([]Competitor, 0, competitorCount)
	used := map[int]bool{}
	for len(out) < competitorCount {
		idx := r.IntN(len(rivalNames))
		if used[idx] {
			continue
		}
		used[idx] = true
		out = append(out, Competitor{
			Name:        rivalNames[idx],
			Strength:    15 + r.IntN(25), // 15..40
			MarketShare: 0.03 + r.Float64()*0.07,
		})
	}
	return out
}

// Advance applies one simulated day of competitor evolution. Strength trends up
// slowly and occasionally a launch boosts strength and share.
func Advance(c Competitor, seed, day int64) Competitor {
	r := rand.New(rand.NewPCG(uint64(seed)^compStreamSalt, uint64(day)^uint64(len(c.Name))))
	strength := c.Strength + r.IntN(3) // +0..2/day
	if strength > 100 {
		strength = 100
	}
	share := c.MarketShare
	// Launch event roughly every ~30 days: a strength and share jump.
	if int(day)-c.LastLaunchDay >= 30 && r.Float64() < 0.5 {
		strength += 5
		if strength > 100 {
			strength = 100
		}
		share += 0.02
		c.LastLaunchDay = int(day)
	}
	if share > 0.9 {
		share = 0.9
	}
	return Competitor{Name: c.Name, Strength: strength, MarketShare: share, LastLaunchDay: c.LastLaunchDay}
}

// Pressure returns the aggregate 0..0.5 acquisition penalty the rivals exert on
// the player, derived from their average strength.
func Pressure(comps []Competitor) float64 {
	if len(comps) == 0 {
		return 0
	}
	var sum int
	for _, c := range comps {
		sum += c.Strength
	}
	avg := float64(sum) / float64(len(comps))
	p := avg / 100.0 * 0.4
	if p > 0.5 {
		p = 0.5
	}
	return p
}
