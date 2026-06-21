package timeline

import "testing"

func TestMonthForDay(t *testing.T) {
	cases := []struct {
		day, want int
	}{
		{0, 0}, {1, 1}, {30, 1}, {31, 2}, {60, 2}, {61, 3}, {90, 3},
	}
	for _, c := range cases {
		if got := MonthForDay(c.day); got != c.want {
			t.Errorf("MonthForDay(%d) = %d, want %d", c.day, got, c.want)
		}
	}
}

func TestMonthBounds(t *testing.T) {
	cases := []struct {
		month, start, end int
	}{
		{1, 1, 30}, {2, 31, 60}, {3, 61, 90}, {0, 0, 0},
	}
	for _, c := range cases {
		s, e := MonthBounds(c.month)
		if s != c.start || e != c.end {
			t.Errorf("MonthBounds(%d) = (%d,%d), want (%d,%d)", c.month, s, e, c.start, c.end)
		}
	}
}

func TestIsInMonth(t *testing.T) {
	if !IsInMonth(15, 1) {
		t.Error("day 15 should be in month 1")
	}
	if IsInMonth(31, 1) {
		t.Error("day 31 should not be in month 1")
	}
	if !IsInMonth(45, 2) {
		t.Error("day 45 should be in month 2")
	}
}

func TestSummarizeMonthsProducesOnePerMonth(t *testing.T) {
	deltas := []int64{-50_000_00, 10_000_00, 25_000_00}
	sums := SummarizeMonths(75, 5_000_00, 80_000_00, deltas) // day 75 = month 3
	if len(sums) != 3 {
		t.Fatalf("got %d summaries, want 3", len(sums))
	}
	if sums[0].Month != 1 || sums[2].Month != 3 {
		t.Errorf("months = %d,%d,%d", sums[0].Month, sums[1].Month, sums[2].Month)
	}
	if sums[0].CashChange != -50_000_00 {
		t.Errorf("month 1 delta = %d, want -5000000", sums[0].CashChange)
	}
	if sums[2].StartDay != 61 || sums[2].EndDay != 90 {
		t.Errorf("month 3 bounds = (%d,%d)", sums[2].StartDay, sums[2].EndDay)
	}
}

func TestSummarizeMonthsHandlesZeroDay(t *testing.T) {
	sums := SummarizeMonths(0, 0, 0, nil)
	if len(sums) != 1 {
		t.Fatalf("got %d summaries for day 0, want 1", len(sums))
	}
}
