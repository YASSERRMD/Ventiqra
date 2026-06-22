package infrastructure

import "testing"

func TestCapacityForTier(t *testing.T) {
	if CapacityForTier(1) != BaseCapacity {
		t.Errorf("tier 1 capacity = %d", CapacityForTier(1))
	}
	if CapacityForTier(3) != BaseCapacity+2*CapacityPerTier {
		t.Errorf("tier 3 capacity = %d", CapacityForTier(3))
	}
	if CapacityForTier(0) != BaseCapacity { // clamped
		t.Errorf("tier 0 should clamp to tier 1")
	}
}

func TestHostingCostForTier(t *testing.T) {
	if HostingCostForTier(1) != BaseHostingCostCents {
		t.Errorf("tier 1 cost = %d", HostingCostForTier(1))
	}
	if HostingCostForTier(2) != BaseHostingCostCents+CostPerTierCents {
		t.Errorf("tier 2 cost = %d", HostingCostForTier(2))
	}
}

func TestLoadRatio(t *testing.T) {
	if LoadRatio(500, 1000) != 0.5 {
		t.Errorf("load 500/1000 = %v", LoadRatio(500, 1000))
	}
	if LoadRatio(1000, 0) != 1 {
		t.Error("zero capacity should be full load")
	}
}

func TestOutageRisk(t *testing.T) {
	if OutageRisk(0.5) != 0.02 {
		t.Errorf("low load risk = %v", OutageRisk(0.5))
	}
	if OutageRisk(1.0) <= 0.02 {
		t.Error("full load should have elevated risk")
	}
	if OutageRisk(1.5) != 0.8 {
		t.Errorf("overload risk = %v, want 0.8", OutageRisk(1.5))
	}
}

func TestRecommendedTier(t *testing.T) {
	if RecommendedTier(500) != 1 {
		t.Errorf("500 customers → tier %d, want 1", RecommendedTier(500))
	}
	if RecommendedTier(50000) != MaxTier {
		t.Errorf("50000 customers → tier %d, want max", RecommendedTier(50000))
	}
}
