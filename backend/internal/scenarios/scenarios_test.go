package scenarios

import (
	"errors"
	"strings"
	"testing"
)

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

func TestCustomInputValidateAcceptsValid(t *testing.T) {
	in := CustomInput{
		Name: "My Scenario", Difficulty: DifficultyHard, Industry: "Fintech",
		StartingCashCents: 1_000_000_00, StartingBurnCents: 100_000_00,
		MarketTAM: 50_000, MarketGrowthRate: 0.1, MarketTrend: 1.0,
	}
	if err := in.Validate(); err != nil {
		t.Errorf("expected valid input, got %v", err)
	}
}

func TestCustomInputValidateRejectsOutOfRange(t *testing.T) {
	cases := []struct {
		name string
		mod  func(CustomInput) CustomInput
		want error
	}{
		{"empty name", func(c CustomInput) CustomInput { c.Name = ""; return c }, ErrNameLength},
		{"name too long", func(c CustomInput) CustomInput { c.Name = strings.Repeat("x", 81); return c }, ErrNameLength},
		{"bad difficulty", func(c CustomInput) CustomInput { c.Difficulty = "nightmare"; return c }, ErrInvalidDifficulty},
		{"cash too low", func(c CustomInput) CustomInput { c.StartingCashCents = 100; return c }, ErrCashRange},
		{"cash too high", func(c CustomInput) CustomInput { c.StartingCashCents = 100_000_000_00; return c }, ErrCashRange},
		{"burn too low", func(c CustomInput) CustomInput { c.StartingBurnCents = 10; return c }, ErrBurnRange},
		{"tam too low", func(c CustomInput) CustomInput { c.MarketTAM = 100; return c }, ErrTAMRange},
		{"growth too high", func(c CustomInput) CustomInput { c.MarketGrowthRate = 0.9; return c }, ErrGrowthRange},
		{"trend too low", func(c CustomInput) CustomInput { c.MarketTrend = 0.1; return c }, ErrTrendRange},
	}
	base := CustomInput{
		Name: "My Scenario", Difficulty: DifficultyNormal,
		StartingCashCents: 1_000_000_00, StartingBurnCents: 100_000_00,
		MarketTAM: 50_000, MarketGrowthRate: 0.1, MarketTrend: 1.0,
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := c.mod(base)
			err := in.Validate()
			if !errors.Is(err, c.want) {
				t.Errorf("Validate() err = %v, want %v", err, c.want)
			}
		})
	}
}

func TestCustomInputValidateDefaultsEmptyDifficulty(t *testing.T) {
	in := CustomInput{
		Name: "Default Diff", Difficulty: "",
		StartingCashCents: 1_000_000_00, StartingBurnCents: 100_000_00,
		MarketTAM: 50_000, MarketGrowthRate: 0.1, MarketTrend: 1.0,
	}
	if err := in.Validate(); err != nil {
		t.Errorf("empty difficulty should be allowed: %v", err)
	}
}

func TestCustomInputNormalizeClampsAndDefaults(t *testing.T) {
	in := CustomInput{
		Name: "  Clamped  ", Difficulty: "",
		StartingCashCents: 1, StartingBurnCents: 1,
		MarketTAM: 1, MarketGrowthRate: -1, MarketTrend: 99,
	}
	s := in.Normalize("custom-1")
	if s.Name != "Clamped" {
		t.Errorf("name not trimmed: %q", s.Name)
	}
	if s.Difficulty != DifficultyNormal {
		t.Errorf("difficulty = %q, want normal", s.Difficulty)
	}
	if s.Category != "custom" {
		t.Errorf("category = %q, want custom", s.Category)
	}
	if s.StartingCashCents != MinStartingCashCents {
		t.Errorf("cash not clamped: %d", s.StartingCashCents)
	}
	if s.Market.TrendMultiplier != MaxTrend {
		t.Errorf("trend not clamped: %v", s.Market.TrendMultiplier)
	}
}
