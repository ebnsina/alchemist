# Alchemist

Multi-tenant video transcoding and delivery. Bangladesh first, international later.
Go + PostgreSQL + River + SeaweedFS + ffmpeg + shaka-packager, self-hosted on bare
metal. No managed transcoding service, no PaaS.

**Scope: one repository, three workspaces.** The engine is Go at the root; `player/`
and `web/` are npm workspaces. They were three repositories until the relative
dependency between two of them made a lone clone unbuildable — see `player/` and
`web/` below, and `internal/modules/README.md` for how a Go module would be pulled
out if one ever needs to be.

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
- `internal/modules/ownership_test.go` — a module names only the tables it owns.
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
| Live ingest, ports, segments | `docs/06-live.md`, `deploy/README.md` |
| Asset states, URL TTLs, headers | `llms.txt` (claims are not test-enforced) |
| Costs, volumes, BD assumptions | `docs/03-cost-model.md`, `docs/05-bangladesh.md` |
| Dashboard routes, copy, components | `web/CLAUDE.md`, `web/DESIGN.md`, `web/CHANGELOG.md` |
| Player behaviour, SDK surface | `player/README.md`, `player/CHANGELOG.md` |

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

## Workspaces

| Path | What |
|---|---|
| `player/` | Embed + JS SDK, published as `@alchemist/player`. Versioned static bundle; embeds pin a major version, so a player change must never require a backend deploy or break live iframes. |
| `web/` | Marketing site and customer dashboard (SvelteKit, adapter-node). Talks to the engine only through the public API. |

npm workspaces from the root: `npm install` once, `npm run build` builds the player
then the web. `web` depends on `@alchemist/player` as a workspace, so a player change
is picked up without publishing. Both share one vite and one TypeScript — two copies
make svelte-check fail on types that are structurally identical.

The player integrates only through the public API and signed playback URLs. If its
signing or beacon payload ever diverges from the engine, playback 403s in production
while looking fine locally — the same failure mode `internal/platform/signing/parity_test.go` guards for the edge.

## Architecture

```
platform/   shared infrastructure (db, storage, media, signing, keys, fetch, httpx)
modules/    bounded contexts; never import each other or concrete infrastructure
adapters/   binds platform to the interfaces modules declare; only cmd/ imports it
api/        control plane, composes modules
```

`internal/modules` holds `delivery` (playback origin) and `live` (broadcast ingest).
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
- **Frame rate is chosen once, in `MezzanineRate`, and carried.** The GOP grid, the
  chunk plan, the params hash and VMAF's frame-to-time mapping are all whole frames of
  that rate, so it cannot be re-derived per stage. 24 and 25 survive, 48/50/60/120
  halve back onto the set, and anything under 24 is left alone rather than inflated.
  High frame rate is not offered: it is bitrate the budget Android fleet cannot spend.
- **Subtitles are sidecar WebVTT, referenced from the manifests, never burned in.**
  Burning in means re-encoding the whole ladder per language and gives the viewer no
  way to turn them off. The consequence is that `signManifest` must know which VTT it
  is holding: the scrubbing index's cue payloads are image URLs and need signing, a
  subtitle's cue payload is the sentence on screen and must not be touched.
- **Expensive stages write to `.partial` and rename.** A file under its final name has
  therefore finished, which is what lets a retry skip it — and the rename is what stops
  a truncated file from a killed process being accepted and stitched. The working
  directory survives between attempts for the same reason, and is removed only when
  there is nothing left to resume.
- **Every dimension comes from the mezzanine, never from the source probe.** ffmpeg
  applies a rotation matrix while building the mezzanine, so a phone clip that probes
  1920x1080 with `rotate:90` is encoded 1080x1920. Deriving a rendition's width from
  the source probe transposes it, and that width is what `republish` writes as the
  HLS `RESOLUTION` — HLS and DASH then disagree about the same rung. `Probe.Rotation`
  is read by nothing and is not the fix; using `MezzProbe` is.
- **Deferred rungs are filtered by source height too.** `Applicable` runs inside
  `Transcode` on the eager set only, so taking `LazyRungs` off the raw profile pends a
  720p row on a 360p upload: the asset never leaves `partially_ready`, its mezzanine is
  retained and billed for life, and first playback queues a JIT job that upscales.
  Nothing errors anywhere, which is why it survived.
- **The packager hoists width/height onto the AdaptationSet when there is exactly one
  Representation.** Parsing an MPD for per-rendition height finds nothing in that
  case, which silently drops deferred rungs from DASH.
- **A deduplicated asset owns no media, no renditions and no content key.** All
  three resolve through `deduplicated_from` to the canonical asset; copying the
  objects instead would make deduplication pointless, and copying the rows meant a
  deferred rung was encoded under the duplicate's id and published under a prefix no
  playback request resolves to. Deleting the canonical asset therefore promotes a
  surviving duplicate and moves those rows onto it: `content_keys` and `renditions`
  cascade from `assets`, so nulling the pointers instead takes the key and the ladder
  with it while the objects, and playback, carry on.
- **Money is integer minor units everywhere it is stored or charged; rates are not.**
  Unit rates are fractions of a paisa per megabyte, so they stay exact `big.Rat` and
  only the line total rounds, half-up, once. A `float64` anywhere in that path bills
  the difference forever: 0.016 has no exact binary representation, and the rate card
  is parsed with `json.Number` for that reason alone.
- **Suspension refuses ingest and never playback.** Taking a customer's viewers offline
  over an unpaid invoice punishes people who are not party to it, and it is the one
  part of a suspension that paying cannot undo.
- **SSLCommerz has no merchant-initiated charge in its published API.** It is hosted
  checkout, so it does not implement `payments.AutoCharger` and a BDT invoice is
  collected by a link the customer opens. Do not "fix" that by inventing a recurring
  endpoint; it needs a separate agreement with SSLCommerz.
- **A verified webhook is not a paid invoice.** SSLCommerz's `verify_sign` proves the
  POST was not edited, not that money moved — the validation API is the authoritative
  answer. And an event whose amount is under the invoice total is refused, or a
  one-taka payment clears a 1,560-taka bill. The gateway's own reference is a unique
  key on `payments`, so an IPN and a redirect describing one payment credit it once.
- **At rest is the storage layer's job, not the video's.** SeaweedFS runs with
  `-s3.encryptVolumeData` in dev and production alike, so a stolen disk or backup
  decodes to nothing for every viewer on every platform. It does not survive leaked S3
  credentials — the API's job is to hand back plaintext to valid credentials — which is
  the one thing playback encryption did cover, and the trade that was taken knowingly.
  Losing the filer's metadata loses the keys: back it up like PostgreSQL.
- **Playback encryption is off by default since 043, and was never DRM.** Clear Key is
  what the W3C EME specification calls the baseline key system for interoperability
  testing, not a content protection system: the key reaches the browser in the clear
  behind the signed URL, so it never stopped an entitled viewer keeping a copy. What it
  buys is encryption at rest, and that threat is answered underneath by encrypting the
  volumes, which costs no viewer anything — rather than at the video layer, which costs
  every Apple viewer. Real content protection is multi-DRM: an Apple certificate, a
  licence vendor, and `cbcs` instead of `cenc`. It belongs to whoever asks for it.
- **Playback encryption, when a tenant turns it on, is still not DRM.** Media is packaged
  `cenc` and `/playback/.../key` answers an EME Clear Key licence, which Chrome,
  Firefox and Edge play with no licence vendor — that is what made the old cbcs
  default unusable and it is fixed. The key still reaches the browser in the clear
  behind the signed URL, so what this buys is encryption at rest: a lifted bucket or
  backup decodes to nothing. It does not stop a viewer who is entitled to watch.
  Never write a comment, doc line or message claiming otherwise. Safari and iOS
  cannot play it at all (FairPlay is WebKit's only key system) and get
  `browser_not_supported` rather than a black screen. Migration 033 changed the
  default only: existing tenants and every published asset keep what they had.
- **A per-tenant rule the edge must enforce has to travel inside the token.** njs
  verifies in the worker with no database, and a warm cache slice never reaches the
  origin at all, so anything looked up per tenant stops applying the moment the cache
  fills. `org` is signed for the same reason `vid` and `wm` are. The corollary: the
  API resolves the origin against `tenants.playback_origins` **before** signing,
  because nothing downstream re-checks it.
- **An origin-locked link cannot play in a native app.** The lock is read from `Origin`
  or `Referer` and an app sends neither; accepting a request with neither would be the
  whole bypass. Empty `playback_origins` is therefore the correct setting for a tenant
  with its own app, not an oversight.
- **Sharing is answered by binding, not by encryption.** `?viewer=` mints a link whose
  signature covers the viewer id and the watermark label, so neither can be edited out,
  and `tenant_limits.max_viewer_devices` caps concurrent devices per viewer at the
  origin. The cap cannot move to the edge: njs validates a signature with no shared
  state, and counting devices needs state. Binding does **not** touch the media cache
  key — `$uri$slice_range` still excludes the query, so viewers share every slice.
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
- **An encoder exiting is not proof a broadcast is over.** A dropped uplink and a
  presenter closing OBS are the same event from the worker's side, so a broadcast that
  has been on air waits `ReconnectGrace` for the encoder to come back. ffmpeg resumes
  into the existing playlist with `-hls_flags append_list` — verified against 9.0.1;
  `discont_start` adds a second DISCONTINUITY at the top of the playlist where there
  is none, and `-start_number` makes it skip to a number neither side expects. ENDLIST
  is stripped from every playlist published before the end, because a player that has
  seen it does not come back when the segments resume.
- **A recording is assembled per encoder connection, never as one byte append.** A
  reconnect restarts timestamps at zero, so appending across the DISCONTINUITY makes a
  file whose timeline runs backwards and everything downstream believes the shorter
  duration: a ten-second test broadcast with one reconnect produced a six-second
  mezzanine, with no error anywhere. `PlaylistRuns` groups the segments and the concat
  demuxer joins them.
- **Live advertises HLS and nothing else.** The live prefix holds a playlist, an init
  segment and media segments — no MPD, no poster, no scrubbing index. It matters most
  in `live_ended`: the conversion writes the content key before it writes any media, so
  `encrypted` flips true while the prefix is still `live/`, and preferring DASH on that
  basis points the player at a manifest that is not there for the whole conversion.
- **One realtime rung, but not the cheapest one.** A core goes on encoding in realtime
  at all, not on the pixel count, so 144p and 360p cost nearly the same and 144p is
  unreadable for the slides and whiteboards this is used for. `LiveMaxHeight` caps it.
- **Live cannot call `media.Package()`.** `Package()` shells out once with
  `CombinedOutput()` over complete files and writes one byte-range CMAF file per
  rendition. A live packager never exits, cannot read a file still being written, and
  cannot append to an S3 object; invoking it per segment restarts the timeline every
  two seconds. Live segments with ffmpeg instead, in `internal/modules/live`. The
  roadmap's "packager as a library" line overstates what the code does.
- **Live objects are a working set, not a library.** Live writes per-segment objects
  under `live/{tenant}/{asset}/`, which is the opposite of the byte-range CMAF rule.
  That holds only because nothing written there survives: the recording converts into
  the ordinary `cmaf/` layout and the live prefix is swept. Never serve the permanent
  library from small objects.
- **A converting recording must stay in `live_ended` until its `cmaf/` objects exist.**
  `DedupResolver.StoragePrefix` switches prefix on that column, so writing `encoding`
  over it points every viewer at objects that are not there yet — a 404 partway through
  a link the customer handed out before the match. `setState` and `markFailed` in
  `internal/pipeline/pipeline.go` both exclude it; the only write that moves the asset
  is the final one, which already runs after the upload.
- **A session becomes a tenant in exactly one place: `auth_session()`.** Platform-admin
  impersonation is resolved there, so every endpoint acts on the customer's account
  with no second code path, and the flag is re-read on every request — revoking it
  ends impersonation on the next call rather than when the session expires. While
  impersonating the session is read-only, enforced in `authenticate`, not per handler.
- **Cross-tenant sweeps need a `SECURITY DEFINER` function.** RLS is forced, so a
  background job with no tenant in scope silently reads zero rows — it does not error.
  `active_bucket_sources()`, `resolve_api_key()` and `live_tenants()` exist for this.

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
make live-journey KEY=...  a real broadcast, with a mid-class encoder drop
```

Local dev needs `packager` (shaka-packager) and `weed` (SeaweedFS) on PATH; neither is
in Homebrew, both are GitHub release binaries.

Config is env-only and fails loudly at boot on anything missing. Never add a default
that masks an absent variable.

`ALCHEMIST_FETCH_ALLOWLIST` bypasses the SSRF address guard for exact `ip:port` pairs
so a local harness can serve from loopback. Never set it in production.
