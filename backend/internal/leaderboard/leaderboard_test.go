package leaderboard

import "testing"

func TestScoreBasic(t *testing.T) {
	s := Score(Input{DaysSurvived: 100, PeakValuation: 10_000_000_00, Customers: 500, Achievements: 2, Outcome: OutcomeThriving})
	// 100*100 + 100000 + 5 + 10000 = 30005
	if s != 30005 {
		t.Errorf("score = %d, want 30005", s)
	}
}

func TestScoreBankruptPenalty(t *testing.T) {
	base := Input{DaysSurvived: 100, PeakValuation: 10_000_000_00}
	thrive := Score(base)
	bankrupt := Score(Input{DaysSurvived: 100, PeakValuation: 10_000_000_00, Outcome: OutcomeBankrupt})
	if bankrupt >= thrive {
		t.Errorf("bankrupt (%d) should be < thriving (%d)", bankrupt, thrive)
	}
	if bankrupt != thrive/2 {
		t.Errorf("bankrupt = %d, want %d", bankrupt, thrive/2)
	}
}

func TestScoreAcquiredBonus(t *testing.T) {
	base := Input{DaysSurvived: 100, PeakValuation: 10_000_000_00}
	thrive := Score(base)
	acquired := Score(Input{DaysSurvived: 100, PeakValuation: 10_000_000_00, Outcome: OutcomeAcquired})
	if acquired != thrive*3/2 {
		t.Errorf("acquired = %d, want %d", acquired, thrive*3/2)
	}
}

func TestScoreNegativeClamped(t *testing.T) {
	s := Score(Input{DaysSurvived: -10, PeakValuation: -100, Customers: -5})
	if s != 0 {
		t.Errorf("negative input score = %d, want 0", s)
	}
}
