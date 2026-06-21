# Ventiqra

> An open-source startup simulation engine.

Ventiqra is a startup simulator. Users create a company, build products, hire
employees, manage funding, tune pricing, fight off competitors, survive market
events, and watch the company grow — or go bankrupt.

**This is NOT an AI product.** It is a deterministic, data-driven simulation
engine. The LLM is only used here as a coding assistant to build it.

---

## Features

- **Company simulation** — found a company and steer its fate day by day.
- **Product lifecycle** — ideate, build, launch, and iterate on products.
- **Hiring & employees** — recruit talent, manage salaries, productivity, and morale.
- **Finance & runway** — track cash, burn rate, revenue, and runway.
- **Funding rounds** — pitch investors, negotiate equity, and manage dilution.
- **Competitors** — react to rival product launches and market-share shifts.
- **Market events** — survive random booms, crises, and trend shifts.
- **Strategic decisions** — weigh risk/reward decision cards with short- and long-term effects.
- **Pricing experiments** — A/B price sensitivity and revenue impact.
- **Customer growth** — acquisition, churn, satisfaction, and MAU.
- **Dashboard** — live metrics, charts, and timeline.
- **Scenario editor** — predefined scenarios and custom difficulty tuning.

---

## Tech Stack

| Layer       | Technology                          |
| ----------- | ----------------------------------- |
| Backend     | Go                                  |
| Frontend    | Next.js + TypeScript + Tailwind CSS |
| Database    | PostgreSQL                          |
| Cache/Queue | Redis                               |
| Realtime    | WebSocket                           |
| Charts      | Recharts                            |
| Deployment  | Docker Compose                      |
| Testing     | Go tests, Playwright, frontend unit |

---

## Repository Structure

```
Ventiqra/
├── backend/      # Go API service (cmd, internal, migrations, pkg)
├── frontend/     # Next.js + TypeScript + Tailwind app
├── infra/        # Docker and deployment assets
├── docs/         # Architecture, API, and simulation docs
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Getting Started

> The full setup guide will be documented in `docs/` as each phase lands.
> Phase 1 establishes the repository scaffolding only.

### Prerequisites

- [Go](https://go.dev/) 1.22+
- [Node.js](https://nodejs.org/) 20+
- [Docker](https://www.docker.com/) & Docker Compose
- [PostgreSQL](https://www.postgresql.org/) 15+ (or use the Docker service)
- [Redis](https://redis.io/) 7+ (or use the Docker service)

### Quick start (Docker Compose)

```bash
cp .env.example .env
docker compose up -d
```

---

## Development Workflow

Development is organized into numbered phases. Each phase lives on its own
branch (`phase_01`, `phase_02`, …), is reviewed via a Pull Request, and is
merged into `main` before the next phase begins.

- Never commit directly to `main`.
- One branch per phase.
- Atomic, focused commits.

---

## License

[MIT](./LICENSE) © Mohamed Yasser
