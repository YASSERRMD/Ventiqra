package decisions

import "testing"

func TestMaybeOfferCadence(t *testing.T) {
	cases := []struct {
		day  int64
		want bool
	}{
		{0, false}, // never on day 0
		{1, false},
		{9, false},
		{10, true}, // first offer
		{11, false},
		{20, true}, // second offer
		{30, true},
	}
	for _, c := range cases {
		_, ok := MaybeOffer(42, c.day)
		if ok != c.want {
			t.Errorf("day %d: offered = %v, want %v", c.day, ok, c.want)
		}
	}
}

func TestMaybeOfferIsDeterministic(t *testing.T) {
	const seed int64 = 1234
	const day int64 = 10
	first, ok1 := MaybeOffer(seed, day)
	second, ok2 := MaybeOffer(seed, day)
	if !ok1 || !ok2 {
		t.Fatalf("expected an offer on day 10")
	}
	if first.ID != second.ID {
		t.Errorf("non-deterministic offer: %q vs %q", first.ID, second.ID)
	}
}

func TestMaybeOfferPicksACatalogDecision(t *testing.T) {
	d, ok := MaybeOffer(7, 10)
	if !ok {
		t.Fatal("expected an offer")
	}
	found := false
	for _, c := range Catalog {
		if c.ID == d.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("offered decision %q not in catalog", d.ID)
	}
}

func TestResolveOutcomeSuccessIsStable(t *testing.T) {
	d, _ := FindDecision("launch_referral_program")
	ch, _ := d.FindChoice("start_program")
	// Run the resolution many times with the same inputs; it must be identical.
	eff, out := ResolveOutcome(99, d.ID, ch.ID, 10, ch)
	for i := 0; i < 20; i++ {
		e2, o2 := ResolveOutcome(99, d.ID, ch.ID, 10, ch)
		if e2 != eff || o2 != out {
			t.Fatalf("ResolveOutcome not deterministic: got (%+v,%v) after (%+v,%v)", e2, o2, eff, out)
		}
	}
	// The outcome must match the effects it produced.
	switch out {
	case OutcomeSuccess:
		if eff != (Effects{CashDelta: ch.CashDelta, ReputationDelta: ch.ReputationDelta, MoraleDelta: ch.MoraleDelta}) {
			t.Errorf("success effects mismatch: %+v", eff)
		}
	case OutcomeFailure:
		if eff != (Effects{CashDelta: ch.FailCashDelta, ReputationDelta: ch.FailReputationDelta, MoraleDelta: ch.FailMoraleDelta}) {
			t.Errorf("failure effects mismatch: %+v", eff)
		}
	default:
		t.Errorf("unknown outcome %q", out)
	}
}

func TestResolveOutcomeRiskDistribution(t *testing.T) {
	// A choice with SuccessChance 0.6 (launch_enterprise on pivot_to_enterprise)
	// across many distinct seeds should land on success roughly 60% of the time.
	// Use a tolerant band to avoid flakiness.
	d, ok := FindDecision("pivot_to_enterprise")
	if !ok {
		t.Fatal("missing pivot_to_enterprise decision")
	}
	ch, ok := d.FindChoice("launch_enterprise")
	if !ok {
		t.Fatal("missing launch_enterprise choice")
	}
	successes := 0
	const trials = 4000
	for i := 0; i < trials; i++ {
		_, out := ResolveOutcome(int64(i), d.ID, ch.ID, 10, ch)
		if out == OutcomeSuccess {
			successes++
		}
	}
	frac := float64(successes) / trials
	if frac < 0.53 || frac > 0.67 {
		t.Errorf("success fraction = %.3f, want ~0.6", frac)
	}
}

func TestResolveOutcomeDeterministicChoiceHasNoRisk(t *testing.T) {
	// A choice with SuccessChance 1 must always succeed.
	d, ok := FindDecision("pivot_to_enterprise")
	if !ok {
		t.Fatal("missing pivot_to_enterprise decision")
	}
	ch, ok := d.FindChoice("stay_self_serve")
	if !ok {
		t.Fatal("missing stay_self_serve choice")
	}
	for i := 0; i < 50; i++ {
		_, out := ResolveOutcome(int64(i), d.ID, ch.ID, 10, ch)
		if out != OutcomeSuccess {
			t.Fatalf("deterministic choice failed on seed %d", i)
		}
	}
}

func TestFindChoiceOnUnknownDecisionIsFalse(t *testing.T) {
	// Guard against the regression where a choice id is mistaken for a decision
	// id (which silently returns the zero Decision and SuccessChance 0).
	if _, ok := FindDecision("launch_enterprise"); ok {
		t.Error("launch_enterprise is a choice id, not a decision id")
	}
}

func TestEachDecisionHasTwoChoicesAndValidRisk(t *testing.T) {
	for _, d := range Catalog {
		if len(d.Choices) != 2 {
			t.Errorf("decision %q has %d choices, want 2", d.ID, len(d.Choices))
		}
		for _, c := range d.Choices {
			if c.SuccessChance < 0 || c.SuccessChance > 1 {
				t.Errorf("decision %q choice %q: success chance %v out of [0,1]", d.ID, c.ID, c.SuccessChance)
			}
			if c.DurationDays < 0 {
				t.Errorf("decision %q choice %q: negative duration %d", d.ID, c.ID, c.DurationDays)
			}
		}
	}
}
