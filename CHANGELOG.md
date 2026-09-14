# Changelog

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
