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
