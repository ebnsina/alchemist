<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Upload01Icon,
		ViewIcon,
		PencilEdit02Icon,
		Scissor01Icon,
		Delete02Icon
	} from '@hugeicons/core-free-icons';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import Seo from '$lib/Seo.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import AssetRow from '$lib/components/AssetRow.svelte';
	import RowMenu from '$lib/components/RowMenu.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import { ASSET_STATE, assetName, bytes, clock, when } from '$lib/assets';
	import {
		listAssets,
		patchAsset,
		deleteAsset,
		ApiError,
		type Asset
	} from '$lib/api';

	let rows = $state<Asset[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');

	let page = $state(0);
	let size = $state(25);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');
	let only = $state('');

	// The dialog owns a plain boolean: passing !!row unbound means Escape closes it and
	// the next render opens it straight back up.
	let renameOpen = $state(false);
	let removeOpen = $state(false);
	let renaming = $state<Asset | null>(null);
	let title = $state('');
	let removing = $state<Asset | null>(null);
	let busyRow = $state(false);

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

	async function rename(e: SubmitEvent) {
		e.preventDefault();
		if (!renaming) return;
		error = '';
		busyRow = true;
		try {
			await patchAsset(renaming.id, title.trim());
			renameOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busyRow = false;
		}
	}

	async function remove() {
		if (!removing) return;
		error = '';
		busyRow = true;
		try {
			await deleteAsset(removing.id);
			removeOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busyRow = false;
		}
	}

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
		},
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const a = c.row.original as Asset;
				return renderComponent(RowMenu, {
					label: `Actions for ${assetName(a)}`,
					actions: [
						{ label: 'View', icon: ViewIcon, href: `/app/videos/${a.id}/` },
						{
							label: 'Rename',
							icon: PencilEdit02Icon,
							onclick: () => {
								renaming = a;
								title = a.title ?? '';
								renameOpen = true;
							}
						},
						{
							label: 'Edit in Studio',
							icon: Scissor01Icon,
							href: `/app/studio/${a.id}/`,
							disabled: !ASSET_STATE[a.state]?.done,
							why: 'Editing opens once the video has finished'
						},
						{ label: 'Delete', icon: Delete02Icon, danger: true, onclick: () => { removing = a; removeOpen = true; } }
					]
				});
			}
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
		searchLabel="Search by name"
	>
		{#snippet toolbar()}
			<label class="flex items-center gap-2">
				<span class="vh">Show</span>
				<select
					class="select w-44"
					value={only}
					onchange={(e) => {
						only = e.currentTarget.value;
						page = 0;
					}}
				>
					{#each FILTERS as f (f.value)}<option value={f.value}>{f.label}</option>{/each}
				</select>
			</label>
		{/snippet}

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

<Dialog
	bind:open={renameOpen}
	title="Rename this video"
	hint="A name for you and your team. Viewers never see it."
>
	<form id="rename-form" onsubmit={rename}>
		<label class="block">
			<span class="vh">Name</span>
			<input bind:value={title} class="field" type="text" maxlength="200" placeholder="e.g. Class 9 — Chapter 3" />
		</label>
		<p class="sub mt-3">
			Leave it empty to go back to showing the first eight characters of its id.
		</p>
	</form>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			<button type="button" class="btn" onclick={() => (renameOpen = false)}>Cancel</button>
			<button type="submit" form="rename-form" class="btn-solid" disabled={busyRow}>
				{busyRow ? 'Saving…' : 'Save the name'}
			</button>
		</div>
	{/snippet}
</Dialog>

<Confirm
	bind:open={removeOpen}
	title="Delete this video?"
	confirm="Yes, delete it"
	destructive
	busy={busyRow}
	onconfirm={remove}
>
	<p class="text-sm">{removing ? assetName(removing) : ''}</p>
	<p class="sub mt-2">
		Every size we made goes with it and any link you have handed out stops playing. This cannot
		be undone.
	</p>
</Confirm>
