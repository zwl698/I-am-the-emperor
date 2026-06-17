# Sanguobaye Web Remake Phase Two Plan

> **For agentic workers:** Continue from phase one. Keep this phase small and test-driven: parse legacy resources first, then use them to replace authored seed data in later slices.

## Goal

Build the first migration bridge from `../sanguobaye_c-master/src/dat.lib.orig` into the Go remake, without changing gameplay behavior yet.

## Scope

- Add a Go package that reads the legacy `dat.lib` archive index, resource headers, fixed-length items, variable-length items, and resource-key decryption.
- Add regression tests against the real sibling `dat.lib.orig` when it is present.
- Add a backend endpoint that reports a resource inventory for development and migration checks.
- Add a compact frontend HUD indicator so the web build visibly confirms that the remake can see the old data source.

## Legacy Facts

- Resource start address is stored at `(resourceID - 1) * 4`.
- `0xffffffff` means a missing resource address.
- `dat.lib.orig` uses little-endian fields for `RCHEAD`:
  - `U32 ResLen`
  - `U16 ResId`
  - `U16 ItmCnt`
  - `U16 ItmLen`
  - `U8 ResKey`
  - `U8 Reserved`
- `dat.lib.orig` uses little-endian fields for `RIDX`:
  - `U16 offset`
  - `U16 rlen`
- The current C port declares wider fields, but the archived binary follows the 12-byte header and 4-byte item index layout above.
- Encrypted resources are stored with `+ResKey`; loading subtracts `ResKey` from each byte.

## Tasks

### Task 1: Progress Checkpoint

- [ ] Create/update a progress md file with the completed phase-one node and current phase-two node.

### Task 2: Legacy Archive Parser

- [ ] Write tests for opening `dat.lib.orig`, listing known resources, reading fixed-length city names, reading fixed-length general scenario data, and reading variable-length items.
- [ ] Run the parser tests and confirm they fail because the package is not implemented.
- [ ] Implement the parser with bounds checking and readable errors.
- [ ] Run parser tests and confirm they pass.

### Task 3: Backend Inventory API

- [ ] Write handler tests for `GET /api/legacy/resources`.
- [ ] Run handler tests and confirm they fail because the route is missing.
- [ ] Implement route and response types.
- [ ] Run backend tests and confirm they pass.

### Task 4: Web Inventory Indicator

- [ ] Write a small unit test for legacy inventory summarization.
- [ ] Run frontend tests and confirm the new test fails before implementation.
- [ ] Implement API typing, fetch wiring, and HUD display.
- [ ] Run frontend tests and build.

### Task 5: Verification

- [ ] Run `go test ./...`.
- [ ] Run frontend tests.
- [ ] Run frontend production build.
- [ ] Start backend and frontend dev servers.
- [ ] Smoke test the browser for map render, month advance, and visible legacy resource status.
- [ ] Update the progress md with the final verified node and any limitations.

## Out Of Scope

- Decoding all legacy GBK strings into gameplay DTOs.
- Replacing phase-one authored city/general seed data.
- Battle map extraction and sprite regeneration.
- Save slots and AI command parity.
