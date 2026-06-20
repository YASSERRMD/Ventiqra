package finance

import "testing"

func TestInfraCostScalesWithCustomers(t *testing.T) {
	zero := InfraCost(0)
	busy := InfraCost(10000)
	if zero != InfraBaseCents {
		t.Errorf("infra floor = %d, want %d", zero, InfraBaseCents)
	}
	want := InfraBaseCents + 10000*InfraPerCustomerCents
	if busy != want {
		t.Errorf("infra(10000) = %d, want %d", busy, want)
	}
}

func TestInfraCostNegativeClamped(t *testing.T) {
	if got := InfraCost(-50); got != InfraBaseCents {
		t.Errorf("infra(-50) = %d, want floor %d", got, InfraBaseCents)
	}
}

func TestMonthlyBreakdownTotals(t *testing.T) {
	b := MonthlyBreakdown(1_000_000, 200_000, 1000)
	if b.Base != BaseMonthlyOperatingCents {
		t.Errorf("base = %d", b.Base)
	}
	if b.Salaries != 1_000_000 {
		t.Errorf("salaries = %d", b.Salaries)
	}
	if b.Marketing != 200_000 {
		t.Errorf("marketing = %d", b.Marketing)
	}
	wantTotal := b.Base + b.Salaries + b.Infra + b.Marketing
	if b.Total() != wantTotal {
		t.Errorf("total = %d, want %d", b.Total(), wantTotal)
	}
}

func TestMonthlyBreakdownClampsNegative(t *testing.T) {
	b := MonthlyBreakdown(-100, -50, 0)
	if b.Salaries != 0 || b.Marketing != 0 {
		t.Errorf("negative inputs not clamped: %+v", b)
	}
}

func TestProfitLoss(t *testing.T) {
	if got := ProfitLoss(1_000_000, 800_000); got != 200_000 {
		t.Errorf("profit = %d, want 200000", got)
	}
	if got := ProfitLoss(500_000, 800_000); got != -300_000 {
		t.Errorf("loss = %d, want -300000", got)
	}
}

func TestMonthlyRevenueFromDaily(t *testing.T) {
	if got := MonthlyRevenueFromDaily(10_000); got != 300_000 {
		t.Errorf("monthly = %d, want 300000", got)
	}
}
