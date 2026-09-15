-- The chunks table has existed since the first migration and nothing ever wrote to
-- it: chunking happened inside one job, in memory, so renditions.chunks_total and
-- chunks_done stayed at zero and no progress was observable from outside.
--
-- A chunk is only interesting once it exists, so completion is a timestamp rather
-- than a boolean: it says when, not just whether.
alter table chunks add column if not exists completed_at timestamptz;

create index if not exists chunks_rendition_id_idx on chunks (rendition_id) where completed_at is null;
