// Package marketing models how a company's marketing spend drives customer
// acquisition: conversion rates, channels, CAC, and the deterministic number of
// conversions a budget produces. Formulas are pure.
package marketing

import "math"

// Channel is an acquisition channel with its relative efficiency and cost.
type Channel struct {
	Name       string
	Weight     float64 // share of the blended conversion pool
	Conversion float64 // base conversion contribution
}

// Channels is the blended acquisition mix. Spend is allocated across these.
var Channels = []Channel{
	{Name: "paid-ads", Weight: 0.40, Conversion: 0.10},
	{Name: "content", Weight: 0.25, Conversion: 0.08},
	{Name: "referral", Weight: 0.20, Conversion: 0.12},
	{Name: "organic", Weight: 0.15, Conversion: 0.06},
}

// BaseConversionRate is the blended conversion rate at saturation spend.
const BaseConversionRate = 0.09

// SaturationCents is the monthly spend at which conversion rate nears its max;
// beyond this, diminishing returns flatten acquisition.
const SaturationCents int64 = 1_000_000

// ConversionRate returns the blended conversion rate for a monthly budget with
// diminishing returns (logistic saturation toward BaseConversionRate).
func ConversionRate(monthlyBudgetCents int64) float64 {
	if monthlyBudgetCents <= 0 {
		return 0
	}
	return BaseConversionRate * (1 - math.Exp(-float64(monthlyBudgetCents)/float64(SaturationCents)))
}

// Reach returns the gross number of people the budget touches (cents → dollars
// of spend, since each dollar reaches one person).
func Reach(monthlyBudgetCents int64) int64 {
	if monthlyBudgetCents <= 0 {
		return 0
	}
	return monthlyBudgetCents / 100
}

// Conversions returns the deterministic number of new customers a monthly budget
// produces in one day (monthly conversions scaled to a 30-day month).
func Conversions(monthlyBudgetCents int64, seed, day int64) int {
	if monthlyBudgetCents <= 0 {
		return 0
	}
	rate := ConversionRate(monthlyBudgetCents)
	monthly := float64(Reach(monthlyBudgetCents)) * rate
	daily := monthly / 30.0
	return int(daily)
}

// CAC returns the customer acquisition cost (cents per customer) for a budget
// and the conversions it produced. Returns 0 when no conversions occurred.
func CAC(monthlyBudgetCents int64, conversions int) int64 {
	if conversions <= 0 {
		return 0
	}
	return monthlyBudgetCents / int64(conversions)
}
