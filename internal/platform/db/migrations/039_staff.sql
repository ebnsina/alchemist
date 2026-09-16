-- A platform administrator: one Alchemist account that can support a customer by
-- assuming their scope, rather than a second app that reads every tenant at once.
--
-- Not a tenant role. owner/admin/member say what a person may do inside one tenant,
-- and this is orthogonal to all three, so it stays out of the users_role_known
-- constraint. That is also what keeps it unreachable from the customer API:
-- /v1/members writes `role` and nothing else, so a tenant owner cannot grant it.
--
-- Impersonation lives on the session, not in a header: the acting tenant is state
-- the server holds, so nothing a browser sends can widen what a support session
-- sees, and revoking the flag stops it on the very next request.

alter table users add column if not exists platform_admin boolean not null default false;

comment on column users.platform_admin is
  'Alchemist staff: may act as any tenant. Not a tenant role, and nothing in /v1 sets it.';

-- Null means the session acts as its own tenant, which is every session but one.
alter table sessions add column if not exists acting_tenant_id uuid
  references tenants(id) on delete set null;

comment on column sessions.acting_tenant_id is
  'Tenant this session is acting as. Honoured only while its user is a platform admin.';

-- auth_session is the single place a session becomes a tenant, so impersonation is
-- resolved here and every existing endpoint follows without a second code path.
-- The platform_admin check is inside the coalesce on purpose: a flag revoked
-- mid-session drops the acting tenant on the next request, with no session to expire
-- and nothing to remember to clean up.
drop function if exists auth_session(bytea);
create or replace function auth_session(hash bytea)
returns table (user_id uuid, tenant_id uuid, email citext, org_name text,
               platform_admin boolean, impersonating boolean)
language sql security definer stable
set search_path = public as $$
  select u.id,
         coalesce(case when u.platform_admin then s.acting_tenant_id end, u.tenant_id),
         u.email,
         coalesce(case when u.platform_admin then acting.name end, t.name),
         u.platform_admin,
         u.platform_admin and s.acting_tenant_id is not null
    from sessions s
    join users u on u.id = s.user_id
    join tenants t on t.id = u.tenant_id
    left join tenants acting on acting.id = s.acting_tenant_id
   where s.token_hash = hash
     and s.revoked_at is null
     and s.expires_at > now()
$$;
revoke all on function auth_session(bytea) from public;
grant execute on function auth_session(bytea) to alchemist_app;

create or replace function staff_is_admin(caller uuid) returns boolean
language sql security definer stable
set search_path = public as $$
  select coalesce((select platform_admin from users where id = caller), false)
$$;
revoke all on function staff_is_admin(uuid) from public;
grant execute on function staff_is_admin(uuid) to alchemist_app;

-- The tenant picker: the one read that crosses tenants, and it sees no customer
-- content -- names and counts only. RLS is forced and alchemist_app is nobypassrls,
-- so it has to be a definer function, and the function checks the flag itself rather
-- than trusting whoever called it: middleware is the part that gets refactored.
create or replace function staff_tenants(caller uuid, p_q text)
returns table (id uuid, name text, assets bigint, live_streams bigint,
               members bigint, created_at timestamptz)
language sql security definer stable
set search_path = public as $$
  select t.id, t.name,
         (select count(*) from assets a where a.tenant_id = t.id),
         (select count(*) from live_streams s where s.tenant_id = t.id),
         (select count(*) from users u where u.tenant_id = t.id),
         t.created_at
    from tenants t
   where staff_is_admin(caller)
     and (p_q = '' or t.name ilike '%' || p_q || '%' or t.id::text like lower(p_q) || '%')
$$;
revoke all on function staff_tenants(uuid, text) from public;
grant execute on function staff_tenants(uuid, text) to alchemist_app;

-- Starting and stopping impersonation writes the staff member's own session row,
-- which their acting tenant no longer matches -- so RLS would refuse the write that
-- ends impersonation and strand them. Definer, with the flag checked here too.
-- A null tenant clears it. False means no live session, no flag, or no such tenant.
create or replace function staff_impersonate(hash bytea, tenant uuid) returns boolean
language sql security definer
set search_path = public as $$
  update sessions s set acting_tenant_id = tenant
   where s.token_hash = hash
     and s.revoked_at is null
     and s.expires_at > now()
     and exists (select 1 from users u where u.id = s.user_id and u.platform_admin)
     and (tenant is null or exists (select 1 from tenants t where t.id = tenant))
  returning true
$$;
revoke all on function staff_impersonate(bytea, uuid) from public;
grant execute on function staff_impersonate(bytea, uuid) to alchemist_app;

-- Bootstrapping: the first platform administrator cannot be granted by one, and
-- users is force-RLS so there is no tenant in scope to insert under. Same shape and
-- same reason as auth_signup. Run twice it promotes the account that is already
-- there and leaves its password alone, rather than creating a second user.
create or replace function staff_grant(user_email citext, hash text, org text)
returns table (user_id uuid, tenant_id uuid, created boolean)
language plpgsql security definer
set search_path = public as $$
declare
  u uuid;
  t uuid;
begin
  select id, users.tenant_id into u, t from users where email = user_email;
  if found then
    update users set platform_admin = true where id = u;
    return query select u, t, false;
    return;
  end if;
  select id into t from tenants where name = org;
  if t is null then
    insert into tenants (name, ladder_profile) values (org, 'bd-mobile') returning id into t;
  end if;
  insert into users (tenant_id, email, password_hash, role, platform_admin)
  values (t, user_email, hash, 'owner', true) returning id into u;
  return query select u, t, true;
end;
$$;
revoke all on function staff_grant(citext, text, text) from public;
grant execute on function staff_grant(citext, text, text) to alchemist_app;
