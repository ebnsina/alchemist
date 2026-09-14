-- Per-tenant limits. Absent rows mean plan defaults apply, so existing tenants keep
-- working without a backfill.
create table tenant_limits (
  tenant_id            uuid primary key references tenants(id) on delete cascade,
  max_concurrent_jobs  int not null default 4,
  max_source_bytes     bigint not null default 34359738368,
  max_ingest_hours_mo  numeric(12,2),
  updated_at           timestamptz not null default now()
);

alter table tenant_limits enable row level security;
alter table tenant_limits force row level security;
create policy tenant_isolation on tenant_limits using (tenant_id = current_tenant());
grant select, insert, update, delete on tenant_limits to alchemist_app;

-- Concurrency is enforced by counting a tenant's in-flight assets, so this index is
-- on the hot path of every ingest request.
create index assets_tenant_inflight on assets (tenant_id)
  where state in ('uploaded','probing','mezzanine','analyzing','encoding','packaging');
