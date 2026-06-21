package morale

import "testing"

func TestDailyDecay(t *testing.T) {
	if got := DailyDecay(60); got != -1 {
		t.Errorf("decay above floor = %d, want -1", got)
	}
	if got := DailyDecay(25); got != 0 {
		t.Errorf("decay at floor = %d, want 0", got)
	}
	if got := DailyDecay(10); got != 0 {
		t.Errorf("decay below floor = %d, want 0", got)
	}
}

func TestBoostClamps(t *testing.T) {
	if got := Boost(80, 30); got != 100 {
		t.Errorf("boost over = %d, want 100", got)
	}
	if got := Boost(50, 10); got != 60 {
		t.Errorf("boost = %d, want 60", got)
	}
}

func TestBurnoutFactor(t *testing.T) {
	if got := BurnoutFactor(0); got != 0 {
		t.Errorf("burnout at 0 = %v, want 0", got)
	}
	if got := BurnoutFactor(50); got != 1.0 {
		t.Errorf("healthy burnout factor = %v, want 1.0", got)
	}
	if got := BurnoutFactor(10); got >= 0.31 {
		t.Errorf("burnout factor at 10 = %v, should be small", got)
	}
}

func TestResignsOnlyWhenBurntOut(t *testing.T) {
	// Healthy employees never resign.
	for i := 0; i < 50; i++ {
		if Resigns(50, int64(i), 1, i) {
			t.Errorf("healthy employee resigned")
		}
	}
}

func TestResignsAtZeroMoraleFrequently(t *testing.T) {
	resigned := 0
	for i := 0; i < 100; i++ {
		if Resigns(0, int64(i), 1, i) {
			resigned++
		}
	}
	if resigned < 40 {
		t.Errorf("expected frequent resignations at zero morale, got %d/100", resigned)
	}
}

func TestResignsDeterministic(t *testing.T) {
	a := Resigns(10, 42, 5, 3)
	b := Resigns(10, 42, 5, 3)
	if a != b {
		t.Errorf("resignation not deterministic")
	}
}
