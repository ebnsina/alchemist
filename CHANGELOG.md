# Changelog

## Unreleased

### Added
- Home, For edtech, For media, Pricing, Docs and About pages, bilingual in English and
  Bangla with an explicit toggle. English is the default and the canonical language.
- Dark and light themes, remembered per visitor.
- Open Graph and Twitter metadata on every route, a sitemap, `robots.txt` and a 404.
- `npm run verify`: overflow, language-swap, toggle-persistence and page-weight checks.

### Fixed
- The 404 is a prerendered page rather than the adapter's SPA fallback, which rendered
  empty with no client runtime to fill it in.
- Asset URLs are absolute, so `404.html` loads its stylesheet at any path.
- The main navigation is a wrapping list; the `<details>` version it replaced never
  rendered its links at desktop width.
