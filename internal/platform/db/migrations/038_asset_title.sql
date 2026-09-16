-- Videos had no name. Every list and every table showed `001b22f2`, so a customer
-- with two lectures recorded on the same day could not tell which was which.
--
-- Nullable on purpose and never backfilled with a placeholder: null means "nobody
-- has named this", and the dashboard falls back to the short id. Writing "Untitled"
-- into the column would make an unnamed video indistinguishable from one a customer
-- deliberately called that.
alter table assets add column if not exists title text;

comment on column assets.title is
  'What a person calls this video. Null means unnamed; show the short id instead.';
