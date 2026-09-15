-- Deleting an asset used to mean manual psql plus manual S3 surgery, which is not a
-- takedown process. The hard part is deduplication: several assets play one copy of
-- the media, stored under the id of whichever asset carried it first. Dropping that
-- row orphans the media and leaves every duplicate playing 404s.
--
-- So the prefix has to outlive the id it was named after. Null keeps the old
-- behaviour -- cmaf/{tenant}/{canonical} -- and is written out only when the asset
-- that named it is deleted while others still play its bytes.
alter table assets add column if not exists media_prefix text;

comment on column assets.media_prefix is
  'Where this asset''s media actually lives. Null means cmaf/{tenant}/{canonical}.';
