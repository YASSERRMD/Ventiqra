// Package achievements defines the milestone catalog and the engine that
// evaluates company state to award them. Pure functions over a state snapshot
// so the server and tests share one source of truth.
package achievements

// Key is the stable identifier for an achievement.
type Key string

const (
	KeyFirstLaunch    Key = "first_launch"
	KeyFirstFunding   Key = "first_funding"
	KeyProfitability  Key = "profitability"
	KeyUnicorn        Key = "unicorn"
	KeyHiredTen       Key = "hired_ten"
	KeyCustomerThousand Key = "thousand_customers"
)

// Definition describes an achievement.
type Definition struct {
	Key         Key
	Name        string
	Description string
}

// Catalog is the full set of achievements.
var Catalog = []Definition{
	{KeyFirstLaunch, "First Launch", "Ship your first product."},
	{KeyFirstFunding, "First Funding", "Close your first funding round."},
	{KeyProfitability, "Profitable", "Reach a month where revenue exceeds burn."},
	{KeyUnicorn, "Unicorn", "Reach a $1,000,000,000 valuation."},
	{KeyHiredTen, "Growing Team", "Employ 10 or more people."},
	{KeyCustomerThousand, "Viral", "Reach 1,000 customers."},
}

// State is the company snapshot the engine evaluates against.
type State struct {
	ProductsLaunched int
	FundingRounds    int
	RevenuePerMonth  int64
	MonthlyBurn      int64
	ValuationCents   int64
	Employees        int
	Customers        int
}

// Awarded is the set of achievement keys already awarded (for dedup).
type Awarded map[Key]bool

// Evaluate returns the keys the company newly qualifies for (not already
// awarded). Pure and deterministic.
func Evaluate(state State, already Awarded) []Key {
	var newly []Key
	check := func(k Key, cond bool) {
		if cond && !already[k] {
			newly = append(newly, k)
		}
	}
	check(KeyFirstLaunch, state.ProductsLaunched >= 1)
	check(KeyFirstFunding, state.FundingRounds >= 1)
	check(KeyProfitability, state.RevenuePerMonth > state.MonthlyBurn && state.RevenuePerMonth > 0)
	check(KeyUnicorn, state.ValuationCents >= 1_000_000_000_00)
	check(KeyHiredTen, state.Employees >= 10)
	check(KeyCustomerThousand, state.Customers >= 1000)
	return newly
}

// FindDefinition returns the definition for a key, or false.
func FindDefinition(k Key) (Definition, bool) {
	for _, d := range Catalog {
		if d.Key == k {
			return d, true
		}
	}
	return Definition{}, false
}
