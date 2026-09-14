# Changelog

## Unreleased

### Changed
- Rebuilt on **Tailwind CSS v4** (`@import "tailwindcss"` with a `@theme` block) and
  **Inter** from Google Fonts, replacing the hand-rolled stylesheet and the
  self-hosted faces.
- New palette: deep sea green on near-black, with white → emerald → gold gradient
  text. `muted` raised from `0.40` to `0.55` opacity and the emerald button given dark
  text, because the specified values fail WCAG AA.
- The landing page is now composed from one component per section under
  `src/lib/components/`, with an `IntersectionObserver` action in
  `src/lib/utils/scroll-reveal.js`.
- Scroll reveals, a scrolling ticker, a shine sweep and an animated progress bar are
  back; all four stop under `prefers-reduced-motion`.
- Client-side rendering is on again, which the reveal action requires. The site is no
  longer zero-JavaScript.

### Removed
- The light theme and its toggle: the brief specifies a single dark palette.
- The self-hosted Google Sans Flex and Geist Mono files.

### Added
- `npm run verify` now also asserts that every scroll reveal completes and that no
  content is hidden when scripting is off.
