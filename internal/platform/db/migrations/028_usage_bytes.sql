-- Billing had only ingest seconds. For a video business egress is the bill, and
-- renditions.bytes and assets.source_bytes have been declared since 001 with nothing
-- ever writing them -- so /v1/usage returned a single line and the asset API reported
-- source_bytes: null forever.

-- What the mezzanine costs. Known when it is uploaded, unrecoverable afterwards
-- without a round trip per asset.
alter table assets add column if not exists mezzanine_bytes bigint;

-- Aggregate rows carry no asset id and there is exactly one per tenant per kind per
-- UTC day. A row per byte-range request would be millions a day and bill the same
-- number. The timezone is pinned because an index expression must be immutable.
create unique index if not exists usage_events_daily
  on usage_events (tenant_id, kind, ((occurred_at at time zone 'UTC')::date))
  where asset_id is null;

-- Stored bytes across every tenant, for the daily GB-hours job. RLS is forced, so a
-- background job with no tenant in scope reads zero rows and does not error -- the
-- same reason active_bucket_sources() exists.
--
-- Deduplicated assets are excluded rather than counted: they own no renditions and
-- their mezzanine_key points at the canonical asset's file, so counting them bills
-- one copy of the bytes several times.
create or replace function tenant_stored_bytes()
returns table (tenant_id uuid, bytes numeric)
language sql security definer stable
set search_path = public as $$
  select id, total from (
    select t.id,
           coalesce((select sum(r.bytes) from renditions r where r.tenant_id = t.id), 0)
         + coalesce((select sum(a.mezzanine_bytes) from assets a
                      where a.tenant_id = t.id and a.deduplicated_from is null), 0)
         + coalesce((select sum(a.source_bytes) from assets a
                      where a.tenant_id = t.id and a.source_key is not null), 0) as total
      from tenants t
  ) s where total > 0
$$;
revoke all on function tenant_stored_bytes() from public;
grant execute on function tenant_stored_bytes() to alchemist_app;
