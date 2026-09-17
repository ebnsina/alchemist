-- Where a customer's videos are allowed to play.
--
-- Playback answered Access-Control-Allow-Origin: * and checked no Referer, so a signed
-- link lifted out of a page played inside any page on any site for the life of the
-- token. Against paid content that is long enough to build a competing site around
-- somebody else's course, and it is the one control every commercial platform ships
-- that this did not.
--
-- Empty means no restriction, which is what every existing tenant has and what a
-- customer with a native app must keep: the lock is enforced from the browser's Origin
-- or Referer, and an app sends neither.
--
-- The list is not read at playback time. The origin a link is locked to is stamped into
-- the token and covered by its signature, because the edge decides this: njs verifies
-- in the worker with no database, and a warm cache slice never reaches the origin at
-- all. A rule the edge cannot read is a rule that stops applying once the cache fills.
alter table tenants add column if not exists playback_origins text[] not null default '{}';

comment on column tenants.playback_origins is
  'Web origins (scheme://host) a playback link may be locked to. Empty means '
  'unrestricted. Enforced from the token, not from this column.';
