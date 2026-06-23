package achievements

import "testing"

func TestEvaluateFirstLaunch(t *testing.T) {
	newly := Evaluate(State{ProductsLaunched: 1}, Awarded{})
	if !contains(newly, KeyFirstLaunch) {
		t.Error("expected first_launch awarded")
	}
}

func TestEvaluateDoesNotReaward(t *testing.T) {
	already := Awarded{KeyFirstLaunch: true}
	newly := Evaluate(State{ProductsLaunched: 1}, already)
	if contains(newly, KeyFirstLaunch) {
		t.Error("should not re-award first_launch")
	}
}

func TestEvaluateProfitabilityRequiresPositiveRevenue(t *testing.T) {
	// Zero revenue + zero burn should NOT count as profitable.
	newly := Evaluate(State{RevenuePerMonth: 0, MonthlyBurn: 0}, Awarded{})
	if contains(newly, KeyProfitability) {
		t.Error("zero revenue should not be profitable")
	}
	// Revenue > burn should.
	newly = Evaluate(State{RevenuePerMonth: 100, MonthlyBurn: 50}, Awarded{})
	if !contains(newly, KeyProfitability) {
		t.Error("revenue>burn should be profitable")
	}
}

func TestEvaluateUnicorn(t *testing.T) {
	newly := Evaluate(State{ValuationCents: 999_999_999_00}, Awarded{})
	if contains(newly, KeyUnicorn) {
		t.Error("just under $1B should not be unicorn")
	}
	newly = Evaluate(State{ValuationCents: 1_000_000_000_00}, Awarded{})
	if !contains(newly, KeyUnicorn) {
		t.Error("$1B should be unicorn")
	}
}

func TestEvaluateMultipleAtOnce(t *testing.T) {
	state := State{ProductsLaunched: 1, FundingRounds: 1, Employees: 12, Customers: 1500}
	newly := Evaluate(state, Awarded{})
	if len(newly) < 4 {
		t.Errorf("expected at least 4 awards, got %d: %v", len(newly), newly)
	}
}

func TestFindDefinition(t *testing.T) {
	d, ok := FindDefinition(KeyUnicorn)
	if !ok || d.Name != "Unicorn" {
		t.Errorf("FindDefinition(unicorn) = %+v ok=%v", d, ok)
	}
	if _, ok := FindDefinition("nope"); ok {
		t.Error("unknown key should return false")
	}
}

func contains(keys []Key, want Key) bool {
	for _, k := range keys {
		if k == want {
			return true
		}
	}
	return false
}
