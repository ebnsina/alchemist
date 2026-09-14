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

**Almost no motion.** A hover state on links and buttons, and the focus ring. No
entrance animations, no scroll-triggered anything, no libraries. `prefers-reduced-motion`
removes what little there is, and `npm run verify` asserts that it actually does.

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

**Warm stone paper, warm near-black ink, one accent.** Light is the primary surface.
Brass because alchemy is al-kimiya and because `alchemist-player` already wears it.
Every text pair clears WCAG AA; the numbers are in the comment at the top of `app.css`.

## Rules this site holds itself to

- No customer logos, testimonials, case studies, uptime figures or compliance badges.
  There are no customers yet, so there is nothing honest to put there.
- No countdown timers, fake scarcity, pre-ticked boxes or asterisked promises.
- Every price is an indicative example, rendered inside `.indic`, which draws a brass
  edge and captions the figure "Indicative" in both languages. The pricing page also
  opens with a notice saying no price has been set. Grep for `class="indic"`.
- Every number is either measured, cited to whoever measured it, or marked as a
  projection or an indicative example. Claims about servers inside Bangladesh describe
  what the system is being built to do; those servers are not running, and the site
  says so on four pages.

## Verification

`npm run verify` serves `build/` and checks what can silently break: no horizontal
scroll at 360px across every route in both languages and themes, no language leaking
through the CSS swap, `prefers-reduced-motion` actually removing motion, the toggles
working and surviving a reload, no module script sneaking back in. It prints the page
weight at the end.

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

- **Set real prices.** Every figure on `/pricing/` is an indicative example worked out
  from the cost model, not a decision. Replace them and remove the `.indic` wrappers
  and the notice at the top of that page together, so a real price is never shown in
  the indicative treatment or an indicative one without it.
- **Register the domain.** Grep for `alchemist.example`: it is in `src/lib/site.ts`
  (site origin, contact address, repository links) and `static/robots.txt`.
- **Re-check the two dated figures**: the BTRC June 2026 user counts, and the ৳135 to
  the euro conversion used on the pricing page.
