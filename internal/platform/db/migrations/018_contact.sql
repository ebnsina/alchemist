-- Contact requests from the public site. Not tenant data: nobody is signed in when
-- one arrives, and no customer should ever be able to read them, so these rows stay
-- outside the tenant-scoped path entirely and are reachable only through the
-- definer function below and the operator surface.

create table contact_requests (
  id         uuid primary key default gen_random_uuid(),
  name       text not null,
  email      citext not null,
  org        text,
  message    text not null,
  source     text,
  created_at timestamptz not null default now(),
  handled_at timestamptz
);

create index on contact_requests (created_at desc);

alter table contact_requests enable row level security;
alter table contact_requests force row level security;
-- No policy: with RLS forced and nothing granted, a tenant-scoped connection sees
-- nothing here even if a query reaches the table by mistake.

create or replace function contact_submit(
  p_name text, p_email citext, p_org text, p_message text, p_source text
)
returns uuid
language sql security definer
set search_path = public as $$
  insert into contact_requests (name, email, org, message, source)
  values (p_name, p_email, nullif(p_org, ''), p_message, p_source)
  returning id
$$;
revoke all on function contact_submit(text, citext, text, text, text) from public;
grant execute on function contact_submit(text, citext, text, text, text) to alchemist_app;

create or replace function admin_list_contact(p_limit int)
returns table (id uuid, name text, email citext, org text, message text,
               source text, created_at timestamptz, handled_at timestamptz)
language sql security definer stable
set search_path = public as $$
  select id, name, email, org, message, source, created_at, handled_at
    from contact_requests order by created_at desc limit p_limit
$$;
revoke all on function admin_list_contact(int) from public;
grant execute on function admin_list_contact(int) to alchemist_app;
