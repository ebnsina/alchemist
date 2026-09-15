-- Alchemist initial schema.
-- Tenant isolation is enforced by RLS, not by application WHERE clauses.

create extension if not exists pgcrypto;

-- Roles. The app role owns nothing and cannot bypass RLS.
do $$ begin
  if not exists (select from pg_roles where rolname = 'alchemist_app') then
    create role alchemist_app login password 'alchemist' nobypassrls;
  end if;
end $$;

-- Ladder profiles are data, not code: this is what makes a new market a row insert.
create table if not exists ladder_profiles (
  name        text primary key,
  description text not null,
  rungs       jsonb not null,
  created_at  timestamptz not null default now()
);

create table if not exists tenants (
  id             uuid primary key default gen_random_uuid(),
  name           text not null,
  ladder_profile text not null references ladder_profiles(name),
  created_at     timestamptz not null default now()
);

create table if not exists api_keys (
  id         uuid primary key default gen_random_uuid(),
  tenant_id  uuid not null references tenants(id) on delete cascade,
  name       text not null,
  key_hash   bytea not null unique,
  created_at timestamptz not null default now(),
  revoked_at timestamptz
);

do $$ begin
  if not exists (select 1 from pg_type where typname = 'asset_state') then
    create type asset_state as enum ('created', 'uploading', 'uploaded', 'probing', 'mezzanine', 'analyzing',
  'encoding', 'packaging', 'ready', 'partially_ready', 'failed');
  end if;
end $$;

create table if not exists assets (
  id             uuid primary key default gen_random_uuid(),
  tenant_id      uuid not null references tenants(id) on delete cascade,
  state          asset_state not null default 'created',
  error_code     text,
  ladder_profile text not null references ladder_profiles(name),
  source_key     text,
  source_bytes   bigint,
  source_sha256  bytea,
  duration_sec   numeric(12,3),
  width          int,
  height         int,
  frame_rate     numeric(8,4),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now()
);
create index if not exists assets_tenant_id_created_at_idx on assets (tenant_id, created_at desc);
-- Dedup is scoped within tenant; cross-tenant dedup would leak content existence.
-- Not unique: several assets may share content. Uploading the same file twice yields
-- two assets with independent lifecycles, sharing only the encoding work.
create index if not exists assets_dedup_lookup on assets (tenant_id, source_sha256)
  where source_sha256 is not null;

create table if not exists renditions (
  id              uuid primary key default gen_random_uuid(),
  asset_id        uuid not null references assets(id) on delete cascade,
  tenant_id       uuid not null references tenants(id) on delete cascade,
  height          int not null,
  codec           text not null,
  bitrate_bps     int not null,
  encoder_version text not null,
  params_hash     text not null,
  state           text not null default 'pending',
  chunks_total    int not null default 0,
  chunks_done     int not null default 0,
  object_key      text,
  bytes           bigint,
  created_at      timestamptz not null default now(),
  unique (asset_id, height, codec)
);

create table if not exists chunks (
  rendition_id uuid not null references renditions(id) on delete cascade,
  idx          int not null,
  tenant_id    uuid not null references tenants(id) on delete cascade,
  start_sec    numeric(12,3) not null,
  end_sec      numeric(12,3) not null,
  object_key   text,
  primary key (rendition_id, idx)
);

-- Billing data cannot be backfilled. Emit from the first request.
create table if not exists usage_events (
  id         bigserial primary key,
  tenant_id  uuid not null references tenants(id) on delete cascade,
  asset_id   uuid,
  kind       text not null,
  quantity   numeric(20,4) not null,
  unit       text not null,
  occurred_at timestamptz not null default now()
);
create index if not exists usage_events_tenant_id_occurred_at_idx on usage_events (tenant_id, occurred_at desc);

-- RLS on every tenant-scoped table.
create or replace function current_tenant() returns uuid
language sql stable as $$
  select nullif(current_setting('app.tenant_id', true), '')::uuid
$$;

do $$
declare t text;
begin
  foreach t in array array['tenants','api_keys','assets','renditions','chunks','usage_events']
  loop
    execute format('alter table %I enable row level security', t);
    execute format('alter table %I force row level security', t);
    -- Dropped first so the whole file stays safe to re-run; policies have no
    -- IF NOT EXISTS and the deploy applies every migration on every release.
    execute format('drop policy if exists tenant_isolation on %I', t);
    if t = 'tenants' then
      execute format(
        'create policy tenant_isolation on %I using (id = current_tenant())', t);
    else
      execute format(
        'create policy tenant_isolation on %I using (tenant_id = current_tenant())', t);
    end if;
  end loop;
end $$;

-- Resolving an API key must happen before a tenant is known, so it cannot be
-- tenant-scoped. SECURITY DEFINER runs as the owner (bypassing RLS) and returns
-- only the tenant id for an exact hash match, leaking nothing else.
create or replace function resolve_api_key(hash bytea) returns uuid
language sql security definer stable
set search_path = public as $$
  select tenant_id from api_keys where key_hash = hash and revoked_at is null
$$;
revoke all on function resolve_api_key(bytea) from public;
grant execute on function resolve_api_key(bytea) to alchemist_app;

grant select, insert, update, delete
  on tenants, api_keys, assets, renditions, chunks, usage_events to alchemist_app;
grant select on ladder_profiles to alchemist_app;
grant usage, select on all sequences in schema public to alchemist_app;

insert into ladder_profiles (name, description, rungs) values
('bd-mobile', 'Bangladesh default: mobile-first, 720p cap, H.264 Main, no AV1', '[
  {"height":144,"codec":"h264","profile":"baseline","preset":"medium","crf":30,"maxrate_bps":180000},
  {"height":240,"codec":"h264","profile":"main","preset":"medium","crf":28,"maxrate_bps":350000},
  {"height":360,"codec":"h264","profile":"main","preset":"medium","crf":26,"maxrate_bps":700000},
  {"height":480,"codec":"h264","profile":"main","preset":"medium","crf":25,"maxrate_bps":1200000},
  {"height":720,"codec":"h264","profile":"main","preset":"medium","crf":24,"maxrate_bps":2200000}
]'::jsonb),
('bd-ott', 'Bangladesh OTT: adds 1080p for BDIX broadband', '[
  {"height":240,"codec":"h264","profile":"main","preset":"medium","crf":28,"maxrate_bps":350000},
  {"height":360,"codec":"h264","profile":"main","preset":"medium","crf":26,"maxrate_bps":700000},
  {"height":480,"codec":"h264","profile":"main","preset":"medium","crf":25,"maxrate_bps":1200000},
  {"height":720,"codec":"h264","profile":"main","preset":"medium","crf":24,"maxrate_bps":2200000},
  {"height":1080,"codec":"h264","profile":"high","preset":"medium","crf":22,"maxrate_bps":4000000}
]'::jsonb),
('intl-default', 'International: 1080p default, higher bitrates', '[
  {"height":240,"codec":"h264","profile":"main","preset":"medium","crf":26,"maxrate_bps":400000},
  {"height":360,"codec":"h264","profile":"main","preset":"medium","crf":25,"maxrate_bps":800000},
  {"height":480,"codec":"h264","profile":"main","preset":"medium","crf":24,"maxrate_bps":1400000},
  {"height":720,"codec":"h264","profile":"high","preset":"medium","crf":23,"maxrate_bps":3000000},
  {"height":1080,"codec":"h264","profile":"high","preset":"medium","crf":22,"maxrate_bps":6000000}
]'::jsonb)
on conflict (name) do nothing;
