// Package support models the customer-support system: an open-ticket backlog
// that grows with the customer base and shrinks as support staff resolve
// tickets. A large unresolved backlog erodes customer satisfaction.
package support

// TicketsPerThousandCustomers is the daily ticket-arrival rate per 1000 customers.
const TicketsPerThousandCustomers = 5.0

// TicketsResolvedPerAgentPerDay is how many tickets one support agent clears
// per simulated day.
const TicketsResolvedPerAgentPerDay = 8

// SatisfactionPenaltyPerHundredOpen returns the satisfaction points subtracted
// per 100 open tickets (clamped >= 0).
func SatisfactionPenaltyPerHundredOpen(openTickets int) int {
	p := openTickets / 100
	if p < 0 {
		return 0
	}
	return p
}

// DailyArrivals returns how many new tickets arrive in a day for a customer base.
func DailyArrivals(customers int) int {
	if customers <= 0 {
		return 0
	}
	return int(float64(customers) / 1000.0 * TicketsPerThousandCustomers)
}

// ResolveForDay returns how many tickets the given agent count resolves in a day.
func ResolveForDay(agents int) int {
	if agents < 0 {
		agents = 0
	}
	return agents * TicketsResolvedPerAgentPerDay
}

// ApplyDay advances the backlog by one day: adds arrivals, subtracts resolutions,
// clamps at zero, and returns (newOpen, resolvedToday, arrivalsToday).
func ApplyDay(openTickets, customers, agents int) (newOpen, resolvedToday, arrivalsToday int) {
	arrivalsToday = DailyArrivals(customers)
	resolvedToday = ResolveForDay(agents)
	if resolvedToday > openTickets+arrivalsToday {
		resolvedToday = openTickets + arrivalsToday
	}
	newOpen = openTickets + arrivalsToday - resolvedToday
	if newOpen < 0 {
		newOpen = 0
	}
	return newOpen, resolvedToday, arrivalsToday
}
