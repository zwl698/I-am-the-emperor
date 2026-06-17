# Sanguobaye Visual Polish Progress

## Current Node

- Date: 2026-06-17
- Phase: Visual polish repair pass
- Active task: Visual polish pass verified; next pass can replace SVG portraits with final generated raster portraits if desired.

## Completed Before This Node

- Phase one gameplay slice exists.
- Legacy resource parser and inventory endpoint exist.
- The current UI is functional but visually too placeholder-like.

## This Node Plan

1. Generate one campaign map bitmap and four anime-styled ruler portrait bitmaps.
2. Copy final assets into `web/public/assets`.
3. Wire the map bitmap into Phaser as the background layer.
4. Wire portrait metadata into HUD owner/general rows.
5. Run tests/build and record browser QA status.

## Completed In This Node

- Added `web/public/assets/map/campaign-map.svg`.
- Added anime-styled ruler portrait images under `web/public/assets/portraits`.
- Wired Phaser to use the map image as a base layer with refined route and city overlays.
- Wired HUD owner/general portraits through `web/src/game/portraitRegistry.ts`.
- Improved HUD visual treatment for dark lacquer, gold, and portrait-first game UI.
- Reduced legacy map label clutter by showing city names only and compacting markers/routes for dense maps.
- Matched legacy ruler records by Chinese name so generated ids such as `ruler-0` still load colored portraits.
- `npm test` passed.
- `npm run build` passed.
- `go test ./...` passed.
- Browser QA loaded `http://127.0.0.1:5173/` with no console errors.
- Browser QA confirmed owner portrait `/assets/portraits/dongzhuo.svg` loads for the selected legacy city.
- Browser QA confirmed advancing one month changes `189年 1月` to `189年 2月`.
- Browser QA confirmed 390px mobile viewport has no horizontal overflow.
- Browser screenshot capture timed out in the Browser runtime after the first screenshot, so final proof used DOM/resource/state checks instead of a final emitted screenshot.
