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
version table, so each is written to be safe to re-run. That claim is now enforceable:
apply them with `ON_ERROR_STOP=1` and a second run must be silent. Sixteen of them
were not re-runnable until 2026-09-16 — the loop above only appeared to work because
it ignored errors, which also meant a genuinely broken migration was indistinguishable
from "already exists". Keep new ones idempotent: `if not exists` on tables, columns and
named indexes, `drop policy if exists` before a policy, a guard around a type, and
`on conflict do nothing` on a seed. `internal/platform/db/migrations/020_members_branding.sql`
also repairs the `users` and `sessions` policies from `internal/platform/db/migrations/017_accounts.sql`, which named a setting
nothing sets and so matched no rows. `internal/platform/db/migrations/027_dedup_renditions.sql` deletes rows rather than
adding a column: deduplicated assets carried copies of the canonical asset's rendition
rows, and those copies are now resolved rather than stored.
`internal/platform/db/migrations/028_usage_bytes.sql` adds `tenant_stored_bytes()`, another
`SECURITY DEFINER` reader for a cross-tenant job, and the unique index the daily egress
and storage rows upsert onto -- without it every flush inserts a new row instead of
folding into the day. `internal/platform/db/migrations/041_live_cleanup.sql` drops `live_streams.ingest_port`,
which nothing has written since the ingest server replaced per-stream ports with one
port dispatched by path.
`internal/platform/db/migrations/040_captions.sql` adds the subtitle table; it holds
only sidecar WebVTT metadata, so there is nothing to backfill and no media to migrate.
`internal/platform/db/migrations/034_live_recording.sql` adds `live_sessions.swept_at`
and `live_tenants()`, both additive and safe to re-run; the reaper needs the definer
function because RLS on `tenant_limits` is forced and a cross-tenant select there
returns zero rows without erroring.
`internal/platform/db/migrations/038_asset_title.sql` adds the nullable
`assets.title` and nothing else; it is additive, safe to re-run, and deliberately
backfills nothing — null means a video nobody has named and the dashboard shows the
short id instead.
`internal/platform/db/migrations/039_staff.sql` adds `users.platform_admin` and
`sessions.acting_tenant_id`, both nullable or defaulted and safe to re-run, and
replaces `auth_session()` so a session resolves to the tenant it is acting as. It is
dropped and recreated rather than replaced in place, because its return type changes.
`internal/platform/db/migrations/033_encryption_default.sql` flips
`tenants.encrypt_playback` to default true and changes **nothing** for tenants that
already exist -- see below.

## The platform administrator

One Alchemist account may act as any tenant, for support. Create it with the binary's
own subcommand, which asks:

```
/opt/alchemist/bin/alchemist create-admin
Email: you@example.com
Password:
Again:
```

The password is not echoed and is asked for twice, because nothing in the product can
reset it yet and a typo locks the only administrator out. Ten characters minimum, the
same rule as signing up. Locally, `make create-admin`.

With no terminal to ask at — a deploy script, CI — pass the address as an argument
and the password in `ALCHEMIST_STAFF_PASSWORD`; the command exits if either is
missing, and never takes the password as an argument, which would put it in the shell
history and in `ps`.

Run it twice and it promotes the account that is already there, leaving the password
alone; it never creates a second user. A new account lands in a tenant called
`Alchemist` — the flag grants everything, the tenant grants nothing, except live,
which is switched on so staff are not sold their own product.

Support then works by impersonation: `POST /v1/staff/impersonate {"tenant_id"}` makes
every ordinary endpoint answer for that customer, and the session becomes read-only
(`403 read_only_session` on every write) until `DELETE /v1/staff/impersonate`. Staff
see one thing a customer does not: `GET /v1/assets/{id}/diagnostics` and
`GET /v1/live-sessions/{id}/diagnostics` return the raw `river_job` error text, which
the customer-facing activity view withholds because it carries paths and internal
arguments.

Changing something on a customer's behalf is still `/admin/*` with
`ALCHEMIST_ADMIN_KEY`. Impersonation reads; it does not write.

`ALCHEMIST_ENCODE_WORKERS` should be roughly the core count. Encoding already uses a
per-job worker pool, so setting it far above that only lengthens the tail.

## Live ingest

Off unless both `ALCHEMIST_LIVE_INGEST_HOST` and `ALCHEMIST_LIVE_PULL_BASE` are set;
one without the other refuses to boot. With neither, the `/v1/live-streams` endpoints
are not served at all rather than served and always failing.

An ingest server sits in front — `deploy/live/mediamtx.yml` configures it as a socket
and nothing else, with HLS and recording off, because playback is the origin's job. It
terminates SRT, RTMP and WebRTC on **one port each**, whatever the number of streams,
and dispatches on the path.

    encoder ──RTMP/SRT────┐
                          ├─> ingest server ──loopback RTSP──> transcoder ──> origin
    browser ──WHIP :8889──┘        │
                                   └── POST /internal/live/authorize (stream key)

WebRTC is on for **publishing only**: a browser POSTs its SDP offer to
`http://host:8889/{stream-id}/whip` and becomes the encoder, which is the whole of
what a customer with a laptop and no OBS needs. The path is the stream id exactly as
RTMP's is, so the same `(key, path)` authorisation and the same loopback RTSP pull
carry it; nothing downstream knows the difference.

**WHIP credentials go in a header, not the query string.** The stream key travels as
`Authorization: Bearer publisher:<key>`, and a key in `?user=&pass=` is refused —
verified against MediaMTX v1.15.6, where the query-string form logs
`authentication failed: server replied with code 401`, indistinguishable from a wrong
key. `Authorization: Basic` works too; Bearer is what the dashboard sends, because a
401 to a Basic request makes the browser pop its own credential dialog.

A browser has no key to paste, so `POST /v1/live-streams/{id}/start` on a `camera`
stream mints one and returns it as `publish_token`. It replaces the stream key, which
is free because nothing else holds it: the credential sitting in page script is good
for one broadcast.

**A public deployment needs TLS on the WHIP port.** `webrtcEncryption: no` and the
`http://` URL the API returns are right for a box where the dashboard is also on
`http://localhost`; the moment the dashboard is served over https, the browser blocks
a plain-http WHIP POST as mixed content and `getUserMedia` refuses an insecure origin
anyway. Set `webrtcEncryption: yes` with `webrtcServerCert`/`webrtcServerKey`, or put
the same certificate in front of `:8889`, and change the scheme in `publishURL` in
`internal/modules/live/ffmpeg.go` alongside the ports it already hardcodes.

**The stream key is the credential.** The ingest server asks the API about every
publish, and a publish is allowed only when the key resolves to a live stream *and*
that stream is the path being published to — checking the key alone would let one
customer's valid key publish over everyone else's broadcast. Verified by refusing a
wrong key: the ingest server logs `authentication failed: server replied with code
401` and the publisher is dropped.

**That endpoint carries no API key and must never be publicly reachable.** It is
served on the same listener as the customer API, which has to be public, so "bind it
privately" is not something the app can do for you: whatever proxies the API must
return 404 for `/internal/`, the way `deploy/edge/nginx.conf` now does for the origin.
Both internal routes fail closed without credentials, but neither is a viewer's
business and nothing rate-limits them. The caller
is the ingest server on the same host, so bind it to loopback or a private interface
and firewall it like a database port.

Reads are asked about too, and are allowed **only from loopback** — that is how the
transcoder pulls the stream back. Allowing them from anywhere would turn the ingest
ports into a second, unsigned way to watch a customer's broadcast. That still holds
with WebRTC on: WHEP is a read, so it is refused by the same rule, and it was verified
after enabling it rather than assumed. Loopback reads stay open to anything on the
box, as they were for RTSP; the box is not a trust boundary.

### Enabling live for a customer

Configuring an ingest host enables live for the **deployment**. It does not give it to
anyone: Live is a separate product, `tenant_limits.live_enabled` is false by default,
and a tenant with no limits row has it off. Turn it on per customer:

    PUT /admin/tenants/{id}/live   {"enabled": true}

Operator surface, behind `ALCHEMIST_ADMIN_KEY`. A VOD-only tenant calling the live
endpoints gets `live_not_enabled` rather than a stream nobody sold them, and
`/v1/whoami` reports `live_enabled` so a dashboard knows whether to offer it.

### Sizing

`ALCHEMIST_LIVE_MAX_STREAMS` is the number of concurrent broadcasts this box accepts.
Size it against cores, not ambition: one live rung is roughly one CPU core held for
the length of the broadcast, so a 16-core box does not run 100 streams. The port range
that used to bound this is gone — one port now serves every stream. See
`docs/06-live.md`.

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

## Playback encryption after migration 033

New tenants get `cenc` encryption on, with the key endpoint answering an EME Clear Key
licence. Chrome, Firefox and Edge play it with no licence vendor.

Two things it deliberately does not do:

- **It does not touch existing tenants.** The migration changes the column default, not
  the rows. Turning it on for a live account is an operator decision, because **Safari
  and iOS cannot play Clear Key** -- WebKit's only key system is FairPlay -- and those
  viewers get `403 browser_not_supported` from `/playback/.../key`. Flip one when its
  audience is not on Apple devices: `PUT /v1/playback-settings {"encrypt_playback":true}`,
  or `update tenants set encrypt_playback = true where id = '...';`
- **It does not re-package anything already published.** Assets keep whatever they were
  encoded with. Re-encrypting a library rewrites every object, changes every ETag, and
  evicts the lot from every edge cache -- real money for protection nobody asked for.
  A deferred rung generated years later reuses the asset's own key, or stays clear if
  the asset has none.

And be honest about what it is: **encryption at rest, not DRM.** The key goes to the
browser in the clear behind the signed URL, so a stolen bucket or backup decodes to
nothing, while a viewer who is entitled to watch can still keep a copy. Password
sharing is answered by viewer-bound tokens and the device cap below, not by this.

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
