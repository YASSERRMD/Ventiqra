// Package difficulty defines difficulty presets and the multipliers they apply
// to the simulation's economy. Pure data so the server, tests, and the
// scenario editor share one source of truth.
package difficulty

// Level is a difficulty label.
type Level string

const (
	LevelEasy    Level = "easy"
	LevelNormal  Level = "normal"
	LevelHard    Level = "hard"
	LevelBrutal  Level = "brutal"
	LevelCustom  Level = "custom"
)

// IsValid reports whether l is a recognized level.
func IsValid(l Level) bool {
	switch l {
	case LevelEasy, LevelNormal, LevelHard, LevelBrutal, LevelCustom:
		return true
	}
	return false
}

// Multipliers are the economy scaling factors a difficulty applies. Values >1
// make things harder (more burn, more churn); <1 make them easier.
type Multipliers struct {
	BurnMultiplier       float64 // applied to monthly burn
	ChurnMultiplier      float64 // applied to customer churn rate
	FundingChanceMult    float64 // applied to funding-round success chance
	EventSeverityMult    float64 // applied to negative-event impact
	StartingCashMult     float64 // applied to starting cash
	AcquisitionRateMult  float64 // applied to customer acquisition rate
}

// Presets is the catalog of difficulty multipliers.
var Presets = map[Level]Multipliers{
	LevelEasy: {
		BurnMultiplier: 0.7, ChurnMultiplier: 0.5, FundingChanceMult: 1.4,
		EventSeverityMult: 0.6, StartingCashMult: 1.5, AcquisitionRateMult: 1.3,
	},
	LevelNormal: {
		BurnMultiplier: 1.0, ChurnMultiplier: 1.0, FundingChanceMult: 1.0,
		EventSeverityMult: 1.0, StartingCashMult: 1.0, AcquisitionRateMult: 1.0,
	},
	LevelHard: {
		BurnMultiplier: 1.3, ChurnMultiplier: 1.3, FundingChanceMult: 0.8,
		EventSeverityMult: 1.3, StartingCashMult: 0.85, AcquisitionRateMult: 0.85,
	},
	LevelBrutal: {
		BurnMultiplier: 1.7, ChurnMultiplier: 1.7, FundingChanceMult: 0.6,
		EventSeverityMult: 1.6, StartingCashMult: 0.7, AcquisitionRateMult: 0.6,
	},
}

// For returns the multipliers for a level, defaulting to Normal when unknown.
func For(l Level) Multipliers {
	if m, ok := Presets[l]; ok {
		return m
	}
	return Presets[LevelNormal]
}
