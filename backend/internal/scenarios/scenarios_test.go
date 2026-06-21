package scenarios

import "testing"

func TestCatalogHasFourPredefinedScenarios(t *testing.T) {
	want := map[string]bool{
		"bootstrap_saas":      false,
		"vc_funded_startup":   false,
		"hardware_startup":    false,
		"marketplace":         false,
	}
	for _, s := range Catalog {
		if _, ok := want[s.ID]; ok {
			want[s.ID] = true
		}
	}
	for id, found := range want {
		if !found {
			t.Errorf("missing predefined scenario %q", id)
		}
	}
	if len(Catalog) < 4 {
		t.Errorf("catalog has %d scenarios, want at least 4", len(Catalog))
	}
}

func TestFindReturnsScenario(t *testing.T) {
	s, ok := Find("bootstrap_saas")
	if !ok {
		t.Fatal("expected to find bootstrap_saas")
	}
	if s.Name == "" || s.Industry == "" {
		t.Errorf("scenario missing name/industry: %+v", s)
	}
}

func TestFindUnknownReturnsFalse(t *testing.T) {
	if _, ok := Find("nope"); ok {
		t.Error("expected false for unknown scenario id")
	}
}

func TestEachScenarioHasValidConfig(t *testing.T) {
	for _, s := range Catalog {
		if s.ID == "" || s.Name == "" || s.Description == "" {
			t.Errorf("scenario %+v has empty required field", s)
		}
		if !IsValidDifficulty(s.Difficulty) {
			t.Errorf("scenario %q has invalid difficulty %q", s.ID, s.Difficulty)
		}
		if s.StartingCashCents <= 0 {
			t.Errorf("scenario %q: non-positive starting cash %d", s.ID, s.StartingCashCents)
		}
		if s.StartingBurnCents <= 0 {
			t.Errorf("scenario %q: non-positive starting burn %d", s.ID, s.StartingBurnCents)
		}
		if s.Market.TAM <= 0 {
			t.Errorf("scenario %q: non-positive TAM %d", s.ID, s.Market.TAM)
		}
	}
}

func TestDifficultyValidation(t *testing.T) {
	valid := []Difficulty{DifficultyEasy, DifficultyNormal, DifficultyHard, DifficultyBrutal}
	for _, d := range valid {
		if !IsValidDifficulty(d) {
			t.Errorf("difficulty %q should be valid", d)
		}
	}
	if IsValidDifficulty("nightmare") {
		t.Error("unknown difficulty should be invalid")
	}
}

func TestIDsMatchesCatalog(t *testing.T) {
	ids := IDs()
	if len(ids) != len(Catalog) {
		t.Fatalf("IDs() len = %d, want %d", len(ids), len(Catalog))
	}
	for i, s := range Catalog {
		if ids[i] != s.ID {
			t.Errorf("IDs()[%d] = %q, want %q", i, ids[i], s.ID)
		}
	}
}

func TestBootstrapperIsCheapestAndBurnsLeast(t *testing.T) {
	// The bootstrap scenario is the low-capital archetype; it should have the
	// smallest starting cash and the lowest burn of the set, framing its
	// "make every dollar count" identity.
	boot, ok := Find("bootstrap_saas")
	if !ok {
		t.Fatal("missing bootstrap_saas")
	}
	for _, s := range Catalog {
		if s.ID == boot.ID {
			continue
		}
		if s.StartingCashCents < boot.StartingCashCents {
			t.Errorf("scenario %q has less cash (%d) than bootstrapper (%d)", s.ID, s.StartingCashCents, boot.StartingCashCents)
		}
		if s.StartingBurnCents < boot.StartingBurnCents {
			t.Errorf("scenario %q has lower burn (%d) than bootstrapper (%d)", s.ID, s.StartingBurnCents, boot.StartingBurnCents)
		}
	}
}
