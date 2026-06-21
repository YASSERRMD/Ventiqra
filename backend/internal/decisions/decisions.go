// Package decisions models strategic decision cards: periodically offered
// choices with short-term and long-term effects, and a risk/reward roll that
// can swing the short-term outcome between success and failure. Everything is
// pure and deterministic so the same company on the same day always faces the
// same card and the same resolution.
package decisions

import (
	"fmt"
	"math/rand/v2"
)

// Category groups decisions by theme for display and balancing.
type Category string

const (
	CategoryGrowth Category = "growth"
	CategoryCost   Category = "cost"
	CategoryPeople Category = "people"
	CategoryRisk   Category = "risk"
)

// Choice is one selectable option on a decision card.
type Choice struct {
	// ID is the stable choice identifier used to record and resolve a choice.
	ID string
	// Label is the short user-facing label shown on the choice button.
	Label string
	// Description explains the choice and its trade-offs.
	Description string
	// Short-term effects applied immediately on success.
	CashDelta       int64 // cents
	ReputationDelta int
	MoraleDelta     int
	// Risk: the probability in [0,1] that the success branch fires. The failure
	// branch applies the Fail* deltas instead. SuccessChance == 1 means the
	// choice is deterministic (no risk).
	SuccessChance float64
	// Short-term effects applied when the risk roll fails.
	FailCashDelta       int64
	FailReputationDelta int
	FailMoraleDelta     int
	// Long-term commitment applied every day for DurationDays after resolution.
	// A negative value is an ongoing cost (e.g. an office lease); a positive
	// value is an ongoing benefit (e.g. a referral program). Cash-only to keep
	// the recurring model simple.
	RecurringCashDelta int64
	DurationDays       int
}

// Decision is a single card the player may be offered.
type Decision struct {
	ID          string
	Category    Category
	Title       string
	Description string
	Choices     []Choice
}

// Catalog is the pool of decisions the engine can offer.
var Catalog = []Decision{
	{
		ID:          "pivot_to_enterprise",
		Category:    CategoryGrowth,
		Title:       "Pivot to enterprise pricing",
		Description: "A handful of large accounts are asking for an enterprise tier. Repositioning takes effort but unlocks bigger deals.",
		Choices: []Choice{
			{
				ID: "stay_self_serve", Label: "Stay self-serve",
				Description:        "Keep the current product motion; avoid the disruption of a pivot.",
				CashDelta:          0,
				SuccessChance:      1,
				RecurringCashDelta: 0,
				DurationDays:       0,
			},
			{
				ID: "launch_enterprise", Label: "Launch enterprise tier",
				Description:        "Invest in an enterprise sales motion. A successful pivot lifts cash and reputation; a botched one burns runway and morale.",
				CashDelta:          300_000,
				ReputationDelta:    3,
				MoraleDelta:        3,
				SuccessChance:      0.6,
				FailCashDelta:      -200_000,
				FailReputationDelta: -2,
				FailMoraleDelta:    -4,
				RecurringCashDelta: 25_000,
				DurationDays:       30,
			},
		},
	},
	{
		ID:          "aggressive_marketing_blitz",
		Category:    CategoryGrowth,
		Title:       "Aggressive marketing blitz",
		Description: "Marketing wants to go all-in on a two-week blitz to steal mindshare. The upside is large; the cost is certain.",
		Choices: []Choice{
			{
				ID: "hold_steady", Label: "Hold steady",
				Description:   "Keep the current marketing budget.",
				CashDelta:     0,
				SuccessChance: 1,
			},
			{
				ID: "run_blitz", Label: "Run the blitz",
				Description:        "Spend big upfront. If it lands, the brand gets a lasting lift; if it flops, the money is gone and morale slips.",
				CashDelta:          -150_000,
				ReputationDelta:    6,
				MoraleDelta:        2,
				SuccessChance:      0.55,
				FailCashDelta:      -150_000,
				FailReputationDelta: -3,
				FailMoraleDelta:    -2,
				RecurringCashDelta: 0,
				DurationDays:       0,
			},
		},
	},
	{
		ID:          "open_remote_office",
		Category:    CategoryPeople,
		Title:       "Open a remote office",
		Description: "Hiring is tight locally. A remote office widens the talent pool but adds coordination overhead and a recurring lease cost.",
		Choices: []Choice{
			{
				ID: "stay_local", Label: "Stay local",
				Description:   "Keep hiring in the home market.",
				CashDelta:     0,
				SuccessChance: 1,
			},
			{
				ID: "open_office", Label: "Open the office",
				Description:        "Set up a satellite office. A smooth rollout lifts morale and reputation; a rough one drains cash and spirits.",
				CashDelta:          -80_000,
				ReputationDelta:    2,
				MoraleDelta:        5,
				SuccessChance:      0.65,
				FailCashDelta:      -80_000,
				FailReputationDelta: 0,
				FailMoraleDelta:    -5,
				RecurringCashDelta: -12_000,
				DurationDays:       60,
			},
		},
	},
	{
		ID:          "acquire_small_competitor",
		Category:    CategoryRisk,
		Title:       "Acquire a small competitor",
		Description: "A struggling rival is for sale cheap. An acquisition could consolidate the market — or saddle you with their baggage.",
		Choices: []Choice{
			{
				ID: "pass", Label: "Pass",
				Description:   "Let the rival fade on its own.",
				CashDelta:     0,
				SuccessChance: 1,
			},
			{
				ID: "acquire", Label: "Acquire them",
				Description:        "Buy the rival. A clean integration brings talent and share; a messy one courts reputation damage and resignations.",
				CashDelta:          -400_000,
				ReputationDelta:    4,
				MoraleDelta:        2,
				SuccessChance:      0.45,
				FailCashDelta:      -400_000,
				FailReputationDelta: -6,
				FailMoraleDelta:    -8,
				RecurringCashDelta: 40_000,
				DurationDays:       45,
			},
		},
	},
	{
		ID:          "cut_salaries_extend_runway",
		Category:    CategoryCost,
		Title:       "Cut salaries to extend runway",
		Description: "Cash is getting tight. A temporary salary cut would slow the burn — but the team may not take it kindly.",
		Choices: []Choice{
			{
				ID: "protect_salaries", Label: "Protect salaries",
				Description:   "Find savings elsewhere; keep the team whole.",
				CashDelta:     0,
				SuccessChance: 1,
			},
			{
				ID: "cut_salaries", Label: "Cut salaries 10%",
				Description:        "Reduce payroll to preserve cash. The savings recur monthly, but morale takes a hit either way.",
				CashDelta:          0,
				ReputationDelta:    0,
				MoraleDelta:        0,
				SuccessChance:      1,
				RecurringCashDelta: 18_000,
				DurationDays:       90,
				FailMoraleDelta:    -8,
			},
		},
	},
	{
		ID:          "launch_referral_program",
		Category:    CategoryGrowth,
		Title:       "Launch a referral program",
		Description: "Rewarding existing customers for referrals is cheap to try and compounds if it works.",
		Choices: []Choice{
			{
				ID: "skip_referrals", Label: "Skip it",
				Description:   "Rely on existing acquisition channels.",
				CashDelta:     0,
				SuccessChance: 1,
			},
			{
				ID: "start_program", Label: "Start the program",
				Description:        "Stand up a referral program. If customers embrace it, recurring revenue climbs; if not, you're out a small setup cost.",
				CashDelta:          -30_000,
				ReputationDelta:    1,
				MoraleDelta:        1,
				SuccessChance:      0.7,
				FailCashDelta:      -30_000,
				FailReputationDelta: 0,
				FailMoraleDelta:    0,
				RecurringCashDelta: 9_000,
				DurationDays:       60,
			},
		},
	},
}

// OfferInterval is the number of days between decision card offers. A new card
// is offered only on days that are positive multiples of this interval and only
// when no card is already pending.
const OfferInterval = 10

// OfferStreamSalt separates the decision-offer RNG stream from other streams.
const offerStreamSalt uint64 = 7777555533331111

// MaybeOffer deterministically decides whether a decision card should be offered
// on the given day and, if so, which one. A card is offered on positive
// multiples of OfferInterval (day 10, 20, 30, …); the specific card is chosen
// deterministically from the catalog. Returns the chosen decision and true when
// an offer is due.
func MaybeOffer(seed, day int64) (Decision, bool) {
	if day <= 0 || day%OfferInterval != 0 {
		return Decision{}, false
	}
	r := rand.New(rand.NewPCG(uint64(seed)^offerStreamSalt, uint64(day)))
	return Catalog[r.IntN(len(Catalog))], true
}

// Effects are the concrete short-term deltas to apply after a risk roll.
type Effects struct {
	CashDelta       int64
	ReputationDelta int
	MoraleDelta     int
}

// Outcome labels the result of a choice's risk roll.
type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
)

// resolveStreamSalt separates the risk-roll RNG stream.
const resolveStreamSalt uint64 = 18182838485868788

// ResolveOutcome deterministically rolls a choice's risk and returns the
// concrete short-term effects to apply, the outcome label, and the long-term
// recurring commitment (cash delta per day and duration). seedKey mixes the
// company seed with the decision and choice so each card resolves independently
// yet reproducibly.
func ResolveOutcome(seed int64, decisionID, choiceID string, day int64, ch Choice) (Effects, Outcome) {
	// Deterministic, choice-specific seed: company seed ^ salt ^ a hash of the
	// decision/choice ids ^ the day.
	seedKey := uint64(seed) ^ resolveStreamSalt ^ hashKey(decisionID, choiceID) ^ uint64(day)
	r := rand.New(rand.NewPCG(seedKey, seedKey>>3))
	if r.Float64() < ch.SuccessChance {
		return Effects{
			CashDelta:       ch.CashDelta,
			ReputationDelta: ch.ReputationDelta,
			MoraleDelta:     ch.MoraleDelta,
		}, OutcomeSuccess
	}
	return Effects{
		CashDelta:       ch.FailCashDelta,
		ReputationDelta: ch.FailReputationDelta,
		MoraleDelta:     ch.FailMoraleDelta,
	}, OutcomeFailure
}

// FindDecision returns the catalog decision with the given id, or false.
func FindDecision(id string) (Decision, bool) {
	for _, d := range Catalog {
		if d.ID == id {
			return d, true
		}
	}
	return Decision{}, false
}

// FindChoice returns the choice with the given id on the decision, or false.
func (d Decision) FindChoice(id string) (Choice, bool) {
	for _, c := range d.Choices {
		if c.ID == id {
			return c, true
		}
	}
	return Choice{}, false
}

// hashKey folds two ids into a stable 64-bit key for RNG seeding. It uses the
// FNV-1a pattern over the concatenated ids.
func hashKey(decisionID, choiceID string) uint64 {
	const (
		offset uint64 = 14695981039346656037
		prime  uint64 = 1099511628211
	)
	h := offset
	for _, b := range []byte(fmt.Sprintf("%s|%s", decisionID, choiceID)) {
		h ^= uint64(b)
		h *= prime
	}
	return h
}
