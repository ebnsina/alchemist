<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { setCrumbs } from '$lib/crumbs.svelte';
	import {
		CloudServerIcon,
		FolderLibraryIcon,
		SquareLock01Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import RegionPicker from '$lib/components/RegionPicker.svelte';
	import { listBucketSources, connectBucket, ApiError, type BucketSource } from '$lib/api';

	let sources = $state<BucketSource[]>([]);
	let loading = $state(true);
	let error = $state('');
	let busy = $state(false);
	let open = $state(false);
	let step = $state(0);
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
	async function load() {
		error = '';
		try {
			sources = (await listBucketSources()).bucket_sources;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

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

{#if loading}
	<div class="mt-6 grid gap-2">
		{#each [0, 1] as i (i)}<div class="sk h-16"></div>{/each}
	</div>
{:else if sources.length === 0}
	<div class="card mt-6 py-8 text-center">
		<p class="title">No bucket connected</p>
		<p class="sub mx-auto mt-2 max-w-md">
			Use <b>Add new</b> if your videos already live in S3-compatible storage
			of your own. Otherwise there is nothing to do here — send videos through
			<a href="/app/upload/" class="link">Upload</a> or the API instead.
		</p>
	</div>
{:else}
	<ul class="mt-6 divide-y divide-sunk border-y border-sunk">
		{#each sources as s (s.id)}
			<li class="py-4">
				<div class="flex flex-wrap items-center gap-x-4 gap-y-2">
					<div class="min-w-48 flex-1">
						<p class="font-mono text-sm">{s.bucket}{s.prefix ? '/' + s.prefix : ''}</p>
						<p class="mono mt-0.5">
							{count(s.imported_objects)} taken in · Last checked {when(s.last_synced_at)}
						</p>
					</div>
					<span class="chip" class:chip-on={s.active}>{s.active ? 'Syncing' : 'Paused'}</span>
				</div>
				{#if s.last_error}
					<p class="mt-2 text-sm text-red">Last scan did not finish. We will try again shortly.</p>
				{/if}
			</li>
		{/each}
	</ul>
{/if}
