// Package hiring implements Ventiqra's deterministic talent market.
//
// The market is pure: given a company's seed and the current simulated day, it
// reproduces an identical candidate pool and the same offer outcomes. This lets
// a reloaded game show the same hiring opportunities and keeps the simulation
// fully reproducible and testable.
package hiring

import "math/rand/v2"

// PoolSize is the number of candidates presented in a single hiring round.
const PoolSize = 6

// offerSalt is XORed into the offer-decision stream so acceptance rolls do not
// correlate with pool generation.
const offerSalt uint64 = 0xD1B54A32D192ED03

// Candidate describes a single hireable prospect in a round.
type Candidate struct {
	Index             int     `json:"index"`
	Role              string  `json:"role"`
	Name              string  `json:"name"`
	Quality           string  `json:"quality"`
	Skill             int     `json:"skill"`
	SalaryExpectation int64   `json:"salary_expectation_cents"`
	HiringFee         int64   `json:"hiring_fee_cents"`
	AcceptanceChance  float64 `json:"acceptance_chance"`
}

var firstNames = []string{
	"Ada", "Grace", "Linus", "Margaret", "Hedy", "Alan", "Katherine",
	"Tim", "Radia", "Donald", "Frances", "Ken", "Barbara", "Vinton",
	"Anita", "Edsger",
}

var lastNames = []string{
	"Lovelace", "Hopper", "Torvalds", "Hamilton", "Lamarr", "Turing",
	"Johnson", "Berners-Lee", "Perlman", "Knuth", "Allen", "Thompson",
	"Liskov", "Cerf", "Borg", "Dijkstra",
}

var roles = []string{
	"engineer", "designer", "sales", "marketing", "support", "operations",
}

// newPoolRand builds the deterministic generator for a (seed, day) round.
func newPoolRand(seed, day int64) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed), uint64(day)^0x51ED2701))
}

// GeneratePool returns the deterministic candidate pool for a round identified
// by the company seed and the current simulated day. The same inputs always
// yield the same candidates in the same order.
func GeneratePool(seed, day int64) []Candidate {
	r := newPoolRand(seed, day)
	pool := make([]Candidate, 0, PoolSize)
	for i := 0; i < PoolSize; i++ {
		pool = append(pool, generateOne(r, i))
	}
	return pool
}

// QualityTier labels a skill score. It drives candidate display and later
// balancing (stronger candidates command higher salaries and fees).
func QualityTier(skill int) string {
	switch {
	case skill >= 80:
		return "strong"
	case skill >= 50:
		return "average"
	default:
		return "weak"
	}
}

func generateOne(r *rand.Rand, index int) Candidate {
	role := roles[r.IntN(len(roles))]
	first := firstNames[r.IntN(len(firstNames))]
	last := lastNames[r.IntN(len(lastNames))]
	// Skill skews toward the middle: 30 + N(0,99) keeps most candidates usable
	// while still producing the occasional star or dud.
	skill := 30 + r.IntN(70)
	if skill > 99 {
		skill = 99
	}
	tier := QualityTier(skill)

	salary := baseSalary(role)
	fee := hiringFee(salary, tier)
	chance := acceptanceChance(tier, salary)

	return Candidate{
		Index:             index,
		Role:              role,
		Name:              first + " " + last,
		Quality:           tier,
		Skill:             skill,
		SalaryExpectation: salary,
		HiringFee:         fee,
		AcceptanceChance:  chance,
	}
}

// baseSalary returns the deterministic monthly salary expectation (cents) for a
// role, independent of randomness so roles stay comparable across rounds.
func baseSalary(role string) int64 {
	switch role {
	case "engineer":
		return 13_000_00
	case "designer":
		return 10_500_00
	case "sales":
		return 9_500_00
	case "marketing":
		return 9_000_00
	case "support":
		return 7_000_00
	case "operations":
		return 8_500_00
	default:
		return 8_000_00
	}
}

// hiringFee is the one-time recruiting cost: a fraction of annual salary that
// scales with quality tier. Strong candidates cost more to land.
func hiringFee(salary int64, tier string) int64 {
	annual := salary * 12
	switch tier {
	case "strong":
		return annual / 6
	case "average":
		return annual / 10
	default:
		return annual / 20
	}
}

// acceptanceChance is the base probability a candidate accepts an offer. Stars
// are pickier; weaker candidates accept more readily.
func acceptanceChance(tier string, salary int64) float64 {
	switch tier {
	case "strong":
		return 0.55
	case "average":
		return 0.75
	default:
		return 0.9
	}
}

// OfferAccepted reports deterministically whether a candidate at the given index
// accepts an offer in the round (seed, day). It is independent of pool ordering
// so the decision for candidate N is stable even if the pool is regenerated.
func OfferAccepted(seed, day int64, index int) bool {
	pool := GeneratePool(seed, day)
	if index < 0 || index >= len(pool) {
		return false
	}
	c := pool[index]
	r := rand.New(rand.NewPCG(uint64(seed), uint64(day)^uint64(index+1)^offerSalt))
	return r.Float64() < c.AcceptanceChance
}
