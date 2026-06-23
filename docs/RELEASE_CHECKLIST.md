# Release Checklist

Pre-flight checklist for cutting a Ventiqra release.

## Version

- [ ] Update `VERSION` to the new semver (e.g. `1.1.0`)
- [ ] Add a `## [X.Y.Z]` section to `CHANGELOG.md` with all changes
- [ ] Update `package.json` version in `frontend/`
- [ ] Commit: `chore: bump version to X.Y.Z`

## Backend

- [ ] `cd backend && go build ./...` passes
- [ ] `cd backend && go vet ./...` passes
- [ ] `cd backend && go test ./...` passes (note: 1 pre-existing config test may fail due to env leakage — verify it's not a new failure)
- [ ] No new migrations without a corresponding down/rollback plan
- [ ] All new `With*` options wired in `main.go` and `auth_test.go`

## Frontend

- [ ] `cd frontend && npm run lint` passes (0 errors)
- [ ] `cd frontend && npm run build` passes
- [ ] `cd frontend && npm run test` passes (vitest)
- [ ] All new components have a smoke test or are covered by E2E
- [ ] No `console.log` left in production code

## E2E

- [ ] `cd frontend && npm run test:e2e` runs (skips gracefully if no stack, or passes against a running stack)

## Documentation

- [ ] `README.md` features list is current
- [ ] `docs/API.md` lists all new endpoints
- [ ] `docs/SIMULATION_FORMULAS.md` reflects current tuning
- [ ] `CHANGELOG.md` is updated

## Docker

- [ ] `docker compose --profile app up -d` builds and starts all services
- [ ] `docker compose -f docker-compose.yml -f docker-compose.prod.yml --profile app up -d` runs in prod mode
- [ ] Backend health check passes (`GET /health`)
- [ ] Frontend loads (`GET /`)

## Git

- [ ] Branch `phase_50` merged to `main`
- [ ] Tag the release: `git tag v1.0.0 -m "Release 1.0.0" && git push origin v1.0.0`
- [ ] GitHub Release created with the changelog section as the body

## Post-release

- [ ] Verify the deployed app loads and registration works
- [ ] Verify the dashboard renders with a company
- [ ] Verify a simulation tick advances the day
- [ ] Announce the release
