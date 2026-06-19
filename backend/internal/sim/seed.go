package sim

import (
	"hash/fnv"
)

// SeedFromCompanyID derives a stable int64 seed from a company identifier.
// The same id always yields the same seed, so a company's simulation sequence
// is reproducible without persisting extra randomness beyond the id.
func SeedFromCompanyID(companyID string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(companyID))
	return int64(h.Sum64())
}
