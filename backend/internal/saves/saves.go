// Package saves models simulation snapshots: the captured state needed to
// persist and later restore a run. Snapshots are pure data structures so they
// can be marshaled to JSON for storage and applied deterministically on load.
package saves

// Snapshot is a restorable capture of a company's simulation run. Fields mirror
// the persisted company and simulation_state rows plus a few derived display
// values, so a save slot can be listed without joins.
type Snapshot struct {
	Day         int    `json:"day"`
	Seed        int64  `json:"seed"`
	CashCents   int64  `json:"cash_cents"`
	Revenue     int64  `json:"revenue"`
	MonthlyBurn int64  `json:"monthly_burn"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Industry    string `json:"industry"`
}

// SlotLimit is the maximum number of named slots a single owner may keep.
const SlotLimit = 5

// MaxSlotNameLen bounds the slot identifier length.
const MaxSlotNameLen = 32

// MaxLabelLen bounds the human-readable label length.
const MaxLabelLen = 80

// IsValidSlot reports whether a slot name is well-formed: 1–32 chars, only
// lowercase letters, digits, hyphens, and underscores.
func IsValidSlot(slot string) bool {
	if len(slot) < 1 || len(slot) > MaxSlotNameLen {
		return false
	}
	for _, c := range slot {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '-' || c == '_':
		default:
			return false
		}
	}
	return true
}

// ClampLabel truncates a label to MaxLabelLen runes, returning the result.
func ClampLabel(label string) string {
	r := []rune(label)
	if len(r) > MaxLabelLen {
		return string(r[:MaxLabelLen])
	}
	return label
}
