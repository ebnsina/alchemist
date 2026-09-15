# Alchemist design system

Monochrome console with a single brand hue. State is carried by weight and contrast;
lime marks the one thing to act on. Taken from the stream-migrate migration console
and applied to both surfaces — the public site and the customer dashboard.

This file is the reference; `src/app.css` is the implementation, and the two are
meant to agree.

## Principles

Two hues, both load-bearing. **Lime is the brand**: the filled button, the focus
ring, the figure worth reading. **Red means loss**, and nothing else. Everything in
between is a row that is important because it is heavier and darker, not because it
is coloured.

Lime fills; lime rarely writes. The bright `#C9F24D` is 1.6:1 on a light ground, so
text-lime is a separate token that darkens to `#3F5C06` in the light theme while the
fill stays the same in both.

Numbers live in mono with tabular figures so columns line up without anybody padding
them.

Dark is the default because the console is a tool that sits open all day.

## Colours

| Token | Dark | Light | Use |
|---|---|---|---|
| `--color-bg` | `#0D0D0F` | `#FAFAFA` | The page |
| `--color-card` | `#161618` | `#FFFFFF` | Panels, rails, dialogs |
| `--color-sunk` | `#1E1E21` | `#F1F1F2` | Hover, rules, inactive fills, skeletons |
| `--color-ink` | `#F2F2F3` | `#0F0F10` | Titles, values, body |
| `--color-dim` | `#9A9AA0` | `#6E6E73` | Secondary prose, captions |
| `--color-faint` | `#82828A` | `#6E6E73` | Labels, units, timestamps |
| `--color-brand` | `#C9F24D` | `#C9F24D` | The lime. Fills and focus rings, never body text |
| `--color-on-brand` | `#0D0D0F` | `#0D0D0F` | Text on lime |
| `--color-accent` | `#C9F24D` | `#3F5C06` | Lime *as text*: figures, ticks, active state |
| `--color-solid` | → brand | → brand | Filled surfaces: primary button, badge, selection |
| `--color-on-solid` | → on-brand | → on-brand | Text on a filled surface |
| `--color-muted` | `#33333A` | `#D8D8DC` | Finished, inactive fills |
| `--color-red` | `#E08A8A` | `#8A1F1F` | Failure, destructive. Nothing else |

Measured. Dark on `#0D0D0F`: ink 17.35:1, dim 6.94:1, faint 5.09:1, red 7.56:1,
accent 15.06:1. Light on `#FAFAFA`: ink 18.36:1, dim 4.86:1, faint 4.86:1, red
8.76:1, accent 7.33:1. Near-black on lime is 15.06:1 either way.

Faint is the one departure from the console this is copied from. There it is
`#6A6A71` and `#A0A0A6`, which measure 3.62:1 and 2.49:1 — under AA. Same role,
lifted until it passes. It is still only for 11px uppercase labels that repeat down
a column, where the word is scaffolding for the number beside it; anything read once
uses ink or dim.

The filled button is lime with near-black text, 15.06:1 in both themes. One per view.

## Typography

Three families, each with one job. **Clash Display** sets headings — a display face
with enough width and character to carry a line on its own. **Archivo** sets
everything read as prose. **Geist Mono** sets everything a machine produced: figures,
labels, ids, timestamps, code.

All three are self-hosted from `static/fonts/`, one variable woff2 per family, latin
subset only — 85 KB for the set, and no CDN in the critical path.

Sizes are fluid: a `clamp()` per role rather than a breakpoint per role.

| Role | Size | Weight | Line | Tracking |
|---|---|---|---|---|
| `.big` | 2.5 → 4rem | 600 | 0.94 | −0.035em |
| `h1` | 2.25 → 3.5rem | 600 | 0.98 | −0.03em |
| `h2` | 1.75 → 2.5rem | 600 | 1.04 | −0.025em |
| `h3` | 1.06 → 1.25rem | 600 | 1.3 | −0.015em |
| `.lead` | 1.06 → 1.25rem | 400 | 1.55 | −0.012em, dim |
| body | 15 → 17px | 500 | 1.65 | −0.005em |
| `.title` | 1 → 1.125rem | 600 display | 1.35 | −0.015em |
| `.sub` | 15px | 400 | 1.6 | −0.005em, dim |
| `.label` | mono 11px | 500 | 1.3 | 0.12em, uppercase, faint |
| `.num` | mono | 700 | — | −0.02em, tabular |
| `.mono` | mono 11px | 400 | — | 0.02em, faint |

Body weight is 500: Archivo at 400 thins out on a dark ground and at 600 it shouts.
Headings hold 600 — Clash Display at 700 closes its counters.

Every number, date, duration, price, id and timestamp goes through `Intl` and is set
in mono.

## Shapes and spacing

Radii: `6px` skeletons and inline chips, `9px` rows, `10px` small buttons, `12px`
buttons and cards, `20px` the dashboard frame, `99px` chips, dots and progress bars.
Every drawn corner is `corner-shape: squircle`, which degrades to a plain round where
it is not supported.

Rows are 8px/12px padded with a 12px gap. Rails are 22px/20px. Landing sections are
`py-24` and separate with a 1px `sunk` rule, never with a shadow.

## Elevation

None. There are no shadow tokens: static content uses `card` against `bg`, or a
`sunk` rule. Add one only when something genuinely floats over the page.

## Components

- **button** — uppercase, 700, 13px at 0.06em, radius 12, `sunk` background and ink
  text; `.btn-solid` fills lime for the one primary action. Hover drops opacity to
  .85. `.btn-sm` is 12px at radius 10.
- **node (row)** — radius 9, hover `sunk`, a 24px round badge, a title, a faint
  sub, an optional 120×4 progress rail, and a chip at the end. A finished row goes
  to 55% opacity and stops responding to hover.
- **chip** — mono 10px, 500, uppercase, 0.1em, radius 99, min-width 54, centred.
  `sunk`+faint for inactive, `ink`+`card` for active.
- **badge** — 24px circle, mono 700. Filled lime for next, outlined for stale,
  `muted` for done.
- **metric** — a faint uppercase mono label and a mono 700 value in `accent`, on one
  baseline.
- **ring** — a 108px donut with the percentage inside and a faint caption under it,
  drawn with one `stroke-dasharray` rather than two arcs.
- **skeleton** — `sunk` block with a shimmer, in the shape of the thing that is
  loading, because a cold query takes long enough that a spinner says nothing.

Every surface that can load, be empty, or fail ships all three states.

## Do and don't

- **Do** carry state with weight, fill and contrast.
- **Do** put every number in mono with tabular figures.
- **Do** keep one filled button per view; everything else is a `sunk` button.
- **Don't** introduce a third hue. Lime and red are the whole palette; if something
  needs to stand out beyond those, make it heavier or fill it.
- **Don't** set body text in lime. Fill with it, or use `accent`, which is a
  different value in the light theme for exactly this reason.
- **Don't** use red for anything but failure and destruction.
- **Don't** add a shadow to static content.
- **Don't** ship framework default styling, an icon set other than Hugeicons, a
  font from a CDN, or a dark-pattern flow.

## Motion

There is almost none, and no animation library. The landing page opens on one screen
and the rest is an ordinary column beneath it — no paging, no pinned stack, no scroll
reveals, no veil. Four short sections under the fold say what it does, how it is
integrated, what it costs and what happens when the customer leaves.

What is left is the hero card cycling through upload, encode, link, which is a
diagram rather than decoration, and hover transitions. Both stop under reduced
motion.

## Navigation

There is no header. The landing page carries its own mark at the top of its opening
screen; the dashboard has its sidebar; everything else is in the footer.
Signing in lives there too — it is for people who already have an account, and it
does not need a permanent slot next to the thing that gets new ones.

## Where it is used

The dashboard overview is the reference layout: the node list on the left, and a
rail on the right holding the ring and the metrics — everything that is a figure
rather than an action.

## The mark

`src/lib/components/Logo.svelte` is the only file holding it — a flask with a play
triangle in it, drawn in `currentColor` set to the brand lime, so it follows the
palette instead of carrying its own. `static/favicon.svg` is the same shape on a
near-black square. Swap those two to change the logo everywhere.

## Checks

`npm run verify` builds, serves `build/`, and asserts AA contrast against the real
composited background, no horizontal scroll at 360px in both themes, reduced motion
stopping every transition, and the page weight per route.

`npm run check:contrast` samples the real pixel under every text node on the public
pages, in both themes, at every viewport step, and fails under 4.5:1. It needs a
server: `npm run preview -- --port 4321` first.
