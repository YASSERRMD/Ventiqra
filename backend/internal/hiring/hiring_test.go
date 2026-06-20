package hiring

import (
	"testing"
)

func TestGeneratePoolDeterministic(t *testing.T) {
	a := GeneratePool(42, 1)
	b := GeneratePool(42, 1)
	if len(a) != PoolSize || len(b) != PoolSize {
		t.Fatalf("pool size = %d/%d, want %d", len(a), len(b), PoolSize)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("candidate %d differs: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestGeneratePoolVariesByDay(t *testing.T) {
	day1 := GeneratePool(42, 1)
	day2 := GeneratePool(42, 2)
	same := true
	for i := range day1 {
		if day1[i].Name != day2[i].Name || day1[i].Skill != day2[i].Skill {
			same = false
			break
		}
	}
	if same {
		t.Errorf("expected different candidates across days, got identical pools")
	}
}

func TestGeneratePoolVariesBySeed(t *testing.T) {
	s1 := GeneratePool(1, 5)
	s2 := GeneratePool(2, 5)
	same := true
	for i := range s1 {
		if s1[i].Name != s2[i].Name {
			same = false
			break
		}
	}
	if same {
		t.Errorf("expected different candidates across seeds, got identical names")
	}
}

func TestCandidateFieldsInRange(t *testing.T) {
	pool := GeneratePool(7, 3)
	roleset := map[string]bool{
		"engineer": true, "designer": true, "sales": true,
		"marketing": true, "support": true, "operations": true,
	}
	for i, c := range pool {
		if c.Index != i {
			t.Errorf("candidate %d index = %d", i, c.Index)
		}
		if !roleset[c.Role] {
			t.Errorf("candidate %d role = %q", i, c.Role)
		}
		if c.Skill < 30 || c.Skill > 99 {
			t.Errorf("candidate %d skill = %d out of [30,99]", i, c.Skill)
		}
		if c.SalaryExpectation <= 0 {
			t.Errorf("candidate %d salary = %d", i, c.SalaryExpectation)
		}
		if c.HiringFee <= 0 {
			t.Errorf("candidate %d fee = %d", i, c.HiringFee)
		}
		if c.AcceptanceChance < 0 || c.AcceptanceChance > 1 {
			t.Errorf("candidate %d chance = %v out of [0,1]", i, c.AcceptanceChance)
		}
		if c.Quality != QualityTier(c.Skill) {
			t.Errorf("candidate %d quality mismatch", i)
		}
	}
}

func TestQualityTier(t *testing.T) {
	cases := map[int]string{
		0: "weak", 49: "weak", 50: "average", 79: "average", 80: "strong", 99: "strong",
	}
	for skill, want := range cases {
		if got := QualityTier(skill); got != want {
			t.Errorf("QualityTier(%d) = %q, want %q", skill, got, want)
		}
	}
}

func TestOfferAcceptedDeterministic(t *testing.T) {
	for _, index := range []int{0, 1, 2, 3, 4, 5} {
		first := OfferAccepted(42, 1, index)
		second := OfferAccepted(42, 1, index)
		if first != second {
			t.Errorf("index %d: offer decision not deterministic (%v vs %v)", index, first, second)
		}
	}
}

func TestOfferAcceptedOutOfRangeIsFalse(t *testing.T) {
	if OfferAccepted(42, 1, -1) {
		t.Errorf("negative index should not be accepted")
	}
	if OfferAccepted(42, 1, PoolSize) {
		t.Errorf("out-of-range index should not be accepted")
	}
}

func TestStrongCandidatesPickier(t *testing.T) {
	// Over many seeds, weak candidates should accept at least as often as strong.
	const trials = 200
	weakAccept, strongAccept := 0, 0
	for seed := int64(0); seed < trials; seed++ {
		pool := GeneratePool(seed, 0)
		for i, c := range pool {
			accepted := OfferAccepted(seed, 0, i)
			switch c.Quality {
			case "weak":
				if accepted {
					weakAccept++
				}
			case "strong":
				if accepted {
					strongAccept++
				}
			}
		}
	}
	if strongAccept >= weakAccept {
		t.Errorf("expected weak >= strong acceptances, got weak=%d strong=%d", weakAccept, strongAccept)
	}
}
