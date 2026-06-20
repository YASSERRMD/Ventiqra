package pricing

import "testing"

func TestDemandMultiplierAtBaseline(t *testing.T) {
	if got := DemandMultiplier(1000, 1000, DefaultElasticity); got < 0.99 || got > 1.01 {
		t.Errorf("at baseline = %v, want ~1.0", got)
	}
}

func TestDemandMultiplierDecreasesWithPrice(t *testing.T) {
	cheap := DemandMultiplier(500, 1000, DefaultElasticity)
	pricey := DemandMultiplier(3000, 1000, DefaultElasticity)
	if pricey >= cheap {
		t.Errorf("higher price should lower demand: cheap=%v pricey=%v", cheap, pricey)
	}
	if cheap <= 1.0 {
		t.Errorf("below-baseline price should boost demand: %v", cheap)
	}
	if pricey >= 1.0 {
		t.Errorf("above-baseline price should cut demand: %v", pricey)
	}
}

func TestDemandMultiplierFreeIsMax(t *testing.T) {
	if got := DemandMultiplier(0, 1000, DefaultElasticity); got != MaxDemandMul {
		t.Errorf("free demand = %v, want %v", got, MaxDemandMul)
	}
}

func TestDemandMultiplierClamped(t *testing.T) {
	// An absurdly high price should clamp to MinDemandMul, not underflow.
	if got := DemandMultiplier(10_000_000, 1000, DefaultElasticity); got != MinDemandMul {
		t.Errorf("extreme price demand = %v, want %v", got, MinDemandMul)
	}
}

func TestDailyRevenue(t *testing.T) {
	// 600 customers × $10/mo → $6000/mo → $200/day.
	got := DailyRevenueCents(600, 1000)
	if got != 20_000 {
		t.Errorf("daily revenue = %d, want 20000", got)
	}
}

func TestDailyRevenueZeroCases(t *testing.T) {
	if got := DailyRevenueCents(0, 1000); got != 0 {
		t.Errorf("no customers: %d", got)
	}
	if got := DailyRevenueCents(600, 0); got != 0 {
		t.Errorf("free product: %d", got)
	}
}
