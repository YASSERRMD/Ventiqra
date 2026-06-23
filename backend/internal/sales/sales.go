// Package sales models the B2B sales pipeline: deals that progress through
// stages, with a close probability that rises as they advance. Sales employees
// advance deals each tick; when a deal reaches the negotiation→close boundary,
// the probability roll decides won vs lost.
package sales

import "math/rand/v2"

// Stage is a deal's position in the pipeline.
type Stage string

const (
	StageLead        Stage = "lead"
	StageQualified   Stage = "qualified"
	StageProposal    Stage = "proposal"
	StageNegotiation Stage = "negotiation"
	StageClosedWon   Stage = "closed_won"
	StageClosedLost  Stage = "closed_lost"
)

// PipelineOrder is the ordered set of pre-close stages.
var PipelineOrder = []Stage{StageLead, StageQualified, StageProposal, StageNegotiation}

// StageProbability is the base close probability (%) for a deal at each stage.
var StageProbability = map[Stage]int{
	StageLead:        10,
	StageQualified:   25,
	StageProposal:    45,
	StageNegotiation: 65,
}

// IsValidStage reports whether s is a recognized stage.
func IsValidStage(s Stage) bool {
	switch s {
	case StageLead, StageQualified, StageProposal, StageNegotiation, StageClosedWon, StageClosedLost:
		return true
	}
	return false
}

// NextStage returns the stage after s in the pipeline, or s if already closed.
func NextStage(s Stage) Stage {
	for i, st := range PipelineOrder {
		if st == s && i+1 < len(PipelineOrder) {
			return PipelineOrder[i+1]
		}
	}
	return s
}

// Advance moves a deal one stage forward. At negotiation→close, it rolls the
// probability and returns closed_won or closed_lost. seed makes the roll
// deterministic. Returns the new stage.
func Advance(current Stage, probability int, seed int64) Stage {
	if current == StageClosedWon || current == StageClosedLost {
		return current
	}
	if current == StageNegotiation {
		// Roll close: compare a deterministic [0,100) draw to the probability.
		r := rand.New(rand.NewPCG(uint64(seed)^0x5a135000, uint64(seed)|1))
		roll := r.IntN(100)
		if roll < probability {
			return StageClosedWon
		}
		return StageClosedLost
	}
	return NextStage(current)
}

// AdvanceChanceBoost returns the probability boost a sales team provides: each
// agent adds a flat 5 points (capped at 95).
func AdvanceChanceBoost(agents int) int {
	if agents < 0 {
		agents = 0
	}
	boost := agents * 5
	if boost > 35 {
		boost = 35
	}
	return boost
}
