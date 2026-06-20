package develop

import (
	"math"
	"testing"
)

func TestDailyProgressSingleBaselineEngineer(t *testing.T) {
	// Skill 100, morale 70 → exactly the base per-engineer daily rate.
	got := DailyProgress([]Employee{{Role: "engineer", Skill: 100, Morale: 70}})
	if !approx(got, BaseDailyProgressPerEngineer) {
		t.Errorf("got %v, want %v", got, BaseDailyProgressPerEngineer)
	}
}

func TestDailyProgressScalesWithSkill(t *testing.T) {
	low := DailyProgress([]Employee{{Role: "engineer", Skill: 25, Morale: 70}})
	high := DailyProgress([]Employee{{Role: "engineer", Skill: 100, Morale: 70}})
	if low >= high {
		t.Errorf("expected higher skill to yield more progress: %v vs %v", low, high)
	}
	if !approx(low, high*0.25) {
		t.Errorf("skill scaling: low=%v want %v (0.25*%v)", low, high*0.25, high)
	}
}

func TestDailyProgressScalesWithMorale(t *testing.T) {
	low := DailyProgress([]Employee{{Role: "engineer", Skill: 100, Morale: 35}})
	high := DailyProgress([]Employee{{Role: "engineer", Skill: 100, Morale: 100}})
	if low >= high {
		t.Errorf("expected higher morale to yield more progress: %v vs %v", low, high)
	}
	// morale 35 → 0.5x of the morale-70 baseline; morale 100 → ~1.43x.
	if !approx(low, high*(35.0/100.0)) {
		t.Errorf("morale scaling mismatch: low=%v high=%v", low, high)
	}
}

func TestDailyProgressRoleWeights(t *testing.T) {
	engineer := DailyProgress([]Employee{{Role: "engineer", Skill: 100, Morale: 70}})
	designer := DailyProgress([]Employee{{Role: "designer", Skill: 100, Morale: 70}})
	ops := DailyProgress([]Employee{{Role: "operations", Skill: 100, Morale: 70}})
	sales := DailyProgress([]Employee{{Role: "sales", Skill: 100, Morale: 100}})

	if !approx(designer, engineer*0.6) {
		t.Errorf("designer weight: %v want %v", designer, engineer*0.6)
	}
	if !approx(ops, engineer*0.2) {
		t.Errorf("operations weight: %v want %v", ops, engineer*0.2)
	}
	if sales != 0 {
		t.Errorf("non-building roles should add no progress, got %v", sales)
	}
}

func TestDailyProgressSumsTeam(t *testing.T) {
	team := []Employee{
		{Role: "engineer", Skill: 100, Morale: 70},
		{Role: "engineer", Skill: 100, Morale: 70},
		{Role: "designer", Skill: 100, Morale: 70},
	}
	got := DailyProgress(team)
	want := 2*BaseDailyProgressPerEngineer + BaseDailyProgressPerEngineer*0.6
	if !approx(got, want) {
		t.Errorf("team progress = %v, want %v", got, want)
	}
}

func TestDailyProgressEmptyIsZero(t *testing.T) {
	if got := DailyProgress(nil); got != 0 {
		t.Errorf("empty roster progress = %v, want 0", got)
	}
}

func TestDailyProgressClampsOutOfRange(t *testing.T) {
	// Skill/morale above 100 or below 0 should be clamped, not panic or explode.
	got := DailyProgress([]Employee{{Role: "engineer", Skill: 9999, Morale: 9999}})
	max := DailyProgress([]Employee{{Role: "engineer", Skill: 100, Morale: 100}})
	if !approx(got, max) {
		t.Errorf("over-range not clamped: %v vs %v", got, max)
	}

	neg := DailyProgress([]Employee{{Role: "engineer", Skill: -50, Morale: -50}})
	if neg != 0 {
		t.Errorf("under-range progress = %v, want 0", neg)
	}
}

func TestMonthlyPayroll(t *testing.T) {
	got := MonthlyPayroll([]int64{12_000_00, 10_000_00, 7_000_00})
	if got != 29_000_00 {
		t.Errorf("payroll = %d, want 2900000", got)
	}
}

func TestMonthlyPayrollIgnoresNonPositive(t *testing.T) {
	got := MonthlyPayroll([]int64{12_000_00, 0, -5, 8_000_00})
	if got != 20_000_00 {
		t.Errorf("payroll = %d, want 2000000", got)
	}
}

func approx(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
