// Package sim implements Ventiqra's deterministic simulation engine.
//
// The engine is intentionally pure: it has no knowledge of the database or
// HTTP layer. A simulation is driven by an Engine applied to a State. Given
// the same seed, two engines produce identical state sequences, which makes
// every company's future reproducible and testable.
package sim

import "math/rand/v2"

// MinDailyDeltaCents is the lower bound (inclusive) of a single tick's random
// cash delta, expressed in cents.
const MinDailyDeltaCents int64 = -100_000

// MaxDailyDeltaCents is the upper bound (inclusive) of a single tick's random
// cash delta, expressed in cents.
const MaxDailyDeltaCents int64 = 50_000

// State holds the mutable simulation state for a single company. The *rand.Rand
// drives the deterministic daily change; callers must not share the same Rand
// across unrelated simulations.
type State struct {
	CompanyID string
	Day       int
	Cash      int64 // cents
	Seed      int64
	Rand      *rand.Rand
}

// Engine applies simulation rules to a State. The zero value is not usable;
// construct one with NewEngine.
type Engine struct {
	seed int64
}

// NewEngine constructs an Engine bound to the given seed. The seed is retained
// so callers can reproduce a fresh, identically seeded Rand for new States via
// NewState. The engine itself holds no mutable randomness.
func NewEngine(seed int64) *Engine {
	return &Engine{seed: seed}
}

// Seed returns the engine's deterministic seed.
func (e *Engine) Seed() int64 { return e.seed }

// NewState creates a fresh State for a company using the engine's seed. The
// returned State carries its own *rand.Rand so that advancing it does not
// affect other simulations sharing the same engine.
func (e *Engine) NewState(companyID string, cash int64) *State {
	return &State{
		CompanyID: companyID,
		Day:       0,
		Cash:      cash,
		Seed:      e.seed,
		Rand:      rand.New(rand.NewPCG(uint64(e.seed), uint64(e.seed)^0x9E3779B97F4A7C15)),
	}
}

// Tick advances the simulation by one day, mutating state in place. The cash
// delta is drawn deterministically from state.Rand in [MinDailyDeltaCents,
// MaxDailyDeltaCents]. The engine receives the state so future rules can be
// added without expanding the State struct.
func (e *Engine) Tick(state *State) {
	state.Day++
	delta := MinDailyDeltaCents + state.Rand.Int64N(MaxDailyDeltaCents-MinDailyDeltaCents+1)
	state.Cash += delta
}

// AdvanceDays runs the engine's Tick n times on the given state, simulating a
// multi-day cycle in one call. n must be non-negative; a negative n leaves the
// state untouched.
func AdvanceDays(e *Engine, state *State, n int) {
	if n < 0 || state == nil {
		return
	}
	for i := 0; i < n; i++ {
		e.Tick(state)
	}
}
