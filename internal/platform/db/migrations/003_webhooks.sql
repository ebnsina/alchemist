-- Webhook endpoints and delivery log. "We never got the callback" must be an
-- answerable question, so every attempt is recorded.
create table webhook_endpoints (
  id         uuid primary key default gen_random_uuid(),
  tenant_id  uuid not null references tenants(id) on delete cascade,
  url        text not null,
  secret     bytea not null,
  events     text[] not null default array['asset.ready','asset.failed'],
  active     boolean not null default true,
  created_at timestamptz not null default now()
);
create index on webhook_endpoints (tenant_id) where active;

create table webhook_deliveries (
  id          bigserial primary key,
  endpoint_id uuid not null references webhook_endpoints(id) on delete cascade,
  tenant_id   uuid not null references tenants(id) on delete cascade,
  event       text not null,
  payload     jsonb not null,
  status_code int,
  error       text,
  attempt     int not null default 1,
  delivered_at timestamptz,
  created_at  timestamptz not null default now()
);
create index on webhook_deliveries (tenant_id, created_at desc);

alter table webhook_endpoints enable row level security;
alter table webhook_endpoints force row level security;
create policy tenant_isolation on webhook_endpoints using (tenant_id = current_tenant());

alter table webhook_deliveries enable row level security;
alter table webhook_deliveries force row level security;
create policy tenant_isolation on webhook_deliveries using (tenant_id = current_tenant());

grant select, insert, update, delete on webhook_endpoints, webhook_deliveries to alchemist_app;
grant usage, select on sequence webhook_deliveries_id_seq to alchemist_app;

-- Pull-from-URL ingest keeps the source URL for audit and re-fetch.
alter table assets add column source_url text;
