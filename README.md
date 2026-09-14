# alchemist-web

Marketing site for Alchemist. SvelteKit + TypeScript + **Tailwind CSS v4**, static
output via `adapter-static`.

```sh
npm install
npm run dev       # local development
npm run build     # static output in build/
npm run verify    # checks the built output; see below
npm run check     # svelte-check
```

## Structure

```
src/app.html                     Inter from Google Fonts, language pre-paint script
src/app.css                      @import "tailwindcss" + @theme + custom utilities
src/routes/+layout.svelte        Nav + Footer
src/routes/+page.svelte          composes the landing sections
src/lib/components/              Logo, Nav, Hero, LogoTicker, Stats, Features,
                                 HowItWorks, Pricing, Testimonials, FAQ, CTA, Footer
src/lib/utils/scroll-reveal.js   IntersectionObserver action
src/routes/{edtech,media,pricing,docs,about,404}/  the platform-facing pages
```

`Logo.svelte` is the only file holding the mark — a flask with a play triangle in it.
Swap that one file to change the logo everywhere.

## Palette and contrast

Deep sea green on near-black, with white → emerald → gold gradient text for the
transformation. Every pair was computed against `#06100d`, not eyeballed:

| token | value | contrast |
|---|---|---|
| text | `#e5e7eb` | 15.59:1 |
| **muted** | **`rgba(255,255,255,0.55)`** | **6.24:1** |
| emerald | `#10b981` | 7.61:1 |
| emerald light | `#34d399` | 10.04:1 |
| gold | `#fbbf24` | 11.57:1 |
| seafoam | `#2dd4bf` | 10.37:1 |

Two values differ from the brief, both because the brief's would fail AA:

- **muted is `0.55`, not `0.40`.** At 0.40 it measures **3.80:1** and fails.
- **the emerald button uses `#06100d` text, not white.** White on `#10b981` is
  **2.54:1**; dark-on-emerald is 7.61:1.

`#064e3b` is 1.99:1 against the body and is used only as a fill behind white text or
as a gradient stop, never as text.

## Motion

`scroll-reveal.js` adds `.scroll-fade` itself, so **with scripting off nothing is
hidden** — the class that sets `opacity: 0` never lands. Under
`prefers-reduced-motion: reduce` the action marks the element visible immediately and
never observes anything, and the CSS stops the ticker, the shine sweep, the progress
bar and every transition. `npm run verify` asserts both directions.

## The three substituted sections

`LogoTicker`, `Stats` and `Testimonials` sit where a template would put invented
customers, invented metrics and invented people. The layout, animation and spacing are
unchanged; the content is not fabricated:

- **LogoTicker** — the kinds of work the tool is for, under "Built for creators,
  teachers, and small businesses". No claim that anyone is a customer.
- **Stats** — 93% smaller, under 60s, 8K, any device. Facts about the product rather
  than a customer count or a star rating.
- **Testimonials** — three use cases in neutral voice under "Who it's for". No names,
  no roles, no avatars, no ratings.

Do not fill these with people, companies, counts or ratings that do not exist.

## Verification

`npm run verify` serves `build/` and checks: WCAG AA contrast for every visible text
node against its real composited background (gradient text is skipped and its stops
checked separately), no horizontal scroll at 360px in both languages, no language
leaking through the CSS swap, every scroll reveal completing, nothing hidden with
scripting off, reduced motion stopping all four animations while they still run when
motion is allowed, the language toggle persisting, and the page weight.

## Before this goes live

- `src/lib/site.ts` still has `alchemist.example` for the origin, contact address and
  repository links. Also in `static/robots.txt`.
- The landing page describes a consumer file converter; `/edtech/`, `/media/`,
  `/pricing/`, `/docs/` and `/about/` describe multi-tenant video infrastructure sold
  to platforms. Those are two products and the site currently claims both.
