<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import { readOnly } from '$lib/me.svelte';
	import { when } from '$lib/assets';
	import {
		getBilling,
		listInvoices,
		payInvoice,
		ApiError,
		type Billing,
		type Invoice
	} from '$lib/api';

	let billing = $state<Billing | null>(null);
	let rows = $state<Invoice[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');
	let paying = $state('');

	let page = $state(0);
	let size = $state(12);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'issued_at', desc: true }]);

	const query = $derived({
		limit: size,
		offset: page * size,
		sort: sorting[0]?.id ?? 'issued_at',
		order: (sorting[0]?.desc ?? true ? 'desc' : 'asc') as 'asc' | 'desc'
	});

	async function load() {
		error = '';
		try {
			const [b, list] = await Promise.all([getBilling(), listInvoices(query)]);
			billing = b;
			rows = list.invoices;
			total = list.total;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		query;
		load();
	});

	// Amounts arrive in minor units and are formatted through Intl, so the taka sign,
	// the digit grouping and the decimal mark all come from the locale rather than
	// from anything typed here.
	const money = (minor: number, currency: string) =>
		new Intl.NumberFormat(currency === 'USD' ? 'en-US' : 'en-BD', {
			style: 'currency',
			currency,
			currencyDisplay: 'narrowSymbol'
		}).format(minor / 100);

	const KIND: Record<string, string> = {
		ingest: 'Video sent in',
		storage: 'Video held for you',
		egress: 'Video watched',
		peak_viewers: 'Most watching at once'
	};

	const STATE: Record<string, { chip: string; tone: 'good' | 'bad' | 'busy' | 'idle' }> = {
		paid: { chip: 'Paid', tone: 'good' },
		open: { chip: 'Due', tone: 'busy' },
		failed: { chip: 'Not paid', tone: 'bad' },
		void: { chip: 'Cancelled', tone: 'idle' }
	};

	async function pay(id: string) {
		error = '';
		paying = id;
		try {
			const { pay_url } = await payInvoice(id);
			// Straight to the gateway's own page. Nothing about a card is ever typed
			// into this dashboard, so there is nothing here for it to leak.
			window.location.href = pay_url;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
			paying = '';
		}
	}

	const columns: ColumnDef<any, Invoice>[] = [
		{
			id: 'period',
			header: 'Month',
			enableSorting: false,
			cell: (c) => {
				const v = c.row.original as Invoice;
				return new Intl.DateTimeFormat('en', { month: 'long', year: 'numeric' }).format(
					new Date(v.period_start)
				);
			}
		},
		{
			accessorKey: 'total_minor',
			header: 'Amount',
			cell: (c) => money(c.getValue() as number, (c.row.original as Invoice).currency)
		},
		{
			accessorKey: 'status',
			header: 'State',
			cell: (c) => {
				const st = STATE[String(c.getValue())];
				return renderComponent(Badge, {
					label: st?.chip ?? String(c.getValue()),
					tone: st?.tone ?? 'idle'
				});
			}
		},
		{ accessorKey: 'issued_at', header: 'Issued', cell: (c) => when(String(c.getValue())) }
	];
</script>

<Seo title="Billing — Alchemist" description="What you have used and what it came to." />

<header class="min-w-0">
	<h1 class="text-2xl font-semibold tracking-tight">Billing</h1>
	<p class="sub mt-1.5 max-w-xl">
		One bill a month, for what you actually used. Every line shows the amount, the rate and
		what it came to, so nothing here has to be taken on trust.
	</p>
</header>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="sk mt-6 h-28"></div>
{:else if billing}
	{#if billing.status !== 'active'}
		<div class="card mt-6 border border-red" role="alert">
			<p class="title text-red">
				{billing.status === 'suspended' ? 'Uploads are paused' : 'A payment did not go through'}
			</p>
			<p class="sub mt-2 max-w-xl">
				{#if billing.status === 'suspended'}
					We could not collect an invoice, so new uploads are paused. Everything you have
					already sent us <strong class="text-ink">keeps playing</strong> — your viewers are
					not affected. Pay the invoice below and uploads start again straight away.
				{:else}
					We will try again over the next few days. Nothing is paused yet, and your videos
					are playing normally.
				{/if}
			</p>
		</div>
	{/if}

	<dl class="mt-6 grid gap-3 sm:grid-cols-3">
		<div class="card">
			<p class="label">Outstanding</p>
			<p class="num mt-2.5 text-[22px] leading-none">
				{money(billing.outstanding_minor, billing.currency)}
			</p>
			<p class="sub mt-2">Across every invoice not yet paid.</p>
		</div>
		<div class="card">
			<p class="label">Billed in</p>
			<p class="num mt-2.5 text-[22px] leading-none">{billing.currency}</p>
			<p class="sub mt-2">
				Paid through {billing.gateway === 'sslcommerz' ? 'SSLCommerz' : 'Stripe'}.
			</p>
		</div>
		<div class="card">
			<p class="label">Account</p>
			<p class="num mt-2.5 text-[22px] leading-none">
				{billing.status === 'active' ? 'Good' : billing.status === 'past_due' ? 'Late' : 'Paused'}
			</p>
			<p class="sub mt-2">Playback never stops, whatever this says.</p>
		</div>
	</dl>
{/if}

<div class="mt-8">
	<DataTable {columns} {rows} {total} {loading} bind:page bind:size bind:sorting>
		{#snippet empty()}
			<p class="title">No bills yet</p>
			<p class="sub mx-auto mt-2 max-w-sm">
				The first one arrives after your first full month. Until then, Usage shows what is
				adding up.
			</p>
		{/snippet}
	</DataTable>
</div>

{#if rows.length}
	<section class="mt-8">
		<h2 class="text-lg font-semibold tracking-tight">What made up each bill</h2>
		<div class="mt-4 grid gap-3">
			{#each rows as inv (inv.id)}
				<div class="card">
					<div class="flex flex-wrap items-baseline justify-between gap-3">
						<p class="text-sm font-semibold">
							{new Intl.DateTimeFormat('en', { month: 'long', year: 'numeric' }).format(
								new Date(inv.period_start)
							)}
						</p>
						<p class="num">{money(inv.total_minor, inv.currency)}</p>
					</div>

					{#if inv.lines.length}
						<dl class="mt-4 grid gap-2 border-t border-sunk pt-4">
							{#each inv.lines as l (l.kind)}
								<div class="flex items-baseline justify-between gap-3">
									<dt class="sub">
										{KIND[l.kind] ?? l.kind}
										<span class="mono ml-1"
											>{Number(l.quantity).toLocaleString('en')} {l.unit}</span
										>
									</dt>
									<dd class="num flex-none text-[15px]">
										{money(l.amount_minor, inv.currency)}
									</dd>
								</div>
							{/each}
						</dl>
					{:else}
						<p class="sub mt-3">Nothing was used this month, so there is nothing to pay.</p>
					{/if}

					{#if inv.last_error}
						<p class="sub mt-3 text-red">
							Last attempt: {inv.last_error}{inv.attempts > 1 ? ` (tried ${inv.attempts} times)` : ''}
						</p>
					{/if}

					{#if inv.status !== 'paid' && inv.total_minor > 0 && !readOnly()}
						<button
							type="button"
							class="btn-solid btn-sm mt-4"
							disabled={paying === inv.id}
							onclick={() => pay(inv.id)}
						>
							{paying === inv.id ? 'Opening…' : `Pay ${money(inv.total_minor, inv.currency)}`}
						</button>
					{/if}
				</div>
			{/each}
		</div>
	</section>
{/if}
