-- Regenerating the MPD by re-running the packager over published media does not
-- work: the files are encrypted and the packager cannot demux them without keys, and
-- decrypting to re-package would rewrite every segment and evict the asset from every
-- edge cache.
--
-- Instead each rendition's <Representation> element is captured when that rendition
-- is packaged, and the manifest is composed from the stored fragments. Nothing is
-- re-read and no media is touched.
alter table renditions add column if not exists dash_representation text;

-- The surrounding MPD, with a placeholder where video representations go.
alter table assets add column if not exists dash_skeleton text;
