-- Two changes to how a playback link is handed out.

-- 1. The link's lifetime becomes the customer's, not a constant.
--
-- Four hours was hardcoded, and it is a floor rather than a preference: every segment
-- request carries the same signature, so a link that expires mid-lecture stops playback
-- mid-lecture. The TTL has to outlast the longest watch session, which is why it could
-- not simply be made short -- and why a leaked link stayed useful for hours. Origin
-- locking (042) is what actually closed that; this lets an account that knows its own
-- content is short stop paying the four-hour window for it.
--
-- Bounded in the API: a minute is the floor because anything less cannot survive a
-- player's own retry, and a day is the ceiling because past that the expiry has stopped
-- being a control at all.
alter table tenants add column if not exists playback_ttl_seconds int not null default 14400;

comment on column tenants.playback_ttl_seconds is
  'How long a minted playback link lasts. Also the ceiling for a per-request ?ttl=. '
  'Must outlast the longest single viewing, because the signature covers every segment.';

-- 2. Playback encryption stops being the default.
--
-- 033 turned it on because cenc plus Clear Key plays in Chrome, Firefox and Edge with
-- no licence vendor. What went unsaid is what it costs: WebKit implements FairPlay and
-- nothing else, so every Apple viewer is refused, and the W3C EME specification
-- defines Clear Key as the baseline key system for interoperability testing rather
-- than as a content protection system.
--
-- What it buys is encryption at rest -- the key is handed to the browser in the clear
-- behind the signed URL, so it never stopped an entitled viewer from keeping a copy.
-- That threat is better answered underneath, by encrypting the volumes, which costs no
-- viewer anything. Real content protection is multi-DRM, needs an Apple certificate
-- and a licence vendor, and belongs to whoever asks and pays for it.
--
-- Existing tenants keep the value they have. This changes the default for new accounts
-- only, and published assets are never re-packaged: they keep what they were made with.
alter table tenants alter column encrypt_playback set default false;
