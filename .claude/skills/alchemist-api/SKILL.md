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

5. **Do not use `/v1/auth/*` from server code.** Those four endpoints back the
   sign-up page in a browser: they set an HttpOnly session cookie. An integration
   authenticates with the API key that signup returned once, as a Bearer token.

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
             "dash":"...","poster":"...","thumbnails":"..."}}
```

Fetch these per viewer, per session. Hand the `hls` URL to any HLS player —
shaka-player and hls.js both handle the SAMPLE-AES encryption without configuration.
HLS is the verified path; DASH is produced but less tested.

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
| `invalid_api_key` | Key is wrong or has been revoked. |

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
