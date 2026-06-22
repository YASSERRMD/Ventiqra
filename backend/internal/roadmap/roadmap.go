// Package roadmap models the product roadmap: backlog features with priority,
// development progress, and a value contribution once shipped. Pure helpers so
// the server and tests share one source of truth.
package roadmap

// Status is a feature's lifecycle state.
type Status string

const (
	StatusBacklog    Status = "backlog"
	StatusDeveloping Status = "developing"
	StatusShipped    Status = "shipped"
)

// IsValidStatus reports whether s is a recognized status.
func IsValidStatus(s Status) bool {
	return s == StatusBacklog || s == StatusDeveloping || s == StatusShipped
}

// Feature is a single roadmap item.
type Feature struct {
	ID          string
	Name        string
	Description string
	Priority    int
	Status      Status
	Progress    int  // 0..100
	ValuePoints int  // product-value contribution when shipped
	StartedDay  *int
	ShippedDay  *int
}

// DevelopProgress advances a feature by the given engineering points, clamping
// at 100. Returns the new progress and whether the feature shipped this tick
// (crossed 100 from below).
func DevelopProgress(f *Feature, points int) (newProgress int, shipped bool) {
	if f == nil {
		return 0, false
	}
	if points < 0 {
		points = 0
	}
	prev := f.Progress
	f.Progress = clamp(f.Progress+points, 0, 100)
	shipped = prev < 100 && f.Progress >= 100
	if shipped {
		f.Status = StatusShipped
	}
	return f.Progress, shipped
}

// SortByPriorityDesc returns the indices that would order features by priority
// descending (stable for equal priorities). Pure helper for display/retrieval.
func SortByPriorityDesc(features []Feature) []int {
	idx := make([]int, len(features))
	for i := range features {
		idx[i] = i
	}
	// Simple insertion sort (n is small for a roadmap) — stable on equal keys.
	for i := 1; i < len(idx); i++ {
		for j := i; j > 0 && features[idx[j-1]].Priority < features[idx[j]].Priority; j-- {
			idx[j-1], idx[j] = idx[j], idx[j-1]
		}
	}
	return idx
}

// TotalShippedValue sums the value points of shipped features.
func TotalShippedValue(features []Feature) int {
	total := 0
	for i := range features {
		if features[i].Status == StatusShipped {
			total += features[i].ValuePoints
		}
	}
	return total
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
