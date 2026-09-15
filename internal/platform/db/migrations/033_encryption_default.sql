-- Playback encryption becomes the default, now that encrypted media plays.
--
-- 025 turned it off because cbcs needs EME and a licence server in Chrome and
-- Firefox, which is nearly the whole BD audience. Packaging is cenc now and the key
-- endpoint answers an EME Clear Key licence, which Chrome, Firefox and Edge all
-- support with no vendor, so the reason for off has gone.
--
-- What this is: encryption at rest. A lifted bucket or a stolen backup decodes to
-- nothing. It is NOT DRM -- the key is handed to the browser in the clear behind the
-- signed URL -- and it does not stop a viewer who is allowed to watch from keeping a
-- copy. Viewer-bound tokens and the device cap are what address sharing.
--
-- Existing tenants keep the value they have: this changes the default for new
-- accounts only. Flip an existing one deliberately, per deploy/README.md, because
-- Safari and iOS cannot play Clear Key and a whole tenant's Apple viewers are not
-- something to switch off by migration. Existing assets are never re-packaged either
-- -- they keep whatever they were encoded with, and re-encoding a library costs real
-- money for no protection anyone asked for.
alter table tenants alter column encrypt_playback set default true;

comment on column tenants.encrypt_playback is
  'cenc playback encryption with an EME Clear Key licence. On by default for new '
  'tenants. Encryption at rest, not DRM: the key is served to the browser behind the '
  'signed URL. Safari and iOS cannot play it.';
