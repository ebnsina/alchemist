-- Peak concurrent viewers, per asset per day.
--
-- Both live SKUs are priced on peak, and peak is the one billable number that cannot
-- be reconstructed afterwards: by the time anyone asks, the requests that would have
-- proved it are long gone. So it is recorded with the work, like ingest seconds.
--
-- Per asset rather than per tenant because a Channel is billed on its own peak. A
-- tenant-level number rolls up from this; the reverse does not.
--
-- 028 made the daily aggregate unique only where asset_id is null. This is the same
-- idea for rows that carry one.
create unique index if not exists usage_events_daily_asset
  on usage_events (tenant_id, asset_id, kind, ((occurred_at at time zone 'UTC')::date))
  where asset_id is not null;
