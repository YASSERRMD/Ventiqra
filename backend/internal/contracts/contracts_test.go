package contracts

import "testing"

func TestAnnualValueForDiscount(t *testing.T) {
	if AnnualValueFor(1_000_000_00, 0) != 1_000_000_00 {
		t.Error("0% discount should be full value")
	}
	if v := AnnualValueFor(1_000_000_00, 20); v != 800_000_00 {
		t.Errorf("20%% discount = %d, want 8000000", v)
	}
	if v := AnnualValueFor(1_000_000_00, 80); v != 500_000_00 {
		t.Errorf("80%% (over cap) = %d, want 5000000 (50%%)", v)
	}
}

func TestDailyRevenue(t *testing.T) {
	if DailyRevenue(360_000_00) != 1_000_00 {
		t.Errorf("360k annual → %d daily, want 1000", DailyRevenue(360_000_00))
	}
	if DailyRevenue(0) != 0 {
		t.Error("zero annual → zero daily")
	}
}

func TestTermForYears(t *testing.T) {
	if TermForYears(1) != DaysPerYear {
		t.Error("1 year term wrong")
	}
	if TermForYears(3) != 3*DaysPerYear {
		t.Error("3 year term wrong")
	}
	if TermForYears(0) != DaysPerYear {
		t.Error("0 years should clamp to 1")
	}
}

func TestRenewalRollDistribution(t *testing.T) {
	// At satisfaction 50 (base 70%), roughly 70% should renew.
	renewed := 0
	const trials = 4000
	for i := int64(0); i < trials; i++ {
		if RenewalRoll(i, 50) {
			renewed++
		}
	}
	frac := float64(renewed) / trials
	if frac < 0.63 || frac > 0.77 {
		t.Errorf("renewal fraction at sat 50 = %.3f, want ~0.7", frac)
	}
}

func TestRenewalRollSatisfactionBoost(t *testing.T) {
	// High satisfaction should renew more often than low.
	high := 0
	low := 0
	const trials = 4000
	for i := int64(0); i < trials; i++ {
		if RenewalRoll(i, 90) {
			high++
		}
		if RenewalRoll(i, 20) {
			low++
		}
	}
	if high <= low {
		t.Errorf("expected high sat to renew more: high=%d low=%d", high, low)
	}
}
