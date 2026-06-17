# Sanguobaye Visual Polish Design

## Goal

Repair the weak visual pass while preserving gameplay behavior. The web version should feel like a modern anime-styled Three Kingdoms tactics game rather than a placeholder dashboard.

## Art Direction

- Map: painterly campaign map of late Han China, warm parchment land, jade rivers, misted mountains, fortified city clusters, road networks, no embedded labels.
- Portraits: handsome anime warlord bust portraits with distinct silhouettes, armor, hair, and palette. They should be readable at small HUD sizes and dramatic at detail-panel sizes.
- UI: dark lacquer, parchment glass, jade, cinnabar, brass, and river-blue accents. Keep the playfield clear, with compact HUD panels at edges.

## Implementation Scope

- Add project-local raster assets under `web/public/assets`.
- Use a generated bitmap as the Phaser map base layer.
- Keep Phaser city markers, route lines, labels, and interactions code-native so gameplay data still controls the map.
- Add portrait metadata to the frontend from ruler/general ids.
- Show a large selected-city owner portrait and smaller garrison portraits in the HUD.
- Improve CSS density, contrast, borders, portrait framing, and mobile behavior.

## Non-Goals

- No gameplay rule changes.
- No extraction of original portrait graphics.
- No battle scene art pass yet.
- No fully transparent sprites in this pass; portraits are framed square images for reliability.

## Verification

- `npm test`
- `npm run build`
- `go test ./...`
- Browser or local screenshot check when dev servers can be started.
