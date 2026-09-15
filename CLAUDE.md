# Alchemist

Multi-tenant video transcoding and delivery. Bangladesh first, international later.
Go + PostgreSQL + River + SeaweedFS + ffmpeg + shaka-packager, self-hosted on bare
metal. No managed transcoding service, no PaaS.

**Scope: this repo is the transcode and delivery engine only.** The player and the
customer dashboard are separate repositories — see `internal/modules/README.md`.

## Keeping docs true

**Any change that moves, renames, or removes a file, flag, env var, endpoint, or
migration updates the docs that mention it, in the same change.** Stale docs are worse
than none: they read plausibly and send the next person somewhere that no longer
exists.

`llms.txt` is what an agent integrating against this reads first, so a wrong statement
there propagates into someone else's code. Its links are checked by `internal/docscheck`
like any other document; its claims are not, so verify them by hand when behaviour
changes — TTLs, state names, header names, and the lazy-rung threshold especially.

Three tests enforce what can be enforced:

- `internal/docscheck` — every path referenced in any `.md`, and in `llms.txt`, must exist.
- `internal/modules/boundary_test.go` — module boundaries.
- `internal/api/openapi_test.go` — every served route is in `api/openapi.yaml`, and
  every documented route is served. Both directions, so neither can drift.
- `internal/platform/signing/parity_test.go` — the edge's njs signing matches Go.

Prose still needs judgement. After changing behaviour, check:

| Changed | Also update |
|---|---|
| Ladder, codecs, rate control | `docs/02-encoding-pipeline.md`, `ladder_profiles` seed |
| Storage layout, object keys | `docs/01-architecture.md`, `docs/03-cost-model.md` |
| Origin headers, cache keys, edge auth | `docs/01-architecture.md`, `deploy/README.md`, `deploy/edge/nginx.conf` |
| Signing, playback URLs | `deploy/edge/playback_auth.js` **and** its parity test |
| Routes | `api/openapi.yaml` (a test enforces this) |
| Error codes | Bangla copy in `internal/platform/httpx/lang.go` |
| Env vars | `.env.example`, `deploy/README.md` |
| Account/session behaviour | `llms.txt`, `api/openapi.yaml`, `.claude/skills/alchemist-api/SKILL.md` |
| Migrations | `deploy/README.md` |
| Phase scope | `docs/04-roadmap.md` |
| Asset states, URL TTLs, headers | `llms.txt` (claims are not test-enforced) |
| Costs, volumes, BD assumptions | `docs/03-cost-model.md`, `docs/05-bangladesh.md` |

`docs/` and `data/` are gitignored — they are working references, not published.

## Consumers talk to the API only

`~/Sites/academy` and any other product integrate through the public API with an API
key. No shared database, no direct object-storage reads, no internal endpoints. If
something a consumer needs is not reachable through `api/openapi.yaml`, the gap is in
the API and gets fixed there.

This is what surfaced the deduplication bug: from the outside the API reported an
asset `ready` while every playback URL returned 404. A consumer with database access
would have seen rows and assumed it worked.

## Agent-facing documentation

Two files, different jobs, both hand-written and neither generated:

- `llms.txt` — orientation. What this is, what surprises an integrator, links onward.
- `.claude/skills/alchemist-api/SKILL.md` — the procedure. Concrete requests, webhook
  verification, error handling. Loads automatically when an agent works on a video
  integration. A copy is installed at `~/Sites/academy/.claude/skills/` so it loads
  there too; that copy says where the original is.

Their links are checked by `internal/docscheck`; their claims are not. When asset
states, URL TTLs, header names, error codes or the lazy-rung threshold change, update
both by hand — a wrong statement here propagates into someone else's code.

## Related repositories

| Repo | What |
|---|---|
| `~/Sites/alchemist-player` | Embed + JS SDK. Versioned static bundle; embeds pin a major version, so a player change must never require a backend deploy or break live iframes. |

The player integrates only through the public API and signed playback URLs. If its
signing or beacon payload ever diverges from this repo, playback 403s in production
while looking fine locally — the same failure mode `internal/platform/signing/parity_test.go` guards for the edge.

## Architecture

```
platform/   shared infrastructure (db, storage, media, signing, keys, fetch, httpx)
modules/    bounded contexts; never import each other or concrete infrastructure
adapters/   binds platform to the interfaces modules declare; only cmd/ imports it
api/        control plane, composes modules
```

`cmd/alchemist` runs everything; `cmd/alchemist-origin` runs the delivery module
alone. The second binary exists to prove the seam is real — keep it building.

## Non-obvious constraints

These were established by measurement and are expensive to rediscover.

- **Rate control is CRF + VBV, never per-chunk two-pass ABR.** Two-pass makes quality
  oscillate on a chunk-length cycle: glaring on a TV, invisible on a laptop. Verified
  by per-chunk VMAF; max adjacent delta must stay near zero.
- **Every rendition shares an identical GOP grid.** ABR switching and no-re-encode
  stitching both depend on it.
- **One CMAF file per rendition, byte-range addressed.** Not per-segment objects.
  Changing this re-packages the entire library.
- **The media cache key excludes the signature.** Including it would give every viewer
  a private copy and collapse hit rate. The consequence is that `auth_request` MUST run
  before the cache lookup, or a warmed slice is served to anyone. See `deploy/README.md`.
- **Scene threshold is 0.1, not the commonly cited 0.3.** Real cuts measured as low as
  0.04. It is an optimisation with a grid fallback, never a correctness requirement.
- **AV1 is excluded for BD, not deferred.** No hardware decode on the budget Android
  fleet; software decode is worse for the viewer than H.264.
- **SSRF guard validates in `Dialer.Control`**, at connect time, not by pre-resolving —
  otherwise DNS rebinding walks straight through. Its tests assert both that private
  ranges are blocked *and* that public ones are reachable: an early version blocked the
  entire internet and a security-only suite would have passed it.
- **MinIO's open-source binaries are archived** (410 Gone). Dev and prod both use
  SeaweedFS.
- **Content hash is not unique per tenant.** Uploading the same file twice yields two
  assets with independent lifecycles that share the encoding, so the dedup index is a
  lookup index, not a constraint. A unique index there rejects a legitimate upload.
- **The original is deleted once the mezzanine is stored**, unless the tenant pays to
  retain it. Safe only because every rendition, including one generated on demand years
  later, is built from the mezzanine and never the original.
- **Per-job working directories must be keyed by rung, not just by asset.** Deferred
  rungs for one asset are queued together and run concurrently; a shared directory
  means two jobs overwriting each other's mezzanine and chunks, and one's deferred
  cleanup deleting the other's working files. It surfaces as unrelated-looking
  stitch/encode/probe failures that pass on retry.
- **The packager hoists width/height onto the AdaptationSet when there is exactly one
  Representation.** Parsing an MPD for per-rendition height finds nothing in that
  case, which silently drops deferred rungs from DASH.
- **A deduplicated asset owns no media.** Playback resolves through
  `deduplicated_from` to the canonical asset's storage prefix; copying the objects
  instead would make deduplication pointless.
- **Playback encryption is off by default and is not DRM.** With `KEYFORMAT="identity"`
  the key is served from the same signed URL as the segments, so it protects nothing
  the signed URL does not — and cbcs makes the stream unplayable outside Safari,
  because Chrome and Firefox need EME. The signed expiring URL is the access control.
  Tenants can opt in; assets keep whatever they were packaged with.
- **JIT renditions must reuse the asset's existing content key and its stored
  complexity.** A fresh key produces a rung nothing can decrypt (and it fails looking
  like a corrupt file, not a key mismatch); skipping complexity leaves one asset with
  half a per-title ladder and half raw profile bitrates.
- **Adding a rendition rewrites only `master.m3u8`,** composed from rendition rows.
  Re-running the packager over everything would rewrite every `.cmfv`, changing ETags
  and evicting the whole asset from every edge cache.
- **River DAG workflows are a paid feature.** An atomic counter does fan-in.
- **River's default `UniqueOpts.ByState` includes `Completed`,** and completed jobs are
  retained 24h. A recurring job left on the default runs once per day, not once per
  interval. Bucket sync lists its states explicitly for this reason.
- **Cross-tenant sweeps need a `SECURITY DEFINER` function.** RLS is forced, so a
  background job with no tenant in scope silently reads zero rows — it does not error.
  `active_bucket_sources()` and `resolve_api_key()` exist for this.

## Working here

```
make db-reset          schema + River migrations + grants
make dev-account       local test account with the quotas lifted (never run remotely)
make storage           SeaweedFS (dev object storage)
make run               control plane + workers
make test              all Go tests
make edge              local nginx edge cache
make customer-servers  stand-ins for a third-party customer
./test/journey.sh KEY  end-to-end as a customer, public API only
```

Local dev needs `packager` (shaka-packager) and `weed` (SeaweedFS) on PATH; neither is
in Homebrew, both are GitHub release binaries.

Config is env-only and fails loudly at boot on anything missing. Never add a default
that masks an absent variable.

`ALCHEMIST_FETCH_ALLOWLIST` bypasses the SSRF address guard for exact `ip:port` pairs
so a local harness can serve from loopback. Never set it in production.
