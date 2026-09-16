<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Upload01Icon } from '@hugeicons/core-free-icons';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import Seo from '$lib/Seo.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import AssetRow from '$lib/components/AssetRow.svelte';
	import { ASSET_STATE, bytes, clock, when } from '$lib/assets';
	import { listAssets, ApiError, type Asset } from '$lib/api';

	let rows = $state<Asset[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');

	let page = $state(0);
	let size = $state(25);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');
	let only = $state('');

	// Anything still moving is worth another look without the customer asking.
	const busy = $derived(rows.some((a) => !ASSET_STATE[a.state]?.done && a.state !== 'failed'));

	// One read of every control the table owns, so a change to any of them refetches
	// exactly once rather than each firing its own request.
	const query = $derived({
		limit: size,
		offset: page * size,
		q,
		sort: sorting[0]?.id ?? 'created_at',
		order: (sorting[0]?.desc ?? true ? 'desc' : 'asc') as 'asc' | 'desc',
		filters: { state: only || undefined }
	});

	async function load() {
		error = '';
		try {
			const r = await listAssets(query);
			rows = r.assets;
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

	$effect(() => {
		if (!busy) return;
		const id = setInterval(load, 5000);
		return () => clearInterval(id);
	});

	const FILTERS = [
		{ value: '', label: 'Everything' },
		{ value: 'ready', label: 'Ready' },
		{ value: 'encoding', label: 'Being made' },
		{ value: 'failed', label: 'Did not work' }
	];

	const columns: ColumnDef<any, Asset>[] = [
		{
			id: 'title',
			header: 'Video',
			enableSorting: false,
			cell: (c) => renderComponent(AssetRow, { asset: c.row.original })
		},
		{
			accessorKey: 'state',
			header: 'State',
			cell: (c) => ASSET_STATE[String(c.getValue())]?.chip ?? String(c.getValue())
		},
		{ accessorKey: 'duration_sec', header: 'Length', cell: (c) => clock(c.getValue() as number) },
		{ accessorKey: 'source_bytes', header: 'Size', cell: (c) => bytes(c.getValue() as number) },
		{
			accessorKey: 'created_at',
			header: 'Added',
			cell: (c) => when(String(c.getValue()))
		}
	];
</script>

<Seo title="Videos — Alchemist" description="Every video you have sent us." />

<header class="flex flex-wrap items-start justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Videos</h1>
		<p class="sub mt-1.5 max-w-xl">
			Everything you have sent us. Open one to watch it, get its playback links, or see what
			is still being made.
		</p>
	</div>
	<a href="/app/upload/" class="btn-solid btn-sm flex-none">
		<HugeiconsIcon icon={Upload01Icon} size={14} strokeWidth={2} />
		Upload
	</a>
</header>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

<div class="mt-6 flex flex-wrap gap-1.5">
	{#each FILTERS as f (f.value)}
		<button
			type="button"
			class="chip"
			class:chip-on={only === f.value}
			aria-pressed={only === f.value}
			onclick={() => {
				only = f.value;
				page = 0;
			}}
		>
			{f.label}
		</button>
	{/each}
</div>

<div class="mt-4">
	<DataTable
		{columns}
		{rows}
		{total}
		{loading}
		bind:page
		bind:size
		bind:sorting
		bind:q
		searchLabel="Search by name"
	>
		{#snippet empty()}
			<p class="title">Nothing here</p>
			<p class="sub mx-auto mt-2 max-w-sm">
				{#if q || only}
					No video matches that. Clear the search, or pick Everything above.
				{:else}
					Send us a video and it appears here while it is still being made.
				{/if}
			</p>
		{/snippet}
	</DataTable>
</div>
