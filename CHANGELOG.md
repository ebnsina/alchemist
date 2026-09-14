# Changelog

## Unreleased

### Added
- Home, Course platforms, News and media, Pricing, For developers and About pages,
  bilingual in English and Bangla with an explicit toggle. English is the default and
  the canonical language.
- Indicative pricing worked out from the published cost model, marked as indicative in
  both languages wherever a figure appears, plus a worked monthly example.
- Dark and light themes, remembered per visitor.
- Open Graph and Twitter metadata on every route, a sitemap, `robots.txt` and a 404.
- `npm run verify`: contrast, overflow, language-swap, reduced-motion,
  toggle-persistence and page-weight checks. Contrast is recomputed for every visible
  text node against its real composited background, in both themes.

### Changed
- Rebuilt the visual system dark-first: `#08090A` ground, one ink at three opacities
  instead of separate text colours, hairline borders at 7% and 11%, 8px radii, large
  negatively-tracked headings over 15px body, and a single soft radial gradient behind
  the opening of the home page. Light is now an explicit choice rather than the
  system's preference.
- Tertiary text is 0.48 opacity, not 0.44, because 0.44 fails WCAG AA on this ground.
- Rewrote every page outside `/docs/` in plain language. No encoding or delivery
  vocabulary reaches a reader who is not a developer.
- New visual system: warm stone paper with warm near-black ink and a single brass
  accent, light as the primary surface, a 17px text size and a narrower type scale.

### Removed
- All entrance and scroll motion. What is left is a hover state and the focus ring.
- The icon set, the card grids, the typographic labels above section headings, and the
  competitor price comparison, which came from figures recorded from memory rather
  than checked against anybody's published prices.

### Fixed
- The 404 is a prerendered page rather than the adapter's SPA fallback, which rendered
  empty with no client runtime to fill it in.
- Asset URLs are absolute, so `404.html` loads its stylesheet at any path.
- A wide code block no longer pushes the page sideways at 360px.
