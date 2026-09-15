-- Live streaming. A live stream IS an asset from the moment it is armed.
--
-- That is the whole trick, and it is what makes live a phase rather than a second
-- product: playback URLs, the signed prefix, the origin, the edge njs, the content
-- key endpoint and the player all key on (tenant, asset) and need no change. When
-- the broadcast ends and the recording is converted, the customer's link keeps
-- working because it was never a live-specific link. See docs/06-live.md.

-- live_ended is not cosmetic: it is the window where the segments still exist but
-- the VOD conversion has not finished, and playback has to keep working throughout.
alter type asset_state add value if not exists 'live';
alter type asset_state add value if not exists 'live_ended';

create table if not exists live_streams (
  id          uuid primary key default gen_random_uuid(),
  tenant_id   uuid not null references tenants(id) on delete cascade,
  name        text not null,
  -- The key itself is shown once and never stored, exactly like an API key.
  key_hash    bytea not null unique,
  protocol    text not null check (protocol in ('srt', 'rtmp')),
  state       text not null default 'idle'
              check (state in ('idle', 'armed', 'live', 'ended')),
  -- Held only while armed or live, so the range is not exhausted by streams nobody
  -- is using.
  ingest_port int unique,
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now()
);

create index if not exists live_streams_tenant_id_created_at_idx on live_streams (tenant_id, created_at desc);

create table if not exists live_sessions (
  id           uuid primary key default gen_random_uuid(),
  stream_id    uuid not null references live_streams(id) on delete cascade,
  tenant_id    uuid not null references tenants(id) on delete cascade,
  asset_id     uuid not null references assets(id) on delete cascade,
  state        text not null default 'waiting'
               check (state in ('waiting', 'live', 'ended', 'failed')),
  error_code   text,
  started_at   timestamptz,
  ended_at     timestamptz,
  -- Touched whenever a segment lands. A crashed encoder otherwise leaves an asset in
  -- 'live' forever, its segments never swept, with no error anywhere.
  last_seen_at timestamptz,
  created_at   timestamptz not null default now()
);

create index if not exists live_sessions_stream_id_created_at_idx on live_sessions (stream_id, created_at desc);

alter table live_streams enable row level security;
alter table live_streams force row level security;
drop policy if exists live_streams_tenant on live_streams;
create policy live_streams_tenant on live_streams using (tenant_id = current_tenant());

alter table live_sessions enable row level security;
alter table live_sessions force row level security;
drop policy if exists live_sessions_tenant on live_sessions;
create policy live_sessions_tenant on live_sessions using (tenant_id = current_tenant());

grant select, insert, update, delete on live_streams to alchemist_app;
grant select, insert, update, delete on live_sessions to alchemist_app;

-- An ingest connection arrives with no tenant in scope, and RLS is forced, so a
-- plain query here returns zero rows rather than erroring. Same shape and same
-- reason as resolve_api_key().
create or replace function resolve_stream_key(hash bytea)
returns uuid language sql security definer stable
set search_path = public as $$
  select id from live_streams where key_hash = hash and state in ('armed', 'live')
$$;

revoke all on function resolve_stream_key(bytea) from public;
grant execute on function resolve_stream_key(bytea) to alchemist_app;
