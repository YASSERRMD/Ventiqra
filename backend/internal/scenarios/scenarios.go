// Package scenarios models predefined startup scenarios: curated starting
// configurations that shape how a new company begins — its industry, starting
// cash, market size, and difficulty. The catalog is pure data so it can be
// listed and applied deterministically.
package scenarios

// Difficulty labels how punishing a scenario's economy is.
type Difficulty string

const (
	DifficultyEasy    Difficulty = "easy"
	DifficultyNormal  Difficulty = "normal"
	DifficultyHard    Difficulty = "hard"
	DifficultyBrutal  Difficulty = "brutal"
)

// MarketConfig describes the addressable market a scenario drops the company
// into. These are starting values; the market model evolves them during play.
type MarketConfig struct {
	// TAM is the initial total addressable market size (number of potential
	// customers).
	TAM int `json:"tam"`
	// GrowthRate is the annual market growth rate as a fraction (0.05 = 5%/yr).
	GrowthRate float64 `json:"growth_rate"`
	// TrendMultiplier is the starting trend multiplier on demand (1.0 = neutral).
	TrendMultiplier float64 `json:"trend_multiplier"`
}

// Scenario is a complete predefined starting configuration.
type Scenario struct {
	// ID is the stable scenario identifier.
	ID string `json:"id"`
	// Name is the display name.
	Name string `json:"name"`
	// Category groups scenarios by archetype (e.g. "saas", "hardware").
	Category string `json:"category"`
	// Description explains the scenario's flavor and challenge.
	Description string `json:"description"`
	// Difficulty labels the scenario's base difficulty.
	Difficulty Difficulty `json:"difficulty"`
	// Industry is the default industry applied to the company.
	Industry string `json:"industry"`
	// StartingCashCents is the initial cash balance (in cents).
	StartingCashCents int64 `json:"starting_cash_cents"`
	// StartingBurnCents is the initial monthly burn (in cents) before revenue.
	StartingBurnCents int64 `json:"starting_burn_cents"`
	// Market is the starting market configuration.
	Market MarketConfig `json:"market"`
}

// Catalog is the pool of predefined scenarios shipped with the game.
var Catalog = []Scenario{
	{
		ID:   "bootstrap_saas",
		Name: "Bootstrap SaaS",
		Category: "saas",
		Description: "A scrappy solo founder building a SaaS tool on savings. Low capital, low burn, slow but steady growth. Make every dollar count.",
		Difficulty: DifficultyNormal,
		Industry:   "SaaS",
		StartingCashCents: 150_000_00,   // $150,000
		StartingBurnCents: 12_000_00,    // $12,000/mo
		Market: MarketConfig{
			TAM: 120_000, GrowthRate: 0.08, TrendMultiplier: 1.05,
		},
	},
	{
		ID:   "vc_funded_startup",
		Name: "VC-Funded Startup",
		Category: "saas",
		Description: "Fresh out of a seed round with money to burn. High capital and high burn — grow fast or run out of runway before the next raise.",
		Difficulty: DifficultyHard,
		Industry:   "SaaS",
		StartingCashCents: 3_000_000_00, // $3,000,000
		StartingBurnCents: 250_000_00,   // $250,000/mo
		Market: MarketConfig{
			TAM: 500_000, GrowthRate: 0.15, TrendMultiplier: 1.2,
		},
	},
	{
		ID:   "hardware_startup",
		Name: "Hardware Startup",
		Category: "hardware",
		Description: "Building physical products means heavy upfront capital, long development cycles, and supply-chain risk. High reward, higher stakes.",
		Difficulty: DifficultyHard,
		Industry:   "Hardware",
		StartingCashCents: 1_500_000_00, // $1,500,000
		StartingBurnCents: 180_000_00,   // $180,000/mo
		Market: MarketConfig{
			TAM: 80_000, GrowthRate: 0.06, TrendMultiplier: 0.95,
		},
	},
	{
		ID:   "marketplace",
		Name: "Marketplace",
		Category: "marketplace",
		Description: "A two-sided marketplace that must solve the cold-start problem. Capital funds subsidies to attract both sides; the network effect is the prize.",
		Difficulty: DifficultyBrutal,
		Industry:   "Marketplace",
		StartingCashCents: 2_000_000_00, // $2,000,000
		StartingBurnCents: 200_000_00,   // $200,000/mo
		Market: MarketConfig{
			TAM: 250_000, GrowthRate: 0.12, TrendMultiplier: 1.1,
		},
	},
}

// Find returns the scenario with the given id, or false.
func Find(id string) (Scenario, bool) {
	for _, s := range Catalog {
		if s.ID == id {
			return s, true
		}
	}
	return Scenario{}, false
}

// IDs returns the stable ids of all catalog scenarios.
func IDs() []string {
	out := make([]string, 0, len(Catalog))
	for _, s := range Catalog {
		out = append(out, s.ID)
	}
	return out
}

// IsValidDifficulty reports whether d is a recognized difficulty label.
func IsValidDifficulty(d Difficulty) bool {
	switch d {
	case DifficultyEasy, DifficultyNormal, DifficultyHard, DifficultyBrutal:
		return true
	}
	return false
}

// Limits bound the editable ranges for custom-scenario fields, keeping saved
// scenarios playable (no infinite cash, no negative markets).
const (
	MinNameLen         = 1
	MaxNameLen         = 80
	MaxDescriptionLen  = 1000
	MaxIndustryLen     = 60

	MinStartingCashCents int64 = 10_000_00      // $1,000
	MaxStartingCashCents int64 = 50_000_000_00  // $50,000,000

	MinStartingBurnCents int64 = 1_000_00       // $100/mo
	MaxStartingBurnCents int64 = 5_000_000_00   // $5,000,000/mo

	MinTAM = 1_000
	MaxTAM = 10_000_000

	MinGrowthRate = 0.0
	MaxGrowthRate = 0.5 // 50%/yr

	MinTrend = 0.5
	MaxTrend = 2.0
)

// CustomInput holds the user-supplied fields for a custom scenario, before
// validation/clamping. Numeric fields are pointers so "omitted" is distinct
// from zero.
type CustomInput struct {
	Name              string
	Description       string
	Difficulty        Difficulty
	Industry          string
	StartingCashCents int64
	StartingBurnCents int64
	MarketTAM         int
	MarketGrowthRate  float64
	MarketTrend       float64
}

// Validate returns an error if the custom-scenario input is out of range or
// missing required fields. Difficulty defaults to Normal when empty.
func (c CustomInput) Validate() error {
	c.Name = trim(c.Name)
	if len(c.Name) < MinNameLen || len(c.Name) > MaxNameLen {
		return ErrNameLength
	}
	if len(c.Description) > MaxDescriptionLen {
		return ErrDescriptionTooLong
	}
	if len(c.Industry) > MaxIndustryLen {
		return ErrIndustryTooLong
	}
	if c.Difficulty == "" {
		c.Difficulty = DifficultyNormal
	}
	if !IsValidDifficulty(c.Difficulty) {
		return ErrInvalidDifficulty
	}
	if c.StartingCashCents < MinStartingCashCents || c.StartingCashCents > MaxStartingCashCents {
		return ErrCashRange
	}
	if c.StartingBurnCents < MinStartingBurnCents || c.StartingBurnCents > MaxStartingBurnCents {
		return ErrBurnRange
	}
	if c.MarketTAM < MinTAM || c.MarketTAM > MaxTAM {
		return ErrTAMRange
	}
	if c.MarketGrowthRate < MinGrowthRate || c.MarketGrowthRate > MaxGrowthRate {
		return ErrGrowthRange
	}
	if c.MarketTrend < MinTrend || c.MarketTrend > MaxTrend {
		return ErrTrendRange
	}
	return nil
}

// Normalize clamps a CustomInput into safe, defaulted values and returns the
// resulting Scenario. It assumes Validate() has passed (or clamps permissively
// when called directly). Difficulty defaults to Normal.
func (c CustomInput) Normalize(id string) Scenario {
	if c.Difficulty == "" || !IsValidDifficulty(c.Difficulty) {
		c.Difficulty = DifficultyNormal
	}
	return Scenario{
		ID: id, Name: trim(c.Name), Category: "custom", Description: trim(c.Description),
		Difficulty: c.Difficulty, Industry: trim(c.Industry),
		StartingCashCents: clampCash(c.StartingCashCents),
		StartingBurnCents: clampBurn(c.StartingBurnCents),
		Market: MarketConfig{
			TAM:             clampTAM(c.MarketTAM),
			GrowthRate:      clampGrowth(c.MarketGrowthRate),
			TrendMultiplier: clampTrend(c.MarketTrend),
		},
	}
}

// Sentinel validation errors. They carry no state so callers can compare with
// errors.Is.
var (
	ErrNameLength       = validationError("name must be 1–80 characters")
	ErrDescriptionTooLong = validationError("description must be 1000 characters or fewer")
	ErrIndustryTooLong  = validationError("industry must be 60 characters or fewer")
	ErrInvalidDifficulty = validationError("difficulty must be easy, normal, hard, or brutal")
	ErrCashRange        = validationError("starting cash must be between $1,000 and $50,000,000")
	ErrBurnRange        = validationError("starting burn must be between $100 and $5,000,000 per month")
	ErrTAMRange         = validationError("market TAM must be between 1,000 and 10,000,000")
	ErrGrowthRange      = validationError("market growth rate must be between 0% and 50%")
	ErrTrendRange       = validationError("market trend must be between 0.5 and 2.0")
)

type validationError string

func (e validationError) Error() string { return string(e) }

func trim(s string) string {
	// Trim leading/trailing whitespace without importing unicode (ASCII space
	// and tabs suffice for form input).
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n') {
		s = s[1:]
	}
	for len(s) > 0 {
		last := s[len(s)-1]
		if last == ' ' || last == '\t' || last == '\n' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}

func clampCash(v int64) int64 {
	if v < MinStartingCashCents {
		return MinStartingCashCents
	}
	if v > MaxStartingCashCents {
		return MaxStartingCashCents
	}
	return v
}

func clampBurn(v int64) int64 {
	if v < MinStartingBurnCents {
		return MinStartingBurnCents
	}
	if v > MaxStartingBurnCents {
		return MaxStartingBurnCents
	}
	return v
}

func clampTAM(v int) int {
	if v < MinTAM {
		return MinTAM
	}
	if v > MaxTAM {
		return MaxTAM
	}
	return v
}

func clampGrowth(v float64) float64 {
	if v < MinGrowthRate {
		return MinGrowthRate
	}
	if v > MaxGrowthRate {
		return MaxGrowthRate
	}
	return v
}

func clampTrend(v float64) float64 {
	if v < MinTrend {
		return MinTrend
	}
	if v > MaxTrend {
		return MaxTrend
	}
	return v
}
