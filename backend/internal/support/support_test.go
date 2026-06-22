package support

import "testing"

func TestDailyArrivals(t *testing.T) {
	if DailyArrivals(0) != 0 {
		t.Error("0 customers → 0 tickets")
	}
	if DailyArrivals(1000) != 5 {
		t.Errorf("1000 customers → %d, want 5", DailyArrivals(1000))
	}
	if DailyArrivals(3000) != 15 {
		t.Errorf("3000 customers → %d, want 15", DailyArrivals(3000))
	}
}

func TestResolveForDay(t *testing.T) {
	if ResolveForDay(0) != 0 {
		t.Error("0 agents → 0 resolved")
	}
	if ResolveForDay(3) != 24 {
		t.Errorf("3 agents → %d, want 24", ResolveForDay(3))
	}
}

func TestApplyDay(t *testing.T) {
	// 0 open, 2000 customers (→10 arrivals), 2 agents (→16 resolved) → 0 open, 10 resolved.
	open, resolved, arrivals := ApplyDay(0, 2000, 2)
	if arrivals != 10 {
		t.Errorf("arrivals = %d, want 10", arrivals)
	}
	if resolved != 10 { // only 10 to resolve
		t.Errorf("resolved = %d, want 10", resolved)
	}
	if open != 0 {
		t.Errorf("open = %d, want 0", open)
	}

	// Large backlog, no agents → grows by arrivals.
	open, _, arrivals = ApplyDay(500, 1000, 0)
	if open != 505 || arrivals != 5 {
		t.Errorf("no agents: open=%d arrivals=%d, want 505/5", open, arrivals)
	}
}

func TestSatisfactionPenalty(t *testing.T) {
	if SatisfactionPenaltyPerHundredOpen(50) != 0 {
		t.Error("50 open → 0 penalty")
	}
	if SatisfactionPenaltyPerHundredOpen(250) != 2 {
		t.Errorf("250 open → %d, want 2", SatisfactionPenaltyPerHundredOpen(250))
	}
}
