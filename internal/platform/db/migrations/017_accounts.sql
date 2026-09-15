-- Self-service accounts. Until now a tenant existed only because an operator made
-- one, so there were no users and nothing to log in as.
--
-- Signup and login both happen before any tenant is in scope — the session is what
-- establishes the tenant — so every lookup here is a definer function, the same
-- shape the admin surface uses. The tables themselves still carry RLS, so anything
-- that reaches them through the tenant-scoped path stays isolated.

create extension if not exists citext;

create table if not exists users (
  id            uuid primary key default gen_random_uuid(),
  tenant_id     uuid not null references tenants(id) on delete cascade,
  email         citext not null unique,
  password_hash text not null,
  role          text not null default 'owner',
  created_at    timestamptz not null default now(),
  last_login_at timestamptz
);

-- Sessions store a hash, never the token, so a database leak does not hand over
-- live sessions.
create table if not exists sessions (
  id         uuid primary key default gen_random_uuid(),
  user_id    uuid not null references users(id) on delete cascade,
  token_hash bytea not null unique,
  created_at timestamptz not null default now(),
  expires_at timestamptz not null,
  revoked_at timestamptz
);

create index if not exists sessions_user_id_idx on sessions (user_id);
create index if not exists sessions_expires_at_idx on sessions (expires_at) where revoked_at is null;

alter table users enable row level security;
alter table users force row level security;
alter table sessions enable row level security;
alter table sessions force row level security;

drop policy if exists users_tenant on users;

create policy users_tenant on users
  using (tenant_id = current_setting('alchemist.tenant_id', true)::uuid);

drop policy if exists sessions_tenant on sessions;

create policy sessions_tenant on sessions
  using (exists (
    select 1 from users u
     where u.id = sessions.user_id
       and u.tenant_id = current_setting('alchemist.tenant_id', true)::uuid
  ));

grant select, insert, update, delete on users, sessions to alchemist_app;

-- One transaction: an org with no owner, or an owner with no org, is not a state
-- worth being able to reach.
create or replace function auth_signup(
  org_name text, user_email citext, hash text, profile text, key_name text, key_hash bytea
)
returns table (user_id uuid, tenant_id uuid, api_key_id uuid)
language plpgsql security definer
set search_path = public as $$
declare
  t uuid;
  u uuid;
  k uuid;
begin
  insert into tenants (name, ladder_profile) values (org_name, profile) returning id into t;
  insert into users (tenant_id, email, password_hash) values (t, user_email, hash) returning id into u;
  insert into api_keys (tenant_id, name, key_hash) values (t, key_name, key_hash) returning id into k;
  return query select u, t, k;
end;
$$;
revoke all on function auth_signup(text, citext, text, text, text, bytea) from public;
grant execute on function auth_signup(text, citext, text, text, text, bytea) to alchemist_app;

-- Returns a row for an unknown email too, with a null hash, so the caller can spend
-- the same time verifying either way and the response cannot be timed to tell
-- whether an address is registered.
create or replace function auth_find_user(user_email citext)
returns table (id uuid, tenant_id uuid, password_hash text)
language sql security definer stable
set search_path = public as $$
  select id, tenant_id, password_hash from users where email = user_email
$$;
revoke all on function auth_find_user(citext) from public;
grant execute on function auth_find_user(citext) to alchemist_app;

create or replace function auth_start_session(u uuid, hash bytea, expires timestamptz)
returns uuid
language sql security definer
set search_path = public as $$
  update users set last_login_at = now() where id = u;
  insert into sessions (user_id, token_hash, expires_at) values (u, hash, expires)
  returning id
$$;
revoke all on function auth_start_session(uuid, bytea, timestamptz) from public;
grant execute on function auth_start_session(uuid, bytea, timestamptz) to alchemist_app;

create or replace function auth_session(hash bytea)
returns table (user_id uuid, tenant_id uuid, email citext, org_name text)
language sql security definer stable
set search_path = public as $$
  select u.id, u.tenant_id, u.email, t.name
    from sessions s
    join users u on u.id = s.user_id
    join tenants t on t.id = u.tenant_id
   where s.token_hash = hash
     and s.revoked_at is null
     and s.expires_at > now()
$$;
revoke all on function auth_session(bytea) from public;
grant execute on function auth_session(bytea) to alchemist_app;

create or replace function auth_end_session(hash bytea) returns boolean
language sql security definer
set search_path = public as $$
  update sessions set revoked_at = now()
   where token_hash = hash and revoked_at is null
  returning true
$$;
revoke all on function auth_end_session(bytea) from public;
grant execute on function auth_end_session(bytea) to alchemist_app;
