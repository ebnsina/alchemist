# Changelog

What changed, in the terms a customer would notice. The workspaces keep their own,
closer to the code: [`web/CHANGELOG.md`](web/CHANGELOG.md) for the dashboard and
[`player/CHANGELOG.md`](player/CHANGELOG.md) for the embed and SDK.

Nothing has been released yet, so everything so far sits under Unreleased.

## Unreleased

### Added

**Subtitles.** Upload a WebVTT file per language and viewers pick one from the player's
menu. None is switched on for them. Adding a language to a video that is already
published costs nothing and changes nothing for anyone watching it.

**Draft subtitles from the audio.** Have a transcript generated and read it before you
publish it — it is offered as a draft and never goes live on its own. Runs against
OpenAI, or against a Whisper server of your own so the audio never leaves your machine.

**An audio-only option** on every video that has sound. About 40 MB an hour against 300
at the smallest picture — enough to finish a lecture on what is left of a data pack.

**Allowed sites.** Name the websites your videos are embedded on and a link copied out
of your page stops working anywhere else.

**Link lifetime is yours to set.** Anything from a minute to a day. Set it from your
longest video, not your shortest: one link carries the whole thing, so if it runs out
while somebody is watching, the video stops there.

**Try a failed video again.** A video that did not process used to be a dead end. The
file you sent is still here, so it can be retried without uploading it again — from the
video's page or from the row menu in your library.

**Billing.** One invoice a month for what you actually used, with every line showing
the quantity, the rate and what it came to. BDT is paid through SSLCommerz, other
currencies through Stripe.

**Live streams survive a dropped signal.** Mobile uplinks drop. If yours does, the
broadcast waits for you instead of ending on everybody watching, and picks up where it
stopped when you are back.

### Changed

**Videos keep their frame rate.** 24 and 25 frames a second used to be converted to 30,
which put a judder on any pan. Film and PAL footage are now left as they are, and high
frame rate footage is halved to something a phone can play.

**Live broadcasts go out at a watchable size.** They were being sent at the smallest
size in your ladder, which is unreadable for slides and whiteboards. They use a size
that still fits one machine's worth of realtime encoding.

**Scrambling stored video is off for new accounts.** It is on for anyone who already
had it. It refuses every iPhone, iPad and Safari viewer, for a key the browser is
handed in the clear anyway — so it never stopped anyone who was allowed to watch from
keeping a copy. Everything we hold is written to disk encrypted regardless.

**Recordings of broadcasts are kept as they were sent**, rather than re-encoded into a
larger, slightly worse copy.

**A stream that is on air cannot be deleted.** Stop it first. Deleting it mid-broadcast
used to take the broadcast down and leave its recording stranded.

**Two things that delete now ask first**: removing your logo, and the bin icon beside
an edit in Studio.

### Fixed

**Videos shot on a phone were the wrong way round.** Anything recorded in portrait was
described with its width and height swapped, so players were told the wrong shape.

**Scrubbing previews were squashed** on anything that was not widescreen.

**Videos smaller than 720p got stuck part-way.** A 360p upload would sit at "partially
ready" forever, waiting for sizes larger than the file it came from — and then spend
your quota making them by stretching the picture.

**A size requested by a viewer took far longer than it should have.** The work was
being done one piece at a time instead of in parallel, on the one path where somebody
is actually waiting.

**Subtitle text could appear on screen with a web address stuck to the end of every
line.** Only the scrubbing index needed that, and it was being done to subtitles too.

**During a broadcast, the player was told about three files that do not exist yet.**
Live now offers only what is actually there, and the rest appear when the recording has
converted.

**A retry starts where it stopped.** It used to re-download the file and re-encode
everything from the beginning — on a large upload, most of the cost of a retry.
