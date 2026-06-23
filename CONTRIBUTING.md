# Contributing to Ventiqra

Thanks for your interest in contributing! This guide covers the workflow.

## Development setup

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for full setup instructions.

## Branch workflow

Ventiqra uses a phase-based branch workflow:

1. **Never push directly to `main`.**
2. Create one branch per phase: `phase_01`, `phase_02`, etc.
3. Every small task gets its own atomic commit with a clear message.
4. At the end of each phase:
   - Run `go build ./... && go test ./...` (backend)
   - Run `npm run lint && npm run build && npm run test` (frontend)
   - Push the branch
   - Create a PR
   - Merge to `main`
   - Delete the branch
   - Continue to the next phase

## Commit messages

Use conventional commit prefixes:

- `feat(scope):` — new feature
- `fix(scope):` — bug fix
- `test(scope):` — tests only
- `docs:` — documentation
- `refactor(scope):` — code restructuring
- `chore:` — tooling, deps

Examples:
```
feat(decisions): add deterministic decision catalog and risk roll
test(api): cover scenario list and apply endpoints
docs: mention strategic decisions in README features
```

## Code style

### Backend (Go)

- Follow the layered pattern: pure logic → repository → server handler → wiring.
- Pure logic packages (`internal/<domain>/`) must have no I/O and be fully unit-tested.
- Repository packages embed `*Repository` and use the shared pgx pool.
- Server handlers use `decodeJSON`/`writeJSON`/`writeError` helpers.
- Every new server file must be wired via an `Option` in `server.go` and added to both `main.go` and `auth_test.go`.

### Frontend (TypeScript/React)

- Use `"use client"` for interactive components.
- Fetch via the `api` client from `@/lib/api`.
- Types live in `@/lib/types.ts`.
- Dashboard panels go in `@/components/dashboard/`.

## Testing

- **Backend unit tests**: `go test ./internal/<domain>/`
- **Backend integration tests**: `go test ./internal/server/` (requires PostgreSQL)
- **Frontend unit tests**: `npm run test` (vitest)
- **E2E tests**: `npm run test:e2e` (playwright; requires running stack)

## Reporting issues

Open a GitHub issue with:
- What you expected
- What happened
- Steps to reproduce
- Backend/frontend version
