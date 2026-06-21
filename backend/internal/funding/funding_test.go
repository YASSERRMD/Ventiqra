package funding

import (
	"math"
	"testing"
)

func TestNextRoundName(t *testing.T) {
	cases := []struct {
		prior int
		want  string
	}{
		{0, "pre-seed"},
		{1, "seed"},
		{2, "series-a"},
		{10, "growth"},
		{-1, "pre-seed"},
	}
	for _, c := range cases {
		if got := NextRoundName(c.prior); got != c.want {
			t.Errorf("NextRoundName(%d) = %q, want %q", c.prior, got, c.want)
		}
	}
}

func TestPreMoneyValuationUsesRevenueMultiple(t *testing.T) {
	// 50k/mo revenue × 12 × 8 = 4.8M; cash 100k → valuation 4.8M.
	got := PreMoneyValuation(100_000, 50_000)
	want := int64(50_000 * 12 * 8)
	if got != want {
		t.Errorf("valuation = %d, want %d", got, want)
	}
}

func TestPreMoneyValuationCashFloor(t *testing.T) {
	// No revenue, but 1M cash → valuation is the cash.
	if got := PreMoneyValuation(1_000_000, 0); got != 1_000_000 {
		t.Errorf("cash floor valuation = %d, want 1000000", got)
	}
}

func TestEquityPercent(t *testing.T) {
	// Raise 2M on 8M pre → 2/10 = 20%.
	got := EquityPercent(2_000_000, 8_000_000)
	if math.Abs(got-20.0) > 0.01 {
		t.Errorf("equity = %v, want 20", got)
	}
}

func TestEquityPercentZeroAmount(t *testing.T) {
	if got := EquityPercent(0, 1_000_000); got != 0 {
		t.Errorf("zero-amount equity = %v, want 0", got)
	}
}

func TestFounderEquityCompounds(t *testing.T) {
	// Start at 100%, sell 20% → 80%.
	if got := FounderEquity(100, 20); math.Abs(got-80) > 0.01 {
		t.Errorf("founder equity = %v, want 80", got)
	}
	// Another 25% round → 80 * 0.75 = 60.
	if got := FounderEquity(80, 25); math.Abs(got-60) > 0.01 {
		t.Errorf("compounded founder equity = %v, want 60", got)
	}
}

func TestInvestorInterestBounds(t *testing.T) {
	if got := InvestorInterest(0, 0); got < 0 || got > 1 {
		t.Errorf("interest = %v out of [0,1]", got)
	}
	if got := InvestorInterest(0, -1); got >= 0.2 {
		t.Errorf("negative cash should reduce interest: %v", got)
	}
	high := InvestorInterest(2_000_000, 5_000_000)
	low := InvestorInterest(0, 0)
	if high <= low {
		t.Errorf("traction should raise interest: high=%v low=%v", high, low)
	}
}
