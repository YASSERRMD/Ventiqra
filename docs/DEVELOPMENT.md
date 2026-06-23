# Development Setup

## Prerequisites

- **Go** 1.25+
- **Node.js** 20+
- **PostgreSQL** 16+
- **Redis** 7+ (for queues, optional for dev)
- **Docker** + **Docker Compose** (for containerized dev)

## Quick start with Docker Compose

```bash
# Clone
git clone https://github.com/YASSERMD/Ventiqra.git
cd Ventiqra

# Copy env
cp .env.example .env

# Start all services
docker compose up -d

# Run migrations (auto-applied on backend startup)
# Backend: http://localhost:8080
# Frontend: http://localhost:3000
```

## Manual setup

### 1. Database

```bash
# Create database
createdb ventiqra

# Set connection string
export DATABASE_URL="postgres://ventiqra:changeme@localhost:5432/ventiqra?sslmode=disable"
```

### 2. Backend

```bash
cd backend

# Install dependencies
go mod download

# Copy and edit env
cp .env.example .env

# Run the API server (migrations auto-apply on startup)
go run ./cmd/api
```

The API runs on `http://localhost:8080`.

### 3. Frontend

```bash
cd frontend

# Install dependencies
npm install

# Copy and edit env
cp .env.example .env.local

# Run dev server
npm run dev
```

The frontend runs on `http://localhost:3000`.

## Environment variables

### Backend (`.env`)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server port |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `JWT_SECRET` | — | JWT signing secret |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `ENV` | `development` | Environment (development/production) |

### Frontend (`.env.local`)

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Backend API base URL |

## Running tests

### Backend

```bash
cd backend

# All tests (requires PostgreSQL)
go test ./...

# Specific package
go test ./internal/decisions/

# With verbose output
go test ./internal/server/ -v -run TestDecision
```

Set `DATABASE_TEST_URL` to use a separate test database.

### Frontend

```bash
cd frontend

# Unit/component tests (vitest)
npm run test

# E2E tests (playwright — requires running stack)
npm run test:e2e

# Lint
npm run lint
```

## Database migrations

Migrations are numbered SQL files in `backend/migrations/`. They auto-apply on
backend startup via `db.Migrate()`. To reset:

```bash
psql $DATABASE_URL -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"
```

## Project structure

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full directory layout and design patterns.
