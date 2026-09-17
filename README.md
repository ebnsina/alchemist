# Alchemist

Video infrastructure you run yourself. Your users upload video or go live; Alchemist
encodes it, stores it, and streams it to their viewers behind links that expire.

Built for Bangladesh first — where most viewers are on a phone, on mobile data, paying
by the gigabyte — which is the reason for most of the decisions in here.

```
┌──────────┐   upload / go live   ┌───────────┐   signed links   ┌────────┐
│ your app │ ───────────────────► │ Alchemist │ ───────────────► │ viewer │
└──────────┘      REST + webhooks └───────────┘   HLS · DASH     └────────┘
```

## What it does

**Video on demand.** Send a file, get back links that play on anything. Alchemist
probes it, works out a bitrate ladder suited to the content, encodes the cheap sizes
immediately and the expensive ones the first time somebody asks for them. A one-hour
lecture is watchable in about a minute.

**Live.** Publish from a browser camera, OBS, or a hardware encoder over SRT or RTMP.
Viewers watch through the same kind of link as any other video, and when the broadcast
ends it becomes an ordinary video at the same address — nothing you handed out breaks.
A dropped uplink does not end the class.

**Editing.** Trim, crop, reshape for phones, add a watermark, add subtitles. Every edit
produces a new video, so the original and every link to it keep working.

**Delivery you control.** Links expire on your clock, can be bound to one viewer, can
be limited to a number of devices, and can be locked to your own website so a copied
link does nothing anywhere else.

**Billing.** Usage is metered as the work happens and invoiced monthly, in BDT through
SSLCommerz or in other currencies through Stripe.

## Try it

You need Go, PostgreSQL, `ffmpeg`, `ffprobe`, `packager` ([shaka-packager]) and `weed`
([SeaweedFS]). The last two are GitHub release binaries, not in Homebrew.

```bash
cp .env.example .env     # every value is required; nothing has a silent default
make db-reset            # schema, job queue, grants
make storage             # object storage, in another terminal
make run                 # API and workers
make dev-account         # an account with the limits lifted
```

Then, as a customer would:

```bash
./test/journey.sh <api-key>          # upload, encode, play — public API only
make live-journey KEY=<api-key>      # a broadcast, including a mid-class encoder drop
```

`make test` runs the Go suite. It needs a database for the isolation tests; without
one they skip rather than fail, which is worth knowing before you trust a green run.

## Integrating

The API is the product. Everything a consumer needs is reachable with an API key and
nothing else — no shared database, no internal endpoints.

| | |
|---|---|
| [`api/openapi.yaml`](api/openapi.yaml) | Every route, with what it answers and why |
| [`llms.txt`](llms.txt) | Orientation, and the things that surprise an integrator |
| [`.claude/skills/alchemist-api/SKILL.md`](.claude/skills/alchemist-api/SKILL.md) | The procedure: real requests, webhook verification, error handling |

Three surprises worth knowing before you start:

- **`partially_ready` is playable.** The expensive sizes are made on first play, so a
  video nobody watches stays `partially_ready` forever, and that is correct.
- **Playback links expire.** Ask for the video again when a viewer presses play. Never
  build a link yourself, cache one, or put one in a database.
- **Uploads do not go through the API.** You get a presigned target and the browser
  puts the bytes straight into storage.

## Repository

One repository, three workspaces. `npm install` once at the root.

| Path | What |
|---|---|
| Go at the root | The engine: API, encoding pipeline, delivery origin, live ingest |
| [`player/`](player/README.md) | The embed and JS SDK, published as `@alchemist/player` |
| [`web/`](web/README.md) | Marketing site and customer dashboard (SvelteKit) |

Inside the engine:

```
internal/platform/   shared infrastructure — storage, media, signing, billing, payments
internal/modules/    bounded contexts — delivery (playback), live (broadcast)
internal/pipeline/   the background work: encode, package, publish, invoice
internal/api/        the control plane
```

[`internal/modules/README.md`](internal/modules/README.md) explains the module
boundaries and how one would be pulled out into its own service.

## Running it for real

[`deploy/README.md`](deploy/README.md) covers the whole deployment: control plane,
storage and its encryption, the BDIX edge cache, the live ingest server, payment
gateway callbacks, and the migrations.

Configuration is environment variables only, and a missing one is a boot failure rather
than a silent default. [`.env.example`](.env.example) lists every one with what it is
for.

## Contributing

[`CLAUDE.md`](CLAUDE.md) is the working agreement: the architecture, the constraints
that were established by measurement and are expensive to rediscover, and which
documents have to be updated alongside which kind of change. Read it before the code —
several things in here look wrong until you know what they are avoiding.

[shaka-packager]: https://github.com/shaka-project/shaka-packager
[SeaweedFS]: https://github.com/seaweedfs/seaweedfs
