package market

import "testing"

func TestDailyGrowth(t *testing.T) {
	m := Model{TAM: 100_000, GrowthRate: 0.03}
	// 3% of 100k monthly = 3000/mo → 100/day.
	if got := DailyGrowth(m); got != 100 {
		t.Errorf("daily growth = %d, want 100", got)
	}
}

func TestAdvanceGrowsTAM(t *testing.T) {
	m := DefaultModel
	out := Advance(m, 42, 1)
	if out.TAM <= m.TAM {
		t.Errorf("TAM should grow: %d -> %d", m.TAM, out.TAM)
	}
}

func TestAdvanceTrendBounded(t *testing.T) {
	m := Model{TAM: 100_000, GrowthRate: 0.01, TrendMultiplier: 1.49}
	for i := 0; i < 200; i++ {
		m = Advance(m, 42, int64(i+1))
		if m.TrendMultiplier < 0.5 || m.TrendMultiplier > 1.5 {
			t.Fatalf("trend out of bounds: %v", m.TrendMultiplier)
		}
	}
}

func TestAdvanceDeterministic(t *testing.T) {
	a := Advance(DefaultModel, 7, 3)
	b := Advance(DefaultModel, 7, 3)
	if a != b {
		t.Errorf("Advance not deterministic: %+v vs %+v", a, b)
	}
}

func TestDemandPenetration(t *testing.T) {
	if got := DemandPenetration(0, 100_000); got != 1.0 {
		t.Errorf("empty penetration = %v, want 1", got)
	}
	if got := DemandPenetration(50_000, 100_000); got > 0.49 || got < 0.51 {
		t.Errorf("half penetration = %v, want ~0.5", got)
	}
	if got := DemandPenetration(200_000, 100_000); got != 0 {
		t.Errorf("over-saturated penetration = %v, want 0", got)
	}
}
