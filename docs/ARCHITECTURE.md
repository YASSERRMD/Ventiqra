# Architecture

Ventiqra is a monorepo with a Go backend, a Next.js frontend, PostgreSQL, and Redis.

```
ventiqra/
├── backend/                 # Go API service
│   ├── cmd/api/             # Entry point (main.go)
│   ├── internal/
│   │   ├── auth/            # JWT tokens, password hashing
│   │   ├── balance/         # Tuned economy formulas + invariant tests
│   │   ├── config/          # Env-based configuration
│   │   ├── contracts/       # Enterprise contract model
│   │   ├── customers/       # Customer acquisition + churn
│   │   ├── db/              # PostgreSQL connection + migrations
│   │   ├── decisions/       # Strategic decision cards
│   │   ├── develop/         # Product development logic
│   │   ├── difficulty/      # Difficulty presets + multipliers
│   │   ├── events/          # Random event engine
│   │   ├── finance/         # Burn, revenue, P&L
│   │   ├── funding/         # Funding rounds + valuation
│   │   ├── hiring/          # Candidate generation + offers
│   │   ├── infrastructure/  # Capacity, hosting cost, scaling
│   │   ├── leaderboard/     # Outcome score computation
│   │   ├── logger/          # Structured logging
│   │   ├── market/          # Market size, growth, demand
│   │   ├── marketing/       # Acquisition channels, CAC
│   │   ├── middleware/       # Auth + request middleware
│   │   ├── morale/          # Employee morale + burnout
│   │   ├── pricing/         # Price sensitivity + experiments
│   │   ├── realtime/        # WebSocket hub
│   │   ├── repository/      # Data access layer (all repos)
│   │   ├── reputation/      # Brand reputation engine
│   │   ├── roadmap/         # Feature backlog + shipping
│   │   ├── sales/           # B2B deal pipeline
│   │   ├── saves/           # Snapshot model for save/load
│   │   ├── scenarios/       # Predefined + custom scenarios
│   │   ├── server/          # HTTP handlers + routing
│   │   ├── sim/             # Simulation tick core
│   │   ├── simctl/          # Speed control (pause/resume/speed)
│   │   ├── support/         # Customer support tickets
│   │   ├── techdebt/        # Technical debt model
│   │   ├── testutil/        # Shared test helpers
│   │   └── timeline/        # Unified company history
│   └── migrations/          # Numbered SQL migration files
├── frontend/                # Next.js + TypeScript + Tailwind
│   ├── src/
│   │   ├── app/             # Next.js App Router pages
│   │   ├── components/      # React components (dashboard panels)
│   │   └── lib/             # API client, types, hooks, utilities
│   ├── e2e/                 # Playwright E2E tests
│   └── src/__tests__/       # Vitest unit/component tests
├── docker-compose.yml       # Local dev orchestration
└── docs/                    # This documentation
```

## Layered backend pattern

Each domain follows four layers:

1. **Pure logic** (`internal/<domain>/`) — deterministic functions, no I/O, fully unit-tested.
2. **Repository** (`internal/repository/<domain>_repo.go`) — SQL queries over the shared pgx pool.
3. **Server handler** (`internal/server/<domain>.go`) — HTTP request/response, validation, orchestration.
4. **Wiring** (`server.go` Option + route, `main.go` construction, `auth_test.go` test wiring).

## Simulation tick

The `/api/v1/companies/me/sim/tick` endpoint advances one simulated day. Per tick:

1. Advance the day counter.
2. Apply finance (burn, revenue).
3. Advance customers (acquisition + churn).
4. Roll random events.
5. Apply active decision effects + offer new cards.
6. Advance support tickets.
7. Accrue contract revenue + roll renewals.
8. Evaluate achievements.
9. Record an analytics snapshot.
10. Broadcast via WebSocket.
11. Detect bankruptcy → finalize leaderboard.

## Frontend data flow

- Server components fetch initial data via the API client.
- Client components use `useEffect` + the `api` client for mutations.
- `useRealtime` opens a WebSocket for live tick updates.
- `ventiqra:tick` window events let panels refresh on push.
