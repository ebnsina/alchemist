-- Live is a separate product from the VOD engine, and a customer may hold either or
-- both. Until now it was a deployment-wide flag: configure an ingest host and every
-- tenant on the box could broadcast, including the ones paying only for VOD.
--
-- Off by default, and an absent tenant_limits row means off, so enabling live is
-- always a deliberate act rather than a side effect of deploying with an ingest host.
alter table tenant_limits add column if not exists live_enabled boolean not null default false;

comment on column tenant_limits.live_enabled is
  'Whether this tenant has bought the Live product. No row means no.';
