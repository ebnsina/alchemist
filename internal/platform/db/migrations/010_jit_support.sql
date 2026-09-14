-- A lazy rendition is generated from the mezzanine long after ingest, so the
-- mezzanine has to survive. It is regenerable in principle, but only from an original
-- we may have already dropped, and re-deriving it costs a full decode.
alter table assets add column if not exists mezzanine_key text;

-- Enough per-rendition detail to compose an HLS master playlist without re-running
-- the packager or touching the media. Adding a rendition then rewrites one small
-- text file, leaving every existing object -- and every edge cache entry -- untouched.
alter table renditions add column if not exists codec_string text;
alter table renditions add column if not exists width int;
alter table renditions add column if not exists avg_bandwidth_bps int;
