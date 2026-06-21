package events

import "testing"

func TestCatalogHasAllKinds(t *testing.T) {
	kinds := map[Kind]bool{}
	for _, e := range Catalog {
		kinds[e.Kind] = true
		if e.Title == "" || e.Description == "" || e.Weight <= 0 {
			t.Errorf("malformed event: %+v", e)
		}
	}
	for _, k := range []Kind{Positive, Negative, Neutral} {
		if !kinds[k] {
			t.Errorf("catalog missing kind %q", k)
		}
	}
}

func TestMaybeRollDeterministic(t *testing.T) {
	aEv, aOk := MaybeRoll(42, 1)
	bEv, bOk := MaybeRoll(42, 1)
	if aOk != bOk {
		t.Errorf("fire decision not deterministic: %v vs %v", aOk, bOk)
	}
	if aOk && aEv != bEv {
		t.Errorf("event not deterministic: %+v vs %+v", aEv, bEv)
	}
}

func TestMaybeRollFiresSometimes(t *testing.T) {
	fired, total := 0, 1000
	for i := 0; i < total; i++ {
		if _, ok := MaybeRoll(int64(i), 1); ok {
			fired++
		}
	}
	if fired == 0 {
		t.Errorf("events never fire")
	}
	// ~15% expected; sanity bounds.
	if fired < total/10 || fired > total/4 {
		t.Errorf("fired %d/%d, expected ~15%%", fired, total)
	}
}

func TestMaybeRollVariesByDay(t *testing.T) {
	seen := map[string]bool{}
	for day := int64(0); day < 200; day++ {
		if ev, ok := MaybeRoll(7, day); ok {
			seen[ev.Title] = true
		}
	}
	if len(seen) < 3 {
		t.Errorf("expected several distinct events over time, saw %d", len(seen))
	}
}

func TestEventEffectsAreBounded(t *testing.T) {
	for _, e := range Catalog {
		// Negative/positive cash deltas should be sane magnitudes (not billions).
		if e.CashDelta < -10_000_000 || e.CashDelta > 10_000_000 {
			t.Errorf("event %+v has out-of-bounds cash delta", e)
		}
		if e.ReputationDelta < -50 || e.ReputationDelta > 50 {
			t.Errorf("event %+v has out-of-bounds reputation delta", e)
		}
	}
}
