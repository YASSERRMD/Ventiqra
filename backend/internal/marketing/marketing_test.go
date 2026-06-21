package marketing

import (
	"math"
	"testing"
)

func TestConversionRateSaturates(t *testing.T) {
	zero := ConversionRate(0)
	small := ConversionRate(100_000)
	large := ConversionRate(10_000_000)
	if zero != 0 {
		t.Errorf("zero budget conversion = %v, want 0", zero)
	}
	if small >= large {
		t.Errorf("conversion should grow with budget: %v vs %v", small, large)
	}
	if large > BaseConversionRate*1.01 {
		t.Errorf("conversion should not exceed base: %v", large)
	}
}

func TestReach(t *testing.T) {
	if got := Reach(1_000_000); got != 10_000 {
		t.Errorf("reach = %d, want 10000", got)
	}
	if Reach(0) != 0 {
		t.Errorf("zero reach should be 0")
	}
}

func TestConversionsScalesWithBudget(t *testing.T) {
	low := Conversions(100_000, 1, 1)
	high := Conversions(5_000_000, 1, 1)
	if high <= low {
		t.Errorf("more budget should yield more conversions: %d vs %d", high, low)
	}
	if Conversions(0, 1, 1) != 0 {
		t.Errorf("zero budget conversions should be 0")
	}
}

func TestConversionsDeterministic(t *testing.T) {
	a := Conversions(500_000, 42, 3)
	b := Conversions(500_000, 42, 3)
	if a != b {
		t.Errorf("conversions not deterministic: %d vs %d", a, b)
	}
}

func TestCAC(t *testing.T) {
	cac := CAC(500_000, 50)
	if cac != 10_000 {
		t.Errorf("CAC = %d, want 10000", cac)
	}
	if CAC(500_000, 0) != 0 {
		t.Errorf("CAC with no conversions should be 0")
	}
}

func TestChannelWeightsSumToOne(t *testing.T) {
	var sum float64
	for _, c := range Channels {
		sum += c.Weight
	}
	if math.Abs(sum-1.0) > 1e-9 {
		t.Errorf("channel weights sum to %v, want 1.0", sum)
	}
}
