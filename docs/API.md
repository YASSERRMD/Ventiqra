# API Reference

Base URL: `/api/v1`. All protected endpoints require a `Authorization: Bearer <jwt>`
header (except WebSocket, which uses `?token=<jwt>`).

## Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Register `{email, password, name}` → `{token, user}` |
| POST | `/auth/login` | Login `{email, password}` → `{token, user}` |

## Company

| Method | Path | Description |
|--------|------|-------------|
| POST | `/companies` | Create company `{name, starting_cash_cents?}` |
| GET | `/companies/me` | Get the owner's latest company |
| GET | `/companies/me/metrics` | Headline metrics (cash, revenue, burn, day) |
| GET | `/companies/me/market` | Market state (TAM, growth, trend) |

## Simulation

| Method | Path | Description |
|--------|------|-------------|
| POST | `/companies/me/sim/tick` | Advance one simulated day |
| GET | `/companies/me/sim/control` | Get speed/mode |
| POST | `/companies/me/sim/pause` | Pause auto-run |
| POST | `/companies/me/sim/resume` | Resume auto-run |
| POST | `/companies/me/sim/speed` | Set speed `{speed: 1|5|30}` |

## Products

| Method | Path | Description |
|--------|------|-------------|
| POST | `/companies/me/products` | Create product `{name}` |
| GET | `/companies/me/products` | List products |
| POST | `/companies/me/products/{id}/launch` | Launch a product |

## Employees & Hiring

| Method | Path | Description |
|--------|------|-------------|
| GET | `/companies/me/employees` | List employees |
| POST | `/companies/me/hire` | Hire `{candidate_id}` |

## Finance & Funding

| Method | Path | Description |
|--------|------|-------------|
| POST | `/companies/me/funding/raise` | Raise a round `{amount_cents}` |

## Pricing & Marketing

| Method | Path | Description |
|--------|------|-------------|
| POST | `/companies/me/products/{id}/price` | Set price `{price_cents}` |
| POST | `/companies/me/marketing` | Set marketing spend |

## Strategic Decisions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/companies/me/decisions/pending` | Get pending decision card |
| POST | `/companies/me/decisions/{id}/resolve` | Resolve `{choice_id}` |

## Scenarios

| Method | Path | Description |
|--------|------|-------------|
| GET | `/scenarios` | List predefined scenarios |
| POST | `/scenarios/{id}/apply` | Apply a scenario |
| GET | `/scenarios/custom` | List custom scenarios |
| POST | `/scenarios/custom` | Create custom scenario |
| PATCH | `/scenarios/custom/{id}` | Update custom scenario |
| DELETE | `/scenarios/custom/{id}` | Delete custom scenario |
| POST | `/scenarios/custom/{id}/apply` | Apply custom scenario |

## Save/Load

| Method | Path | Description |
|--------|------|-------------|
| GET | `/saves` | List save slots |
| POST | `/saves` | Save `{slot, label?}` |
| POST | `/saves/{slot}/load` | Load a slot |
| DELETE | `/saves/{slot}` | Delete a slot |

## Timeline & Analytics

| Method | Path | Description |
|--------|------|-------------|
| GET | `/companies/me/timeline` | Unified timeline + monthly summary |
| GET | `/companies/me/analytics` | Daily metric series |

## Roadmap & Engineering

| Method | Path | Description |
|--------|------|-------------|
| GET | `/companies/me/features` | List roadmap features |
| POST | `/companies/me/features` | Create feature `{name, priority?, value_points?}` |
| POST | `/companies/me/features/{id}/develop` | Advance feature `{points?}` |
| DELETE | `/companies/me/features/{id}` | Delete feature |
| GET | `/companies/me/tech-debt` | Get debt/quality state |
| POST | `/companies/me/tech-debt/refactor` | Pay down debt |
| GET | `/companies/me/infrastructure` | Get capacity/load |
| POST | `/companies/me/infrastructure/scale` | Scale up a tier |

## Support & Sales

| Method | Path | Description |
|--------|------|-------------|
| GET | `/companies/me/support` | Ticket backlog state |
| GET | `/companies/me/deals` | List sales deals |
| POST | `/companies/me/deals` | Create deal `{name, value_cents?}` |
| POST | `/companies/me/deals/{id}/advance` | Advance deal stage |
| GET | `/companies/me/contracts` | List enterprise contracts |
| POST | `/companies/me/contracts` | Sign contract |

## Achievements & Leaderboard

| Method | Path | Description |
|--------|------|-------------|
| GET | `/companies/me/achievements` | List awarded achievements |
| GET | `/leaderboard` | Local high scores |

## Difficulty

| Method | Path | Description |
|--------|------|-------------|
| GET | `/companies/me/difficulty` | Get level + multipliers |
| POST | `/companies/me/difficulty` | Set level `{level}` |

## Realtime

| Method | Path | Description |
|--------|------|-------------|
| GET (WS) | `/realtime?token=<jwt>` | WebSocket for live updates |
