# Ventiqra

> An open-source startup simulation engine.

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black?logo=next.js)](https://nextjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

Ventiqra is a startup simulator. Users create a company, build products, hire
employees, manage funding, tune pricing, fight off competitors, survive market
events, and watch the company grow — or go bankrupt.

**This is NOT an AI product.** It is a deterministic, data-driven simulation
engine.

---

## Features

### Core simulation
- **Company simulation** — found a company and steer its fate day by day.
- **Product lifecycle** — ideate, build, launch, and iterate on products.
- **Hiring & employees** — recruit talent, manage salaries, productivity, and morale.
- **Finance & runway** — track cash, burn rate, revenue, and runway.
- **Funding rounds** — pitch investors, negotiate equity, and manage dilution.
- **Competitors** — react to rival product launches and market-share shifts.
- **Market events** — survive random booms, crises, and trend shifts.
- **Pricing experiments** — A/B price sensitivity and revenue impact.
- **Customer growth** — acquisition, churn, satisfaction, and MAU.

### Gameplay systems
- **Strategic decisions** — risk/reward decision cards with short- and long-term effects.
- **Product roadmap** — feature backlog with development progress and value shipping.
- **Technical debt** — debt accumulates as you ship; refactor to reduce outage risk.
- **Infrastructure** — scale capacity and hosting to match customer growth.
- **Customer support** — ticket backlog that erodes satisfaction if unchecked.
- **Sales pipeline** — B2B deals through lead → qualified → proposal → negotiation → close.
- **Enterprise contracts** — recurring revenue with renewal and churn.
- **Achievements** — 6 milestones (first launch, first funding, profitability, unicorn, growing team, viral).
- **Difficulty levels** — easy, normal, hard, brutal, and custom with economy multipliers.

### Scenarios
- **Predefined scenarios** — Bootstrap SaaS, VC-Funded, Hardware, Marketplace.
- **Scenario editor** — author and save custom scenarios with tuned cash, market, and difficulty.

### Player experience
- **Save/load** — capture a run into named slots and restore it later.
- **Timeline** — unified company history with monthly summaries.
- **Analytics** — cash, revenue, burn, valuation, and customer charts (Recharts).
- **Realtime** — WebSocket push of live tick updates with auto-reconnect.
- **Speed controls** — pause/resume, 1x/5x/30x speed, and manual step.
- **Leaderboard** — local high scores with outcome-based scoring.

---

## Tech Stack

| Layer       | Technology                          |
| ----------- | ----------------------------------- |
| Backend     | Go 1.25                             |
| Frontend    | Next.js 16 + TypeScript + Tailwind v4 |
| Database    | PostgreSQL 16                       |
| Cache/Queue | Redis 7                             |
| Realtime    | WebSocket (coder/websocket)         |
| Charts      | Recharts                            |
| Testing     | Go testing, Vitest, Playwright      |
| Deployment  | Docker Compose                      |

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/) 1.25+
- [Node.js](https://nodejs.org/) 20+
- [Docker](https://www.docker.com/) & Docker Compose
- [PostgreSQL](https://www.postgresql.org/) 16+ (or use the Docker service)
- [Redis](https://redis.io/) 7+ (or use the Docker service)

### Quick start (Docker Compose)

```bash
git clone https://github.com/YASSERMD/Ventiqra.git
cd Ventiqra
cp .env.example .env
docker compose up -d
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

### Manual setup

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for full instructions.

---

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — directory layout, layered pattern, tick flow
- [Simulation Formulas](docs/SIMULATION_FORMULAS.md) — all economic formulas
- [API Reference](docs/API.md) — full endpoint documentation
- [Development Setup](docs/DEVELOPMENT.md) — prerequisites, env vars, testing
- [Contributing](CONTRIBUTING.md) — branch workflow, commit conventions, code style
- [Release Checklist](docs/RELEASE_CHECKLIST.md) — pre-flight checklist for releases
- [Changelog](CHANGELOG.md) — version history

---

## Repository Structure

```
Ventiqra/
├── backend/              # Go API service
│   ├── cmd/api/          # Entry point
│   ├── internal/         # Domain logic, repository, server, middleware
│   └── migrations/       # SQL migrations
├── frontend/             # Next.js + TypeScript + Tailwind
│   ├── src/app/          # App Router pages
│   ├── src/components/   # React components
│   ├── src/lib/          # API client, types, hooks
│   └── e2e/              # Playwright E2E tests
├── docs/                 # Architecture, API, formulas, setup docs
├── docker-compose.yml    # Dev orchestration
├── docker-compose.prod.yml # Production override
├── VERSION               # Current version
├── CHANGELOG.md          # Version history
└── README.md
```

---

## Development Workflow

Development is organized into 50 numbered phases. Each phase lives on its own
branch (`phase_01` … `phase_50`), is reviewed via a Pull Request, and is merged
into `main` before the next phase begins.

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## Testing

```bash
# Backend
cd backend && go test ./...

# Frontend unit tests
cd frontend && npm run test

# E2E (requires running stack)
cd frontend && npm run test:e2e
```

---

## License

[MIT](./LICENSE) © Mohamed Yasser
