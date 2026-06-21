package reputation

import "testing"

func TestGrowthMultiplier(t *testing.T) {
	if got := GrowthMultiplier(50); got != 1.0 {
		t.Errorf("neutral multiplier = %v, want 1.0", got)
	}
	if got := GrowthMultiplier(100); got < 1.29 || got > 1.31 {
		t.Errorf("top multiplier = %v, want ~1.3", got)
	}
	if got := GrowthMultiplier(0); got < 0.69 || got > 0.71 {
		t.Errorf("bottom multiplier = %v, want ~0.7", got)
	}
	if high, low := GrowthMultiplier(90), GrowthMultiplier(10); high <= low {
		t.Errorf("higher reputation should boost growth more: %v vs %v", high, low)
	}
}

func TestGrowthMultiplierClamps(t *testing.T) {
	if got := GrowthMultiplier(200); got < 1.29 || got > 1.31 {
		t.Errorf("over-range multiplier = %v, want ~1.3", got)
	}
	if got := GrowthMultiplier(-50); got < 0.69 || got > 0.71 {
		t.Errorf("under-range multiplier = %v, want ~0.7", got)
	}
}

func TestSatisfactionDrift(t *testing.T) {
	cases := map[int]int{
		80:  1,
		75:  1,
		60:  0,
		40:  0,
		30:  -1,
		0:   -1,
	}
	for sat, want := range cases {
		if got := SatisfactionDrift(sat); got != want {
			t.Errorf("SatisfactionDrift(%d) = %d, want %d", sat, got, want)
		}
	}
}

func TestHealthDelta(t *testing.T) {
	if got := HealthDelta("bankrupt"); got != -5 {
		t.Errorf("bankrupt delta = %d, want -5", got)
	}
	if got := HealthDelta("critical"); got != -1 {
		t.Errorf("critical delta = %d, want -1", got)
	}
	if got := HealthDelta("healthy"); got != 0 {
		t.Errorf("healthy delta = %d, want 0", got)
	}
}

func TestClamp(t *testing.T) {
	if Clamp(-5) != 0 || Clamp(150) != 100 || Clamp(42) != 42 {
		t.Errorf("clamp broken")
	}
}
