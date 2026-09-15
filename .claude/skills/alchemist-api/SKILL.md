---
name: alchemist-api
description: Integrate video upload, transcoding and playback using the Alchemist API. Use when adding video to an application - uploading, importing from a URL or a bucket, polling asset state, embedding playback, handling webhooks, or reading usage. Also use when debugging an Alchemist integration that returns 403, 404 or an asset that never becomes ready.
---

# Integrating with Alchemist

Alchemist transcodes video and serves adaptive-bitrate playback. You talk to it over
HTTP with an API key. There is no shared database, no direct object-storage access
and no internal endpoints — if something is not in the API, that is a gap in the API.

Canonical contract: `api/openapi.yaml` in the Alchemist repository. A runnable
end-to-end example using nothing but an API key: `test/journey.sh`.

## The five things that go wrong

1. **Waiting for `ready` hangs forever.** Ingest encodes a low ladder and returns
   `partially_ready`; 480p and above are generated when someone first plays the
   video. **Treat `partially_ready` and `ready` as equally playable.** An asset
   nobody watches stays `partially_ready` indefinitely, and that is correct.
2. **Playback URLs expire after 4 hours** and carry `exp`, `kid` and `sig`. Fetch
   them from `GET /v1/assets/{id}` when a viewer asks. Never build them, cache them
   long-term, or store them in a database.
3. **403 on playback is almost always an expired or mismatched signature.** One
   signature covers the manifest, every segment, the content key, the poster and the
   thumbnails — so if the manifest works and segments 403, you are mixing signatures
   from two different fetches.
4. **Uploads do not go through the API.** `POST /v1/uploads` returns a presigned
   target; `PUT` the bytes straight there, then call `/complete`. Proxying the file
   through your own server wastes your bandwidth and will time out on large files.
5. **Branch on `error.code`, never on `error.message`.** Codes are stable; messages
   are prose and are returned in Bangla when the request sends `Accept-Language: bn`.

5. **Do not use `/v1/auth/*` from server code.** Those endpoints back the sign-up
   page in a browser: they set an HttpOnly session cookie. An integration
   authenticates with the API key that signup returned once, as a Bearer token.
6. **`/v1/members/*` and `/v1/branding/logo` are session-only.** An API key gets
   `403 session_required` there, deliberately: a key is something a server holds, and
   a leaked one that could invite an owner would be a permanent way back in. There is
   nothing to work around here — team changes happen in the dashboard.

## Getting a key

Operator-only, behind `ALCHEMIST_ADMIN_KEY` (a different credential from customer
keys, so a leaked customer key cannot mint more):

```bash
curl -X POST -H "Authorization: Bearer $ADMIN_KEY" -H 'Content-Type: application/json' \
  -d '{"name":"My App","ladder_profile":"bd-mobile"}' \
  $ALCHEMIST/admin/tenants
# -> {"tenant_id":"...","name":"My App","ladder_profile":"bd-mobile"}

curl -X POST -H "Authorization: Bearer $ADMIN_KEY" -H 'Content-Type: application/json' \
  -d '{"name":"production"}' \
  $ALCHEMIST/admin/tenants/$TENANT_ID/keys
# -> {"api_key":"alc_...","note":"Store this now. It is not shown again."}
```

The plaintext key is shown once. Store it before moving on.

Ladder profiles: `bd-mobile` (144p-720p, mobile-first), `bd-ott` (adds 1080p),
`intl-default`. Pick at tenant creation; it decides what every upload is encoded to.

## Uploading

```bash
# 1. ask for a target
curl -X POST -H "Authorization: Bearer $KEY" $ALCHEMIST/v1/uploads
# -> {"asset_id":"...","upload_url":"https://...","expires_in_seconds":21600}

# 2. bytes go straight to storage, not through the API
curl -X PUT --upload-file lecture.mp4 "$UPLOAD_URL"

# 3. tell Alchemist the bytes landed
curl -X POST -H "Authorization: Bearer $KEY" $ALCHEMIST/v1/assets/$ASSET_ID/complete
```

Check the `PUT` status. A failed upload followed by `/complete` produces an asset
that fails with `source_unreadable` several minutes later.

## Importing instead of uploading

```bash
curl -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/lecture.mp4"}' $ALCHEMIST/v1/assets
```

The URL is fetched server-side and re-validated on every redirect. Addresses inside
private, loopback or link-local ranges are refused with `source_url_not_allowed` —
this includes anything on your own network, which is the usual cause when a local
test file is rejected.

For a whole library, register the bucket once and it is scanned continuously:

```bash
curl -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' -d '{
  "endpoint":"https://s3.example.com","bucket":"media","prefix":"lectures/",
  "access_key_id":"...","secret_access_key":"..."}' \
  $ALCHEMIST/v1/bucket-sources
```

Only video extensions under the prefix are taken, and re-scanning never re-imports.

## Knowing when it is playable

Prefer webhooks to polling:

```bash
curl -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d '{"url":"https://yourapp.example/hooks/alchemist","events":["asset.ready","asset.failed"]}' \
  $ALCHEMIST/v1/webhooks
# -> {"secret":"..."}   shown once
```

The events are `asset.ready`, `asset.failed`, `rendition.ready`, and for live
`live.started`, `live.ended`, `live.failed` (which carries `error_code`). A broadcast
is an asset, so its recording is reported by `asset.ready` on the same asset id --
nothing live-specific to handle.

Deliveries carry `X-Alchemist-Signature: sha256=<hmac-sha256 of the raw body>`.
Verify it with a constant-time compare before trusting anything in the payload:

```go
mac := hmac.New(sha256.New, secret)
mac.Write(body)
expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
if !hmac.Equal([]byte(expected), []byte(r.Header.Get("X-Alchemist-Signature"))) {
    http.Error(w, "bad signature", http.StatusUnauthorized)
    return
}
```

The webhook URL must be https and publicly reachable — it is validated the same way
import URLs are, so a localhost endpoint is refused.

If you poll instead, accept both playable states:

```go
switch asset.State {
case "ready", "partially_ready":
    // playable
case "failed":
    // asset.ErrorCode says why, in stable form
default:
    // created, uploading, uploaded, probing, mezzanine, analyzing, encoding, packaging
}
```

## Playing it

```bash
curl -H "Authorization: Bearer $KEY" $ALCHEMIST/v1/assets/$ASSET_ID
```

```json
{"id":"...","state":"partially_ready","duration_seconds":2400,
 "playback":{"hls":"/playback/.../master.m3u8?exp=...&kid=...&sig=...",
             "dash":"...","poster":"...","thumbnails":"...",
             "encrypted":true,"preferred":"dash"}}
```

Fetch these per viewer, per session. **Hand the player whichever URL `preferred`
names.** Both are always present, but an encrypted asset is packaged `cenc` and HLS
has no `cenc` -- its fMP4 encryption is the SAMPLE-AES family -- so an encrypted
asset plays over DASH and a clear one over HLS. `encrypted` says which kind it is;
`preferred` saves you having to know the rule.

### Binding a link to one student

```bash
curl -H "Authorization: Bearer $KEY" \
  "$ALCHEMIST/v1/assets/$ASSET_ID?viewer=student-8842&watermark=01712345678"
# playback URLs now carry &vid=student-8842&wm=01712345678
```

`viewer` is your own id for whoever is watching. It stays opaque to Alchemist: signed,
echoed back, never stored. `watermark` is a short label your player draws on screen —
render it drifting, so nobody reposts a dump with their own number on it. Both are
inside the signature, so a viewer who edits either out of the URL gets a 403. Letters,
digits and `- . _ ~ @` only; 64 and 48 characters. Anything else is `400 invalid_viewer`.
There is no server-side burn-in: that is a re-encode per viewer.

Set `max_viewer_devices` (`PUT /v1/playback-settings`, 0–20, 0 means no cap) and a bound
viewer watching from more devices than that gets `403 viewer_limit_reached` on the
newest one — the sessions already playing are left alone. Unbound links are never
counted. This, not encryption, is what answers one login shared with a class.

### Encryption

**On by default for new accounts.** Media is `cenc` and the key endpoint returns an EME
Clear Key licence:

```json
{"keys":[{"kty":"oct","kid":"<base64url>","k":"<base64url>"}],"type":"temporary"}
```

Unpadded base64url. Point shaka-player's `drm.servers['org.w3.clearkey']` at the asset's
`/key` URL with the same `exp`, `kid` and `sig`; shaka POSTs its challenge there and the
endpoint answers the licence. Chrome, Firefox and Edge all play it.

**Safari and iOS cannot.** WebKit's only key system is FairPlay, which needs an Apple
certificate and a licence server. Those viewers get `403 browser_not_supported` from the
key endpoint — show that message rather than an unexplained black player. Turn
encryption off for the account if the audience is mostly on Apple devices.

Be accurate with customers about what this is: **encryption at rest, not DRM.** The key
is delivered to the browser behind the signed URL, so a stolen bucket or backup is
useless, while a viewer entitled to watch can still keep a copy.

`thumbnails` is a WebVTT file for scrubbing previews; `poster` is a JPEG.

## Usage

```bash
curl -H "Authorization: Bearer $KEY" "$ALCHEMIST/v1/usage?from=2026-09-01T00:00:00Z"
# -> {"lines":[{"kind":"ingest","quantity":2400,"unit":"seconds"}]}
```

Defaults to the current calendar month.

## Error codes worth handling

| Code | Meaning |
|---|---|
| `quota_exceeded` | Too many videos processing, or the monthly limit is reached. Retry later. |
| `source_url_not_allowed` | The URL resolved to a private or internal address. |
| `source_too_large` | Above the tenant's size limit. |
| `source_unreadable` / `no_video_stream` | The file is corrupt, or is not video. |
| `playback_not_authorized` | Signature expired or invalid. Re-fetch the asset. |
| `viewer_limit_reached` | This viewer is already streaming from the maximum number of devices. |
| `browser_not_supported` | Safari or iOS asking for a Clear Key licence. Tell them to use Chrome, Firefox or Edge. |
| `invalid_viewer` | `viewer` or `watermark` has characters or a length that will not survive a URL. |
| `invalid_api_key` | Key is wrong or has been revoked. |
| `session_required` | Account administration. Not reachable with an API key, by design. |
| `not_permitted` | The signed-in user is not an owner or admin. |
| `last_owner` | Refused: an account cannot be left with no owner. |
| `invite_not_found` | Expired, already redeemed, or never existed — the three are one answer on purpose. |

`encode_failed`, `stitch_failed`, `package_failed` and `processing_failed` are ours,
not yours — surface them as a generic failure and check the Alchemist logs.

## Running Alchemist locally

```bash
make db-reset && make storage   # separate terminals
make run
```

Needs `ffmpeg`, `packager` (shaka-packager) and `weed` (SeaweedFS) on PATH. The last
two are GitHub release binaries, not available through Homebrew.

To let it fetch from a loopback test server, set `ALCHEMIST_FETCH_ALLOWLIST` to the
exact `ip:port`. It disables an SSRF defence for those hosts only — never set it in
production.
