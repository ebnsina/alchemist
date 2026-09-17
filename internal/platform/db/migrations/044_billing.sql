-- Billing. Usage has been metered since 028; nothing has ever charged for it.
--
-- Postpaid: usage accrues, an invoice is issued for the closed month, and it is
-- collected. How it is collected differs by gateway and cannot be made uniform --
-- Stripe charges a saved card off-session, while SSLCommerz's published API is hosted
-- checkout, so a BDT invoice is collected through a payment link the customer opens.
-- The invoice, the retries and the suspension are the same either way.

-- Rates live in a row, like ladder_profiles, so a price change or a negotiated rate is
-- not a deploy. Unit amounts are numeric rather than minor units because they are
-- fractions of a paisa per megabyte; only line and invoice totals round to money.
create table if not exists rate_cards (
  name        text primary key,
  currency    char(3) not null,
  description text not null,
  -- [{"kind":"ingest","unit":"seconds","per":60,"amount":2.00}, ...]
  -- `per` is how many units the amount buys: a minute of ingest, a GB of egress.
  rates       jsonb not null,
  created_at  timestamptz not null default now()
);

insert into rate_cards (name, currency, description, rates) values
('bd-standard', 'BDT', 'Bangladesh, published rates', '[
  {"kind":"ingest",  "unit":"seconds",  "per":60,         "amount":2.00},
  {"kind":"storage", "unit":"gb_hours", "per":720,        "amount":1.20},
  {"kind":"egress",  "unit":"bytes",    "per":1000000000, "amount":1.20}
]'::jsonb),
('intl-standard', 'USD', 'International, published rates', '[
  {"kind":"ingest",  "unit":"seconds",  "per":60,         "amount":0.016},
  {"kind":"storage", "unit":"gb_hours", "per":720,        "amount":0.010},
  {"kind":"egress",  "unit":"bytes",    "per":1000000000, "amount":0.010}
]'::jsonb)
on conflict (name) do nothing;

-- Which card a tenant is on, and therefore which currency they are billed in and which
-- gateway collects. One gateway per account: holding a balance in one currency while
-- charging in another is an accounting problem nobody asked for.
alter table tenants add column if not exists rate_card text references rate_cards(name);
alter table tenants add column if not exists billing_email text;
-- active: normal. past_due: an invoice is unpaid and being retried. suspended: ingest
-- is refused. Playback is never suspended -- taking a customer's viewers offline over
-- an unpaid invoice punishes people who are not party to it.
alter table tenants add column if not exists billing_status text not null default 'active'
  check (billing_status in ('active', 'past_due', 'suspended'));

create table if not exists invoices (
  id            uuid primary key default gen_random_uuid(),
  tenant_id     uuid not null references tenants(id) on delete cascade,
  period_start  date not null,
  period_end    date not null,
  currency      char(3) not null,
  -- Minor units, always integers. Money in a float is how a cent goes missing.
  total_minor   bigint not null,
  status        text not null default 'open'
                check (status in ('open', 'paid', 'failed', 'void')),
  attempts      int not null default 0,
  last_error    text,
  -- The gateway's own id for the attempt in flight, so a webhook finds its invoice.
  gateway       text,
  gateway_ref   text,
  issued_at     timestamptz not null default now(),
  paid_at       timestamptz,
  -- One invoice per tenant per period, so a retried issuing job cannot bill twice.
  unique (tenant_id, period_start)
);
create index if not exists invoices_tenant_id_issued_at_idx on invoices (tenant_id, issued_at desc);
create index if not exists invoices_open_idx on invoices (status) where status in ('open', 'failed');

create table if not exists invoice_lines (
  id            uuid primary key default gen_random_uuid(),
  invoice_id    uuid not null references invoices(id) on delete cascade,
  tenant_id     uuid not null references tenants(id) on delete cascade,
  kind          text not null,
  unit          text not null,
  quantity      numeric(20,4) not null,
  unit_amount   numeric(14,6) not null,
  per           numeric(20,4) not null,
  amount_minor  bigint not null
);
create index if not exists invoice_lines_invoice_id_idx on invoice_lines (invoice_id);

-- What the gateway gave us back for a card. Never a card number: Stripe hands a
-- payment method id, and the brand and last four are for showing the customer which
-- card it is, nothing more.
create table if not exists payment_methods (
  id          uuid primary key default gen_random_uuid(),
  tenant_id   uuid not null references tenants(id) on delete cascade,
  gateway     text not null check (gateway in ('stripe', 'sslcommerz')),
  external_id text not null,
  customer_id text,
  brand       text,
  last4       text,
  exp_month   int,
  exp_year    int,
  created_at  timestamptz not null default now(),
  unique (tenant_id, gateway, external_id)
);

-- Every attempt, successful or not. A payment that is not written down is one that
-- gets made twice.
create table if not exists payments (
  id           uuid primary key default gen_random_uuid(),
  invoice_id   uuid references invoices(id) on delete set null,
  tenant_id    uuid not null references tenants(id) on delete cascade,
  gateway      text not null,
  external_id  text not null,
  amount_minor bigint not null,
  currency     char(3) not null,
  status       text not null check (status in ('pending', 'succeeded', 'failed')),
  error        text,
  created_at   timestamptz not null default now(),
  -- The gateway's id is the idempotency key: a webhook delivered twice, or an IPN and
  -- a redirect describing the same payment, must credit an invoice once.
  unique (gateway, external_id)
);
create index if not exists payments_invoice_id_idx on payments (invoice_id);

do $$
declare t text;
begin
  foreach t in array array['invoices','invoice_lines','payment_methods','payments']
  loop
    execute format('alter table %I enable row level security', t);
    execute format('alter table %I force row level security', t);
    execute format('drop policy if exists tenant_isolation on %I', t);
    execute format(
      'create policy tenant_isolation on %I using (tenant_id = current_tenant())', t);
  end loop;
end $$;

grant select on rate_cards to alchemist_app;
grant select, insert, update, delete
  on invoices, invoice_lines, payment_methods, payments to alchemist_app;

-- Invoicing runs across tenants before any is in scope, so it needs a definer reader
-- like every other cross-tenant sweep. Returns only what the job needs to bill.
create or replace function billable_tenants(period_start date)
returns table (tenant_id uuid, rate_card text, currency char(3))
language sql security definer stable
set search_path = public as $$
  select t.id, coalesce(t.rate_card, 'bd-standard'), c.currency
    from tenants t
    join rate_cards c on c.name = coalesce(t.rate_card, 'bd-standard')
   where not exists (
     select 1 from invoices i
      where i.tenant_id = t.id and i.period_start = billable_tenants.period_start)
$$;
revoke all on function billable_tenants(date) from public;
grant execute on function billable_tenants(date) to alchemist_app;

-- Collecting and retrying also runs with no tenant in scope.
create or replace function collectable_invoices(max_attempts int)
returns table (invoice_id uuid, tenant_id uuid, currency char(3), total_minor bigint,
               attempts int)
language sql security definer stable
set search_path = public as $$
  select i.id, i.tenant_id, i.currency, i.total_minor, i.attempts
    from invoices i
   where i.status in ('open', 'failed')
     and i.total_minor > 0
     and i.attempts < max_attempts
   order by i.issued_at
   limit 200
$$;
revoke all on function collectable_invoices(int) from public;
grant execute on function collectable_invoices(int) to alchemist_app;

-- A gateway holds no API key, so a webhook arrives with no tenant in scope and RLS
-- would return nothing. This resolves the invoice a verified callback is about --
-- either by our own id in tran_id/metadata, or by the reference we stored when the
-- checkout session was created -- and returns only what crediting it needs.
create or replace function resolve_invoice(p_invoice_id text, p_gateway text, p_ref text)
returns table (invoice_id uuid, tenant_id uuid, currency char(3), total_minor bigint,
               status text)
language sql security definer stable
set search_path = public as $$
  select i.id, i.tenant_id, i.currency, i.total_minor, i.status
    from invoices i
   where (p_invoice_id <> '' and i.id::text = p_invoice_id)
      or (p_ref <> '' and i.gateway = p_gateway and i.gateway_ref = p_ref)
   limit 1
$$;
revoke all on function resolve_invoice(text, text, text) from public;
grant execute on function resolve_invoice(text, text, text) to alchemist_app;
