-- Migrating a library in from another video host.
--
-- The shape is deliberately the same as bucket_sources: a credential we wrap, a
-- cursor we page with, and a per-item table keyed on the remote id so a re-run never
-- imports the same video twice. What differs is only how a list of videos and a
-- download URL are obtained, which is what the provider adapters do.
--
-- Not every provider can be migrated from. DRM platforms hand out no source file
-- through their API by design, so `provider` is a closed set of the ones that can.

create table migration_sources (
  id             uuid primary key default gen_random_uuid(),
  tenant_id      uuid not null references tenants(id) on delete cascade,
  provider       text not null check (provider in ('vimeo', 'bunny')),
  -- Non-secret settings the adapter needs: a Bunny library id, for instance.
  config         jsonb not null default '{}'::jsonb,
  wrapped_secret bytea not null,
  secret_nonce   bytea not null,
  state          text not null default 'scanning'
                 check (state in ('scanning', 'done', 'failed', 'paused')),
  cursor         text not null default '',
  discovered     int not null default 0,
  imported       int not null default 0,
  last_error     text,
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now()
);

create index on migration_sources (tenant_id);

-- One row per video seen on the far side, whether or not it imported. A failure is
-- recorded rather than retried forever: a video the provider will not hand over is a
-- fact about that video, and hiding it makes the count lie.
create table migration_items (
  id         uuid primary key default gen_random_uuid(),
  source_id  uuid not null references migration_sources(id) on delete cascade,
  remote_id  text not null,
  title      text,
  asset_id   uuid references assets(id) on delete set null,
  state      text not null default 'pending'
             check (state in ('pending', 'imported', 'skipped', 'failed')),
  reason     text,
  created_at timestamptz not null default now()
);

create unique index migration_items_remote on migration_items (source_id, remote_id);
create index on migration_items (source_id, state);

alter table migration_sources enable row level security;
alter table migration_sources force row level security;
create policy migration_sources_tenant on migration_sources
  using (tenant_id = current_tenant());

alter table migration_items enable row level security;
alter table migration_items force row level security;
create policy migration_items_tenant on migration_items
  using (exists (
    select 1 from migration_sources s
     where s.id = migration_items.source_id
       and s.tenant_id = current_tenant()
  ));

grant select, insert, update, delete on migration_sources, migration_items to alchemist_app;

-- The worker runs with no tenant in scope, exactly like active_bucket_sources(), so
-- it needs a definer function or RLS silently returns nothing.
create or replace function active_migration_sources()
returns table (id uuid, tenant_id uuid)
language sql security definer stable
set search_path = public as $$
  select id, tenant_id from migration_sources where state = 'scanning'
$$;
revoke all on function active_migration_sources() from public;
grant execute on function active_migration_sources() to alchemist_app;
