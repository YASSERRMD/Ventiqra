# Changelog

All notable changes to Ventiqra are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/) and the project uses
[semantic versioning](https://semver.org/).

## [1.0.0] — 2026-06-23

### Added — Core simulation
- Company creation with starting cash, industry, and status lifecycle
- Daily simulation tick with deterministic seeded randomness
- Product lifecycle: idea → building → launched
- Customer model: acquisition, churn, satisfaction, MAUs
- Finance engine: monthly expenses, salary/burn, P&L, runway
- Bankruptcy detection, warning states, and restart
- Funding rounds with valuation, dilution, and negotiation
- Competitor system with market share impact
- Market model: TAM, growth, demand, trend multiplier

### Added — Gameplay systems
- **Strategic decisions** — risk/reward decision cards with short/long-term effects (Phase 27)
- **Predefined scenarios** — Bootstrap SaaS, VC-Funded, Hardware, Marketplace (Phase 28)
- **Scenario editor** — custom user-authored scenarios with validation (Phase 29)
- **Save/load** — named save slots with full simulation snapshots (Phase 30)
- **Timeline** — unified company history with monthly summaries (Phase 31)
- **Analytics** — daily metric snapshots + cash/revenue/burn/valuation charts (Phase 32)
- **WebSocket realtime** — live dashboard updates with reconnect (Phase 33)
- **Speed controls** — pause/resume, 1x/5x/30x, manual step (Phase 34)
- **Product roadmap** — feature backlog with development and shipping (Phase 35)
- **Technical debt** — debt accumulation, refactoring, outage risk (Phase 36)
- **Infrastructure** — capacity, hosting cost, scaling, outage risk (Phase 37)
- **Customer support** — ticket backlog, support agents, satisfaction link (Phase 38)
- **Sales pipeline** — B2B deals with stages and close probability (Phase 39)
- **Enterprise contracts** — recurring revenue with renewal/churn (Phase 40)
- **Achievements** — 6 milestone awards with tick evaluation (Phase 41)
- **Leaderboard** — local high scores with outcome scoring (Phase 42)
- **Difficulty levels** — easy/normal/hard/brutal/custom with economy multipliers (Phase 43)
- **Game balancing** — tuned formulas with invariant test suite (Phase 44)

### Added — Quality
- Comprehensive backend unit, integration, and repository tests (Phase 45)
- Frontend vitest suite: format, API mock, and component smoke tests (Phase 46)
- Playwright E2E test suite with auth and simulation flow specs (Phase 47)
- Architecture, simulation formula, API, contributing, and setup docs (Phase 48)
- Empty/loading/error states, error boundary, responsive panel wrapper (Phase 49)

### Tech stack
- Backend: Go 1.25, PostgreSQL 16, Redis 7
- Frontend: Next.js 16, TypeScript, Tailwind CSS v4, Recharts
- Realtime: WebSocket (coder/websocket)
- Testing: Go testing, Vitest, Playwright
- Deployment: Docker Compose
