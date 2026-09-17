# Changelog

## Unreleased

### Added
- **How long a playback link lasts is a setting now**, not a hardcoded four hours, with
  the warning that matters said where the number is chosen: pick it from your longest
  video, because the same link carries the whole thing and a link that runs out mid-
  lecture stops the lecture. If the worry is a link being passed around rather than kept
  too long, the allowed-sites list is the control for that instead.

### Changed
- **Scrambling stored video is off for new accounts.** It was on, and it refuses every
  iPhone, iPad and Safari viewer — for a key that is handed to the browser in the clear
  anyway, so it never stopped anyone who was allowed to watch from keeping a copy.
  Accounts that already have it on keep it, and videos already made are untouched.
- The player and the video page no longer claim links last four hours.
- **Playback settings can name the sites your videos are allowed to play on.** A link
  copied out of your page then does nothing anywhere else, which is the control every
  other video platform ships and this one did not — playback answered
  `Access-Control-Allow-Origin: *` and checked no referer, so a leaked link worked
  inside any page on any site for its whole four hours. The page says the limit out
  loud: only browsers can be checked this way, so an account whose viewers use its own
  app should leave the list empty.
- **Settings has a Playback tab**, which is where the scrambling switch lives now. It
  had an API and no screen anywhere in the product, so a customer whose iPhone viewers
  got a refusal had no way to turn it off themselves — and it is on by default for new
  accounts. The page says what the trade actually is: Apple devices cannot play a
  scrambled video and are told so plainly rather than shown a black screen, the change
  applies to videos sent from now on, and live broadcasts are never scrambled. The
  per-viewer device limit moved here too, because one call writes both and a screen
  that sent only the toggle would have quietly removed the limit.
- **The marketing page sells live and editing, which it had never mentioned.** Two new
  sections: going live from a browser camera or an encoder, and what can be done to a
  video after it is uploaded — cutting, reshaping, watermarking and subtitles. Both say
  the thing a reader would otherwise find out the hard way: a broadcast goes out at one
  quality with the full ladder made afterwards from the recording, and live is switched
  on per account rather than sold with a plan. No data-saving claim is made about live,
  because BDIX is FUP-exempt on fixed broadband only and the mobile majority pays for
  the bytes.
- The middle pricing tier is **Standard**, not Studio. The product already has a Studio
  — it is the editing surface in the dashboard — and a plan sharing its name made the
  new section read as a plan feature.
- **A stream's encoder settings now show a Stream Key you can actually paste.** The
  field was always there and always empty, with the key buried in the server address
  as a placeholder, so nobody following the page could get OBS connected. Server and
  Stream Key are now the two halves an encoder asks for. SRT and camera streams carry
  everything in one address and still say to leave the key empty, and now say why.

### Changed
- **Deleting a stream that is on air is refused** rather than silently taking the
  broadcast down and stranding its segments. Stop it, then delete it.
- **Subtitles.** A video's page takes a WebVTT file per language, names it as viewers
  will see it in the player's menu, and lists what is already there. Uploading a
  language you already have replaces it; removing one asks first and says the video
  keeps playing either way. Nothing is switched on by default — the menu is offered,
  the viewer chooses. The language box suggests the language's own name (বাংলা for
  `bn`), because that is what someone scanning a menu is looking for.
- **A failed video can be tried again.** The failure card on a video's page now offers
  "Try again" and says that the file we were sent is still here, so nothing has to be
  uploaded a second time. It is not offered for the three codes where the same bytes
  fail the same way — no video track, too large, or a link we will not fetch from —
  because a button that spends ingest quota to reach the same screen is worse than no
  button. There is no confirmation step: it costs a wait and nothing else. The page
  resumes its own polling once the video moves back into the queue.

  The Videos table offers the same thing from a row's menu, on a failed row only — a
  permanently dead item on every ready video's menu is worse than one that is not
  there. On the three codes that cannot be retried it is shown but disabled, and says
  why on hover, because someone opening that menu has come looking for it.

- **Two destructive actions that never asked now ask.** Removing your logo under
  Branding, and the bin icon beside an edit in Studio — the second sat next to the
  state chip with nothing between a mis-click and a deleted record. Both use the same
  confirm dialog as every other delete, and Studio's wording matches the one already
  on the Studio list for the same action. Cancelling a migration and stopping a
  broadcast already confirmed, inline where the control is, and were left that way: a
  modal over a live preview is worse than the question sitting in the control bar.

### Changed
- **Live streams is a real table** — name, state, source, made — with a state filter and
  a source filter, and a row menu for Open, Rename, Start, Watch and Delete. Renaming a
  stream was in the API and had no control anywhere in the dashboard, and deleting one
  now asks first. The stream key panel, the ingest address panel and the embedded player
  below the list are unchanged.
- **Team is two tables.** People — email, role, joined, last signed in — paged, searched
  and sorted by the server, with a row menu for changing a role and removing somebody.
  Your own row still offers neither and still says why. Pending invites keep their own
  table below; the API always returns them whole, so that one searches and pages the
  list it already has.
- **Studio's "What you have made" is a table** — what we made, which video it came from,
  state and when — paged and sorted by the server instead of showing the newest ten and
  hiding the rest, with a row menu that opens the new video once it is finished and
  removes the record. Removing says plainly that the video it made stays. The picker no
  longer counts how many edits came from each video: with the list paged that count
  would only have been of the edits on screen.
- **Connected buckets is a table with controls.** Bucket, prefix, state, how many videos
  it has taken in and when it was last checked, with a row menu that pauses, resumes or
  disconnects it. Disconnecting a bucket whose videos are still being made is refused by
  the API, and the refusal now reaches the page in plain words instead of disappearing:
  "Videos from this bucket are still being made. Wait for them to finish, then
  disconnect it."
- **Webhook endpoints can be changed, not only made.** The list was a row, a Live/Paused
  chip and nothing that could move either. It is now a table — address, events, state —
  with a Live/Paused filter and a row menu that edits the address, pauses or resumes the
  endpoint, and deletes it behind a confirm that says the signing secret goes too.
- **API keys is a real table** — name, made, state — with a Live/Switched off filter and
  a row menu that renames a key as well as switching it off. Renaming a key was in the
  API and had no control anywhere in the dashboard. The last live key still cannot be
  switched off and still says why; that count now comes from the account rather than
  from whichever ten keys are on screen.
- **Recordings is a real table.** Same five columns as Videos — name, state, length,
  size, added — paged, searched and sorted by the server with `state=live_ended`
  pinned, and a row menu for Watch, Rename, Edit in Studio and Delete. The old
  `AssetList` component had one page left to serve and is gone.

### Added
- **Stop a broadcast.** Until now a stream ended only when the encoder disconnected or
  the engine timed it out after thirty minutes. The stream page now has a Stop control
  behind a confirm step, for both an encoder stream and a camera one; stopping a camera
  broadcast tells the engine and closes the connection and the camera together.

### Added
- **An Overview page at `/app/`.** The videos list had been doing both jobs: it opened
  with a greeting, four figures and a month of uploads, and the list itself was below
  all of it. Overview keeps the figures and now also names any broadcast on air or
  waiting, above everything else. `/app/videos/` is the list and nothing else.

### Changed
- **The overview leads with a running clock**, on the right where the Upload button
  was. It ticks every second and shows them, with the unlit segments drawn behind the
  lit ones. Upload belongs on Videos, one row away, not on a page you are reading
  rather than acting on.
- **The sidebar is ordered by what you came to do**: Videos, Live, then Develop and
  Account. Getting video in is one job with three doors, so Upload, Connected buckets
  and Connected buckets no longer take a row each — they are reached from the Upload
  button, and the Upload page names the bucket route where somebody looks for it.
  Moving a library in from another service keeps its own row: it is a job of its own,
  not a way to upload.
- **The Refresh buttons are gone** from Videos and Live. Both pages already reload
  themselves every five seconds while anything is in flight, so the button could only
  ask again for a list that was either already coming or could not have changed.
- **Every form opens in a dialog, from a button in the header.** Live streams, API keys,
  Webhooks, Team and Connected buckets follow Move a library: the primary action sits
  top right level with the title, and the form it opens is a modal rather than a panel
  that pushes the page down and leaves you scrolling back to what you were reading. The
  steps, the fields and the wording inside each form are unchanged. Where a page told
  you to use "the button above" it now names the button.
- **A key, an invite link or a stream key still takes the cursor when it appears.** A
  native dialog hands focus back to the button that opened it, which would have left the
  cursor on "Make a key" rather than on the secret shown exactly once.

- **Moved into the Alchemist repository as the `web/` workspace**, with its history.
  The player is now a workspace dependency rather than `file:../alchemist-player`, so
  a clone builds on its own. `npm install` runs once at the repository root.

### Fixed
- **Only one sidebar row lights at a time.** On Recordings both Streams and Recordings
  were highlighted, because the row test asked "does the address start with mine" and
  `/app/live/` prefixes `/app/live/recordings/`. The breadcrumb already took the
  longest match; now the rows follow it rather than deciding again.
- **The API reference is in the repository.** `docs` in `.gitignore` had no leading
  slash, so git ignored any directory of that name at any depth — including
  `src/routes/app/docs/`, the reference the sidebar has linked to all along. A fresh
  clone built a dashboard whose Docs link went nowhere.
- **The trail no longer misnames the page it is on.** The videos list called itself
  "Video" and linked "Dashboard" at itself; Recordings called itself "Streams", because
  the first sidebar entry whose address prefixed the URL won and `/app/live/` prefixes
  `/app/live/recordings/`. A stream's own page is titled with the stream's name instead
  of eight characters of its id.
- **An armed stream stopped offering a button that could only fail.** The ingest address
  is minted once per broadcast and never returned again, so "Start this stream" on a
  stream already waiting answered "that stream is already waiting for an encoder" — in
  red, at the foot of the page, three panels below the button that caused it. The page
  now says the address cannot be shown again and points at the only way to get a fresh
  one, and every error on that page appears beside the controls rather than at the end.
- **An armed broadcast stops pretending to measure a file that has not arrived.** Its
  four figures shimmered for ever. A size made on demand no longer reads as stalled at
  0%, and anything under ten megabytes no longer prints as "0 MB".
- **A video, stream or edit that will not load offers a way back** instead of a red
  sentence and nothing else. Studio drew its whole toolbar and a Make it button over a
  video it had failed to open.
- **One-time secrets take the attention.** Minting an API key cleared the name field, so
  the browser put the cursor back in the form while the key sat above it, on screen for
  the only time it ever will be. API keys, webhook signing secrets and invite links now
  behave like stream keys: the panel takes focus when it appears.
- **Destructive actions ask first.** Switching off an API key, removing a colleague and
  deleting a live stream each happened on one click. Each now asks in place and says what
  is lost. Your own row in Team no longer offers a role menu and a remove button that the
  API refuses every time.
- **Making a camera stream says so.** There is no key to show for one, so the wizard
  emptied itself and nothing on screen acknowledged the stream existed. Watching a stream
  scrolls the player into view rather than opening it below the fold.
- **A migration preview that would not load** shimmered for ever and asked again every
  five seconds. It says so once and offers to try again.
- **The API reference stopped describing a response the engine does not send.** The usage
  example promised storage and delivery lines that are not metered yet, POST badges
  painted a colour that is not a token in this palette, and the closing line pointed at a
  specification with nothing to open.

- **A live broadcast keeps playing.** The watch panel and the video page poll the asset
  while a stream is on air, and every poll returns a freshly signed playback URL. The
  player was rebuilt on each one, so a broadcast played for a few seconds, jumped back
  to a paused play button, and did it again five seconds later. The player is now kept
  across a re-signed URL and only replaced when the video itself changes; it also takes
  a fresh signature in place when its own expires, so a long broadcast does not die at
  the end of the signature's life.
- **The LIVE badge shows during a broadcast.** The player was left to work out for
  itself whether a stream was live and always decided it was not, so a broadcast
  played with no badge and a clock counting towards a total that kept moving. The
  dashboard already knows the asset is on air and now says so.
- **An armed stream no longer reads as On air.** A stream waiting for its encoder sat
  in the video list saying On air, over a broadcast nothing had been sent to yet. The
  new `live_armed` state gets its own chip and its own sentence, so On air means a
  picture is actually going out.

### Changed
- **Making a key, adding a webhook and inviting somebody are step-by-step now,** like
  upload, connecting a bucket, moving a library and making a stream. Each ends on a
  screen saying what is about to happen, and the role menu became three cards with a
  sentence each.
- **Empty pages name the next step.** Connected buckets and Move a library ended on a
  sentence with nowhere to go.
- **Served by `adapter-node` instead of `adapter-static`.** A provider key cannot
  live in page script and a static build has no server to keep one in. Everything
  that can still be static still is: the marketing pages and the dashboard shell
  prerender and are served as files; only `/api/*` runs per request.
- **Lime is the brand.** `--color-brand` (`#C9F24D`) drives the filled button, the
  badge, the focus ring and the selection. A separate `--color-accent` carries lime
  *as text* and darkens to `#3F5C06` in the light theme, because the bright lime is
  1.6:1 on a light ground. Red still means loss and nothing else.
- **New typefaces.** Clash Display for headings, Archivo for prose, Geist Mono kept
  for figures, labels, ids and code. Geist sans is gone.
- **Self-hosted fonts.** One variable woff2 per family in `static/fonts/`, latin
  subset, preloaded — 85 KB for the set, down from 218 KB over the wire. Nothing is
  fetched from Google Fonts or Fontshare any more.
- **Rebuilt type scale.** A `clamp()` per role instead of a fixed size per role, with
  line heights and tracking set for the new faces: body 500/1.65, headings 600 at
  −0.025em and tighter, labels in mono at 0.12em.
- The landing page is an ordinary column again: one opening screen, then four
  sections separated by a rule. No pinned stack, no paging, no scroll reveals.
- Landing features and use cases are cards now, each feature with a Hugeicon on the
  right of its heading.
- Form fields rest on a 1px `muted` border and take the page's own focus ring —
  2px lime, 2px offset — instead of an inset shadow. The global `:focus-visible` no
  longer forces a 6px radius, which used to snap rounded fields square on focus.
- The mark follows the palette: `Logo.svelte` draws in `currentColor` set to the
  brand lime, and the favicon matches. The emerald-to-gold gradient is gone.

### Added
- **Go live from the browser.** Making a stream now asks where the picture comes from
  before it asks anything technical, and someone who picks their own camera is never
  shown the SRT-or-RTMP question at all — the step counter drops from four to three in
  front of them. The stream's page then asks for the camera and microphone at the
  moment they press the button and not on load, offers a picker when there is more
  than one of either, shows a preview only they can see, and puts them on air with one
  WHIP POST. Stopping closes the connection and releases the camera, so the light goes
  out. Permission refused, no camera, a camera another app is holding, an insecure
  address and a refused publish each get their own sentence.
- **Live**: `/app/live/` makes a stream, shows its key once — the only time it is on
  screen — and hands you the address to paste into OBS after you start it. The
  broadcast plays back in the page as it happens, over whichever URL the API says to
  use, and `/app/live/recordings/` lists what each finished broadcast left behind,
  opening in the ordinary video page. The sidebar asks the account whether it has
  Live at all: without it the group says plainly that it is not on this plan and
  points at /contact/, rather than promising it soon.
- **Move a library**: bring videos in from Vimeo, Bunny Stream, or a pasted list of
  links. Nothing is imported until you have seen the list and said yes — a migration
  starts by listing only, and waits. Stop, carry on and cancel are available
  throughout; cancelling keeps the videos already brought across and never touches
  anything on the far side.
- **Studio**: cut, crop, reframe and resize a video. Every edit makes a *new* video
  and never changes the one it came from, so a link already handed out keeps working.
  Reframing covers and centre-crops rather than letterboxing, so a reel fills the
  screen. Renders from the mezzanine, so editing still works after the original upload
  has been deleted.
- **Text, logos and watermarks on the frame.** Laid out the way an editor is: tools
  across the top as icons, the stage in the middle, the timeline under both, and one
  inspector on the right that is about whatever is selected. Put a thing anywhere by
  dragging it on the video, or snap it to a corner from the nine-box picker.
- **AI**, provider-agnostic through TanStack AI. `AI_PROVIDER` and `AI_MODEL` choose
  who answers, and an unset provider fails loudly rather than defaulting to one.
- A **Built for** section, giving the three use cases their own block instead of one
  trailing line under the features.

### Removed
- **`gsap`** — a dependency nothing imported.
- The pinned-stack CSS (`.screen`, `.stack--on`), the scroll-reveal classes and the
  marquee ticker, none of which anything used.
- Dead design tokens and utilities: `--color-done`, `--shadow-bulk`,
  `--shadow-dialog`, `@utility field-area`, `@utility btn-red`.
- `src/lib/DataPanel.svelte`, unreferenced.
- The hero eyebrow: the heading says what it is.

### Fixed
- `npm run verify` and `npm run check:contrast` start the real `adapter-node` server
  instead of hand-serving `build/` as files — which no longer exists in that shape and
  never exercised `/api`. `npm run preview` is replaced by `npm run start`: `vite
  preview` cannot serve an adapter-node build at all.
- The contrast sampler had been walking `/pricing/`, `/docs/`, `/about/` and
  `/edtech/` — four routes that have not existed for months. It was measuring the 404
  page five times and calling it coverage.
- `npm run verify` ran against a site that no longer exists — it asserted a Bangla
  language toggle, six routes that are not built, and three animations that were
  removed. It now checks the real routes, weighs the self-hosted fonts off disk, and
  passes.
