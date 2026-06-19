package metrics

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}

func TestComputeNoBurnInfiniteRunway(t *testing.T) {
	m := Compute(1_000_00, 0, 0, 0, 0)
	if m.RunwayMonths != InfiniteRunway {
		t.Errorf("runway = %v, want %v (infinite)", m.RunwayMonths, InfiniteRunway)
	}
	// With no revenue, valuation floors to cash.
	if m.ValuationCents != 1_000_00 {
		t.Errorf("valuation = %d, want cash 100000", m.ValuationCents)
	}
	if m.BurnCentsPerMonth != 0 || m.RevenueCents != 0 || m.CashCents != 1_000_00 {
		t.Errorf("unexpected echoed fields: %+v", m)
	}
}

func TestComputeRunwayRatio(t *testing.T) {
	// $10,000 cash (1_000_000 cents) and $1,000/month burn (100_000 cents)
	// yields exactly 10 months of runway.
	m := Compute(1_000_000, 0, 100_000, 0, 0)
	if !almostEqual(m.RunwayMonths, 10.0) {
		t.Errorf("runway = %v, want 10", m.RunwayMonths)
	}
}

func TestComputeNegativeBurnIsInfiniteRunway(t *testing.T) {
	// A non-positive burn means the company cannot run out of cash.
	m := Compute(500_00, 0, -5, 0, 0)
	if m.RunwayMonths != InfiniteRunway {
		t.Errorf("runway = %v, want %v for non-positive burn", m.RunwayMonths, InfiniteRunway)
	}
}

func TestComputeValuationMaxOfCashAndRevenue(t *testing.T) {
	// revenue = $1,000 (100_000 cents). Annualized x12 x4 multiplier =
	// 4_800_000 cents, which exceeds the cash floor of 10_000 cents.
	m := Compute(10_000, 100_000, 100_000, 0, 0)
	wantValuation := int64(100_000) * MonthsPerYear * ValuationMultiplier
	if m.ValuationCents != wantValuation {
		t.Errorf("valuation = %d, want %d", m.ValuationCents, wantValuation)
	}
}

func TestComputeValuationFloorsToCash(t *testing.T) {
	// Revenue is zero, so the annualized value is 0 and valuation floors to
	// cash, which is larger.
	m := Compute(750_00, 0, 50_000, 0, 0)
	if m.ValuationCents != 750_00 {
		t.Errorf("valuation = %d, want cash 75000", m.ValuationCents)
	}
}

func TestComputeIsPure(t *testing.T) {
	// Identical inputs must yield identical outputs across calls.
	a := Compute(123_45, 6_00, 7_00, 0, 9)
	b := Compute(123_45, 6_00, 7_00, 0, 9)
	if a != b {
		t.Errorf("Compute not pure: %+v != %+v", a, b)
	}
}

func TestComputeIgnoresReservedInputsForNow(t *testing.T) {
	// employeesCount and day are reserved for later phases and must not
	// change today's result.
	base := Compute(100_00, 0, 10_00, 0, 0)
	withReserved := Compute(100_00, 0, 10_00, 42, 99)
	if base != withReserved {
		t.Errorf("reserved inputs changed result: %+v != %+v", base, withReserved)
	}
}
