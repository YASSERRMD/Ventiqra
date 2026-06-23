# Simulation Formulas

This document describes the economic formulas that drive the simulation. All
constants live in `internal/balance/` and are guarded by invariant tests.

## Revenue

```
daily_revenue = customers × ARPU × price_multiplier
```

- **ARPU** (average revenue per user per day): $0.50 (`balance.ARPUCents = 50`)
- **Price multiplier**: clamped to [0.5, 2.0]; 1.0 is neutral.

## Churn

```
daily_churn = customers × base_churn_rate × satisfaction_factor × difficulty_multiplier
```

- **Base churn rate**: 1%/day (`balance.BaseChurnRate = 0.01`)
- **Satisfaction factor**: `1.0 + (50 - satisfaction) / 50` — 100 sat → 0.5×, 50 sat → 1×, 0 sat → 2×
- **Difficulty multiplier**: from the company's difficulty preset.

## Burn

Monthly burn = salaries + infrastructure + marketing. Applied as
`daily_burn = monthly_burn / 30`.

## Funding

```
funding_chance = base_chance × (0.5 + 0.5 × traction) × ask_factor × difficulty_mult
```

- **Base chance**: 0.5 (`balance.FundingBaseChance`)
- **Traction**: `customers / 500`, capped at 2×
- **Ask factor**: `2_000_000 / ask_cents`, capped at 2×
- Clamped to [0.05, 0.95].

## Valuation

```
valuation = max(cash, monthly_revenue × 12 × 8)
```

An 8× annualized revenue multiple, floored at cash balance.

## Technical debt

- Each shipped feature adds 8 debt points (`techdebt.ShipDebtGain`).
- Debt clamped to [0, 100].
- Quality = 100 − debt.
- Outage risk: 0 below 40 debt, ramping to 0.9 at 90+ debt.
- Refactor: −25 debt for $20,000.

## Infrastructure

- Capacity = 1000 + (tier−1) × 2000.
- Hosting cost = $500 + (tier−1) × $1,500/month.
- Scale-up cost: $30,000 per tier.
- Outage risk rises above 80% load.

## Customer support

- Arrivals: 5 tickets/day per 1000 customers.
- Resolution: 8 tickets/day per support agent.
- Satisfaction penalty: 1 point per 100 open tickets.

## Sales pipeline

- Stages: lead → qualified → proposal → negotiation → close.
- Base close probabilities: 10%, 25%, 45%, 65%.
- Each sales agent adds +5% (capped at +35%).

## Enterprise contracts

- Daily revenue = annual_value / 360.
- Renewal chance: 70% base + satisfaction bonus (up to +20%).
- Discount capped at 50%.

## Difficulty multipliers

| Level  | Burn  | Churn | Funding | Events | Start Cash | Acquisition |
|--------|-------|-------|---------|--------|------------|-------------|
| Easy   | 0.7×  | 0.5×  | 1.4×    | 0.6×   | 1.5×       | 1.3×        |
| Normal | 1.0×  | 1.0×  | 1.0×    | 1.0×   | 1.0×       | 1.0×        |
| Hard   | 1.3×  | 1.3×  | 0.8×    | 1.3×   | 0.85×      | 0.85×       |
| Brutal | 1.7×  | 1.7×  | 0.6×    | 1.6×   | 0.7×       | 0.6×        |

## Leaderboard score

```
score = days_survived×100 + peak_valuation/$10k + customers/100 + achievements×5000
```

- Bankrupt: ×0.5 penalty.
- Acquired: ×1.5 bonus.
