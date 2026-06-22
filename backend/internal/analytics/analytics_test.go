package analytics

import "testing"

func TestValuation(t *testing.T) {
	cases := []struct {
		cash, revenue, want int64
	}{
		{1_000_000_00, 0, 1_000_000_00},                         // cash floor
		{100_000_00, 100_000_00, 9_600_000_00},                  // revenue-backed (100k*12*8)
		{50_000_000_00, 10_000_00, 50_000_000_00},               // cash dominates
	}
	for _, c := range cases {
		if got := Valuation(c.cash, c.revenue); got != c.want {
			t.Errorf("Valuation(%d,%d) = %d, want %d", c.cash, c.revenue, got, c.want)
		}
	}
}

func TestFromSnapshotsCarriesDay(t *testing.T) {
	points := []Point{{Day: 1, CashCents: 100}, {Day: 2, CashCents: 90}}
	s := FromSnapshots(points, 2)
	if s.Day != 2 {
		t.Errorf("Day = %d, want 2", s.Day)
	}
	if len(s.Cash) != 2 {
		t.Errorf("Cash series len = %d, want 2", len(s.Cash))
	}
}

func TestLatestEmpty(t *testing.T) {
	if (Latest(nil) != Point{}) {
		t.Error("Latest of empty should be zero Point")
	}
}

func TestLatestReturnsLast(t *testing.T) {
	points := []Point{{Day: 1}, {Day: 5}, {Day: 9}}
	if Latest(points).Day != 9 {
		t.Errorf("Latest day = %d, want 9", Latest(points).Day)
	}
}
