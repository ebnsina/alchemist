# Changelog

## Unreleased

- A `live` option on the player. The LIVE badge never appeared during a broadcast:
  the live playlist is `#EXT-X-PLAYLIST-TYPE:EVENT` so a viewer can seek back to its
  start, and shaka reads EVENT as a growing VOD, so its `isLive()` stayed false for
  the whole thing. A host that knows the asset is on air now says so, and `isLive`
  falls back to shaka's answer when the option is absent, so the embed and every
  existing SDK caller are unchanged.

- EME Clear Key playback. The content key is fetched from `{signed prefix}/key` with
  the same `exp`, `kid` and `sig` as the manifest; shaka POSTs to that URI verbatim,
  so no request filter is needed to keep the signature attached. Clear Key is
  encryption, not DRM — the key reaches the browser in the clear.
- Safari and iOS, whose only key system is FairPlay, now get their own plain-language
  message in both languages instead of a generic protected-video failure. Error code
  `key_system_unavailable`.
- A drifting viewer watermark, drawn when the signed URL carries a `vl` label and
  moved by a CSS transform every seven seconds. It attributes a leak to an account; it
  does not prevent recording. Honours `prefers-reduced-motion`, takes no pointer
  events, survives fullscreen, and costs nothing when the label is absent.
- `player.viewerLabel` on the SDK. Chrome plus core is now ~16.5 KB gzipped, up from
  ~15.9 KB.

## 1.0.0 — 2026-09-14

First release.

- Iframe embed at `/e/{asset}?t={signed playback URL}`, pinned to the `/v1/` bundle so
  embeds already deployed never break on a player change.
- ES-module SDK (`AlchemistPlayer`) for hosts that want their own chrome, plus
  `AlchemistPlayerCore` for no chrome at all.
- Data Saver: a labelled control in the bar showing live MB/hour, a hard 360p ceiling,
  default on for cellular and Save-Data connections.
- Bangla UI with an English toggle; Bangla digits through `Intl`.
- BD-mobile quality menu — 144p to 720p, each rung labelled with its hourly data cost.
- Plain-language handling for loading, buffering, expired signature, offline, not
  found, unsupported browser and DRM failure, in both languages.
- `needs-refresh` two minutes before the signature expires and again if a request
  403s, over `postMessage` for embeds; `refresh()` resumes from the same second.
- QoE beacons on session end and on error, via `sendBeacon`.
- Scrub previews from the signed `sprite.vtt`, with the signature copied onto the
  sprite image the origin does not rewrite.
- Keyboard shortcuts, ARIA labels in the active language, visible focus, captions.
- UI typeface is Google Sans Flex, loaded as one inlined latin-only variable face
  (35 KB, `font-display: swap`). Bangla continues to fall through to the system
  Noto Sans Bengali at zero bytes.
