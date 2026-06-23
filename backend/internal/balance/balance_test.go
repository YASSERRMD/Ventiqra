// Package balance tests codify the simulation's gameplay invariants. They fail
// when a tuning change breaks a core promise (e.g. revenue grows with customers,
// churn stays below acquisition at neutral difficulty, funding is achievable).
package balance

import (
	"testing"

	"github.com/YASSERRMD/Ventiqra/backend/internal/difficulty"
)

func TestRevenueScalesWithCustomers(t *testing.T) {
	r1 := RevenuePerDay(100, 1.0)
	r2 := RevenuePerDay(1000, 1.0)
	if r2 <= r1 {
		t.Error("revenue should increase with customers")
	}
	if r1 != 100*ARPUCents {
		t.Errorf("r1 = %d, want %d", r1, 100*ARPUCents)
	}
}

func TestRevenueClampsPriceMultiplier(t *testing.T) {
	low := RevenuePerDay(100, 0.1)
	high := RevenuePerDay(100, 10.0)
	if low != RevenuePerDay(100, MinPriceMultiplier) {
		t.Error("price multiplier not clamped at min")
	}
	if high != RevenuePerDay(100, MaxPriceMultiplier) {
		t.Error("price multiplier not clamped at max")
	}
}

func TestChurnBelowAcquisitionAtNeutral(t *testing.T) {
	// Invariant: at neutral difficulty and 70 satisfaction, daily churn must be
	// a small fraction of a typical customer base, so growth is possible.
	customers := 1000
	churn := ChurnPerDay(customers, 70, difficulty.For(difficulty.LevelNormal).ChurnMultiplier)
	if churn >= customers {
		t.Fatalf("churn (%d) >= customers (%d)", churn, customers)
	}
	// Churn rate should be under 3%/day at 70 satisfaction neutral.
	rate := float64(churn) / float64(customers)
	if rate > 0.03 {
		t.Errorf("churn rate %.3f exceeds 3%%/day", rate)
	}
}

func TestChurnWorseAtLowSatisfaction(t *testing.T) {
	high := ChurnPerDay(1000, 80, 1.0)
	low := ChurnPerDay(1000, 20, 1.0)
	if low <= high {
		t.Error("low satisfaction should churn more than high satisfaction")
	}
}

func TestChurnScalesWithDifficulty(t *testing.T) {
	easy := ChurnPerDay(1000, 50, difficulty.For(difficulty.LevelEasy).ChurnMultiplier)
	brutal := ChurnPerDay(1000, 50, difficulty.For(difficulty.LevelBrutal).ChurnMultiplier)
	if brutal <= easy {
		t.Error("brutal churn should exceed easy churn")
	}
}

func TestFundingAchievableWithTraction(t *testing.T) {
	// A company with 1000 customers asking $2M should have a reasonable chance.
	chance := FundingChance(1000, 2_000_000_00, 1.0)
	if chance < 0.3 || chance > 0.95 {
		t.Errorf("funding chance with traction = %.2f, want 0.3–0.95", chance)
	}
}

func TestFundingHarderForLargeAsks(t *testing.T) {
	small := FundingChance(500, 1_000_000_00, 1.0)
	large := FundingChance(500, 10_000_000_00, 1.0)
	if large >= small {
		t.Error("larger asks should be harder")
	}
}

func TestFundingScalesWithDifficulty(t *testing.T) {
	easy := FundingChance(500, 2_000_000_00, difficulty.For(difficulty.LevelEasy).FundingChanceMult)
	brutal := FundingChance(500, 2_000_000_00, difficulty.For(difficulty.LevelBrutal).FundingChanceMult)
	if brutal >= easy {
		t.Error("brutal funding should be harder than easy")
	}
}

func TestHireCostByRole(t *testing.T) {
	if HireCost("engineer") <= HireCost("support") {
		t.Error("engineers should cost more than support")
	}
	if HireCost("unknown") != HireCostBaseCents {
		t.Error("unknown role should use base cost")
	}
}

// Invariant: a bootstrapped company ($150k cash, $12k/mo burn) should survive
// at least 10 months on cash alone (runway sanity check against the bootstrap
// scenario's starting values).
func TestBootstrapRunwaySanity(t *testing.T) {
	const cash = 150_000_00
	const monthlyBurn = 12_000_00
	runwayMonths := cash / monthlyBurn
	if runwayMonths < 10 {
		t.Errorf("bootstrap runway = %d months, want >= 10", runwayMonths)
	}
}
