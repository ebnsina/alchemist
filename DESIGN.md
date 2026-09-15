# Alchemist design system

Monochrome console. State is carried by weight and contrast, not by hue. Taken from
the stream-migrate migration console and applied to both surfaces — the public site
and the customer dashboard.

This file is the reference; `src/app.css` is the implementation, and the two are
meant to agree.

## Principles

There is no accent colour. A row is important because it is heavier and darker, not
because it is blue. The only hue in the system is red, and it means loss.

Everything is dense: 8px row padding, 15px base text, and a base weight of 600 —
this is a working surface, not a brochure. Numbers live in mono with tabular figures
so columns line up without anybody padding them.

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
| `--color-solid` | `#F2F2F3` | `#0F0F10` | Filled surfaces: primary button, active chip |
| `--color-on-solid` | `#0D0D0F` | `#FFFFFF` | Text on a filled surface |
| `--color-muted` | `#33333A` | `#D8D8DC` | Finished, inactive fills |
| `--color-done` | `#1B1B1E` | `#F4F4F5` | The wash on a finished row |
| `--color-red` | `#E08A8A` | `#8A1F1F` | Failure, destructive. Nothing else |

Measured. Dark on `#0D0D0F`: ink 17.35:1, dim 6.94:1, faint 5.09:1, red 7.56:1.
Light on `#FAFAFA`: ink 18.36:1, dim 4.86:1, faint 4.86:1, red 8.76:1.

Faint is the one departure from the console this is copied from. There it is
`#6A6A71` and `#A0A0A6`, which measure 3.62:1 and 2.49:1 — under AA. Same role,
lifted until it passes. It is still only for 11px uppercase labels that repeat down
a column, where the word is scaffolding for the number beside it; anything read once
uses ink or dim.

The filled button inverts: `solid` background, `on-solid` text, 17:1 either way.

## Typography

Geist for everything human, Geist Mono for everything counted — one family in two
voices, drawn for product interfaces, which is what this is. The base weight is 600;
headings and values are 800.

| Role | Size / weight |
|---|---|
| `big` | 44px / 800 / 1.0 |
| `h1` | 28px / 800 / −0.02em |
| `h2` | 20px / 800 / −0.02em |
| body | 15px / 600 |
| `.title` | 14px / 700 / −0.01em |
| `.sub` | 12px / 600, faint |
| `.label` | 11px / 700, uppercase, 0.07em, faint |
| `.metric b` | mono 17px / 800, tabular |
| `.mono` | mono 11px / 500, faint |

Every number, date, duration, price, id and timestamp goes through `Intl` and is set
in mono.

## Shapes and spacing

Radii: `6px` skeletons and inline chips, `9px` rows, `10px` small buttons and bulk
bars, `12px` buttons, `18px` the dock, `20px` dialogs, `99px` chips, dots and
progress bars. Every drawn corner is `corner-shape: squircle`, which degrades to a
plain round where it is not supported.

Rows are 8px/12px padded with a 12px gap. Rails are 22px/20px. Sections separate
with a 1px `sunk` rule, never with a shadow.

## Elevation

Two shadows, both for things that float: the bulk bar
(`0 -1px 0 sunk, 0 8px 24px rgba(0,0,0,.12)`) and the dialog
(`0 18px 50px rgba(20,30,50,.2)`). Static content uses `card` against `bg`, or a
`sunk` rule.

## Components

- **button** — uppercase, 800, 14px, radius 12, `sunk` background and ink text;
  `.solid` inverts to `solid`/`on-solid` for the one primary action. Hover drops
  opacity to .85. `.sm` is 12px at radius 10.
- **node (row)** — radius 9, hover `sunk`, a 24px round badge, a title, a faint
  sub, an optional 120×4 progress rail, and a chip at the end. A finished row goes
  to 55% opacity and stops responding to hover.
- **chip** — 10px, 800, uppercase, radius 99, min-width 54, centred. `sunk`+faint
  for inactive, `ink`+`card` for active.
- **badge** — 24px circle, 800. Filled `solid` for next, outlined for stale, `muted`
  for done.
- **metric** — a faint uppercase label and a mono 800 value on one baseline.
- **ring** — a 108px donut with the percentage inside and a faint caption under it,
  drawn with one `stroke-dasharray` rather than two arcs.
- **skeleton** — `sunk` block with a shimmer, in the shape of the thing that is
  loading, because a cold query takes long enough that a spinner says nothing.

Every surface that can load, be empty, or fail ships all three states.

## Do and don't

- **Do** carry state with weight, fill and contrast.
- **Do** put every number in mono with tabular figures.
- **Do** keep one filled button per view; everything else is a `sunk` button.
- **Don't** introduce a hue. If something needs to stand out, make it heavier or
  fill it.
- **Don't** use red for anything but failure and destruction.
- **Don't** add a shadow to static content.
- **Don't** ship framework default styling, an icon set other than Hugeicons, or a
  dark-pattern flow.

## Motion

There is almost none. The landing page opens on one screen and the rest scrolls
normally beneath it — no paging, no pinned stack, no veil. Nine full-height sections
were nine ways of saying the same thing; four short ones under the fold say it once,
and a buyer can still read what it does, how it is integrated, what it costs and
what happens when they leave.

What is left is the hero card cycling through upload, encode, link, which is a
diagram rather than decoration, and it stops under reduced motion.

## Navigation

There is no header. The landing page carries its own mark, centred, at the top of
the single screen; the dashboard has its sidebar; everything else is in the footer.
Signing in lives there too — it is for people who already have an account, and it
does not need a permanent slot next to the thing that gets new ones.

## Where it is used

The dashboard overview is the reference layout: the node list on the left, and a
rail on the right holding the ring and the metrics — everything that is a figure
rather than an action.

## Checks

`npm run check:contrast` samples the real pixel under every text node on the public
pages, in both themes, at every viewport step, and fails under 4.5:1.
