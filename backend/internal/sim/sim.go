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

// BaseMonthlyBurnCents is the small fixed monthly operating cost (in cents)
// every company pays before it has employees or products. It keeps the burn
// non-zero so runway is meaningful from day one while remaining deterministic:
// the same value is applied to every company regardless of seed.
const BaseMonthlyBurnCents int64 = 500_000

// DaysPerMonth is the fixed convention for converting a monthly burn into a
// daily accrual. Using a constant (instead of a calendar) keeps the engine pure
// and its output reproducible.
const DaysPerMonth int = 30

// State holds the mutable simulation state for a single company. The *rand.Rand
// drives the deterministic daily change; callers must not share the same Rand
// across unrelated simulations.
type State struct {
	CompanyID   string
	Day         int
	Cash        int64 // cents
	Revenue     int64 // cents accrued per day (0 until products exist)
	MonthlyBurn int64 // cents per month
	Seed        int64
	Rand        *rand.Rand
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

// pcgStreamSalt is XORed against the seed to derive the second PCG stream so
// that two different callers seeding with related ints still diverge.
const pcgStreamSalt uint64 = 0x9E3779B97F4A7C15

// NewRand creates a deterministic *rand.Rand for the given seed, advanced past
// `day` already-applied ticks. Recreating a Rand with the same seed and day
// reproduces the exact stream position, so reloaded state produces the same
// next delta as a continuous run.
func NewRand(seed int64, day int) *rand.Rand {
	r := rand.New(rand.NewPCG(uint64(seed), uint64(seed)^pcgStreamSalt))
	for i := 0; i < day; i++ {
		_ = r.Int64N(MaxDailyDeltaCents - MinDailyDeltaCents + 1)
	}
	return r
}

// NewState creates a fresh State for a company using the engine's seed. The
// returned State carries its own *rand.Rand so that advancing it does not
// affect other simulations sharing the same engine.
func (e *Engine) NewState(companyID string, cash int64) *State {
	return &State{
		CompanyID:   companyID,
		Day:         0,
		Cash:        cash,
		Revenue:     0,
		MonthlyBurn: BaseMonthlyBurnCents,
		Seed:        e.seed,
		Rand:        NewRand(e.seed, 0),
	}
}

// DailyBurn returns the deterministic daily operating cost (in cents) for the
// given monthly burn. It is derived purely from the monthly figure divided by
// DaysPerMonth (truncated), so the same monthly burn always yields the same
// daily deduction.
func DailyBurn(monthlyBurn int64) int64 { return monthlyBurn / int64(DaysPerMonth) }

// Tick advances the simulation by one day, mutating state in place. The cash
// delta has three deterministic components: a random delta drawn from
// state.Rand in [MinDailyDeltaCents, MaxDailyDeltaCents], the day's accrued
// revenue (state.Revenue), and the daily operating cost derived from
// state.MonthlyBurn. Revenue is 0 until products exist (Phase 9); the burn
// keeps the model realistic while remaining fully reproducible.
func (e *Engine) Tick(state *State) {
	state.Day++
	delta := MinDailyDeltaCents + state.Rand.Int64N(MaxDailyDeltaCents-MinDailyDeltaCents+1)
	state.Cash += delta + state.Revenue - DailyBurn(state.MonthlyBurn)
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
