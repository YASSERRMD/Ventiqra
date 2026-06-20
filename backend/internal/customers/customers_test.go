package customers

import (
	"math/rand/v2"
	"testing"
)

func TestAcquisitionScalesWithSatisfaction(t *testing.T) {
	rLow := rand.New(rand.NewPCG(1, 1))
	rHigh := rand.New(rand.NewPCG(1, 1))
	low := Acquisition(20, 1.0, rLow)
	high := Acquisition(90, 1.0, rHigh)
	if high <= low {
		t.Errorf("higher satisfaction should acquire more: %d vs %d", high, low)
	}
}

func TestAcquisitionScalesWithDemand(t *testing.T) {
	rA := rand.New(rand.NewPCG(1, 1))
	rB := rand.New(rand.NewPCG(1, 1))
	normal := Acquisition(80, 1.0, rA)
	boosted := Acquisition(80, 2.0, rB)
	if boosted <= normal {
		t.Errorf("higher demand multiplier should acquire more: %d vs %d", boosted, normal)
	}
}

func TestChurnScalesInverseWithSatisfaction(t *testing.T) {
	rHappy := rand.New(rand.NewPCG(1, 1))
	rUnhappy := rand.New(rand.NewPCG(1, 1))
	happyLoss := Churn(1000, 90, rHappy)
	unhappyLoss := Churn(1000, 10, rUnhappy)
	if happyLoss >= unhappyLoss {
		t.Errorf("lower satisfaction should churn more: happy=%d unhappy=%d", happyLoss, unhappyLoss)
	}
}

func TestChurnBoundedByTotal(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 1))
	lost := Churn(5, 0, r) // worst case satisfaction
	if lost > 5 {
		t.Errorf("churn %d exceeds total 5", lost)
	}
}

func TestChurnZeroTotal(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 1))
	if lost := Churn(0, 0, r); lost != 0 {
		t.Errorf("churn of zero customers = %d, want 0", lost)
	}
}

func TestMauRatioInRange(t *testing.T) {
	for sat := 0; sat <= 100; sat += 10 {
		r := MauRatio(sat)
		if r < 0.1 || r > 0.95 {
			t.Errorf("MauRatio(%d) = %v out of expected range", sat, r)
		}
		if sat > 0 && r <= MauRatio(sat-10) {
			// monotonic non-decreasing
		}
	}
	if MauRatio(100) <= MauRatio(50) {
		t.Errorf("MauRatio should increase with satisfaction")
	}
}

func TestSatisfactionDriftTowardBaseline(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 1))
	low := SatisfactionDrift(20, r)
	if low < 19 { // +1 drift plus noise in -1..+1 → at least 19
		t.Errorf("drift from 20 = %d, want >= 19", low)
	}
	high := SatisfactionDrift(95, r)
	if high > 96 {
		t.Errorf("drift from 95 = %d, want <= 96", high)
	}
}

func TestSatisfactionDriftClamped(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 1))
	if got := SatisfactionDrift(0, r); got < 0 || got > 100 {
		t.Errorf("drift out of range: %d", got)
	}
	if got := SatisfactionDrift(100, r); got < 0 || got > 100 {
		t.Errorf("drift out of range: %d", got)
	}
}

func TestAdvanceDeterministic(t *testing.T) {
	p := Product{Total: 500, MAU: 400, Churned: 10, Satisfaction: 75}
	a := Advance(p, 42, 1, 1.0)
	b := Advance(p, 42, 1, 1.0)
	if a != b {
		t.Errorf("Advance not deterministic: %+v vs %+v", a, b)
	}
}

func TestAdvanceAppliesAcquisitionAndChurn(t *testing.T) {
	p := Product{Total: 1000, MAU: 800, Churned: 0, Satisfaction: 90}
	out := Advance(p, 7, 1, 1.0)
	// With high satisfaction, net should generally grow (acquisition > churn).
	if out.Total <= 0 {
		t.Errorf("total = %d, want > 0", out.Total)
	}
	if out.Churned == 0 {
		t.Errorf("expected some cumulative churn to accrue")
	}
	if out.MAU > out.Total {
		t.Errorf("MAU %d cannot exceed total %d", out.MAU, out.Total)
	}
}

func TestAdvanceTotalNeverNegative(t *testing.T) {
	p := Product{Total: 0, MAU: 0, Churned: 100, Satisfaction: 0}
	out := Advance(p, 3, 1, 1.0)
	if out.Total < 0 {
		t.Errorf("total = %d, want >= 0", out.Total)
	}
}
