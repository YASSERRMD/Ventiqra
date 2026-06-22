package techdebt

import "testing"

func TestQuality(t *testing.T) {
	cases := []struct{ debt, want int }{{0, 100}, {50, 50}, {100, 0}, {120, 0}}
	for _, c := range cases {
		if got := Quality(c.debt); got != c.want {
			t.Errorf("Quality(%d) = %d, want %d", c.debt, got, c.want)
		}
	}
}

func TestOutageRisk(t *testing.T) {
	if OutageRisk(0) != 0 {
		t.Error("debt 0 should have 0 risk")
	}
	if OutageRisk(40) != 0 {
		t.Error("debt 40 should have 0 risk")
	}
	if OutageRisk(70) <= 0 {
		t.Error("debt 70 should have positive risk")
	}
	if OutageRisk(90) != 0.9 {
		t.Errorf("debt 90 risk = %v, want 0.9", OutageRisk(90))
	}
}

func TestAccumulateClamps(t *testing.T) {
	if d := Accumulate(95); d != 100 {
		t.Errorf("accumulate at 95 = %d, want 100", d)
	}
	if d := Accumulate(20); d != 28 {
		t.Errorf("accumulate at 20 = %d, want 28", d)
	}
}

func TestRefactorClamps(t *testing.T) {
	if d := Refactor(20); d != 0 {
		t.Errorf("refactor at 20 = %d, want 0", d)
	}
	if d := Refactor(60); d != 35 {
		t.Errorf("refactor at 60 = %d, want 35", d)
	}
}
