package competitors

import "testing"

func TestGenerateDeterministic(t *testing.T) {
	a := Generate(42)
	b := Generate(42)
	if len(a) != competitorCount {
		t.Fatalf("count = %d, want %d", len(a), competitorCount)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("competitor %d differs: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestGenerateDistinctNames(t *testing.T) {
	comps := Generate(7)
	seen := map[string]bool{}
	for _, c := range comps {
		if seen[c.Name] {
			t.Errorf("duplicate rival name %q", c.Name)
		}
		seen[c.Name] = true
		if c.Strength < 0 || c.Strength > 100 {
			t.Errorf("strength %d out of range", c.Strength)
		}
	}
}

func TestAdvanceStrengthGrows(t *testing.T) {
	c := Competitor{Name: "X", Strength: 20, MarketShare: 0.05}
	advanced := Advance(c, 42, 1)
	if advanced.Strength < 20 {
		t.Errorf("strength should not decrease: %d", advanced.Strength)
	}
}

func TestAdvanceClampsStrength(t *testing.T) {
	c := Competitor{Name: "X", Strength: 99, MarketShare: 0.05}
	for i := 0; i < 50; i++ {
		c = Advance(c, 42, int64(i+1))
		if c.Strength > 100 {
			t.Fatalf("strength exceeded 100: %d", c.Strength)
		}
	}
}

func TestPressureInRange(t *testing.T) {
	comps := Generate(3)
	p := Pressure(comps)
	if p < 0 || p > 0.5 {
		t.Errorf("pressure = %v out of [0,0.5]", p)
	}
	if Pressure(nil) != 0 {
		t.Errorf("empty pressure should be 0")
	}
}

func TestPressureScalesWithStrength(t *testing.T) {
	weak := []Competitor{{Name: "a", Strength: 10}}
	strong := []Competitor{{Name: "a", Strength: 90}}
	if Pressure(strong) <= Pressure(weak) {
		t.Errorf("stronger rivals should exert more pressure")
	}
}
