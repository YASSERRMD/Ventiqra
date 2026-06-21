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
	Crisis   Kind = "crisis"
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

// CrisisChance is the (much lower) probability a severe crisis fires on a given
// day. A crisis preempts the regular event roll when it triggers.
const CrisisChance = 0.02

// CrisisCatalog holds the severe crisis events. Each carries large penalties.
var CrisisCatalog = []Event{
	{Kind: Crisis, Title: "Server outage", Description: "A prolonged outage took the product offline for hours.", CashDelta: -400_000, ReputationDelta: -8, MoraleDelta: -10, Weight: 3},
	{Kind: Crisis, Title: "Bad PR scandal", Description: "A scandal drew intense negative coverage.", CashDelta: -250_000, ReputationDelta: -15, MoraleDelta: -8, Weight: 3},
	{Kind: Crisis, Title: "Competitor attack", Description: "A rival launched an aggressive campaign targeting your customers.", CashDelta: -150_000, ReputationDelta: -5, MoraleDelta: -6, Weight: 4},
	{Kind: Crisis, Title: "Funding collapse", Description: "A committed investor pulled out at the last minute.", CashDelta: -500_000, ReputationDelta: -4, MoraleDelta: -10, Weight: 2},
	{Kind: Crisis, Title: "Employee resignation wave", Description: "Several key employees resigned in quick succession.", CashDelta: -100_000, ReputationDelta: -3, MoraleDelta: -20, Weight: 2},
}

// eventStreamSalt separates the event RNG stream.
const eventStreamSalt uint64 = 4242424242424242

// MaybeRoll deterministically decides whether an event fires on the given day
// and, if so, which one. A crisis roll is checked first; if no crisis fires,
// the regular daily roll decides. Returns the chosen event and true when fired.
func MaybeRoll(seed, day int64) (Event, bool) {
	r := rand.New(rand.NewPCG(uint64(seed)^eventStreamSalt, uint64(day)))
	if r.Float64() < CrisisChance {
		return pick(r, CrisisCatalog), true
	}
	if r.Float64() >= DailyChance {
		return Event{}, false
	}
	return pick(r, Catalog), true
}

// pick makes a weighted deterministic selection from a catalog using r.
func pick(r *rand.Rand, catalog []Event) Event {
	total := 0
	for _, e := range catalog {
		total += e.Weight
	}
	p := r.IntN(total)
	running := 0
	for _, e := range catalog {
		running += e.Weight
		if p < running {
			return e
		}
	}
	return catalog[len(catalog)-1]
}
