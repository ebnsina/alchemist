# Alchemist design system

House style for Fajr Labs products. Ink on warm paper, one deep teal for action,
alta red only for loss. This file is the reference; `src/app.css` is the
implementation, and the two are meant to agree.

## Principles

It reads like a well-set printed page rather than a dashboard. The surfaces are
data-heavy — lists, usage tables, billing lines — so numbers are set in mono and
aligned, prose is set in Mona Sans and ragged-right. No gradients, no glass, no
shadow theatre. If an element is not carrying meaning it is removed rather than
styled down.

## Colours

| Token | Value | Use |
|---|---|---|
| `--color-primary` | `#15181B` | Near-black ink: headings, body, icons |
| `--color-on-primary` | `#FBF9F5` | Paper, for text on inked surfaces |
| `--color-secondary` | `#666E74` | Slate: captions, metadata, timestamps, placeholders. Never body copy |
| `--color-tertiary` | `#1F5F5B` | Neel Teal — the only interactive colour |
| `--color-tertiary-container` | `#174744` | Hover and pressed state of the same teal |
| `--color-neutral` | `#FBF9F5` | Warm paper: the page ground |
| `--color-surface` | `#FFFFFF` | Cards, inputs, table rows — lifted by colour, not shadow |
| `--color-outline` | `#E3DFD8` | Hairlines, dividers, borders. One weight, 1px |
| `--color-danger` | `#A8342A` | Alta red. Destructive actions and failure only |
| `--color-success` | `#2F6B3A` | Terminal healthy states: playable, paid, verified |

Measured, not eyeballed. On the paper ground `#FBF9F5`: ink 16.95:1, secondary
4.93:1, teal 7.02:1, danger 6.26:1, success 6.08:1. Paper on teal is 7.02:1, paper
on ink 16.95:1. Secondary is the tightest of them, which is why it is barred from
body copy — it passes AA for text but has no headroom to spare.

Dark mode would invert primary and neutral and lift the teal one step. It does not
introduce new hues, and it is not built yet.

## Typography

Two families. Mona Sans for everything human, Geist Mono for everything counted.

| Token | Size / weight | Use |
|---|---|---|
| `display` | 3.5rem / 600 / 1.05 / −0.03em | One per page, at most |
| `h1` | 2.25rem / 600 / 1.15 / −0.02em | Page title |
| `h2` | 1.5rem / 600 / 1.25 / −0.01em | Section |
| `body-md` | 1rem / 400 / 1.6 | Default prose, max measure 68ch |
| `body-sm` | 0.875rem / 400 / 1.55 | Secondary prose, help text, table cells |
| `label-caps` | 0.75rem / 600 / 0.08em, uppercase | Buttons, table headers, eyebrows. Never a sentence |
| `numeric` | Geist Mono 0.875rem / 500, `tnum` | Every byte count, duration, price, ID, timestamp |

Headings stop at three levels. A fourth means the page needs splitting.

Every number, date, duration and currency goes through `Intl`. No hand-rolled
formatters, and no localisation logic inside components.

## Layout

An 8px scale, used strictly: `xs` 4, `sm` 8, `md` 16, `lg` 24, `xl` 40, `xxl` 64.

- `xs`/`sm` inside a control — icon to label, badge padding
- `md` between related elements in a group
- `lg` between groups inside a card
- `xl`/`xxl` between page sections

Content column caps at 1200px; reading prose caps at 68ch at any viewport. Tables
run 40px rows with 12px cell padding. Whitespace separates sections, not rows.

## Elevation

There is no elevation scale. Depth is surface colour against warm paper plus a 1px
outline. Exactly one shadow exists, for things that float above the page — menus,
dialogs, toasts:

```
0 8px 24px rgba(21, 24, 27, 0.10)
```

Anything else reaching for a shadow is probably a card, and should use the outline.

## Shapes

Four radii: `sm` 4px for inline chips and code, `md` 8px for buttons, inputs and
menus, `lg` 14px for cards and dialogs, `full` 999px for avatars, pills and status
dots. Radii do not scale with size — a large card and a small card share `lg`.

## Components

- **button-primary** — teal, paper text, label-caps, 40px tall, `md` radius. One per
  view. Hover darkens to tertiary-container; focus is a 2px teal ring offset 2px.
- **button-secondary** — white on paper, 1px outline, ink text. The default for
  anything that is not the single primary action.
- **button-danger** — alta red, and only behind a confirmation for anything
  irreversible.
- **card** — white, `lg` radius, 24px padding, 1px outline, no shadow.
- **input** — white, `md` radius, 40px tall, 1px outline turning teal on focus.
  Errors add red helper text below, never a red placeholder.
- **badge-inverse** — ink pill, paper text, for counts and primary status.
- **status** — a `full` dot plus a label-caps word. Never a coloured background block.
- **stat-value** — mono, tabular, right-aligned in tables.
- **divider** — 1px outline, full bleed inside cards.

Every surface that can load, be empty, or fail ships all three: a skeleton matching
the final layout's dimensions, an empty state with one sentence and one action, and
an error state in plain language with a retry.

## Do and don't

- **Do** use teal for exactly one thing: something the user can act on.
- **Do** set every number, ID and timestamp in Geist Mono with tabular figures.
- **Do** write error copy as a plain sentence — what happened, what to do next, with
  stable API error codes mapped to that copy on the client.
- **Do** keep one primary action per view. Everything else is secondary or a link.
- **Don't** introduce a new colour, radius or shadow. If the system cannot express
  it, the layout is wrong.
- **Don't** use red for emphasis, warnings or counts. Red means loss.
- **Don't** use shadows to separate static content. Use the outline.
- **Don't** ship framework default styling, an icon set other than Hugeicons, or any
  dark-pattern flow — hidden pricing, pre-checked upsells, a disguised cancel.

## Checks

`npm run check:contrast` samples the real pixel under every text node on the public
pages, at every viewport step, and fails under 4.5:1. It exists because the usual
walk-up-the-DOM check reads an ancestor's background rather than what is painted.
