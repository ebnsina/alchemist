# alchemist-web

The public site for Alchemist — video hosting for apps and websites in Bangladesh.

SvelteKit + TypeScript, `adapter-static`. The output in `build/` is plain files: copy
it to any static host or serve it from nginx on a VPS.

```sh
npm install
npm run dev       # local development
npm run build     # static output in build/
npm run verify    # checks the built output; see below
npm run check     # svelte-check
```

## Decisions worth knowing before you change something

**Write for the buyer, not the engineer.** The person deciding to use this runs a
course platform or a news site. Outside `/docs/`, there is no encoding vocabulary on
this site — no bitrate ladders, no adaptive streaming, no CMAF, no SAMPLE-AES, no
just-in-time packaging. Those are all real and all beside the point to the reader.
Say what they worry about instead: the video starts instead of buffering, it does not
eat the viewer's data allowance, and the course they sell cannot simply be passed
around. `/docs/` is the one page where precise terms are correct, because developers
read it.

**Almost no motion.** A hover state on links and buttons and the focus ring, all at
140ms, plus one very slow drift on the hero background: four layered
`radial-gradient`s on a single element, transform-only, 120 seconds per cycle. It is
CSS rather than a canvas because a shader costs battery and bytes on exactly the
phones this product exists for. `prefers-reduced-motion` removes all of it, and
`npm run verify` asserts both that it does and that the drift exists when motion is
allowed.

**No client framework ships.** `csr = false` in `src/routes/+layout.ts`. The language
and theme toggles are about twenty lines of delegated DOM in `src/app.html`. A first
visit to the home page is about 57 KB gzipped including both fonts; every other page
is about 9 KB once they are cached.

**Both languages are in the HTML; CSS hides one.** `src/lib/T.svelte` renders the
English and the Bangla, and `html[lang]` rules in `src/app.css` display one. No flash,
works with JavaScript off, and `display: none` keeps the hidden copy out of the
accessibility tree. English is canonical and is what goes in `<head>`. Language is
never guessed from IP or `navigator.language`; the toggle decides, and the choice is
remembered in `localStorage` inside a `try`/`catch`.

**Fonts are self-hosted and subset to latin.** Google Sans Flex for text, Geist Mono
for columns of figures and code, 46 KB together. The `unicode-range` deliberately
excludes Bengali codepoints so the browser never tries a latin face for them and falls
through to the system's Noto Sans Bengali, which shapes যুক্তাক্ষর correctly. Same
reasoning as `alchemist-player`'s `src/styles.ts`.

**Dark first.** `#08090A` ground, `#0F1011` raised, `#141516` elevated. Light is an
explicit choice from the header, not the system's — there is no
`prefers-color-scheme` block, which is one less thing to keep in sync and a few bytes
less CSS.

**Hierarchy is one ink at three opacities, not three colours.** Primary 0.95,
secondary 0.63, tertiary 0.48. Tertiary sits at 0.48 rather than the 0.44 the look
wants because 0.44 measures 4.33:1 against `#08090A` and fails AA; 0.48 is 4.98:1.
Borders are 0.07 and 0.11 — dividers you sense before you see. The brass appears
about once a screen: a link, the focus ring, the default row of the data table.
`npm run verify` recomputes every visible text/background pair in both themes and
fails the build under AA, so this cannot drift.

**Type carries it.** Headings are large and negatively tracked (h1 -0.032em at
34–56px, line-height 1.06); body is small at 15px/1.6; weights are 400, 500 and 600
only. Bangla gets 16px and more leading, because Bengali needs the room.

## Rules this site holds itself to

- No customer logos, testimonials, case studies, uptime figures or compliance badges.
  There are no customers yet, so there is nothing honest to put there.
- No countdown timers, fake scarcity, pre-ticked boxes or asterisked promises.
- **No invented people or organisations.** No testimonial quotes with names, no
  customer logo wall, no "trusted by N customers", no star rating. Those slots are
  filled instead by a line naming the kinds of organisation this is for, by the
  megabytes-per-hour diagram, and — where testimonials would go — by nothing.
- Product capabilities are written in present tense, like any product page. Figures
  that are not yet fixed carry one quiet line rather than a badge on every number:
  "Pricing is not final until launch" appears once on the pricing page, once under the
  worked examples on the home page, and once in the footer.
- Rollout status lives in one quiet place, the "Where it stands" block on `/about/`.
  It is deliberately not on the home page or the audience pages.

## Verification

`npm run verify` serves `build/` and checks what can silently break: WCAG AA contrast
for every visible text node against its actual composited background in both themes,
no horizontal scroll at 360px across every route in both languages and themes, no
language leaking through the CSS swap, `prefers-reduced-motion` actually removing
motion, the toggles working and surviving a reload, and no module script sneaking back
in. It prints the page weight at the end.

The contrast check composites translucent ancestors down to the first opaque one, and
handles `color(srgb …)` as well as `rgb()` — `color-mix()` computes to the former with
0–1 channels, and reading those as 0–255 turns a pale translucent header into
near-black and invents failures that are not there.

## Deploying

`build/` is static files. Point a web root at it and map unknown paths to `404.html`:

```nginx
root /srv/alchemist-web;
error_page 404 /404.html;
location / { try_files $uri $uri/ $uri/index.html =404; }
```

Asset paths are absolute (`kit.paths.relative = false`) because `404.html` is served at
whatever path the visitor asked for, so relative hrefs would resolve against that path
and 404 as well.

## Before this goes live

- **Set real prices.** Every figure on `/pricing/`, and the three worked examples on
  the home page, come from the cost model rather than a decision. When they are fixed,
  drop the three "not final until launch" lines in the same change.
- **Decide the commercial terms the copy already states**: no minimum, no setup fee,
  no exit fee, cancel any time, price changes announced before they take effect.
  Nothing on the site promises a free tier or a trial, because neither has been
  decided — do not add one to the page before it exists in the billing code.
- **Register the domain.** Grep for `alchemist.example`: it is in `src/lib/site.ts`
  (site origin, contact address, repository links) and `static/robots.txt`.
- **Re-check the two dated figures**: the BTRC June 2026 user counts, and the ৳135 to
  the euro conversion used on the pricing page.
