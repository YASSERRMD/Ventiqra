package sim

import (
	"testing"
)

func TestTickAdvancesDayAndCash(t *testing.T) {
	e := NewEngine(42)
	state := e.NewState("company-1", 1_000_00)

	e.Tick(state)
	if state.Day != 1 {
		t.Errorf("Day = %d, want 1", state.Day)
	}
	if state.Cash == 1_000_00 {
		t.Errorf("Cash unchanged after tick; expected a deterministic delta")
	}
	if state.Revenue != 0 {
		t.Errorf("Revenue = %d, want 0 (no products yet)", state.Revenue)
	}
	if state.MonthlyBurn != BaseMonthlyBurnCents {
		t.Errorf("MonthlyBurn = %d, want %d", state.MonthlyBurn, BaseMonthlyBurnCents)
	}
}

func TestNewStateInitializesBurn(t *testing.T) {
	e := NewEngine(3)
	state := e.NewState("co", 0)
	if state.MonthlyBurn != BaseMonthlyBurnCents {
		t.Errorf("MonthlyBurn = %d, want %d", state.MonthlyBurn, BaseMonthlyBurnCents)
	}
	if state.Revenue != 0 {
		t.Errorf("Revenue = %d, want 0", state.Revenue)
	}
}

func TestTickDeterminismSameSeed(t *testing.T) {
	// Two engines with the same seed must produce identical sequences.
	e1 := NewEngine(99)
	e2 := NewEngine(99)
	s1 := e1.NewState("co", 500_00)
	s2 := e2.NewState("co", 500_00)

	for i := 0; i < 50; i++ {
		e1.Tick(s1)
		e2.Tick(s2)
		if s1.Day != s2.Day || s1.Cash != s2.Cash || s1.Revenue != s2.Revenue || s1.MonthlyBurn != s2.MonthlyBurn {
			t.Fatalf("divergence at tick %d: s1={day:%d cash:%d rev:%d burn:%d} s2={day:%d cash:%d rev:%d burn:%d}",
				i+1, s1.Day, s1.Cash, s1.Revenue, s1.MonthlyBurn, s2.Day, s2.Cash, s2.Revenue, s2.MonthlyBurn)
		}
	}
}

func TestTickDifferentSeedDiverges(t *testing.T) {
	a := NewEngine(1)
	b := NewEngine(2)
	sa := a.NewState("co", 0)
	sb := b.NewState("co", 0)

	diverged := false
	for i := 0; i < 20; i++ {
		a.Tick(sa)
		b.Tick(sb)
		if sa.Cash != sb.Cash {
			diverged = true
			break
		}
	}
	if !diverged {
		t.Errorf("different seeds produced identical sequences; expected divergence")
	}
}

func TestTickDeltaInRange(t *testing.T) {
	e := NewEngine(7)
	state := e.NewState("co", 0)
	burn := DailyBurn(state.MonthlyBurn)

	for i := 0; i < 1000; i++ {
		prev := state.Cash
		e.Tick(state)
		// Net cash change is (random delta + revenue(0) - daily burn); recover
		// the random component and assert it stays within the documented range.
		randomDelta := (state.Cash - prev) - state.Revenue + burn
		if randomDelta < MinDailyDeltaCents || randomDelta > MaxDailyDeltaCents {
			t.Fatalf("tick %d random delta = %d, out of [%d, %d]", i+1, randomDelta, MinDailyDeltaCents, MaxDailyDeltaCents)
		}
	}
}

func TestTickAccruesDailyBurn(t *testing.T) {
	// With revenue 0 and a known burn, each tick must deduct exactly the daily
	// burn from the random delta.
	e := NewEngine(11)
	state := e.NewState("co", 0)
	burn := DailyBurn(state.MonthlyBurn)

	// Compare against a twin with no burn to isolate the deduction.
	twin := NewEngine(11).NewState("co", 0)
	twin.MonthlyBurn = 0

	for i := 0; i < 20; i++ {
		beforeGap := state.Cash - twin.Cash
		e.Tick(state)
		e.Tick(twin)
		afterGap := state.Cash - twin.Cash
		if afterGap-beforeGap != -burn {
			t.Fatalf("tick %d burn deduction = %d, want %d", i+1, beforeGap-afterGap, burn)
		}
	}
}

func TestDailyBurnIsDeterministic(t *testing.T) {
	want := BaseMonthlyBurnCents / int64(DaysPerMonth)
	if got := DailyBurn(BaseMonthlyBurnCents); got != want {
		t.Errorf("DailyBurn(%d) = %d, want %d", BaseMonthlyBurnCents, got, want)
	}
	// Same input always yields the same output.
	if DailyBurn(123_456) != DailyBurn(123_456) {
		t.Errorf("DailyBurn not deterministic for identical input")
	}
}

func TestAdvanceDays(t *testing.T) {
	e := NewEngine(123)
	state := e.NewState("co", 0)

	AdvanceDays(e, state, 10)
	if state.Day != 10 {
		t.Errorf("Day = %d, want 10", state.Day)
	}

	// Compare against manual single-tick sequence with a fresh state.
	e2 := NewEngine(123)
	ref := e2.NewState("co", 0)
	for i := 0; i < 10; i++ {
		e2.Tick(ref)
	}
	if state.Day != ref.Day || state.Cash != ref.Cash || state.MonthlyBurn != ref.MonthlyBurn {
		t.Errorf("AdvanceDays mismatch: got {day:%d cash:%d burn:%d}, want {day:%d cash:%d burn:%d}",
			state.Day, state.Cash, state.MonthlyBurn, ref.Day, ref.Cash, ref.MonthlyBurn)
	}
}

func TestAdvanceDaysNegativeNoOp(t *testing.T) {
	e := NewEngine(1)
	state := e.NewState("co", 100)
	AdvanceDays(e, state, -5)
	if state.Day != 0 || state.Cash != 100 {
		t.Errorf("negative n mutated state: day=%d cash=%d", state.Day, state.Cash)
	}
}

func TestAdvanceDaysNilSafe(t *testing.T) {
	e := NewEngine(1)
	AdvanceDays(e, nil, 3) // must not panic
}

func TestNewRandReproducesStream(t *testing.T) {
	// A reloaded Rand at day=N must yield the same next deltas as a continuous
	// run that already performed N ticks.
	e := NewEngine(555)
	continuous := e.NewState("co", 0)
	AdvanceDays(e, continuous, 7)

	reloaded := &State{
		CompanyID:   "co",
		Day:         7,
		Cash:        continuous.Cash,
		Revenue:     0,
		MonthlyBurn: BaseMonthlyBurnCents,
		Seed:        555,
		Rand:        NewRand(555, 7),
	}
	for i := 0; i < 5; i++ {
		e.Tick(continuous)
		e.Tick(reloaded)
		if continuous.Cash != reloaded.Cash {
			t.Fatalf("reloaded stream diverged at step %d: %d != %d", i+1, reloaded.Cash, continuous.Cash)
		}
	}
}

func TestSeedFromCompanyIDStable(t *testing.T) {
	if SeedFromCompanyID("abc") != SeedFromCompanyID("abc") {
		t.Errorf("seed not stable for same id")
	}
	if SeedFromCompanyID("abc") == SeedFromCompanyID("abd") {
		t.Errorf("distinct ids produced identical seed")
	}
}
