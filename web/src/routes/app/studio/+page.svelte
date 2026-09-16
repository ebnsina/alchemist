<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Scissor01Icon,
		ArrowRight01Icon,
		ViewIcon,
		Delete02Icon
	} from '@hugeicons/core-free-icons';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import Seo from '$lib/Seo.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import RowMenu from '$lib/components/RowMenu.svelte';
	import Badge, { type Tone } from '$lib/components/Badge.svelte';
	import { when as stamp } from '$lib/assets';
	import { listAssets, listEdits, deleteEdit, ApiError, type Asset, type Edit } from '$lib/api';

	let assets = $state<Asset[]>([]);
	let edits = $state<Edit[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');
	let picking = $state(false);

	let page = $state(0);
	let size = $state(10);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');

	// The dialog owns a plain boolean: passing !!row unbound means Escape closes it and
	// the next render opens it straight back up.
	let removeOpen = $state(false);
	let removing = $state<Edit | null>(null);
	let busyRow = $state(false);

	// One read of every control the table owns, so a change to any of them refetches
	// exactly once rather than each firing its own request.
	const query = $derived({
		limit: size,
		offset: page * size,
		q,
		sort: sorting[0]?.id ?? 'created_at',
		order: (sorting[0]?.desc ?? true ? 'desc' : 'asc') as 'asc' | 'desc'
	});

	async function load() {
		error = '';
		try {
			const r = await listEdits(query);
			edits = r.edits;
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

	// The picker is a separate list: it names videos to start from, not edits made.
	$effect(() => {
		listAssets({ limit: 100 })
			.then((a) => (assets = a.assets))
			.catch((e) => (error = e instanceof ApiError ? e.message : 'Something went wrong.'));
	});

	// Keep asking while something is rendering, and stop the moment nothing is.
	$effect(() => {
		if (!edits.some((e) => e.state === 'queued' || e.state === 'rendering')) return;
		const t = setInterval(load, 4000);
		return () => clearInterval(t);
	});

	async function remove() {
		if (!removing) return;
		error = '';
		busyRow = true;
		try {
			await deleteEdit(removing.id);
			removeOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busyRow = false;
		}
	}

	// Only finished videos can be edited — there is nothing to cut before the
	// mezzanine exists, and offering it would only produce a refusal.
	const editable = $derived(
		assets.filter((a) => a.state === 'ready' || a.state === 'partially_ready')
	);
	const waiting = $derived(
		assets.filter((a) => !['ready', 'partially_ready', 'failed'].includes(a.state)).length
	);

	const length = (secs: number | null) => {
		if (secs == null) return '—';
		const m = Math.floor(secs / 60);
		return `${m}:${String(Math.round(secs % 60)).padStart(2, '0')}`;
	};
	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium' }).format(new Date(iso));
	// The list endpoint reports height only; the full shape lives on the asset itself.
	const shape = (a: Asset) => (a.height ? `${a.height}p` : '—');

	const STATE: Record<string, { chip: string; tone: Tone }> = {
		queued: { chip: 'Waiting to start', tone: 'idle' },
		rendering: { chip: 'Making it', tone: 'busy' },
		done: { chip: 'Ready', tone: 'good' },
		failed: { chip: 'Did not work', tone: 'bad' }
	};
	const describe = (o: Edit['ops']) => {
		const bits: string[] = [];
		if (o.start_sec || o.end_sec) bits.push('trimmed');
		if (o.crop_w) bits.push('cropped');
		if (o.aspect) bits.push(o.aspect);
		if (o.width) bits.push(`${o.width}px`);
		if (o.mute) bits.push('muted');
		return bits.length ? bits.join(' · ') : 'a copy';
	};
	const columns: ColumnDef<any, Edit>[] = [
		{
			id: 'what',
			header: 'What we made',
			enableSorting: false,
			cell: (c) => describe((c.row.original as Edit).ops)
		},
		{
			accessorKey: 'source_asset_id',
			header: 'From',
			enableSorting: false,
			cell: (c) => String(c.getValue()).slice(0, 8)
		},
		{
			accessorKey: 'state',
			header: 'State',
			cell: (c) => {
				const st = STATE[String(c.getValue())];
				return renderComponent(Badge, {
					label: st?.chip ?? String(c.getValue()),
					tone: st?.tone ?? 'idle'
				});
			}
		},
		{ accessorKey: 'created_at', header: 'Made', cell: (c) => stamp(String(c.getValue())) },
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const e = c.row.original as Edit;
				return renderComponent(RowMenu, {
					label: `Actions for the edit made ${stamp(e.created_at)}`,
					actions: [
						{
							label: 'Open',
							icon: ViewIcon,
							href: e.output_asset_id ? `/app/videos/${e.output_asset_id}/` : undefined,
							disabled: !e.output_asset_id,
							why: 'The new video opens once this edit has finished'
						},
						{
							label: 'Remove',
							icon: Delete02Icon,
							danger: true,
							onclick: () => {
								removing = e;
								removeOpen = true;
							}
						}
					]
				});
			}
		}
	];
</script>

<Seo title="Studio — Alchemist" description="Cut, crop and reshape your videos." />

<header class="flex flex-wrap items-start justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Studio</h1>
		<p class="sub mt-1 max-w-xl">
			Cut a video down, crop it, or turn it upright for phones. Everything you make here is a
			<strong class="text-ink">new</strong> video — the one you started from never changes, so
			a link you have already handed out keeps working.
		</p>
	</div>
	{#if editable.length > 0}
		<button type="button" aria-label="Pick a video to edit" class="btn-solid flex-none" onclick={() => (picking = true)}>
			<HugeiconsIcon icon={Scissor01Icon} size={15} strokeWidth={2} />
			Add new
		</button>
	{/if}
</header>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="mt-6 grid gap-2">
		{#each [0, 1, 2] as i (i)}<div class="sk h-16"></div>{/each}
	</div>
{:else if editable.length === 0}
	<div class="card mt-6 text-center">
		<p class="title">Nothing to edit yet</p>
		<p class="sub mx-auto mt-2 max-w-sm">
			{#if waiting > 0}
				{waiting} video{waiting === 1 ? ' is' : 's are'} still being processed. Editing opens as
				soon as one is ready.
			{:else}
				Send us a video first. Once it has finished processing you can cut it up here.
			{/if}
		</p>
		<a href="/app/upload/" class="btn-solid mt-6">Upload a video</a>
	</div>
{:else}
	<Dialog bind:open={picking} title="Pick a video to edit">
		<ul class="-mx-6 divide-y divide-sunk border-y border-sunk">
			{#each editable as a (a.id)}
				<li>
					<a href="/app/studio/{a.id}/" class="flex items-center gap-4 px-6 py-3.5 hover:bg-sunk">
						<span class="grid h-9 w-9 flex-none place-items-center rounded-md bg-sunk text-faint">
							<HugeiconsIcon icon={Scissor01Icon} size={16} strokeWidth={1.8} />
						</span>
						<span class="min-w-0 flex-1">
							<span class="block truncate font-mono text-sm">{a.id.slice(0, 8)}</span>
							<span class="mono mt-0.5 block">
								{length(a.duration_sec)} · {shape(a)} · added {when(a.created_at)}
							</span>
						</span>
						<HugeiconsIcon
							icon={ArrowRight01Icon}
							size={15}
							strokeWidth={2.2}
							class="flex-none text-faint"
						/>
					</a>
				</li>
			{/each}
		</ul>
	</Dialog>

	<h2 class="mt-8 text-lg font-semibold tracking-tight">What you have made</h2>

	<div class="mt-4">
		<DataTable
			columns={columns}
			rows={edits}
			{total}
			{loading}
			bind:page
			bind:size
			bind:sorting
			bind:q
			searchLabel="Search by the video's ID"
		>
			{#snippet empty()}
				<p class="title">Nothing made yet</p>
				<p class="sub mx-auto mt-2 max-w-sm">
					{#if q}
						No edit came from a video with that id. Clear the search to see them all.
					{:else}
						Use <b>Add new</b> to pick one of your {editable.length} finished videos and
						cut it down. The original is never touched.
					{/if}
				</p>
			{/snippet}
		</DataTable>
	</div>

	<Confirm
		bind:open={removeOpen}
		title="Remove this edit from the list?"
		confirm="Yes, remove it"
		destructive
		busy={busyRow}
		onconfirm={remove}
	>
		<p class="text-sm">{removing ? describe(removing.ops) : ''}</p>
		<p class="sub mt-2">
			Only this record goes. The video it made stays in your library and keeps playing —
			delete that under Videos if you want it gone too.
		</p>
	</Confirm>
{/if}
