<p align="center">
  <img src="docs/images/ventiqra-logo.png" width="140" alt="Ventiqra" />
</p>

<h1 align="center">Ventiqra</h1>

<p align="center">
  <strong>An open-source startup simulation engine.</strong><br/>
  Deterministic. Data-driven. Transparent.</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25-1B2A4A?style=flat-square&logo=go" alt="Go" /></a>
  <a href="https://nextjs.org/"><img src="https://img.shields.io/badge/Next.js-16-1B2A4A?style=flat-square&logo=next.js" alt="Next.js" /></a>
  <a href="https://www.postgresql.org/"><img src="https://img.shields.io/badge/PostgreSQL-16-1B2A4A?style=flat-square&logo=postgresql" alt="PostgreSQL" /></a>
  <a href="https://redis.io/"><img src="https://img.shields.io/badge/Redis-7-1B2A4A?style=flat-square&logo=redis" alt="Redis" /></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-MIT-C5A55A?style=flat-square" alt="License: MIT" /></a>
  <a href="./VERSION"><img src="https://img.shields.io/badge/version-1.0.0-C5A55A?style=flat-square" alt="Version" /></a>
</p>

---

## Overview

**Ventiqra** is a deterministic, data-driven startup simulator. Players found a
company, build and launch products, hire and manage employees, raise funding,
experiment with pricing, outmaneuver competitors, navigate market events, and
steer their company toward growth — or bankruptcy.

It is **not** an AI product. Every outcome is computed by a transparent,
tunable simulation engine with documented formulas and deterministic randomness.

---

## Table of Contents

- [Key Features](#key-features)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Manual Setup](#manual-setup)
- [Architecture](#architecture)
- [Documentation](#documentation)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## Key Features

### Simulation Engine

| System | Description |
|--------|-------------|
| **Daily tick** | A deterministic, seeded simulation cycle that advances the company one day at a time. |
| **Company lifecycle** | Found → build → launch → grow → fund → scale, with bankruptcy detection. |
| **Finance** | Cash, revenue, burn rate, runway, profit/loss, and valuation tracking. |
| **Products** | Full lifecycle from idea through development to launch and iteration. |
| **Customers** | Acquisition, churn, satisfaction, and monthly active users. |
| **Funding** | Investor offers, valuation, dilution, and negotiation. |
| **Market** | Total addressable market, growth rate, demand, and trend multipliers. |
| **Competitors** | AI rivals that launch products and compete for market share. |

### Gameplay Systems

| System | Description |
|--------|-------------|
| **Strategic decisions** | Risk/reward decision cards with short- and long-term effects. |
| **Product roadmap** | Feature backlog with development progress and value shipping. |
| **Technical debt** | Debt accumulates as you ship; refactoring reduces outage risk. |
| **Infrastructure** | Capacity tiers, hosting costs, and load-based outage risk. |
| **Customer support** | Ticket backlog with support-agent resolution and satisfaction impact. |
| **Sales pipeline** | B2B deals progressing through lead → qualified → proposal → negotiation → close. |
| **Enterprise contracts** | Multi-year recurring revenue with renewal and churn rolls. |
| **Achievements** | Six milestone awards (first launch, first funding, profitability, unicorn, growing team, viral). |
| **Random & crisis events** | Positive, negative, neutral, and severe low-probability crisis events. |

### Scenarios & Difficulty

| Feature | Description |
|---------|-------------|
| **Predefined scenarios** | Bootstrap SaaS, VC-Funded Startup, Hardware Startup, Marketplace. |
| **Custom scenario editor** | Author and save scenarios with tuned cash, burn, market, and difficulty. |
| **Difficulty levels** | Easy, Normal, Hard, Brutal, and Custom — each applies economy multipliers. |
| **Game balancing** | Centralized, invariant-tested formula constants in `internal/balance/`. |

### Player Experience

| Feature | Description |
|---------|-------------|
| **Save / load** | Up to five named save slots with full simulation snapshots. |
| **Timeline** | Unified company history with monthly summaries and milestone tracking. |
| **Analytics** | Interactive Recharts visualizations for cash, revenue, burn, valuation, and customers. |
| **Realtime** | WebSocket push of live tick updates with automatic reconnection. |
| **Speed controls** | Pause, resume, 1×/5×/30× speed, and manual single-step mode. |
| **Leaderboard** | Local high scores with outcome-based scoring (bankrupt, thriving, acquired). |

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | [Go](https://go.dev/) 1.25 |
| **Frontend** | [Next.js](https://nextjs.org/) 16 · TypeScript · [Tailwind CSS](https://tailwindcss.com/) v4 |
| **Database** | [PostgreSQL](https://www.postgresql.org/) 16 |
| **Cache / Queue** | [Redis](https://redis.io/) 7 |
| **Realtime** | WebSocket ([coder/websocket](https://github.com/coder/websocket)) |
| **Charts** | [Recharts](https://recharts.org/) |
| **Unit testing** | Go `testing` · [Vitest](https://vitest.dev/) |
| **E2E testing** | [Playwright](https://playwright.dev/) |
| **Deployment** | Docker Compose |

---

## Quick Start

The fastest way to run Ventiqra is with Docker Compose, which starts the
backend, frontend, PostgreSQL, and Redis together.

### Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- Git

### Steps

```bash
git clone https://github.com/YASSERMD/Ventiqra.git
cd Ventiqra
cp .env.example .env
docker compose up -d
```

Once the containers are healthy:

| Service | URL |
|---------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |

Database migrations run automatically on backend startup.

---

## Manual Setup

For local development without Docker, see the full guide in
[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md). Summary:

**Backend**

```bash
cd backend
go mod download
cp .env.example .env   # edit DATABASE_URL, JWT_SECRET, etc.
go run ./cmd/api
```

**Frontend**

```bash
cd frontend
npm install
cp .env.example .env.local
npm run dev
```

### Environment Variables

Key variables (see `.env.example` for the complete list):

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | Secret used to sign JWT tokens |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `PORT` | `8080` | Backend API port |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Backend URL (frontend) |

---

## Architecture

Ventiqra follows a strict layered architecture. Each domain is split into four
layers with a clear dependency direction — pure logic never depends on I/O,
repositories never contain business rules, and handlers never touch the database
directly.

<p align="center">
  <img src="docs/images/ventiqra-architecture.png" alt="Ventiqra layered architecture" />
</p>

The simulation tick (`POST /api/v1/companies/me/sim/tick`) advances one
simulated day and runs the full cycle: finance → customers → events → decisions
→ support → contracts → achievements → analytics → realtime broadcast →
bankruptcy detection.

For the full directory layout and design patterns, see
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

### Repository Structure

```
Ventiqra/
├── backend/
│   ├── cmd/api/              # Application entry point
│   ├── internal/
│   │   ├── auth/             # JWT tokens, password hashing
│   │   ├── balance/          # Tuned economy formulas + invariant tests
│   │   ├── config/           # Environment-based configuration
│   │   ├── contracts/        # Enterprise contract model
│   │   ├── customers/        # Customer acquisition and churn
│   │   ├── db/               # PostgreSQL connection and migrations
│   │   ├── decisions/        # Strategic decision cards
│   │   ├── difficulty/       # Difficulty presets and multipliers
│   │   ├── events/           # Random event engine
│   │   ├── finance/          # Burn, revenue, P&L
│   │   ├── funding/          # Funding rounds and valuation
│   │   ├── infrastructure/   # Capacity, hosting, scaling
│   │   ├── leaderboard/      # Outcome score computation
│   │   ├── market/           # Market size, growth, demand
│   │   ├── marketing/        # Acquisition channels, CAC
│   │   ├── morale/           # Employee morale and burnout
│   │   ├── pricing/          # Price sensitivity
│   │   ├── realtime/         # WebSocket hub
│   │   ├── repository/       # Data access layer (all repos)
│   │   ├── reputation/       # Brand reputation
│   │   ├── roadmap/          # Feature backlog and shipping
│   │   ├── sales/            # B2B deal pipeline
│   │   ├── saves/            # Simulation snapshot model
│   │   ├── scenarios/        # Predefined and custom scenarios
│   │   ├── server/           # HTTP handlers and routing
│   │   ├── sim/              # Simulation tick core
│   │   ├── simctl/           # Speed control (pause/resume)
│   │   ├── support/          # Customer support tickets
│   │   ├── techdebt/         # Technical debt model
│   │   └── timeline/         # Unified company history
│   └── migrations/           # 33 numbered SQL migrations
├── frontend/
│   ├── src/
│   │   ├── app/              # Next.js App Router pages (10 routes)
│   │   ├── components/       # React components (27 panels/widgets)
│   │   └── lib/              # API client, types, hooks, utilities
│   ├── e2e/                  # Playwright E2E tests
│   └── src/__tests__/        # Vitest unit and component tests
├── docs/                     # Architecture, formulas, API, setup
├── docker-compose.yml        # Development orchestration
├── docker-compose.prod.yml   # Production override
├── CHANGELOG.md              # Version history
├── CONTRIBUTING.md           # Contribution guide
└── VERSION                   # Current release version
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/ARCHITECTURE.md) | Directory layout, layered pattern, tick flow, data flow |
| [Simulation Formulas](docs/SIMULATION_FORMULAS.md) | All economic formulas with constants and difficulty tables |
| [API Reference](docs/API.md) | Complete REST + WebSocket endpoint documentation |
| [Development Setup](docs/DEVELOPMENT.md) | Prerequisites, env vars, manual and Docker setup, test commands |
| [Contributing](CONTRIBUTING.md) | Branch workflow, commit conventions, code style, testing |
| [Release Checklist](docs/RELEASE_CHECKLIST.md) | Pre-flight checklist for cutting a release |
| [Changelog](CHANGELOG.md) | Versioned history of all notable changes |

---

## Testing

Ventiqra has three test layers:

### Backend (Go)

```bash
cd backend
go test ./...                          # all packages
go test ./internal/decisions/ -v       # specific package
go test ./internal/server/ -v -run TestSimulationEngine  # specific test
```

> Requires a running PostgreSQL instance. Set `DATABASE_TEST_URL` or use the
> default `postgres://ventiqra:changeme@localhost:5432/ventiqra?sslmode=disable`.

### Frontend (Vitest)

```bash
cd frontend
npm run test     # unit and component tests
```

### End-to-End (Playwright)

```bash
cd frontend
npm run test:e2e   # requires a running frontend + backend stack
```

> E2E tests gracefully skip when the application is not reachable.

---

## Contributing

Contributions are welcome. The project uses a phase-based branch workflow:

1. Never push directly to `main`.
2. Create a descriptively named branch (e.g., `feat/new-event-type`).
3. Make atomic commits with [conventional commit](https://www.conventionalcommits.org/) messages.
4. Ensure `go build`, `go vet`, `npm run lint`, and `npm run build` all pass.
5. Open a Pull Request and request review.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide.

---

## License

This project is licensed under the [MIT License](./LICENSE).

---

<p align="center">
  <sub>Built by <strong>Mohamed Yasser</strong> · Solutions Architect</sub>
</p>
