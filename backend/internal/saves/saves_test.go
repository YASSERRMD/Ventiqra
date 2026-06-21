package saves

import (
	"strings"
	"testing"
)

func TestIsValidSlot(t *testing.T) {
	cases := []struct {
		slot string
		want bool
	}{
		{"slot1", true},
		{"auto-save", true},
		{"run_a", true},
		{"", false},
		{"Slot1", false},   // uppercase
		{"slot 1", false},  // space
		{"slot.1", false},  // dot
		{strings.Repeat("a", 32), true},
		{strings.Repeat("a", 33), false},
	}
	for _, c := range cases {
		if got := IsValidSlot(c.slot); got != c.want {
			t.Errorf("IsValidSlot(%q) = %v, want %v", c.slot, got, c.want)
		}
	}
}

func TestClampLabel(t *testing.T) {
	if ClampLabel("short") != "short" {
		t.Error("short label should pass through")
	}
	long := strings.Repeat("x", 100)
	got := ClampLabel(long)
	if len([]rune(got)) != MaxLabelLen {
		t.Errorf("clamped len = %d, want %d", len([]rune(got)), MaxLabelLen)
	}
}

func TestSnapshotFieldsRoundTrip(t *testing.T) {
	// A snapshot's fields should be plain JSON-marshallable data; this guards
	// the struct tag spelling used for persistence.
	s := Snapshot{Day: 10, Seed: 42, CashCents: 1_000_00, Revenue: 5_00, MonthlyBurn: 3_000_00, Status: "active", Name: "Co", Industry: "SaaS"}
	if s.Day != 10 || s.Status != "active" || s.Name != "Co" {
		t.Errorf("snapshot fields not as set: %+v", s)
	}
}
