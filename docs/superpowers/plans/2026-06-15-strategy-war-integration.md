# Strategy War Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Integrate the 三国志式战略地图 into the existing war module so old war tactics, diplomacy pressure, event cards, battle reports, and mini-game entries operate on real cities, roads, armies, supply, and fronts.

**Architecture:** Go remains authoritative for simulation and action resolution. The old `WarCampaign` model stays as compatibility/readout state, but `war_tactic` orders now delegate to strategic city and army changes when `Strategy` exists. Vanilla JS renders richer strategic war affordances without adding a build step.

**Tech Stack:** Go backend, standard library tests, vanilla JavaScript UI tests, CSS/DOM strategic map.

---

### Task 1: War Tactics Drive Strategic Entities

**Files:**
- Modify: `backend/game/actions_test.go`
- Modify: `backend/game/engine.go`
- Create or modify helper code in: `backend/game/strategic_actions.go`

- [x] Add failing tests proving `war_tactic` modes update strategic armies/cities as well as `WarCampaign`.
- [x] Run `env GOCACHE=/Users/zhaowenliang/GolandProjects/I-am-the-emperor/.gocache go test ./backend/game -run 'TestWarTactic' -count=1` and verify the new tests fail because `war_tactic` still only updates the abstract campaign.
- [x] Add a strategic bridge for `OrderMobilize`, `OrderCampaign`, `OrderFortify`, and `OrderTruce`.
- [x] `mobilize` should resupply and train the most relevant court army near the campaign front.
- [x] `campaign` should choose an adjacent hostile strategic city and resolve an assault when practical; otherwise it should pressure the target city through losses, siege, and battle logs.
- [x] `fortify` should improve the best matching court front city and lower nearby road/front risk through existing state effects.
- [x] `truce` should lower matching strategic faction threat/relation pressure and set enemy armies to observing/withdrawing status.
- [x] Re-run the focused Go tests and verify they pass.

### Task 2: Strategic Pressure Feeds War, Foreign, and Events

**Files:**
- Modify: `backend/game/strategic_ai_test.go`
- Modify: `backend/game/strategic_ai.go`
- Modify: `backend/game/foreign.go`
- Modify: `backend/game/event_deck.go`
- Modify: `backend/game/events.go`

- [x] Add failing tests proving strategic faction threat contributes to event hand pressure and foreign state pressure.
- [x] Run focused Go tests and verify they fail before implementation.
- [x] Sync strategic faction threat into `ForeignStates` for matching IDs.
- [x] Make war event pressure include strategic battle/front pressure, not only `WarCampaign`.
- [x] Re-run focused Go tests and verify they pass.

### Task 3: Battle Reports Explain Decision Factors

**Files:**
- Modify: `backend/game/strategic_battle_test.go`
- Modify: `backend/game/types.go`
- Modify: `backend/game/strategic_battle.go`
- Modify: `web/strategy-map-ui.js`
- Modify: `web/strategy-map-ui.test.js`

- [x] Add failing tests expecting battle reports to expose factor summaries such as attack score, defense score, supply, terrain, and key reason.
- [x] Run focused Go and JS tests and verify they fail.
- [x] Extend `BattleReport` with a `Factors []string` JSON field.
- [x] Populate factors for capture, repelled, surrender, and enemy capture reports.
- [x] Render factors in the strategic map war report card.
- [x] Re-run focused tests and verify they pass.

### Task 4: Frontend War Entrypoints Prefer Strategic Commands

**Files:**
- Modify: `web/mini-games.test.js`
- Modify: `web/event-hand-ui.test.js`
- Modify: `web/mini-games.js`
- Modify: `web/event-hand-ui.js`

- [x] Add failing tests for low-supply war cards choosing `supply`, hostile adjacency choosing `assault`, and no hostile adjacency choosing `march` or `train`.
- [x] Run `node web/mini-games.test.js` and `node web/event-hand-ui.test.js` and verify the new assertions fail.
- [x] Improve action selection to sort court armies by front urgency instead of choosing the first court army.
- [x] Prefer `supply` for low-grain armies, `assault` for adjacent hostile cities, `march` toward threatened fronts, otherwise `train`.
- [x] Re-run focused JS tests and verify they pass.

### Task 5: Final Verification

**Files:**
- No new files.

- [x] Run `gofmt` on changed Go files.
- [x] Run `env GOCACHE=/Users/zhaowenliang/GolandProjects/I-am-the-emperor/.gocache go test ./...`.
- [x] Run all existing frontend Node tests in `web/*.test.js`.
- [x] Inspect `git diff --stat` and `git status --short` to report exactly what changed.
