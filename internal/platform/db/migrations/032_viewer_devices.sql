-- A cap on how many devices one viewer may stream from at once.
--
-- One student's login shared with twenty friends is the leak edtech customers
-- actually have, and every one of those viewers is authenticated -- so encryption
-- and DRM do nothing about it and counting devices does. It applies only to links
-- minted with a viewer id (`GET /v1/assets/{id}?viewer=...`); an unbound link is
-- unchanged.
--
-- 0 means no cap, and an absent tenant_limits row means no cap, so this changes
-- nothing for a tenant who has not asked for it.
alter table tenant_limits add column if not exists max_viewer_devices int not null default 0;

comment on column tenant_limits.max_viewer_devices is
  'Concurrent devices allowed per bound viewer. 0 means no cap.';
