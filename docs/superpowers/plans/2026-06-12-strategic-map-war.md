# Strategic Map War Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a playable 三国志-style strategic map layer where foreign war enters city, road, army, supply, and siege gameplay.

**Architecture:** Go owns the simulation state and action resolution. Frontend DOM renders the strategic playfield and sends explicit `ActionRequest` commands to the existing `/actions` endpoint. Event cards and old war panels remain, but military actions route into the strategic map.

**Tech Stack:** Go 1.26 backend, standard `net/http`, vanilla JavaScript, CSS, Node-based frontend tests.

---

### Task 1: Strategic State Model

**Files:**
- Create: `backend/game/strategic_map.go`
- Create: `backend/game/strategic_map_test.go`
- Modify: `backend/game/types.go`
- Modify: `backend/game/engine.go`
- Modify: `backend/game/migration.go`

- [ ] Write tests expecting `GameState.Strategy` to contain at least 12 cities, 14 roads, 5 factions, and 4 armies after coronation.
- [ ] Run the focused Go test and verify it fails because the strategy model is missing.
- [ ] Implement strategic structs, starting state, map connectivity helpers, and `ensureStrategicSystems`.
- [ ] Initialize strategy in new games, coronation, force coronation, and migration repair.
- [ ] Run the focused Go test and verify it passes.

### Task 2: Strategic Actions

**Files:**
- Create: `backend/game/strategic_actions.go`
- Create: `backend/game/strategic_actions_test.go`
- Modify: `backend/game/actions.go`

- [ ] Write tests for `city_develop` relief/farm/fortify and `army_command` train/march/assault.
- [ ] Run focused tests and verify they fail because action kinds are missing.
- [ ] Add action kinds to `ActionCatalog`.
- [ ] Implement `ApplyAction` routing for strategic actions with command point spending and history entries.
- [ ] Sync strategic war outcomes into `Stats.BorderThreat`, `Wars`, and `Crisis`.
- [ ] Run focused tests and verify they pass.

### Task 3: Strategic AI and Seasonal Pressure

**Files:**
- Create: `backend/game/strategic_ai.go`
- Create: `backend/game/strategic_ai_test.go`
- Modify: `backend/game/engine.go`

- [ ] Write tests proving enemy factions pressure front cities and low-supply armies decay during seasonal advancement.
- [ ] Run focused tests and verify they fail.
- [ ] Implement `applyStrategicPressure(domain Domain)` and call it from `applyWorldPressure`.
- [ ] Add strategic log entries for AI pressure, supply loss, and front-line shifts.
- [ ] Run focused tests and verify they pass.

### Task 4: Strategic Map UI

**Files:**
- Create: `web/strategy-map-ui.js`
- Create: `web/strategy-map-ui.test.js`
- Create: `web/strategy-map.css`
- Modify: `web/index.html`
- Modify: `web/app.js`
- Modify: `web/panel-renderers.js`
- Modify: `web/app-contract.test.js`

- [ ] Write JS tests expecting city nodes, road lines, army pieces, city action buttons, and army action buttons.
- [ ] Run focused frontend test and verify it fails because the UI module is missing.
- [ ] Implement the map renderer, action buttons, and responsive CSS.
- [ ] Wire the module into `index.html`, `els`, `normalizeGame`, and `renderExternalPanels`.
- [ ] Run focused frontend tests and verify they pass.

### Task 5: Foreign War Entry

**Files:**
- Modify: `web/event-hand-ui.js`
- Modify: `web/event-hand-ui.test.js`
- Modify: `web/mini-games.js`
- Modify: `web/mini-games.test.js`

- [ ] Write tests proving war event cards and war mini-game buttons emit strategic `army_command` or `siege_command` actions.
- [ ] Run tests and verify they fail with old action routing.
- [ ] Update war suggestions to target strategic armies and front cities.
- [ ] Run tests and verify they pass.

### Task 6: Verification

**Files:**
- No new files.

- [ ] Run `/opt/homebrew/bin/gofmt` on changed Go files.
- [ ] Run `env GOCACHE=/Users/zhaowenliang/GolandProjects/I-am-the-emperor/.gocache /opt/homebrew/bin/go test -count=1 ./...`.
- [ ] Run all frontend Node tests with `/Users/zhaowenliang/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin/node`.
- [ ] Start the Go server and browser-smoke test: create game, advance to emperor, verify map, click a war action, verify command count and map state change.
