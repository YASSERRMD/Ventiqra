// Package events implements the random event engine: a deterministic catalog of
// positive, negative, and neutral events, and a daily roll that decides whether
// one fires. Effects are described declaratively so the server can apply them.
package events

import "math/rand/v2"

// Kind classifies an event's valence.
type Kind string

const (
	Positive Kind = "positive"
	Negative Kind = "negative"
	Neutral  Kind = "neutral"
)

// Event is a single declarative event the engine can fire.
type Event struct {
	Kind            Kind
	Title           string
	Description     string
	CashDelta       int64 // cents, applied to company cash
	ReputationDelta int   // applied to reputation score
	MoraleDelta     int   // applied to team morale
	Weight          int   // relative selection weight
}

// Catalog is the pool of events the engine draws from.
var Catalog = []Event{
	// Positive.
	{Kind: Positive, Title: "Viral moment", Description: "A launch video went viral, driving a wave of goodwill.", CashDelta: 200_000, ReputationDelta: 4, MoraleDelta: 5, Weight: 3},
	{Kind: Positive, Title: "Industry award", Description: "Your product won a respected industry award.", ReputationDelta: 3, MoraleDelta: 6, Weight: 3},
	{Kind: Positive, Title: "Great press", Description: "A major publication ran a glowing profile.", ReputationDelta: 3, MoraleDelta: 2, Weight: 4},
	{Kind: Positive, Title: "Partner deal", Description: "A strategic partner prepaid for a year.", CashDelta: 150_000, ReputationDelta: 1, Weight: 3},
	// Negative.
	{Kind: Negative, Title: "Supplier price hike", Description: "A key supplier raised prices unexpectedly.", CashDelta: -120_000, Weight: 4},
	{Kind: Negative, Title: "Negative review", Description: "A influential reviewer panned the product.", ReputationDelta: -3, MoraleDelta: -3, Weight: 4},
	{Kind: Negative, Title: "Bug in production", Description: "A production bug frustrated customers.", CashDelta: -40_000, ReputationDelta: -2, MoraleDelta: -4, Weight: 3},
	{Kind: Negative, Title: "Key hire counter-offered", Description: "A competitor tried to poach staff.", MoraleDelta: -5, CashDelta: -50_000, Weight: 2},
	// Neutral.
	{Kind: Neutral, Title: "Competitor IPO", Description: "A rival went public, drawing attention to the market.", Weight: 3},
	{Kind: Neutral, Title: "Regulatory rumor", Description: "A rumored regulation is making the rounds; nothing firm yet.", Weight: 3},
	{Kind: Neutral, Title: "Conference season", Description: "It's conference season; the team is traveling.", MoraleDelta: -1, Weight: 2},
}

// DailyChance is the probability an event fires on any given day.
const DailyChance = 0.15

// eventStreamSalt separates the event RNG stream.
const eventStreamSalt uint64 = 4242424242424242

// MaybeRoll deterministically decides whether an event fires on the given day
// and, if so, which one. Returns the chosen event and true when one fires.
func MaybeRoll(seed, day int64) (Event, bool) {
	r := rand.New(rand.NewPCG(uint64(seed)^eventStreamSalt, uint64(day)))
	if r.Float64() >= DailyChance {
		return Event{}, false
	}
	total := 0
	for _, e := range Catalog {
		total += e.Weight
	}
	pick := r.IntN(total)
	running := 0
	for _, e := range Catalog {
		running += e.Weight
		if pick < running {
			return e, true
		}
	}
	return Catalog[len(Catalog)-1], true
}
