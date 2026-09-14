# Alchemist Player

The video player for [Alchemist](https://github.com/ebnsina/alchemist) — an iframe
embed for people who want zero integration, and an ES-module SDK for people who want
their own chrome. Both wrap [shaka-player](https://github.com/shaka-project/shaka-player),
so HLS, DASH and EME are one library and adding studio DRM later is configuration
rather than a rewrite.

Built for Bangladesh: **89% of BD internet users are on metered mobile**, so every
default here spends the viewer's data as if it were money, because it is.

---

## What makes this player different

| | |
|---|---|
| **Data Saver is a headline control** | A labelled pill in the control bar showing live MB/hour, not a buried menu item. Default **on** for cellular and Save-Data connections, **off** on everything else. It is a hard ceiling in shaka's `restrictions`, not an ABR hint, and it flushes the buffer so the viewer stops paying immediately. |
| **Bangla first, English one tap away** | Default language is Bangla. Numbers go through `Intl` so Bangla gets Bangla digits. Timecodes stay Latin in a monospace column. |
| **720p is the top rung** | The BD ladder tops out at 720p with 144p and 240p rungs underneath. The quality menu shows what the asset actually has and what each rung costs per hour. |
| **Small bundle, old WebView** | Player chrome and core are ~15 KB gzipped. Shaka loads on demand and split by manifest type, so an HLS asset never downloads the DASH parser. Build target is ES2019 / Chrome 70 / Safari 12. |
| **No dead ends** | Loading, buffering, expired signature, network lost, not found, unsupported browser and DRM failure each have plain-language copy in both languages and a route out. |

---

## Embed

```html
<iframe
  src="https://play.example/e/{asset}?t={signed playback URL}"
  allow="autoplay; fullscreen; encrypted-media"
  allowfullscreen
  style="width:100%;aspect-ratio:16/9;border:0"></iframe>
```

`t` is the `playback.hls` (or `playback.dash`) value your backend got from
`GET /v1/assets/{id}`, percent-encoded. **The player never sees your API key** — the
signature in that URL authorizes the whole asset prefix: manifest, every segment, the
content key, poster and thumbnails.

| Query param | Default | What it does |
|---|---|---|
| `t` | — | **Required.** The signed playback URL. |
| `lang` | browser, then `bn` | `bn` or `en`. |
| `saver` | `auto` | `1` forces Data Saver on, `0` off, `auto` decides from the connection. |
| `autoplay` | `0` | `1` also forces `muted`, because no browser autoplays with sound. |
| `muted`, `loop` | `0` | |
| `max` | `720` | Top rung, for a tenant whose ladder goes higher. |
| `poster` | derived from `t` | |
| `thumbs` | derived from `t` | Signed `sprite.vtt`. |
| `network` | observed | `bdix`, `cellular`, `wifi` or `other`, for the QoE beacon. The player never *guesses* BDIX — only your origin knows. |
| `beacon` | `1` | `0` sends no telemetry at all. |

### Talking to the iframe

An iframe cannot hand a JS object to its host, so events cross by `postMessage`.

```js
const frame = document.querySelector('iframe');

window.addEventListener('message', async (e) => {
  if (e.data?.alchemist !== '1.0.0') return;

  // The one you must handle: the signature is about to die (or already has).
  if (e.data.type === 'needs-refresh') {
    const fresh = await fetch(`/my-api/playback/${assetId}`).then(r => r.json());
    frame.contentWindow.postMessage(
      { alchemist: 1, command: 'refresh', value: fresh.hls }, '*');
  }
});
```

Events posted out: `embed-ready`, `ready`, `playing`, `pause`, `ended`, `buffering`,
`statechange`, `timeupdate`, `error`, `needs-refresh`, `datasaverchange`,
`languagechange`, `qualitychange`.

Commands accepted in: `play`, `pause`, `seek`, `mute`, `volume`, `dataSaver`, `lang`,
`quality`, `refresh`, `destroy`.

`needs-refresh` fires twice over a session's life: **two minutes before** `exp` with
`{reason:'expiring'}`, so you can swap the URL without the viewer noticing, and again
with `{reason:'expired'}` if a request actually 403s. Handling the first one means
playback never breaks.

---

## SDK

```bash
npm install @alchemist/player
```

```js
import { AlchemistPlayer } from '@alchemist/player';

const player = new AlchemistPlayer(document.getElementById('stage'), {
  src: asset.playback.hls,
  poster: asset.playback.poster,
  thumbnails: asset.playback.thumbnails,
  lang: 'auto',          // 'bn' | 'en' | 'auto'
  dataSaver: 'auto',     // true | false | 'auto'
  maxHeight: 720,
  country: 'BD',
});

player.addEventListener('needs-refresh', async () => {
  const fresh = await getFreshPlaybackURL();
  await player.refresh(fresh);   // keeps the playhead, keeps playing
});
```

### API

```ts
// state
player.state          // 'idle'|'loading'|'ready'|'playing'|'paused'|'buffering'|'ended'|'error'
player.lang           // 'bn' | 'en'
player.dataSaver      // boolean
player.quality        // number | 'auto'
player.error          // { kind, code, messageKey, retryable } | null
player.currentTime / duration / paused / muted / volume
player.src

// control
await player.play(); player.pause(); player.seek(seconds);
player.setMuted(bool); player.setVolume(0..1);
player.setLang('bn' | 'en');
player.setDataSaver(bool);
player.setQuality(360 | 'auto');
player.setCaptions(trackId | null);
await player.refresh(newSignedURL);
await player.destroy();

// introspection
player.qualities();          // [{ height, bandwidth, mbPerHour, id }]
player.captionTracks();
player.estimatedMbPerHour(); // what Data Saver promises right now
player.getStats();           // the QoE beacon body, before it is sent
player.thumbnailTiles;
```

`AlchemistPlayer` is an `EventTarget`; every event above is a `CustomEvent` with the
same name and detail. Pass `controls: false` to get native browser controls and drive
everything yourself, or import `AlchemistPlayerCore` for the player with no chrome
and no stylesheet at all.

### Errors

Errors never reach the viewer as codes. `player.error.kind` is one of
`expired | network | offline | notFound | unsupported | drm | generic`, and
`player.error.code` is the stable string that also goes in the QoE beacon
(`playback_not_authorized`, `shaka-1001`, …). Map `kind` yourself if you are drawing
your own error UI.

---

## QoE beacons

On session end and on error the player POSTs to
`/playback/{tenant}/{asset}/beacon` with the same signature as playback:

```json
{"session_id":"…","startup_ms":420,"rebuffer_count":1,"rebuffer_ms":380,
 "avg_bitrate_bps":689930,"error_code":null,"network":"cellular","country":"BD"}
```

Sent with `navigator.sendBeacon` so it survives page unload, as a `text/plain` blob so
it stays a CORS-simple request — the origin's beacon route answers no preflight, and
we never read the response. The origin always answers 204, so nothing here retries or
is ever visible to the viewer.

> **One thing worth fixing on the origin:** the `POST …/beacon` route sends no CORS
> headers at all (`delivery.setCORS` covers only `GET, HEAD, OPTIONS` on the media
> routes). The `text/plain` trick means beacons land anyway, but any future beacon
> change that needs a real `Content-Type` or a readable response will need
> `Access-Control-Allow-Origin` and an `OPTIONS` handler on that route.

---

## Versioning

**An embed already deployed in the wild must never break.** So:

- The embed bundle is served under a **major-version prefix**: `/v1/`. `vite.config.ts`
  sets `base: '/v1/'` and builds into `dist/v1/`.
- Host rewrite: `/e/*` → `/v1/embed.html`. The pretty URL is what customers paste; the
  versioned path is what actually loads. `embed.html` carries `<base href="/v1/">` so
  its assets resolve against the bundle root, not the pretty URL.
- Anything that changes embed behaviour — query params removed or repurposed,
  `postMessage` shapes, default Data Saver policy — ships as `/v2/` at a new path.
  `/v1/` keeps being served, unchanged, indefinitely.
- Inside a major, deploys are in place: bug fixes and new *optional* params only.
- The SDK follows normal semver; it is not pinned by path because consumers pin it in
  their `package.json`.

Deploying is copying `dist/v1/` next to the previous majors and adding the rewrite:

```nginx
location /e/ { rewrite ^/e/.*$ /v1/embed.html last; }
location /v1/assets/ { expires 1y; add_header Cache-Control "public, immutable"; }
location = /v1/embed.html { expires 5m; }
```

---

## Running it

### Against a real Alchemist

```bash
npm install
npm run dev            # http://localhost:5180/v1/
```

The control plane sends no CORS headers, and the player holds no API key, so the test
bench cannot fetch a playback URL for you — that is the design, not a gap. Get one
from your own shell:

```bash
curl -H "Authorization: Bearer $KEY" http://localhost:8099/v1/assets/$ASSET_ID
```

Paste the `playback.hls` value into the test bench and press **Load**. The bench shows
live stats, the exact beacon body that would be sent, every event, the palette, and a
Bangla conjunct rendering check.

### Against the offline origin stand-in

No backend, no database, no SeaweedFS — one fixture asset and a 200-line origin that
mirrors the real one byte for byte (same HMAC over `prefix|exp` truncated to 32 hex
chars, same manifest rewriting, same raw 16-byte content key, same always-204 beacon):

```bash
bash scripts/make-fixture.sh          # needs ffmpeg + shaka-packager on PATH
python3 scripts/origin.py             # serves ./fixture on :8099, prints a signed URL
npm run dev
```

The fixture is the real BD-mobile ladder — 144p/240p/360p/480p/720p, H.264 Main,
aligned GOPs, one byte-range CMAF file per rendition, SAMPLE-AES (cbcs) encrypted with
`KEYFORMAT="identity"` — so what plays here plays against production.

Use `--port` if a real Alchemist already has 8099.

### Tests

```bash
npm test          # node --test, no framework
npm run typecheck
npm run build     # dist/sdk (ES module + .d.ts) and dist/v1 (embed + test bench)
```

Covered: Data Saver bitrate capping and default policy, language negotiation, beacon
payload construction and clamping, signature expiry and sibling-URL derivation,
sprite-VTT parsing, error classification. No framework — `node --test` strips the
types itself.

---

## Design

Brass on near-black — the alchemist's gold, not another blue video player.

| | |
|---|---|
| `#E8A33D` | brass — accent, progress, primary action |
| `#3FBF8F` | Data Saver active |
| `#E8785F` | errors |
| `#0C0C0E` / `#17171B` | surface, raised surface |

Type is **Google Sans Flex** for UI, Geist Mono for timecodes and MB figures, with
Bangla falling through to **Noto Sans Bengali**.

Google Sans Flex is loaded as a single latin variable face (weights 400–700),
`@font-face` inlined in `src/styles.ts` rather than linked from
`fonts.googleapis.com`: that saves a cold BD mobile connection one round trip, and
makes it impossible to accidentally pull the latin-ext, vietnamese, math or syriac
subsets Google also serves off that family. **35 KB on the wire** (36,096 bytes woff2,
already compressed — gzip adds nothing), cached a year, with a `preconnect` to
`fonts.gstatic.com` in the page head.

That 35 KB is more than twice the player chunk, so it is worth being honest about: it
is a one-time, cache-for-a-year cost, it never blocks playback, and `font-display:
swap` means text is readable in the system font from the first frame. If you decide a
BD-mobile viewer should not pay it at all, delete the `@font-face` block in
`src/styles.ts` — the stack degrades to `system-ui` with no other change.

The face's `unicode-range` is latin only, and that is what protects the Bangla
rendering: Bengali codepoints are not in the range, so the browser never tries Google
Sans Flex for them and falls straight through to Noto Sans Bengali, which Android has
shipped since 4.1 and which shapes conjuncts correctly with the system text engine. So
Bangla still costs the viewer zero bytes and যুক্তাক্ষর, শিক্ষা and বাংলাদেশ still render
correctly — verified after the font change, not assumed.

Geist Mono is a fallback stack only; no mono webfont is downloaded.

Icons are hand-authored inline SVG on the Hugeicons stroke grid (24px, 1.6 stroke,
round caps). The icon package is 72 MB across 12,000 files; twelve paths are not worth
that on a player whose entire argument is small bytes. Swap the path bodies in
`src/icons.ts` for the real set if you prefer.

Keyboard: `space`/`k` play, `←`/`→` seek 5s (`shift` 30s on the scrubber), `↑`/`↓`
volume, `m` mute, `f` fullscreen, `s` Data Saver, `c` captions, `esc` close menu.
Every control has an ARIA label in the active language, the scrubber is a real
`role="slider"` with `aria-valuetext`, and focus is visible everywhere.

---

## Layout

```
src/
  player.ts      core: shaka + Data Saver + expiry + QoE. No DOM chrome.
  ui.ts          the chrome, drawn against player.ts
  styles.ts      one injected stylesheet, so the SDK is a single module
  icons.ts       inline SVG
  i18n.ts        bn/en copy, language negotiation, Intl formatting
  errors.ts      shaka codes -> plain copy, in both languages
  network.ts     connection classification, Data Saver policy, MB/hour
  signed-url.ts  expiry, sibling URLs, signature propagation
  beacon.ts      QoE payload + sendBeacon
  thumbnails.ts  sprite.vtt parsing
  shaka.ts       lazy, split HLS/DASH import
  embed.ts       iframe entry + postMessage bridge
  testpage.ts    the test bench (not shipped)
scripts/
  make-fixture.sh  builds a local stand-in asset with ffmpeg + shaka-packager
  origin.py        offline origin mirroring internal/modules/delivery
```
