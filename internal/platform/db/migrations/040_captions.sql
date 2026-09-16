-- Subtitles, uploaded rather than generated.
--
-- A sidecar WebVTT per language, referenced from the manifests: the media is never
-- touched, so adding a Bangla track to a published lecture rewrites two small
-- playlists and evicts nothing from any edge cache. Burning captions into the picture
-- would mean re-encoding the whole ladder per language and leaving the viewer no way
-- to turn them off.
--
-- One track per language per asset. Re-uploading the same language replaces it, which
-- is what correcting a typo in a transcript actually is.
create table if not exists captions (
  id         uuid primary key default gen_random_uuid(),
  asset_id   uuid not null references assets(id) on delete cascade,
  tenant_id  uuid not null references tenants(id) on delete cascade,
  -- BCP-47, as the manifests carry it: "bn", "en", "bn-BD".
  language   text not null,
  -- What a viewer picks from the menu. "বাংলা", not "bn".
  label      text not null,
  object_key text not null,
  bytes      bigint,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (asset_id, language)
);
create index if not exists captions_asset_id_idx on captions (asset_id);

alter table captions enable row level security;
alter table captions force row level security;
drop policy if exists tenant_isolation on captions;
create policy tenant_isolation on captions using (tenant_id = current_tenant());

grant select, insert, update, delete on captions to alchemist_app;
