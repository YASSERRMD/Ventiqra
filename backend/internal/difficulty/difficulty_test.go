package difficulty

import "testing"

func TestIsValid(t *testing.T) {
	for _, l := range []Level{LevelEasy, LevelNormal, LevelHard, LevelBrutal, LevelCustom} {
		if !IsValid(l) {
			t.Errorf("%q should be valid", l)
		}
	}
	if IsValid("nightmare") {
		t.Error("nightmare should be invalid")
	}
}

func TestPresetsExist(t *testing.T) {
	for _, l := range []Level{LevelEasy, LevelNormal, LevelHard, LevelBrutal} {
		if _, ok := Presets[l]; !ok {
			t.Errorf("missing preset for %q", l)
		}
	}
}

func TestForDefaultsToNormal(t *testing.T) {
	m := For("unknown")
	if m != Presets[LevelNormal] {
		t.Error("unknown level should fall back to normal")
	}
}

func TestEasyIsEasierThanBrutal(t *testing.T) {
	easy := For(LevelEasy)
	brutal := For(LevelBrutal)
	if easy.BurnMultiplier >= brutal.BurnMultiplier {
		t.Error("easy burn should be lower than brutal")
	}
	if easy.StartingCashMult <= brutal.StartingCashMult {
		t.Error("easy cash should be higher than brutal")
	}
	if easy.AcquisitionRateMult <= brutal.AcquisitionRateMult {
		t.Error("easy acquisition should be higher than brutal")
	}
}

func TestNormalIsNeutral(t *testing.T) {
	m := For(LevelNormal)
	if m.BurnMultiplier != 1.0 || m.ChurnMultiplier != 1.0 || m.FundingChanceMult != 1.0 {
		t.Error("normal should be all 1.0 multipliers")
	}
}
