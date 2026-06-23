// Package leaderboard computes a company's outcome score from its run metrics.
// Pure so the server and tests share one source of truth.
package leaderboard

// Outcome labels how a run ended.
type Outcome string

const (
	OutcomeBankrupt Outcome = "bankrupt"
	OutcomeThriving Outcome = "thriving"
	OutcomeAcquired Outcome = "acquired"
)

// Input is the run summary used to compute the score.
type Input struct {
	DaysSurvived  int
	PeakValuation int64 // cents
	Customers     int
	Achievements  int
	Outcome       Outcome
}

// Score returns a single integer outcome score. Higher is better. Weighting:
//   - days survived: 100/day
//   - peak valuation: 1 point per $10k
//   - customers: 1 point per 100
//   - achievements: 5000 each
// Bankrupt runs take a 50% penalty; acquired runs get a 1.5x bonus.
func Score(in Input) int64 {
	if in.DaysSurvived < 0 {
		in.DaysSurvived = 0
	}
	if in.PeakValuation < 0 {
		in.PeakValuation = 0
	}
	if in.Customers < 0 {
		in.Customers = 0
	}
	s := int64(in.DaysSurvived)*100 +
		in.PeakValuation/1_000_00 +
		int64(in.Customers)/100 +
		int64(in.Achievements)*5000
	switch in.Outcome {
	case OutcomeBankrupt:
		s = s / 2
	case OutcomeAcquired:
		s = s * 3 / 2
	}
	if s < 0 {
		s = 0
	}
	return s
}
