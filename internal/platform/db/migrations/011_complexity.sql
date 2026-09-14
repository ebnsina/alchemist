-- Per-title analysis runs once, at ingest, against the whole timeline. A rendition
-- generated later must reuse that result, or the same asset ends up with some rungs
-- tuned to its content and some at raw profile bitrates -- an inconsistent ladder
-- where a viewer stepping up a rung pays more than the content needs.
alter table assets add column if not exists complexity numeric(6,3);
