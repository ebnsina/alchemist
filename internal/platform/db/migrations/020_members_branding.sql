-- Teammates and tenant branding.
--
-- Until now a tenant had exactly one user, created at signup, and nothing to change
-- about how it presents itself. Both gaps were only reachable through the database.
--
-- An invite is a token, not an account: the row carries only a hash, so a leaked
-- table hands over nothing that can be redeemed. Accepting one happens before any
-- tenant is in scope — the same bootstrapping problem signup has — so the lookup is
-- a definer function, exactly as auth_find_user is.

create table if not exists invites (
  id           uuid primary key default gen_random_uuid(),
  tenant_id    uuid not null references tenants(id) on delete cascade,
  email        citext not null,
  role         text not null default 'member',
  token_hash   bytea not null unique,
  invited_by   uuid references users(id) on delete set null,
  created_at   timestamptz not null default now(),
  expires_at   timestamptz not null,
  accepted_at  timestamptz
);

-- One live invite per address per tenant. A resend replaces rather than accumulates.
create unique index if not exists invites_pending on invites (tenant_id, email)
  where accepted_at is null;

alter table invites enable row level security;
alter table invites force row level security;
drop policy if exists invites_tenant on invites;
create policy invites_tenant on invites using (tenant_id = current_tenant());

create table if not exists tenant_branding (
  tenant_id     uuid primary key references tenants(id) on delete cascade,
  logo_key      text,
  logo_type     text,
  updated_at    timestamptz not null default now()
);

alter table tenant_branding enable row level security;
alter table tenant_branding force row level security;
drop policy if exists branding_tenant on tenant_branding;
create policy branding_tenant on tenant_branding using (tenant_id = current_tenant());

grant select, insert, update, delete on invites, tenant_branding to alchemist_app;

-- Roles are a closed set. A typo that writes 'admn' would silently grant nothing and
-- read as a permissions bug for whoever is holding it.
alter table users drop constraint if exists users_role_known;
alter table users add constraint users_role_known
  check (role in ('owner', 'admin', 'member'));

-- Redeeming an invite happens with no tenant in scope, so RLS would read zero rows
-- and report the token as unknown rather than erroring. Definer, like auth_signup.
create or replace function invite_redeem(hash bytea, user_hash text)
returns table (user_id uuid, tenant_id uuid, email citext)
language plpgsql security definer
set search_path = public as $$
declare
  inv invites%rowtype;
  u   uuid;
begin
  select * into inv from invites
   where token_hash = hash and accepted_at is null and expires_at > now()
   for update;
  if not found then
    return;
  end if;

  insert into users (tenant_id, email, password_hash, role)
  values (inv.tenant_id, inv.email, user_hash, inv.role)
  returning id into u;

  update invites set accepted_at = now() where id = inv.id;
  return query select u, inv.tenant_id, inv.email;
end;
$$;
revoke all on function invite_redeem(bytea, text) from public;
grant execute on function invite_redeem(bytea, text) to alchemist_app;

-- The address on an invite is shown to whoever opens the link, before they have an
-- account, so it cannot come from a tenant-scoped read.
create or replace function invite_preview(hash bytea)
returns table (email citext, role text, org_name text)
language sql security definer stable
set search_path = public as $$
  select i.email, i.role, t.name
    from invites i
    join tenants t on t.id = i.tenant_id
   where i.token_hash = hash
     and i.accepted_at is null
     and i.expires_at > now()
$$;
revoke all on function invite_preview(bytea) from public;
grant execute on function invite_preview(bytea) to alchemist_app;

-- 017 wrote its policies against `alchemist.tenant_id`, a setting nothing has ever
-- set: AsTenant sets `app.tenant_id`, which is what current_tenant() reads. The
-- policies therefore matched nothing and every tenant-scoped read of users or
-- sessions returned zero rows. It stayed invisible because every account path so far
-- goes through a SECURITY DEFINER function, which bypasses RLS. Team management is
-- the first code to read these tables as a tenant, and it read nothing.
drop policy if exists users_tenant on users;
drop policy if exists users_tenant on users;
create policy users_tenant on users using (tenant_id = current_tenant());

drop policy if exists sessions_tenant on sessions;
drop policy if exists sessions_tenant on sessions;
create policy sessions_tenant on sessions
  using (exists (
    select 1 from users u where u.id = sessions.user_id and u.tenant_id = current_tenant()
  ));
