-- An asset ingested from a customer bucket has to be read back with that bucket's
-- credentials, so it needs to remember where it came from.
alter table assets add column if not exists bucket_source_id uuid references bucket_sources(id) on delete set null;
alter table assets add column if not exists source_object_key text;

-- The transcode worker resolves credentials before any tenant is in scope on its
-- connection, so this is a definer function like the others. It returns only what is
-- needed to read one object.
create or replace function bucket_source_credentials(source uuid)
returns table (endpoint text, region text, bucket text,
               access_key_id text, wrapped_secret bytea, secret_nonce bytea)
language sql security definer stable
set search_path = public as $$
  select endpoint, region, bucket, access_key_id, wrapped_secret, secret_nonce
    from bucket_sources where id = source and active
$$;
revoke all on function bucket_source_credentials(uuid) from public;
grant execute on function bucket_source_credentials(uuid) to alchemist_app;
