<script lang="ts">
	import {
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
	import Badge from '$lib/components/Badge.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import { ASSET_STATE, assetName, bytes, clock, when } from '$lib/assets';
	import { listAssets, patchAsset, deleteAsset, ApiError, type Asset } from '$lib/api';

	let rows = $state<Asset[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');

	let page = $state(0);
	let size = $state(10);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');

	let renameOpen = $state(false);
	let removeOpen = $state(false);
	let renaming = $state<Asset | null>(null);
	let removing = $state<Asset | null>(null);
	let title = $state('');
	let busyRow = $state(false);

	// A finished broadcast is an ordinary asset in live_ended, so this page is the
	// videos list with that one state pinned, and each row opens the same video page.
	const query = $derived({
		limit: size,
		offset: page * size,
		q,
		sort: sorting[0]?.id ?? 'created_at',
		order: (sorting[0]?.desc ?? true ? 'desc' : 'asc') as 'asc' | 'desc',
		filters: { state: 'live_ended' }
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
			header: 'Recording',
			enableSorting: false,
			cell: (c) => renderComponent(AssetRow, { asset: c.row.original })
		},
		{
			accessorKey: 'state',
			header: 'State',
			cell: (c) => {
				const st = ASSET_STATE[String(c.getValue())];
				return renderComponent(Badge, {
					label: st?.chip ?? String(c.getValue()),
					tone: st?.tone ?? 'idle'
				});
			}
		},
		{ accessorKey: 'duration_sec', header: 'Length', cell: (c) => clock(c.getValue() as number) },
		{ accessorKey: 'source_bytes', header: 'Size', cell: (c) => bytes(c.getValue() as number) },
		{ accessorKey: 'created_at', header: 'Added', cell: (c) => when(String(c.getValue())) },
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const a = c.row.original as Asset;
				return renderComponent(RowMenu, {
					label: `Actions for ${assetName(a)}`,
					actions: [
						{ label: 'Watch', icon: ViewIcon, href: `/app/videos/${a.id}/` },
						{
							label: 'Rename',
							icon: PencilEdit02Icon,
							onclick: () => {
								renaming = a;
								title = a.title ?? '';
								renameOpen = true;
							}
						},
						{ label: 'Edit in Studio', icon: Scissor01Icon, href: `/app/studio/${a.id}/` },
						{
							label: 'Delete',
							icon: Delete02Icon,
							danger: true,
							onclick: () => {
								removing = a;
								removeOpen = true;
							}
						}
					]
				});
			}
		}
	];
</script>

<Seo title="Recordings — Alchemist" description="What your live broadcasts left behind." />

<h1 class="text-2xl font-semibold tracking-tight">Recordings</h1>
<p class="sub mt-1.5 max-w-xl">
	Every broadcast is kept as an ordinary video. Open one to watch it, share the link, or edit it.
</p>

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
		searchLabel="Search by name or ID"
	>
		{#snippet empty()}
			<p class="title">Nothing recorded yet</p>
			<p class="sub mx-auto mt-2 max-w-sm">
				{#if q}
					No recording matches that. Clear the search to see them all.
				{:else}
					A recording appears here when a broadcast ends. Start one under
					<a href="/app/live/" class="link">Streams</a>.
				{/if}
			</p>
		{/snippet}
	</DataTable>
</div>

<Dialog
	bind:open={renameOpen}
	title="Rename this recording"
	hint="A name for you and your team. Viewers never see it."
>
	<form id="rename-form" onsubmit={rename}>
		<label class="block">
			<span class="vh">Name</span>
			<input
				bind:value={title}
				class="field"
				type="text"
				maxlength="200"
				placeholder="e.g. Friday physics — week 3"
			/>
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
	title="Delete this recording?"
	confirm="Yes, delete it"
	destructive
	busy={busyRow}
	onconfirm={remove}
>
	<p class="text-sm">{removing ? assetName(removing) : ''}</p>
	<p class="sub mt-2">
		The broadcast it kept is gone with it, every size we made goes too, and any link you have
		handed out stops playing. This cannot be undone.
	</p>
</Confirm>
