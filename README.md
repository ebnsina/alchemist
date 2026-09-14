# alchemist-web

The public marketing site for Alchemist — video transcoding and delivery sold as an
API, hosted inside BDIX, encoded for viewers who pay per megabyte.

SvelteKit + TypeScript, `adapter-static`. The output in `build/` is plain files:
copy it to any static host or serve it from nginx on a VPS.

```sh
npm install
npm run dev       # local development
npm run build     # static output in build/
npm run verify    # builds must exist first; see below
npm run check     # svelte-check
```

## Decisions worth knowing before you change something

**No client framework ships.** `csr = false` in `src/routes/+layout.ts`. The only
interactive things on the site are the language and theme toggles, and both are a
dozen lines of delegated DOM in `src/app.html`. Shipping ~40 KB of hydration to a page
whose whole argument is saving the viewer's data would be self-refuting. A first visit
to the home page is about 60 KB gzipped including both fonts; every other route is
about 9 KB once they are cached.

**Both languages are in the HTML; CSS hides one.** `src/lib/T.svelte` renders the
English and the Bangla, and `html[lang]` rules in `src/app.css` display one. There is
no flash, it works with JavaScript off, and `display: none` keeps the hidden copy out
of the accessibility tree. English is canonical and is what goes in `<head>` — a
document has one title. Language is never guessed from IP or `navigator.language`; the
toggle is the only thing that decides, and the choice is remembered in `localStorage`
inside a `try`/`catch`.

**Fonts are self-hosted and subset to latin.** Google Sans Flex for UI, Geist Mono for
numbers and code, 46 KB together. The `unicode-range` deliberately excludes Bengali
codepoints so the browser never tries a latin face for them and falls through to the
system's Noto Sans Bengali, which shapes যুক্তাক্ষর correctly. Same reasoning as
`alchemist-player`'s `src/styles.ts`.

**Icons are hand-authored** on the Hugeicons 24px grid in `src/lib/icons.ts`. The free
icon package is ~72 MB; fourteen path strings are under 2 KB.

**Brass on near-black.** Alchemy is transmutation, the player already wears this
palette, and it is not the blue every other video API reaches for.

## Rules this site holds itself to

- No customer logos, testimonials, case studies, uptime figures or compliance badges.
  There are no customers yet, so there is nothing honest to put there.
- No countdown timers, fake scarcity, pre-ticked boxes or asterisked promises.
- Every price is a placeholder and is rendered inside `.ph`, which draws a dashed
  brass box and the literal word "placeholder" in both themes and both languages.
  Grep for `class="ph"` before launch.
- Every number is either measured, cited to whoever measured it, or marked as a
  projection. Claims about BDIX describe what the system is being built to do; the
  BDIX edge is not running yet and the site says so on four pages.

## Verification

`npm run verify` serves `build/` and checks the things that can silently break: no
horizontal scroll at 360px across every route in both languages and themes, no
language leaking through the CSS swap, the toggles working and surviving a reload,
no module script sneaking back in, and it prints the page weight.

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

Grep for `alchemist.example` — the domain, the contact address and the repository links
are all placeholders. Replace `SITE.origin` in `src/lib/site.ts`, `static/robots.txt`,
and set real prices in `src/routes/pricing/+page.svelte`, removing the `.ph` wrappers
and the placeholder banner at the top of that page as you do.
