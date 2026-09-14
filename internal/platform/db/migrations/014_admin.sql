-- Creating tenants and issuing keys happens before any tenant is in scope, so it
-- cannot go through the tenant-scoped path. These run as the owner and are reachable
-- only from the admin surface.
create or replace function admin_create_tenant(tenant_name text, profile text)
returns uuid
language sql security definer
set search_path = public as $$
  insert into tenants (name, ladder_profile) values (tenant_name, profile) returning id
$$;
revoke all on function admin_create_tenant(text, text) from public;
grant execute on function admin_create_tenant(text, text) to alchemist_app;

create or replace function admin_issue_key(tenant uuid, key_name text, hash bytea)
returns uuid
language sql security definer
set search_path = public as $$
  insert into api_keys (tenant_id, name, key_hash) values (tenant, key_name, hash)
  returning id
$$;
revoke all on function admin_issue_key(uuid, text, bytea) from public;
grant execute on function admin_issue_key(uuid, text, bytea) to alchemist_app;

create or replace function admin_list_tenants()
returns table (id uuid, name text, ladder_profile text, created_at timestamptz)
language sql security definer stable
set search_path = public as $$
  select id, name, ladder_profile, created_at from tenants order by created_at desc
$$;
revoke all on function admin_list_tenants() from public;
grant execute on function admin_list_tenants() to alchemist_app;

create or replace function admin_revoke_key(key uuid) returns boolean
language sql security definer
set search_path = public as $$
  update api_keys set revoked_at = now() where id = key and revoked_at is null
  returning true
$$;
revoke all on function admin_revoke_key(uuid) from public;
grant execute on function admin_revoke_key(uuid) to alchemist_app;
