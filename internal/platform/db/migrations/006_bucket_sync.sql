-- Customer-owned buckets we ingest from.
--
-- Credentials are wrapped with the platform KEK exactly like content keys, so a
-- database dump does not hand over access to customers' storage.
create table if not exists bucket_sources (
  id             uuid primary key default gen_random_uuid(),
  tenant_id      uuid not null references tenants(id) on delete cascade,
  endpoint       text not null,
  region         text not null default 'us-east-1',
  bucket         text not null,
  prefix         text not null default '',
  access_key_id  text not null,
  wrapped_secret bytea not null,
  secret_nonce   bytea not null,
  -- Where the last reconcile stopped. Without it every pass re-lists the whole
  -- bucket, which is unaffordable once a customer has a real library.
  cursor         text,
  active         boolean not null default true,
  last_synced_at timestamptz,
  last_error     text,
  created_at     timestamptz not null default now()
);
create index if not exists bucket_sources_tenant_id_idx on bucket_sources (tenant_id) where active;

-- One row per object ever ingested, so a reconcile never re-imports what it already
-- took. Keyed by object key plus etag: a replaced object has a new etag and is
-- treated as new content.
create table if not exists bucket_objects (
  source_id   uuid not null references bucket_sources(id) on delete cascade,
  object_key  text not null,
  etag        text not null,
  asset_id    uuid references assets(id) on delete set null,
  ingested_at timestamptz not null default now(),
  primary key (source_id, object_key, etag)
);

alter table bucket_sources enable row level security;
alter table bucket_sources force row level security;
drop policy if exists tenant_isolation on bucket_sources;
create policy tenant_isolation on bucket_sources using (tenant_id = current_tenant());

-- bucket_objects has no tenant column; it is reachable only through its source,
-- which is already tenant-scoped.
alter table bucket_objects enable row level security;
alter table bucket_objects force row level security;
drop policy if exists tenant_isolation on bucket_objects;
create policy tenant_isolation on bucket_objects using (
  exists (select 1 from bucket_sources s
           where s.id = bucket_objects.source_id and s.tenant_id = current_tenant())
);

grant select, insert, update, delete on bucket_sources, bucket_objects to alchemist_app;
