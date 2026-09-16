<script lang="ts">
	import { goto } from '$app/navigation';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { ViewIcon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import RowMenu from '$lib/components/RowMenu.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import { setCrumbs } from '$lib/crumbs.svelte';
	import { me } from '$lib/me.svelte';
	import { staffTenants, impersonate, ApiError, type StaffTenant } from '$lib/api';

	let rows = $state<StaffTenant[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');

	let page = $state(0);
	let size = $state(10);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');

	let picked = $state<StaffTenant | null>(null);
	let confirming = $state(false);
	let busy = $state(false);

	const query = $derived({
		limit: size,
		offset: page * size,
		q,
		sort: sorting[0]?.id ?? 'created_at',
		order: (sorting[0]?.desc ?? true ? 'desc' : 'asc') as 'asc' | 'desc'
	});

	$effect(() => {
		setCrumbs([{ label: 'Accounts' }]);
	});

	async function load() {
		error = '';
		try {
			const r = await staffTenants(query);
			rows = r.tenants;
			total = r.total;
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

	async function view() {
		if (!picked) return;
		busy = true;
		try {
			await impersonate(picked.tenant_id);
			// A full load, not a client navigation: every page holds data fetched as
			// the previous account and none of it is theirs any more.
			location.assign('/app/');
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
			busy = false;
		}
	}

	const num = (n: number) => new Intl.NumberFormat('en').format(n);
	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium' }).format(new Date(iso));

	const columns: ColumnDef<any, StaffTenant>[] = [
		{
			accessorKey: 'name',
			header: 'Account',
			cell: (c) => String(c.getValue())
		},
		{ accessorKey: 'assets', header: 'Videos', enableSorting: true, cell: (c) => num(c.getValue() as number) },
		{
			accessorKey: 'live_streams',
			header: 'Streams',
			enableSorting: false,
			cell: (c) => num(c.getValue() as number)
		},
		{
			accessorKey: 'members',
			header: 'People',
			enableSorting: false,
			cell: (c) => num(c.getValue() as number)
		},
		{ accessorKey: 'created_at', header: 'Joined', cell: (c) => when(String(c.getValue())) },
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const t = c.row.original as StaffTenant;
				return renderComponent(RowMenu, {
					label: `Actions for ${t.name}`,
					actions: [
						{
							label: 'View as',
							icon: ViewIcon,
							onclick: () => {
								picked = t;
								confirming = true;
							}
						}
					]
				});
			}
		}
	];
</script>

<Seo title="Accounts — Alchemist" description="Every account on this installation." />

<h1 class="text-2xl font-semibold tracking-tight">Accounts</h1>
<p class="sub mt-1.5 max-w-xl">
	Every account here, so you can open one when somebody asks for help. Nothing of theirs is on
	this page — you see what they see by viewing their account, and only ever read it.
</p>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if !me()?.platform_admin && !loading}
	<div class="card mt-6 py-8 text-center">
		<p class="title">This is for Alchemist staff</p>
		<p class="sub mx-auto mt-2 max-w-sm">Your account cannot open this page.</p>
		<a href="/app/" class="btn-solid mt-5">Back to your dashboard</a>
	</div>
{:else}
	<div class="mt-6">
		<DataTable
			{columns}
			{rows}
			{total}
			{loading}
			bind:page
			bind:size
			bind:sorting
			bind:q
			searchLabel="Search by name or ID"
		>
			{#snippet empty()}
				<p class="title">No accounts</p>
				<p class="sub mx-auto mt-2 max-w-sm">
					{q ? 'Nothing matches that. Clear the search.' : 'Nobody has signed up yet.'}
				</p>
			{/snippet}
		</DataTable>
	</div>
{/if}

<Confirm
	bind:open={confirming}
	title="View this account?"
	confirm="Yes, view it"
	{busy}
	onconfirm={view}
>
	<p class="text-sm">{picked?.name ?? ''}</p>
	<p class="sub mt-2">
		The whole dashboard becomes theirs until you stop. You can read everything and change
		nothing — every write is refused while you are viewing.
	</p>
</Confirm>
