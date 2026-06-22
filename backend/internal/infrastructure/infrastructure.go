// Package infrastructure models the company's hosting capacity and scaling.
// Capacity caps how many customers the platform can serve; exceeding it raises
// outage risk. Scaling raises both capacity and monthly hosting cost.
package infrastructure

// BaseCapacity is the capacity at tier 1.
const BaseCapacity = 1000

// CapacityPerTier is the capacity added per tier level.
const CapacityPerTier = 2000

// BaseHostingCostCents is the monthly hosting cost at tier 1.
const BaseHostingCostCents int64 = 500_00 // $500

// CostPerTierCents is the added monthly cost per tier.
const CostPerTierCents int64 = 1_500_00 // $1,500

// ScaleUpCostCents is the one-time cost to raise the tier by one.
const ScaleUpCostCents int64 = 30_000_00 // $30,000

// MaxTier is the highest infrastructure tier.
const MaxTier = 10

// CapacityForTier returns the customer capacity at a given tier.
func CapacityForTier(tier int) int {
	if tier < 1 {
		tier = 1
	}
	return BaseCapacity + (tier-1)*CapacityPerTier
}

// HostingCostForTier returns the monthly hosting cost (cents) at a tier.
func HostingCostForTier(tier int) int64 {
	if tier < 1 {
		tier = 1
	}
	return BaseHostingCostCents + int64(tier-1)*CostPerTierCents
}

// LoadRatio returns the fraction of capacity in use (customers/capacity),
// clamped to >= 0.
func LoadRatio(customers, capacity int) float64 {
	if capacity <= 0 {
		return 1
	}
	r := float64(customers) / float64(capacity)
	if r < 0 {
		return 0
	}
	return r
}

// OutageRisk returns the probability [0,1] of an infrastructure-caused outage
// given the load ratio. Below 80% load the risk is low; above 100% it's high.
func OutageRisk(loadRatio float64) float64 {
	if loadRatio <= 0.8 {
		return 0.02
	}
	if loadRatio >= 1.2 {
		return 0.8
	}
	// Linear ramp from 0.02 at 0.8 to 0.8 at 1.2.
	return 0.02 + (loadRatio-0.8)*(0.8-0.02)/(1.2-0.8)
}

// RecommendedTier returns the minimum tier that comfortably serves the given
// customer count (keeping load under 70%).
func RecommendedTier(customers int) int {
	for tier := 1; tier <= MaxTier; tier++ {
		if float64(customers) < 0.7*float64(CapacityForTier(tier)) {
			return tier
		}
	}
	return MaxTier
}
