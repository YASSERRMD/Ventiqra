// Package develop models how employees drive product development and payroll.
//
// The formulas are pure and deterministic: given a roster (roles, skill, morale)
// they compute a daily contribution to product progress and a monthly payroll
// burn. The server layer applies these during each simulation tick so that
// hiring speeds up development while salaries raise burn, exactly as specified
// by the simulation design.
package develop

// Employee is the minimal roster input needed to compute development output.
type Employee struct {
	Role   string
	Skill  int // 0..100
	Morale int // 0..100
}

// RoleWeight is how much each role contributes to product progress, relative to
// an engineer. Builders (engineers, designers) move the product forward; non-
// building roles do not add development progress.
var RoleWeight = map[string]float64{
	"engineer":   1.0,
	"designer":   0.6,
	"operations": 0.2,
	"sales":      0.0,
	"marketing":  0.0,
	"support":    0.0,
}

const (
	// BaselineMorale is the morale value that yields a 1.0 morale multiplier.
	BaselineMorale = 70
	// BaseDailyProgressPerEngineer is the percentage points a perfectly skilled,
	// fully baseline-motivated engineer adds per day.
	BaseDailyProgressPerEngineer = 0.5
)

// SkillFactor scales output by skill as a fraction of 100.
func SkillFactor(skill int) float64 {
	return float64(clampInt(skill, 0, 100)) / 100.0
}

// MoraleFactor scales output by morale, with BaselineMorale as the 1.0 point.
// Higher morale boosts output (up to ~1.43x at 100); low morale throttles it.
func MoraleFactor(morale int) float64 {
	return float64(clampInt(morale, 0, 100)) / float64(BaselineMorale)
}

// DailyProgress returns the percentage points of product progress a roster
// contributes in one simulated day. It is the sum of each builder's weighted,
// skill- and morale-adjusted output.
func DailyProgress(employees []Employee) float64 {
	var total float64
	for _, e := range employees {
		w := RoleWeight[e.Role]
		if w <= 0 {
			continue
		}
		total += BaseDailyProgressPerEngineer * w * SkillFactor(e.Skill) * MoraleFactor(e.Morale)
	}
	if total < 0 {
		return 0
	}
	return total
}

// MonthlyPayroll sums a list of monthly salaries (in cents) into the total
// monthly payroll burn.
func MonthlyPayroll(salaryCents []int64) int64 {
	var total int64
	for _, s := range salaryCents {
		if s > 0 {
			total += s
		}
	}
	return total
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
