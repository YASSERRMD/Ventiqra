package sales

import "testing"

func TestNextStage(t *testing.T) {
	cases := []struct{ from, want Stage }{
		{StageLead, StageQualified},
		{StageQualified, StageProposal},
		{StageProposal, StageNegotiation},
		{StageNegotiation, StageNegotiation}, // terminal pre-close
		{StageClosedWon, StageClosedWon},
	}
	for _, c := range cases {
		if got := NextStage(c.from); got != c.want {
			t.Errorf("NextStage(%s) = %s, want %s", c.from, got, c.want)
		}
	}
}

func TestAdvanceThroughPipeline(t *testing.T) {
	s := Advance(StageLead, 10, 1)
	if s != StageQualified {
		t.Errorf("advance lead = %s, want qualified", s)
	}
}

func TestAdvanceAtNegotiationCloses(t *testing.T) {
	// High probability → should usually win; low → usually lose. Check both
	// branches fire across seeds.
	won, lost := 0, 0
	for i := int64(0); i < 200; i++ {
		s := Advance(StageNegotiation, 80, i)
		if s == StageClosedWon {
			won++
		} else if s == StageClosedLost {
			lost++
		} else {
			t.Errorf("negotiation advanced to %s, want closed", s)
		}
	}
	if won < 100 {
		t.Errorf("expected mostly wins at 80%%, got %d/200", won)
	}
	if lost == 0 {
		t.Error("expected at least one loss")
	}
}

func TestAdvanceDoesNotReopenClosed(t *testing.T) {
	if s := Advance(StageClosedWon, 90, 1); s != StageClosedWon {
		t.Errorf("reopened won: %s", s)
	}
	if s := Advance(StageClosedLost, 90, 1); s != StageClosedLost {
		t.Errorf("reopened lost: %s", s)
	}
}

func TestAdvanceChanceBoost(t *testing.T) {
	if AdvanceChanceBoost(0) != 0 {
		t.Error("0 agents → 0 boost")
	}
	if AdvanceChanceBoost(3) != 15 {
		t.Errorf("3 agents → %d, want 15", AdvanceChanceBoost(3))
	}
	if AdvanceChanceBoost(20) != 35 {
		t.Errorf("20 agents → %d, want cap 35", AdvanceChanceBoost(20))
	}
}
