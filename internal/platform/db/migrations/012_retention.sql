-- Source retention. The original is ~40% of storage growth and is not needed once a
-- mezzanine exists: every rendition, including one generated on demand years later,
-- is built from the mezzanine. Keeping originals is therefore a paid choice, not a
-- default.
alter table tenant_limits add column if not exists retain_original boolean not null default false;

-- What happened to the original, so "where did my upload go" has an answer.
alter table assets add column if not exists source_deleted_at timestamptz;

-- A deduplicated asset keeps its own id and lifecycle; only the encoding is shared.
alter table assets add column if not exists deduplicated_from uuid references assets(id) on delete set null;

-- A failed cleanup is recorded rather than swallowed: a leaked original is precisely
-- the cost this feature exists to remove.
alter table assets add column if not exists last_retention_error text;
