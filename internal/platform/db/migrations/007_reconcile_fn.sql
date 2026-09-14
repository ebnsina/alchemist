-- The reconcile sweep runs across all tenants, before any one tenant is in scope, so
-- it cannot be tenant-scoped. Without this it silently sees zero rows -- RLS is
-- forced, current_tenant() is NULL, and the policy matches nothing.
--
-- SECURITY DEFINER runs as the owner and returns only the ids needed to queue work.
-- No credentials, no cursors, no customer data.
create or replace function active_bucket_sources()
returns table (source_id uuid, tenant_id uuid)
language sql security definer stable
set search_path = public as $$
  select id, tenant_id from bucket_sources where active
$$;
revoke all on function active_bucket_sources() from public;
grant execute on function active_bucket_sources() to alchemist_app;
