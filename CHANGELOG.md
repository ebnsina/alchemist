# Changelog

## Unreleased

### Changed
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
