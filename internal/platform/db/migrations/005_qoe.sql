-- Playback quality telemetry. Written by viewers' players, so every field is
-- untrusted input and is bounded on write.
create table qoe_events (
  id            bigserial primary key,
  tenant_id     uuid not null references tenants(id) on delete cascade,
  asset_id      uuid,
  session_id    text not null,
  startup_ms    int,
  rebuffer_count int,
  rebuffer_ms   int,
  avg_bitrate_bps int,
  error_code    text,
  -- Split BDIX from international traffic: this is the number that shows whether
  -- local delivery is actually doing what it is supposed to.
  network       text,
  country       text,
  occurred_at   timestamptz not null default now()
);
create index on qoe_events (tenant_id, occurred_at desc);
create index on qoe_events (asset_id, occurred_at desc);

alter table qoe_events enable row level security;
alter table qoe_events force row level security;
create policy tenant_isolation on qoe_events using (tenant_id = current_tenant());
grant select, insert on qoe_events to alchemist_app;
grant usage, select on sequence qoe_events_id_seq to alchemist_app;
