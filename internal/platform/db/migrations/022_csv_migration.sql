-- A CSV of links is the way across from every host that will not be automated: DRM
-- platforms, anything with no API, or a library already sitting on a file server.
-- It is the same machinery — dedupe, per-item reasons, progress — with the list
-- supplied by the customer instead of fetched from an API.
alter table migration_sources drop constraint migration_sources_provider_check;
alter table migration_sources add constraint migration_sources_provider_check
  check (provider in ('vimeo', 'bunny', 'csv'));
