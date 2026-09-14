-- Originals whose deletion failed. Nothing retried them, so a transient storage
-- error meant paying for that file forever -- the exact cost the feature removes.
create or replace function assets_pending_cleanup(limit_n int)
returns table (asset_id uuid, tenant_id uuid, source_key text)
language sql security definer stable
set search_path = public as $$
  select id, tenant_id, source_key
    from assets
   where source_key is not null
     and source_deleted_at is null
     and mezzanine_key is not null
     and last_retention_error is not null
   order by updated_at
   limit limit_n
$$;
revoke all on function assets_pending_cleanup(int) from public;
grant execute on function assets_pending_cleanup(int) to alchemist_app;
