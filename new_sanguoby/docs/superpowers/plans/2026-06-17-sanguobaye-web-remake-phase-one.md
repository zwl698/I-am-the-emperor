# Sanguobaye Web Remake Phase One Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first testable web-only remake slice: Go backend campaign state, React/Vite/Phaser client, strategic map, HUD, and month advancement.

**Architecture:** Go owns deterministic simulation and JSON APIs. React owns app shell and dense HUD. Phaser owns campaign map rendering and reports city selection back to React.

**Tech Stack:** Go 1.26, net/http, React 19, Vite 8, TypeScript 6, Phaser 4, Vitest, Playwright for smoke checks when Browser tooling is unavailable.

---

## File Structure

- Create `go.mod`: Go module definition.
- Create `backend/cmd/server/main.go`: backend entrypoint.
- Create `backend/internal/game/types.go`: campaign structs and API snapshot types.
- Create `backend/internal/game/seed.go`: phase-one scenario data.
- Create `backend/internal/game/rules.go`: month advancement rules.
- Create `backend/internal/game/rules_test.go`: backend rule tests.
- Create `backend/internal/server/server.go`: HTTP router and handlers.
- Create `web/package.json`, `web/tsconfig.json`, `web/vite.config.ts`, `web/index.html`: frontend tooling.
- Create `web/src/api/client.ts`, `web/src/api/types.ts`: typed API layer.
- Create `web/src/game/mapProjection.ts`, `web/src/game/mapProjection.test.ts`: renderer-independent map math.
- Create `web/src/phaser/CampaignMap.tsx`, `web/src/phaser/scenes/CampaignScene.ts`: Phaser map bridge.
- Create `web/src/ui/AppShell.tsx`, `web/src/ui/Hud.tsx`: React shell and HUD.
- Create `web/src/main.tsx`, `web/src/styles.css`: frontend entry and visual system.
- Create `web/public/assets/city-marker.svg`, `web/public/assets/army-banner.svg`: first replaceable visual assets.

## Tasks

### Task 1: Backend Rule Foundation

- [ ] Write `backend/internal/game/rules_test.go` with tests for month increment, quarterly money, harvest months, population growth, stamina recovery, and famine upkeep.
- [ ] Run `go test ./backend/internal/game` and confirm it fails because the package is not implemented.
- [ ] Implement `types.go`, `seed.go`, and `rules.go`.
- [ ] Run `go test ./backend/internal/game` and confirm it passes.

### Task 2: Backend HTTP API

- [ ] Write handler tests for health, new game, current game, and advance month.
- [ ] Run `go test ./backend/internal/server` and confirm it fails because handlers are missing.
- [ ] Implement `backend/internal/server/server.go` and `backend/cmd/server/main.go`.
- [ ] Run `go test ./backend/internal/server ./backend/internal/game` and confirm it passes.

### Task 3: Frontend Tooling And Map Math

- [ ] Create frontend package files.
- [ ] Write `mapProjection.test.ts` for stable city-grid to viewport projection.
- [ ] Run frontend tests and confirm the projection test fails before implementation.
- [ ] Implement `mapProjection.ts`.
- [ ] Run frontend tests and confirm they pass.

### Task 4: React And Phaser Campaign Screen

- [ ] Implement typed API client.
- [ ] Implement React app shell and HUD.
- [ ] Implement Phaser campaign scene that draws terrain, routes, city markers, and army banners from backend state.
- [ ] Wire city selection from Phaser to React.
- [ ] Wire advance-month button to backend API.

### Task 5: Verification

- [ ] Run `go test ./...`.
- [ ] Run frontend unit tests.
- [ ] Run frontend production build.
- [ ] Start the Go server and frontend dev server.
- [ ] Smoke test desktop viewport: app loads, map visible, selected city changes, advance month changes date.
- [ ] Smoke test mobile viewport for no major overlap.
- [ ] Record any limitations that remain outside phase-one scope.

## Self-Review

- Spec coverage: phase-one backend state, API, React shell, Phaser map, HUD, tests, and smoke testing are covered by the tasks.
- Placeholder scan: no task depends on undefined later work.
- Type consistency: frontend and backend both use the same snapshot shape: `GameSnapshot`, `City`, `Ruler`, `General`, and `GameDate`.
