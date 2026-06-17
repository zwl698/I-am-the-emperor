# Sanguobaye Web Remake Design

## Goal

Build a modern web-only remake of the legacy `sanguobaye_c-master` game in the current `new_sanguoby` directory. The remake must preserve the original gameplay rules one subsystem at a time while replacing the old electronic-dictionary presentation with a readable modern browser UI.

## Source Of Truth

The sibling legacy project remains the gameplay reference:

- Main engine loop: `../sanguobaye_c-master/src/gamEng.c`
- Campaign turn loop: `../sanguobaye_c-master/src/tactic.c`
- Monthly city/person updates: `../sanguobaye_c-master/src/infdeal.c`
- Battle loop and command execution: `../sanguobaye_c-master/src/Fight.c`
- Core data structures: `../sanguobaye_c-master/src/baye/attribute.h`, `../sanguobaye_c-master/src/baye/order.h`, `../sanguobaye_c-master/src/baye/fight.h`
- Resource archive: `../sanguobaye_c-master/src/dat.lib.orig`

The remake should not run `eval`-style mod scripts from the legacy HTML5 port. Data and rules are migrated into explicit Go and TypeScript modules.

## Phase One Scope

Phase one creates the playable foundation, not the whole game:

- Go 1.26 backend with deterministic campaign state.
- HTTP API for health, new game, current game state, and month advancement.
- Web client using React, Vite, TypeScript, and Phaser.
- DOM HUD for date, ruler, selected city, resources, and command buttons.
- Phaser-rendered campaign map with city markers, routes, terrain color, and selection.
- A small authored seed scenario that follows the original structure: rulers, cities, resources, population, agriculture, commerce, soldiers, and month progression.
- Tests for backend monthly rules and frontend map projection.
- Build and smoke-test proof that the app launches and the primary UI renders.

Full command parity, complete resource extraction, battle tactics, AI diplomacy, save slots, and generated portrait/tile asset replacement are later phases.

## Gameplay Rules In Phase One

The first implemented rule slice mirrors the legacy monthly update pattern:

- Month increases by one.
- After month 12, year increases and month resets to 1.
- Every month, generals recover 4 stamina up to 100.
- Every month, each populated city gains 50 population up to its limit.
- Every third month, city money increases by half of commerce, capped at 30000.
- In months 6 and 10, city food increases by one quarter of farming, capped at 30000.
- Each month, city food pays troop upkeep based on total soldiers divided by 50.
- If food cannot cover upkeep, city food becomes 0, city state becomes famine, and city generals lose half their soldiers.

The exact C-compatible edge cases will be expanded through regression tests as each subsystem is ported.

## Architecture

The backend owns simulation state and rules. The frontend never mutates gameplay state directly; it sends commands to the backend and renders the returned snapshot.

```text
backend/
  cmd/server/            HTTP entrypoint
  internal/game/         campaign state, rules, API DTOs
  internal/server/       routes and JSON handlers

web/
  src/api/               typed backend client
  src/game/              renderer-independent helpers
  src/phaser/            Phaser scene bridge and map rendering
  src/ui/                React DOM HUD and shell
  src/styles.css         design tokens and responsive layout
```

The renderer boundary is intentional: Phaser owns visual map objects and camera behavior, React owns text-heavy UI, and Go owns gameplay truth.

## Visual Direction

The first web screen is a modern campaign command surface:

- The map is a warm parchment strategic map with rivers, mountains, roads, and city markers.
- HUD uses ink, jade, cinnabar, brass, and river-blue accents.
- The center of the map remains clear; status and commands live on the right and bottom edges.
- UI panels are compact and game-like, not a SaaS dashboard.
- Text is code-native Chinese so it stays readable and accessible.
- Old black-and-white graphics are not reused as final art.

Phase one uses code-native terrain and SVG-style markers so the project can self-test without waiting on the full art pipeline. Later phases replace these with generated map tiles, city emblems, army sprites, and portraits behind the same manifest keys.

## API Contract

- `GET /api/health`
  - Returns service name and status.
- `POST /api/games`
  - Creates a new game for a requested scenario and ruler.
  - Returns the full game snapshot.
- `GET /api/games/current`
  - Returns the current in-memory game snapshot.
- `POST /api/games/current/advance-month`
  - Advances the current game by one month and returns the new snapshot.

Phase one stores a single in-memory game. Persistent save slots are a later phase.

## Testing Strategy

- Go unit tests cover deterministic campaign rules.
- TypeScript unit tests cover renderer-independent map coordinate projection.
- Build checks compile backend and frontend.
- Browser smoke testing verifies the web app loads, is not blank, and primary controls update visible state.

## Acceptance Criteria

Phase one is acceptable when:

- `go test ./...` passes.
- Frontend unit tests pass.
- Frontend production build passes.
- The Go server serves API snapshots.
- The web client renders the campaign map and DOM HUD.
- Selecting a city updates the selected city panel.
- Advancing a month updates the visible date/resources from backend state.
- A desktop and mobile viewport smoke test shows no blank screen, framework overlay, or obvious layout overlap.
