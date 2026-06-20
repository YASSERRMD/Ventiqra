package metrics

import "testing"

func TestHealthBankruptDominates(t *testing.T) {
	if got := Health(-1, 100); got != HealthBankrupt {
		t.Errorf("negative cash with long runway = %q, want bankrupt", got)
	}
}

func TestHealthBands(t *testing.T) {
	cases := []struct {
		cash    int64
		runway  float64
		want    string
	}{
		{1_000_000, InfiniteRunway, HealthHealthy},
		{1_000_000, 12, HealthHealthy},
		{100_000, 5, HealthWarning},
		{10_000, 2, HealthCritical},
		{1_000, 0, HealthCritical},
	}
	for _, c := range cases {
		if got := Health(c.cash, c.runway); got != c.want {
			t.Errorf("Health(cash=%d, runway=%v) = %q, want %q", c.cash, c.runway, got, c.want)
		}
	}
}

func TestHealthCriticalAtOrBelowThreshold(t *testing.T) {
	if got := Health(1, CriticalRunwayMonths); got != HealthCritical {
		t.Errorf("at critical threshold = %q, want critical", got)
	}
	if got := Health(1, WarningRunwayMonths); got != HealthWarning {
		t.Errorf("at warning threshold = %q, want warning", got)
	}
}
