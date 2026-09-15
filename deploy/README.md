# Deploying Alchemist

Bare metal, systemd, no container runtime in production. Three roles; at BD phase-1
volume the first two share one machine.

| Role | Where | What runs |
|---|---|---|
| Control plane + workers | Singapore | `alchemist.service`, PostgreSQL |
| Storage | Singapore | SeaweedFS (master, volume, filer, S3) |
| Edge cache | Dhaka, inside BDIX | nginx (`deploy/edge/`) |

## Why the origin is not in Dhaka

Durable, stateful, hard-to-replace things go where operations are boring. The Dhaka
node is a stateless cache: if its power fails, BD traffic falls back to Singapore --
slower and metered, but working. That is an acceptable failure mode for a cache and
would not be for storage.

Invert this only if outward remittance makes paying a Singapore provider impractical,
in which case Dhaka colo paid in BDT wins despite the operational cost. See
`docs/05-bangladesh.md`.

## Splitting the origin out

`alchemist` runs everything in one process, which is right at BD phase-1 volume. When
request volume outgrows the control plane, `alchemist-origin` serves the delivery
plane alone from the same codebase and config — no code change, no schema change:

```
# Singapore, beside storage; or wherever the edge fills from
ALCHEMIST_HTTP_ADDR=:8080 /opt/alchemist/bin/alchemist-origin
```

It mounts only `/playback/...` and `/internal/verify-playback`, and needs the same
`ALCHEMIST_PLAYBACK_KEYS` as the control plane so signatures verify. Point the edge's
`upstream origin` at it and scale it on request volume independently of the API.
`internal/modules/README.md` covers moving it to its own repository entirely.

## Control plane

```
useradd --system --home /opt/alchemist alchemist
install -d -o alchemist -g alchemist /opt/alchemist/bin /var/lib/alchemist /etc/alchemist
install -m 0755 bin/alchemist /opt/alchemist/bin/
install -m 0640 -o root -g alchemist alchemist.env /etc/alchemist/
for m in internal/platform/db/migrations/*.sql; do psql -d alchemist -f "$m"; done
river migrate-up --database-url "$ALCHEMIST_DATABASE_URL"
systemctl enable --now alchemist
```

Migrations are plain SQL applied in filename order; there is no migration tool and no
version table, so each is written to be safe to re-run. `internal/platform/db/migrations/020_members_branding.sql`
also repairs the `users` and `sessions` policies from `internal/platform/db/migrations/017_accounts.sql`, which named a setting
nothing sets and so matched no rows. `internal/platform/db/migrations/027_dedup_renditions.sql` deletes rows rather than
adding a column: deduplicated assets carried copies of the canonical asset's rendition
rows, and those copies are now resolved rather than stored.
`internal/platform/db/migrations/028_usage_bytes.sql` adds `tenant_stored_bytes()`, another
`SECURITY DEFINER` reader for a cross-tenant job, and the unique index the daily egress
and storage rows upsert onto -- without it every flush inserts a new row instead of
folding into the day.

`ALCHEMIST_ENCODE_WORKERS` should be roughly the core count. Encoding already uses a
per-job worker pool, so setting it far above that only lengthens the tail.

## Live ingest

Off unless both `ALCHEMIST_LIVE_INGEST_HOST` and `ALCHEMIST_LIVE_PORT_RANGE` are set;
one without the other refuses to boot. With neither, the `/v1/live-streams` endpoints
are not served at all rather than served and always failing.

`ALCHEMIST_LIVE_PORT_RANGE` is `low-high`, one port per armed stream, so the range
size is the number of concurrent broadcasts this box accepts. ffmpeg in listener mode
takes one connection and cannot dispatch on SRT's streamid, which is why streams
cannot share a port.

**Open the range to your customers' encoders and to nobody else.** With the direct
ffmpeg listener, ingest is authorised by the port rather than by the stream key,
because ffmpeg never exposes the streamid it was handed. Both TCP (RTMP) and UDP (SRT)
need the range open, and SRT needs an ffmpeg built with `--enable-libsrt` -- the
distribution and Homebrew builds generally are not, and without it only RTMP works.

### Enabling live for a customer

Configuring an ingest host enables live for the **deployment**. It does not give it to
anyone: Live is a separate product, `tenant_limits.live_enabled` is false by default,
and a tenant with no limits row has it off. Turn it on per customer:

    PUT /admin/tenants/{id}/live   {"enabled": true}

Operator surface, behind `ALCHEMIST_ADMIN_KEY`. A VOD-only tenant calling the live
endpoints gets `live_not_enabled` rather than a stream nobody sold them.

### Authorising publishers properly

`deploy/live/mediamtx.yml` puts an ingest server in front, which is what makes the
stream key the credential instead of the port. It reads SRT's streamid and RTMP's
query at handshake and asks the API about every publish:

    POST /internal/live/authorize   ->  204 allowed, 401 refused

**That endpoint carries no API key and must never be publicly reachable.** The caller
is the ingest server on the same host, so bind it to loopback or a private interface
and firewall it like a database port. The stream key in the request body is the
credential, and a publish is allowed only when the key resolves to a live stream *and*
that stream is the path being published to -- checking the key alone would let one
customer's valid key publish over everyone else's broadcast.

With the ingest server in front, one RTMP port and one SRT port serve every stream, so
`ALCHEMIST_LIVE_PORT_RANGE` stops being the concurrency ceiling. Cores still are.

**Running a different ingest server.** The rule -- the key resolves to a live stream,
and that stream is the path being published to -- is one function, and each server gets
a small handler around it. They cannot share an endpoint: MediaMTX reads `204` as yes
and `401` as no, while SRS requires `200` with a body of `0` and treats a bare `204` as
a refusal. So swapping means one new handler on its own route plus a config file for
that server, not a rewrite. There is no selector and no plugin layer, because running
two at once is not a thing anyone wants.

One live rung is roughly one CPU core held for the length of the broadcast. Size
`ALCHEMIST_LIVE_PORT_RANGE` against cores, not against ambition: a box with 16 cores
does not run 100 concurrent streams. See `docs/06-live.md`.

## Edge

The edge needs njs, which ships as a standard package -- do not build nginx by hand:

```
apt install nginx libnginx-mod-http-js
install -D -m 0644 deploy/edge/playback_auth.js /etc/nginx/njs/playback_auth.js
install -D -m 0644 deploy/edge/nginx.conf /etc/nginx/nginx.conf
sed -i "s/ORIGIN_HOST/origin.internal/" /etc/nginx/nginx.conf
install -d -o www-data /var/cache/nginx/media /var/cache/nginx/manifest
nginx -t && systemctl reload nginx
```

Size `proxy_cache_path max_size` to about 80% of the cache disk. The working set is
far smaller than the library: a small fraction of assets drives most views, which is
the same power law that justifies JIT packaging.

## Viewer binding, the device cap, and the cache

`GET /v1/assets/{id}?viewer=<opaque-id>&watermark=<label>` mints links carrying `vid`
and `wm`. Both are inside the HMAC, so a viewer cannot edit their id or their watermark
out of the URL, and the njs at the edge hashes them the same way -- `internal/platform/signing/parity_test.go`
covers bound links as well as plain ones.

A link with no binding signs the original `{prefix}|{exp}` and nothing more, so tokens
issued before this existed keep verifying. That is what stops a deploy 403ing every
session in flight for the length of a token TTL.

**This does not change the media cache key.** Media is still keyed on
`$uri$slice_range`, so two students watching the same lecture share every cached slice
no matter what their tokens say -- the hit ratio is untouched. Only manifests are keyed
per full URI, and those were already per-viewer because every mint carries its own
signature; they hold for two seconds, which is what absorbs a burst.

The device cap is enforced at the **origin**, on playlist requests only. It cannot live
at the edge: njs validates a signature with no shared state, and counting devices needs
state shared across viewers. The consequence is that the edge's two-second manifest
cache can serve one poll without the origin seeing it. A player re-reads its playlist
every segment duration, so the next poll counts, and the window is a couple of seconds.

`tenant_limits.max_viewer_devices` is 0 (no cap) by default. The count lives in the
origin process, so a restart forgives everyone until each device polls again, and
splitting `alchemist-origin` across machines gives each its own count -- both are
deliberate. **ponytail:** the ceiling is one process; move the counter into Postgres or
Redis only when several origins actually run at once.

## The cache key excludes the signature -- authorization must run first

Media is cached under `$uri$slice_range`, deliberately **without** the signature. If
the signature were in the key, every viewer would get a private copy of every slice
and the hit rate would collapse to zero, which defeats the entire point of the edge.

The consequence is that a cached object is served to anyone who requests that URI.
**Authorization therefore has to happen before the cache lookup**, via `auth_request`.
Remove that directive and the first valid request warms the cache and every later
unsigned request is served straight out of it. This was verified by warming a slice
with a valid signature and then re-requesting it tampered, unsigned, with an unknown
key id, and expired -- all four must return 403 while the valid request still shows
`X-Cache-Status: HIT`.

Two ways to satisfy it:

- **njs (preferred).** `deploy/edge/playback_auth.js` validates in the nginx worker, so a cache
  hit touches neither the origin nor the database.
- **Subrequest to the origin.** `auth_request` against `/internal/verify-playback`,
  for edges or commodity CDNs that cannot run njs. Correct, but adds an origin
  round trip per request.

If the auth endpoint is unreachable, nginx returns 500 rather than serving content --
it fails closed, which is the behaviour you want.

## Rotating playback keys

Signatures are HMAC-SHA256 over `{prefix}|{exp}`, or `{prefix}|{exp}|{vid}|{wm}` when
the link is bound to a viewer, verified in the nginx worker so a cache hit never touches
the control plane. `internal/platform/signing/parity_test.go` runs the
real `deploy/edge/playback_auth.js` under Node and asserts it matches Go byte for byte -- if those
diverge, every playback URL 403s at the edge while looking valid at the origin.

Rotation never breaks live links, because every listed key is accepted and only the
first one signs:

1. Add the new key to `$playback_keys` on **every** edge, `nginx -t && systemctl reload nginx`.
2. Put the new key **first** in `ALCHEMIST_PLAYBACK_KEYS`, restart the control plane.
   New links now sign with it; links already issued still verify.
3. After the longest playback token TTL has elapsed, drop the old key from both.

## BDIX

Start peering early -- membership and port provisioning is the long pole, far slower
than racking the server. This is what makes delivery both fast (1-5ms rather than
80-200ms) and, on most BD broadband packages, exempt from the subscriber's FUP.
