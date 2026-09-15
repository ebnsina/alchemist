-- Content keys for AES encryption. Stored under envelope encryption with a KEK
-- from the environment, so a database dump alone does not decrypt the library.
create table if not exists content_keys (
  asset_id   uuid primary key references assets(id) on delete cascade,
  tenant_id  uuid not null references tenants(id) on delete cascade,
  key_id     bytea not null,
  wrapped_key bytea not null,
  nonce      bytea not null,
  created_at timestamptz not null default now()
);

alter table content_keys enable row level security;
alter table content_keys force row level security;
drop policy if exists tenant_isolation on content_keys;
create policy tenant_isolation on content_keys using (tenant_id = current_tenant());

grant select, insert, update, delete on content_keys to alchemist_app;
