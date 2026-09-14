-- The original unique index on (tenant_id, source_sha256) was wrong.
--
-- It assumed one asset per distinct file, but a customer may legitimately upload the
-- same video twice and expect two assets with two ids and independent lifecycles.
-- Deduplication shares the encoding, not the identity -- so several assets can and
-- should reference the same content hash.
--
-- A plain index is what this was always for: finding an already-encoded match.
drop index if exists assets_tenant_id_source_sha256_idx;
create index if not exists assets_dedup_lookup
  on assets (tenant_id, source_sha256)
  where source_sha256 is not null;
