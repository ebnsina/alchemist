# Modules

A modular monolith. One binary today, separable services when volume or team shape
demands it — without a rewrite.

```
internal/
  platform/    shared infrastructure: db, storage, media, signing, keys, fetch, httpx
  modules/     bounded contexts, each owning its domain
  adapters/    binds platform implementations to interfaces modules declare
  api/         control-plane HTTP surface, composes modules
cmd/
  alchemist         everything in one process (the default)
  alchemist-origin  delivery module alone (proof the seam is real)
```

## The rules

1. **A module never imports another module.** Cross-module needs go through an
   interface the *consumer* declares, wired in `cmd/` via `internal/adapters`.
2. **A module never imports concrete infrastructure.** It declares the narrow slice it
   needs (`ObjectStore`, `ContentKeys`) and an adapter satisfies it. Two exceptions are
   allowed because they carry no dependencies: `platform/httpx` and `platform/signing`.
3. **Only `cmd/` imports `adapters`.** A module importing adapters inverts the
   dependency and makes it unextractable.

`boundary_test.go` enforces all three at build time. Without it the boundary erodes
within weeks: one "just this once" import turns a copy-out into a refactor, quietly,
with nothing failing. The test is the boundary; the README is only a description of it.

## Extracting a module into its own repository

Delivery is the worked example, because it is the likeliest split: it serves every
byte, holds almost no business logic, and wants to run near storage and the edge — a
completely different scaling profile from the control plane.

1. Copy `internal/modules/<name>/`.
2. Copy the adapters it uses from `internal/adapters/` (delivery needs `ObjectStore`
   and `ContentKeys`).
3. Copy the platform packages those adapters touch — or better, publish `platform/` as
   a shared library and depend on it from both repos.
4. Copy `cmd/alchemist-<name>/main.go`, which already exists for delivery and composes
   exactly these pieces.
5. Point it at the same PostgreSQL and object storage. No schema change is required:
   modules already read and write only their own tables.

What makes this mechanical rather than a refactor is that step 1 pulls in no
surprises. The boundary test is what guarantees that.

## Current modules

| Module | Owns | Tables | Extract when |
|---|---|---|---|
| `delivery` | playback origin, manifests, content keys, edge auth | `content_keys` | Request volume outgrows the API, or it needs to sit beside the BDIX edge. **Most likely first split.** |

Everything else still lives in `api/` and `pipeline/`. Carving out `assets`,
`webhooks`, `usage` and `transcode` follows the same shape; delivery went first
because it is the one whose scaling profile actually diverges.
