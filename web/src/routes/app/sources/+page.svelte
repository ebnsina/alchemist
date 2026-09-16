<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { setCrumbs } from '$lib/crumbs.svelte';
	import {
		CloudServerIcon,
		FolderLibraryIcon,
		SquareLock01Icon,
		PlayIcon,
		PauseIcon,
		Unlink01Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon
	} from '@hugeicons/core-free-icons';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import Seo from '$lib/Seo.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import RowMenu from '$lib/components/RowMenu.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import RegionPicker from '$lib/components/RegionPicker.svelte';
	import {
		listBucketSources,
		patchBucketSource,
		deleteBucketSource,
		connectBucket,
		ApiError,
		type BucketSource
	} from '$lib/api';

	let rows = $state<BucketSource[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');
	let busy = $state(false);
	let open = $state(false);
	let step = $state(0);

	let page = $state(0);
	let size = $state(10);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');

	// The dialog owns a plain boolean: passing !!row unbound means Escape closes it and
	// the next render opens it straight back up.
	let removeOpen = $state(false);
	let removing = $state<BucketSource | null>(null);
	let busyRow = $state(false);
	let form = $state({
		endpoint: '',
		region: 'us-east-1',
		bucket: '',
		prefix: '',
		access_key_id: '',
		secret_access_key: ''
	});

	const steps = [
		{
			key: 'where',
			icon: CloudServerIcon,
			title: 'Where does it live?',
			hint: 'Any S3-compatible storage — AWS, Backblaze, Wasabi, your own MinIO.'
		},
		{
			key: 'what',
			icon: FolderLibraryIcon,
			title: 'Which bucket?',
			hint: 'Narrow it with a prefix if only part of the bucket is video.'
		},
		{
			key: 'keys',
			icon: SquareLock01Icon,
			title: 'How do we read it?',
			hint: 'Read access is enough. Nothing here writes to your bucket.'
		}
	];

	// Only that there is something to send. Whether the endpoint resolves, whether the
	// key works, and whether the bucket exists are all the API's ruling.
	const filled = $derived(
		[
			form.endpoint.trim().length > 0,
			form.bucket.trim().length > 0,
			form.access_key_id.trim().length > 0 && form.secret_access_key.length > 0
		][step]
	);

	function back() {
		error = '';
		if (step > 0) step -= 1;
	}

	function next() {
		error = '';
		if (!filled) return;
		if (step < steps.length - 1) step += 1;
		else connect();
	}
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
			const r = await listBucketSources(query);
			rows = r.bucket_sources;
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

	async function toggleActive(b: BucketSource) {
		error = '';
		try {
			await patchBucketSource(b.id, { active: !b.active });
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		}
	}

	// A bucket with videos still being made answers 409 bucket_in_use, and that refusal
	// is the whole answer the customer needs — it is shown, not swallowed.
	async function remove() {
		if (!removing) return;
		error = '';
		busyRow = true;
		try {
			await deleteBucketSource(removing.id);
			removeOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
			removeOpen = false;
		} finally {
			busyRow = false;
		}
	}

	async function connect() {
		error = '';
		busy = true;
		try {
			await connectBucket({ ...form });
			form = { endpoint: '', region: 'us-east-1', bucket: '', prefix: '', access_key_id: '', secret_access_key: '' };
			open = false;
			step = 0;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}

	const when = (iso: string | null) =>
		iso ? new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso)) : 'Not yet';
	const count = (n: number) => new Intl.NumberFormat('en').format(n);

	$effect(() => {
		setCrumbs([{ label: 'Videos', href: '/app/videos/' }, { label: 'Connected buckets' }]);
	});
	const columns: ColumnDef<any, BucketSource>[] = [
		{ accessorKey: 'bucket', header: 'Bucket' },
		{
			accessorKey: 'prefix',
			header: 'Prefix',
			enableSorting: false,
			cell: (c) => (c.getValue() as string) || 'The whole bucket'
		},
		{
			id: 'state',
			header: 'State',
			enableSorting: false,
			cell: (c) => {
				const b = c.row.original as BucketSource;
				// A failed scan is not a state you can filter on, but it is the one a
				// customer needs to see: syncing and failing looks identical otherwise.
				const label = !b.active ? 'Paused' : b.last_error ? 'Scan failed' : 'Syncing';
				return renderComponent(Badge, {
					label,
					tone: !b.active ? 'idle' : b.last_error ? 'bad' : 'good'
				});
			}
		},
		{
			accessorKey: 'imported_objects',
			header: 'Taken in',
			enableSorting: false,
			cell: (c) => count(c.getValue() as number)
		},
		{
			accessorKey: 'last_synced_at',
			header: 'Last checked',
			enableSorting: false,
			cell: (c) => when(c.getValue() as string | null)
		},
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const b = c.row.original as BucketSource;
				return renderComponent(RowMenu, {
					label: `Actions for ${b.bucket}`,
					actions: [
						{
							label: b.active ? 'Pause' : 'Resume',
							icon: b.active ? PauseIcon : PlayIcon,
							onclick: () => toggleActive(b)
						},
						{
							label: 'Disconnect',
							icon: Unlink01Icon,
							danger: true,
							onclick: () => {
								removing = b;
								removeOpen = true;
							}
						}
					]
				});
			}
		}
	];
</script>

<Seo title="Connected buckets — Alchemist" description="Point us at a bucket and we take what lands in it." />

<header class="flex flex-wrap items-start justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Connected buckets</h1>
		<p class="sub mt-1 max-w-xl">
			Point us at a bucket you already own and every video that appears in it gets taken
			in. We re-list it every fifteen minutes, so nothing is missed when a notification
			goes astray.
		</p>
	</div>
	<button
		type="button"
		aria-label="Connect a bucket"
		class="btn-solid flex-none"
		onclick={() => {
			step = 0;
			open = true;
		}}
	>
		Add new
	</button>
</header>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

<Dialog bind:open title="Connect a bucket">
	<Steps {steps} {step}>
			<div class="mt-5 grid gap-4">
				{#if step === 0}
					<label class="block">
						<span class="label mb-1.5 block">Endpoint</span>
						<input
							bind:value={form.endpoint}
							class="field"
							type="url"
							placeholder="https://s3.eu-central-1.amazonaws.com"
							required
						/>
						<span class="sub mt-1.5 block">
							The address of the storage service, not of the bucket itself.
						</span>
					</label>
					<div class="block">
						<span class="label mb-1.5 block">Region</span>
						<RegionPicker bind:value={form.region} />
					</div>
				{:else if step === 1}
					<label class="block">
						<span class="label mb-1.5 block">Bucket</span>
						<input bind:value={form.bucket} class="field" type="text" placeholder="lectures" required />
					</label>
					<label class="block">
						<span class="label mb-1.5 block">Prefix (optional)</span>
						<input bind:value={form.prefix} class="field" type="text" placeholder="2026/term-1/" />
						<span class="sub mt-1.5 block">
							Leave it empty to watch the whole bucket. We only take video files either way —
							PDFs and images in the same place are ignored.
						</span>
					</label>
				{:else}
					<label class="block">
						<span class="label mb-1.5 block">Access key ID</span>
						<input
							bind:value={form.access_key_id}
							class="field"
							type="text"
							autocomplete="off"
							spellcheck="false"
							required
						/>
					</label>
					<label class="block">
						<span class="label mb-1.5 block">Secret access key</span>
						<input
							bind:value={form.secret_access_key}
							class="field"
							type="password"
							autocomplete="off"
							spellcheck="false"
							required
						/>
					</label>
					<p class="sub">
						The secret is encrypted before it is stored, tied to this one connection, and never
						shown again or sent back to this page. Give it a read-only key — nothing here needs
						to write to your bucket.
					</p>
					<dl class="grid gap-2 rounded-md border border-sunk p-4">
						<div class="flex items-baseline justify-between gap-4">
							<dt class="label">Bucket</dt>
							<dd class="mono truncate">{form.bucket}{form.prefix ? '/' + form.prefix : ''}</dd>
						</div>
						<div class="flex items-baseline justify-between gap-4">
							<dt class="label">Region</dt>
							<dd class="mono">{form.region}</dd>
						</div>
					</dl>
				{/if}

				{#if error}
					<p class="text-sm text-red" role="alert">{error}</p>
				{/if}

			</div>
	</Steps>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			{#if step > 0 && !busy}
				<button type="button" class="btn" onclick={back}>
					<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2.2} />
					Back
				</button>
			{/if}
			<button
				type="button"
				class="btn-solid"
				onclick={next}
				disabled={busy || !filled}
				aria-disabled={busy || !filled}
			>
				{#if busy}
					Connecting…
				{:else if step < steps.length - 1}
					Next
					<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2.2} />
				{:else}
					Connect and scan
				{/if}
			</button>
		</div>
	{/snippet}
</Dialog>

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
		searchLabel="Search by bucket or prefix"
	>
		{#snippet empty()}
			<p class="title">No bucket connected</p>
			<p class="sub mx-auto mt-2 max-w-md">
				{#if q}
					No bucket matches that. Clear the search to see them all.
				{:else}
					Use <b>Add new</b> if your videos already live in S3-compatible storage of your
					own. Otherwise there is nothing to do here — send videos through
					<a href="/app/upload/" class="link">Upload</a> or the API instead.
				{/if}
			</p>
		{/snippet}
	</DataTable>
</div>

<Confirm
	bind:open={removeOpen}
	title="Disconnect this bucket?"
	confirm="Yes, disconnect it"
	destructive
	busy={busyRow}
	onconfirm={remove}
>
	<p class="font-mono text-sm">
		{removing ? removing.bucket + (removing.prefix ? '/' + removing.prefix : '') : ''}
	</p>
	<p class="sub mt-2">
		We stop watching it and forget which files we have already taken in. Videos we made from
		it stay in your library and keep playing. Connecting the same bucket again takes
		everything in it in a second time.
	</p>
</Confirm>
