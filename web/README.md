# alchemist-web

Public marketing site and customer dashboard for Alchemist. SvelteKit + TypeScript +
**Tailwind CSS v4**, served by **`adapter-node`**.

It was static until the AI features arrived. A provider key cannot live in page
script, and a static build has no server to keep one in — so there is a Node process
now. Everything that can still be static still is: the marketing pages and the
dashboard shell prerender at build time and are served as files. Only `/api/*` runs
per request.

One `npm install` at the repository root installs this and the player together; there
is no install to run in here. Every command below works from this directory, or from
the root with `-w alchemist-web`.

```sh
npm install        # at the repository root, once, for both workspaces
npm run dev        # local development on :5173
npm run build      # build/ — a Node server plus prerendered pages
node build/index.js            # run it: PORT and ORIGIN come from the environment
npm run check      # svelte-check
npm run verify     # builds nothing; serves build/ and checks it — see below
npm run check:contrast           # samples the real pixel under every text node
```

## Structure

```
src/app.html                  three self-hosted font preloads, nothing remote
src/lib/server/ai.ts          provider-agnostic AI config; never bundled to the client
src/routes/api/ai/+server.ts  the one route that runs per request
src/app.css                   @import "tailwindcss" + @theme + custom utilities
src/routes/+layout.svelte     skip link + Footer; the dashboard opts out of chrome
src/routes/+page.svelte       the landing page, all of it, in one file
src/routes/{login,signup,contact,404}/
src/routes/app/               the dashboard: overview, keys, upload, videos, docs
src/lib/components/           Logo, Footer, Player, ConversionFlow, DataTable, …
src/lib/site.ts               origin, contact address, sitemap routes
static/fonts/                 Archivo, Clash Display, Geist Mono — one woff2 each
```

`Logo.svelte` is the only file holding the mark — a flask with a play triangle in it.
Swap that one file and `static/favicon.svg` to change the logo everywhere.

## AI

Provider-agnostic through TanStack AI. `AI_PROVIDER` and `AI_MODEL` choose who
answers; the adapter reads its own key from the environment and refuses to accept one
in code, so a key cannot be captured by application code by accident. Adapters for
Gemini and Ollama exist upstream and are deliberately not installed until wanted —
adding one is a package and a case in `adapterFor`.

Nothing is guessed: an unset `AI_PROVIDER` fails the request with a message naming the
variable, rather than falling back to a default nobody chose.

## Design

`DESIGN.md` is the reference and `src/app.css` is the implementation; the two are
meant to agree. In short: monochrome console, lime brand, red only for loss. Dark is
the default and the light theme is a token swap, not a second stylesheet.

## Fonts

Self-hosted from `static/fonts/`, one variable woff2 per family, latin subset only —
85 KB for all three, preloaded, with no CDN in the critical path.

| Family | Role | Size |
|---|---|---|
| Clash Display | headings | 29 KB |
| Archivo | body | 35 KB |
| Geist Mono | figures, labels, ids, code | 23 KB |

To change one: fetch the woff2, drop it in `static/fonts/`, update the `@font-face`
at the top of `app.css` and the `<link rel="preload">` in `app.html`.

## Motion

There is no animation library — `gsap` was a dependency nothing imported and it is
gone. What is left is CSS transitions, the skeleton shimmer and the hero card cycling
through its three states. All of it stops under `prefers-reduced-motion: reduce`, and
`npm run verify` asserts that.

## Verification

`npm run verify` serves `build/` and checks WCAG AA contrast for every visible text
node against its real composited background, no horizontal scroll at 360px in either
theme, that reduced motion stops every transition while they still run when motion is
allowed, and the page weight per route.

`npm run check:contrast` is the harder one: it hides the glyphs, screenshots each
viewport down the page and samples the actual pixel under every text node, so text
sitting over artwork rather than over a background-color is measured honestly. It starts the
server itself; `npm run build` is the only prerequisite.

## Content rules

Facts about the product, never invented customers. No counts, no logos, no
testimonials, no star ratings, no names — none of those exist yet and none of them go
on the page until they do.

## Before this goes live

- `src/lib/site.ts` still has `alchemist.example` for the origin, contact address and
  repository links. Also in `static/robots.txt`.
- Pricing is marked "not final until launch" on the page. Take that down or make the
  rates real.
