package launch

import "testing"

func TestReadinessDominantFactor(t *testing.T) {
	// High dev progress, no team: readiness still dominated by dev (0.6*100 = 60).
	got := Readiness(Inputs{DevProgress: 100, AvgSkill: 0, TeamSize: 0})
	if got < 59.9 || got > 60.1 {
		t.Errorf("readiness = %v, want ~60", got)
	}
	if !CanLaunch(got) {
		t.Errorf("60 should clear the launch threshold")
	}
}

func TestReadinessBelowThreshold(t *testing.T) {
	got := Readiness(Inputs{DevProgress: 10, AvgSkill: 0, TeamSize: 0})
	if got > MinReadiness {
		t.Errorf("readiness = %v, want below %v", got, MinReadiness)
	}
	if CanLaunch(got) {
		t.Errorf("low readiness should not allow launch")
	}
}

func TestReadinessCombinesInputs(t *testing.T) {
	got := Readiness(Inputs{DevProgress: 50, AvgSkill: 80, TeamSize: 5})
	// 0.6*50 + 0.3*80 + 0.1*5*10 = 30 + 24 + 5 = 59
	if got < 58.9 || got > 59.1 {
		t.Errorf("readiness = %v, want ~59", got)
	}
}

func TestReadinessClampsTo100(t *testing.T) {
	got := Readiness(Inputs{DevProgress: 200, AvgSkill: 200, TeamSize: 999})
	if got != 100 {
		t.Errorf("readiness = %v, want 100 (clamped)", got)
	}
}

func TestReadinessTeamSizeCappedAt10(t *testing.T) {
	small := Readiness(Inputs{DevProgress: 0, AvgSkill: 0, TeamSize: 10})
	large := Readiness(Inputs{DevProgress: 0, AvgSkill: 0, TeamSize: 100})
	if small != large {
		t.Errorf("team size bonus should cap at 10: %v vs %v", small, large)
	}
}

func TestInitialCustomersScalesWithReadiness(t *testing.T) {
	low := InitialCustomers(20, 1)
	high := InitialCustomers(90, 1)
	if high <= low {
		t.Errorf("higher readiness should yield more customers: %d vs %d", high, low)
	}
}

func TestInitialCustomersDeterministic(t *testing.T) {
	a := InitialCustomers(75, 42)
	b := InitialCustomers(75, 42)
	if a != b {
		t.Errorf("initial customers not deterministic: %d vs %d", a, b)
	}
}

func TestInitialCustomersNonNegative(t *testing.T) {
	if got := InitialCustomers(0, 1); got < 0 {
		t.Errorf("customers = %d, want >= 0", got)
	}
}
