-- Playback encryption becomes a choice, and the default is off.
--
-- Superseded by 033: this was a measure for as long as the media was packaged cbcs.
-- With cenc and an EME Clear Key licence it plays in Chrome and Firefox, and the
-- default is on.
--
-- As configured it was never DRM. The manifest declares
-- `METHOD=SAMPLE-AES ... KEYFORMAT="identity"`, and the key is served as raw bytes
-- from the same signed URL as the segments — so anyone able to fetch a segment can
-- fetch the key and decrypt it. It added no protection the expiring signed URL does
-- not already give, while making playback impossible in Chrome and Firefox, which
-- need EME and a licence server for cbcs. That is most of the audience, and for
-- Bangladesh on Android it is nearly all of it.
--
-- Access control is, and was, the signed expiring URL. Tenants that genuinely need
-- DRM can turn this back on and accept Safari-only until a licence server exists.
alter table tenants add column if not exists encrypt_playback boolean not null default false;

comment on column tenants.encrypt_playback is
  'cbcs playback encryption. Off by default: it is not DRM without a licence server, '
  'and it blocks Chrome and Firefox. Signed expiring URLs are the access control.';
