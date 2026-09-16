# Alchemist web

Marketing site and customer dashboard for Alchemist, a video infrastructure API.
SvelteKit 2 + Svelte 5 (runes) + TypeScript + Tailwind CSS v4, served by
`adapter-node`. One npm workspace of the Alchemist repository; the engine is Go at
the root and the player is `player/`, which this depends on as a workspace.

`DESIGN.md` is the design reference and `src/app.css` is its implementation — read
both before changing anything visual, and keep them in agreement.

## Commands

Run them here, or from the repository root with `-w alchemist-web`. The single
`npm install` lives at the root and installs this and the player together.

```sh
npm run dev                      # :5173
npm run build                    # build/ — a Node server plus prerendered pages
npm run check                    # svelte-check — must stay at 0 errors
npm run verify                   # serves build/ and checks it; run after build
npm run preview -- --port 4321   # then: npm run check:contrast
```

`verify` and `check:contrast` are the gate. Both must pass before anything ships.

## Rules specific to this workspace

- **No new dependencies without asking.** `gsap` was carried for months without a
  single import. Reach for CSS, then a platform feature, then what is installed.
- **No CDN in the critical path.** Fonts are self-hosted in `static/fonts/`, one
  variable woff2 per family, latin only. Adding a `<link>` to a font host undoes a
  deliberate 130 KB saving.
- **Two hues, both load-bearing.** Lime is the brand, red means loss. Anything else
  that needs to stand out gets weight or fill, not colour. Lime fills; use
  `--color-accent` when lime has to be text, because it is a different value in the
  light theme.
- **Icons are Hugeicons.** `@hugeicons/svelte` + `@hugeicons/core-free-icons`.
- **Formatting goes through `Intl`.** No hand-rolled formatters, ever. Money renders
  through `src/lib/Money.svelte`.
- **Validation is Valibot** where a schema is needed.
- **Comments: one line, two at most.** The existing ones explain *why*, not what.
  Match that.
- **Every state ships.** Loading, empty and error, in plain language. No raw backend
  error text reaches a user; the API's error codes map to sentences.
- **No invented social proof.** No customer counts, logos, testimonials, names or
  ratings. None of them exist. Facts about the product instead.
- **The FAQ answers worries, not architecture.** Every question on the landing page is
  one a buyer would actually ask out loud — *will it play on my viewers' phones*, *what
  does it cost me*, *what if I want to leave*. Never *how isolated is one account from
  another* or *do we have to poll*: those are our vocabulary, not theirs. Answer in
  plain language, lead with the reassurance, and let the consequence carry the
  technical fact rather than naming it. Row-level security becomes "the database
  enforces it, even a query written wrongly returns only your own videos". Webhooks
  become "we tell you when it is ready". If an answer needs a term from `docs/`, it is
  written for the wrong reader — the API reference is where that belongs.
- **No dark patterns.** Clear pricing, no hidden fees, nothing that pressures.

## Shape of the code

- `src/routes/+page.svelte` is the whole landing page, content arrays at the top and
  markup below. It is meant to stay one file.
- `src/routes/app/` is the dashboard and opts out of the public chrome in
  `+layout.svelte`.
- `src/lib/site.ts` holds the origin, contact address and sitemap routes, all still
  `alchemist.example` placeholders.
- `scripts/check.mjs` and `scripts/check-contrast-over-art.mjs` are the checks. When
  you remove a feature, remove its assertion in the same change — both scripts had
  rotted into asserting a site that no longer existed.

## Git

- Author as `ebnsina <ebnsina.me@gmail.com>`. No `Co-Authored-By` trailer.
- Prose commit subjects, matching the existing history. Not Conventional Commits.
- Keep `CHANGELOG.md` current with every user-facing change.
- `docs/` and `data/` are gitignored and never committed.
