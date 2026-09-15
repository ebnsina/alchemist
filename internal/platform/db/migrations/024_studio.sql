-- Studio: an edit produces a NEW asset and never mutates the one it came from.
--
-- That is the whole design. A video someone has already published a link to must not
-- silently change under them, and an edit that turned out wrong has to be discardable
-- without having destroyed the original. The source asset is read-only throughout.
--
-- Edits render from the mezzanine, not the customer's upload: it is already a clean
-- constant-frame-rate master with keyframes on a fixed grid, so a trim lands where it
-- was asked to and the original is never touched again (it may not even still exist).

create table if not exists edits (
  id              uuid primary key default gen_random_uuid(),
  tenant_id       uuid not null references tenants(id) on delete cascade,
  source_asset_id uuid not null references assets(id) on delete cascade,
  output_asset_id uuid references assets(id) on delete set null,
  ops             jsonb not null,
  state           text not null default 'queued'
                  check (state in ('queued', 'rendering', 'done', 'failed')),
  error_code      text,
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now()
);

create index if not exists edits_tenant_id_idx on edits (tenant_id);
create index if not exists edits_source_asset_id_idx on edits (source_asset_id);

alter table edits enable row level security;
alter table edits force row level security;
drop policy if exists edits_tenant on edits;
create policy edits_tenant on edits using (tenant_id = current_tenant());

grant select, insert, update, delete on edits to alchemist_app;
