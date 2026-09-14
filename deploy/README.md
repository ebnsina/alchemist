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

`ALCHEMIST_ENCODE_WORKERS` should be roughly the core count. Encoding already uses a
per-job worker pool, so setting it far above that only lengthens the tail.

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

Signatures are HMAC-SHA256 over `{prefix}|{exp}`, verified in the nginx worker so a
cache hit never touches the control plane. `internal/platform/signing/parity_test.go` runs the
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
