-- Live recordings become ordinary VOD assets, and the live segments are reclaimed.
--
-- swept_at is the only state the handover needs. Without it a broadcast whose
-- segments have already been deleted is re-listed by every reaper pass forever, and
-- a finished conversion is indistinguishable from one that never ran.
alter table live_sessions add column if not exists swept_at timestamptz;

-- The reaper looks for sessions that are stale or unswept, per tenant and under RLS.
create index if not exists live_sessions_unswept
  on live_sessions (tenant_id) where swept_at is null;

-- The reaper runs across tenants with none in scope, and RLS on tenant_limits is
-- forced: a plain select there returns zero rows and no error, so every abandoned
-- broadcast is skipped silently. Same shape and same reason as resolve_api_key().
create or replace function live_tenants()
returns table (tenant_id uuid) language sql security definer stable
set search_path = public as $$
  select l.tenant_id from tenant_limits l where l.live_enabled
$$;

revoke all on function live_tenants() from public;
grant execute on function live_tenants() to alchemist_app;
