<script lang="ts">
	import { page } from '$app/state';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { ArrowLeft01Icon, Copy01Icon, Tick02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Player from '$lib/components/Player.svelte';
	import Advanced from '$lib/components/Advanced.svelte';
	import { getAsset, ApiError, type AssetDetail } from '$lib/api';

	let asset = $state<AssetDetail | null>(null);
	let error = $state('');
	let loading = $state(true);
	let copied = $state('');

	const id = $derived(page.params.id ?? '');

	async function load() {
		try {
			asset = await getAsset(id);
			error = '';
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (id) load();
	});

	// Keep asking while work is outstanding, and stop the moment it is not.
	const working = $derived(
		!!asset &&
			(asset.renditions.some((r) => r.state !== 'ready' && !r.lazy) ||
				!['ready', 'partially_ready', 'failed'].includes(asset.state))
	);
	$effect(() => {
		if (!working) return;
		const t = setInterval(load, 4000);
		return () => clearInterval(t);
	});

	const STATE: Record<string, string> = {
		created: 'Waiting for the file',
		uploading: 'Uploading',
		uploaded: 'Queued',
		probing: 'Looking it over',
		mezzanine: 'Preparing',
		analyzing: 'Preparing',
		encoding: 'Making the sizes',
		packaging: 'Almost there',
		partially_ready: 'Ready to watch',
		ready: 'Ready to watch',
		failed: 'Did not work'
	};

	const mbps = (bps: number) =>
		new Intl.NumberFormat('en', { maximumFractionDigits: 2 }).format(bps / 1_000_000) + ' Mbps';

	const size = (bytes: number | null | undefined) =>
		bytes == null
			? '—'
			: bytes >= 1_000_000_000
				? new Intl.NumberFormat('en', { style: 'unit', unit: 'gigabyte', maximumFractionDigits: 1 }).format(bytes / 1_000_000_000)
				: new Intl.NumberFormat('en', { style: 'unit', unit: 'megabyte', maximumFractionDigits: 0 }).format(bytes / 1_000_000);

	const length = (secs: number | undefined) => {
		if (secs == null) return '—';
		const m = Math.floor(secs / 60);
		const s = Math.round(secs % 60);
		return `${m}:${String(s).padStart(2, '0')}`;
	};

	const pct = (r: { chunks_done: number; chunks_total: number; state: string }) =>
		r.state === 'ready' ? 100 : r.chunks_total > 0 ? Math.round((r.chunks_done / r.chunks_total) * 100) : 0;

	async function copy(text: string, what: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied = what;
			setTimeout(() => (copied = ''), 2000);
		} catch {
			copied = '';
		}
	}
</script>

<Seo title="Video — Alchemist" description="What we made from this video." />

<a href="/app/" class="inline-flex items-center gap-1.5 text-sm text-secondary transition-colors hover:text-primary">
	<HugeiconsIcon icon={ArrowLeft01Icon} size={15} strokeWidth={2} />
	All videos
</a>

{#if loading}
	<p class="mt-6 text-sm text-secondary">Loading…</p>
{:else if error}
	<p class="mt-6 text-sm text-danger" role="alert">{error}</p>
{:else if asset}
	<header class="mt-4 flex flex-wrap items-start justify-between gap-4">
		<div class="min-w-0">
			<h1 class="truncate font-mono text-xl font-semibold tracking-tight">{asset.id}</h1>
			<p class="mt-1 text-sm text-secondary">
				{STATE[asset.state] ?? asset.state}
				{#if asset.duration_seconds} · {length(asset.duration_seconds)}{/if}
				{#if asset.width && asset.height} · {asset.width}×{asset.height}{/if}
				{#if asset.source_bytes} · sent {size(asset.source_bytes)}{/if}
			</p>
		</div>
		{#if working}
			<span class="flex items-center gap-2 text-xs text-secondary">
				<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-secondary"></span>
				Still working — this page keeps itself up to date
			</span>
		{/if}
	</header>

	{#if asset.error_code}
		<p class="mt-4 rounded-xl bg-danger/10 px-4 py-3 text-sm text-danger" role="alert">
			This one did not work: <code class="font-mono text-xs">{asset.error_code}</code>
		</p>
	{/if}

	{#if asset.playback}
		<section class="mt-6">
			<Player hls={asset.playback.hls} poster={asset.playback.poster} />
		</section>
	{/if}

	<section class="mt-8">
		<h2 class="text-lg font-semibold tracking-tight">The sizes we made</h2>
		<p class="mt-1 text-sm text-secondary">
			Each one is encoded in chunks, in parallel. The ones marked on demand are only made when
			somebody actually asks for that size.
		</p>

		{#if asset.renditions.length === 0}
			<p class="mt-4 text-sm text-secondary">Nothing yet — the ladder is worked out after we look the file over.</p>
		{:else}
			<ul class="mt-4 grid gap-3">
				{#each asset.renditions as r (r.height + r.codec)}
					{@const done = pct(r)}
					<li class="card p-5">
						<div class="flex flex-wrap items-center justify-between gap-3">
							<div>
								<p class="font-semibold">
									{r.height}p
									<span class="ml-1 font-normal text-secondary">{r.codec.toUpperCase()} · {mbps(r.bitrate_bps)}</span>
								</p>
								<p class="mt-0.5 text-xs text-secondary">
									{#if r.lazy && r.state !== 'ready'}
										Made on demand, when someone plays it
									{:else if r.chunks_total > 0}
										{r.chunks_done} of {r.chunks_total} chunks
									{:else}
										{r.state}
									{/if}
									{#if r.bytes} · {size(r.bytes)}{/if}
								</p>
							</div>
							<span class="tabular-nums text-sm {r.state === 'ready' ? 'text-tertiary' : 'text-secondary'}">
								{r.state === 'ready' ? 'Ready' : `${done}%`}
							</span>
						</div>
						<div class="mt-3 h-1.5 overflow-hidden rounded-full bg-outline">
							<div
								class="h-full rounded-full transition-[width] duration-500 ease-out {r.state === 'ready'
									? 'bg-tertiary'
									: 'bg-secondary'}"
								style="width: {done}%"
							></div>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	{#if asset.playback}
		<section class="mt-8">
			<h2 class="text-lg font-semibold tracking-tight">The links</h2>
			<p class="mt-1 text-sm text-secondary">
				Signed, good for four hours. Ask for the asset again when you need fresh ones rather
				than building them yourself.
			</p>
			<div class="card mt-4 divide-y divide-outline">
				{#each Object.entries(asset.playback) as [kind, url] (kind)}
					<div class="flex items-center gap-3 p-4">
						<span class="w-20 flex-none text-xs text-secondary">{kind}</span>
						<code class="min-w-0 flex-1 truncate font-mono text-xs text-secondary">{url}</code>
						<button
							type="button"
							class="flex flex-none items-center gap-1.5 text-xs text-tertiary"
							onclick={() => copy(url, kind)}
						>
							<HugeiconsIcon icon={copied === kind ? Tick02Icon : Copy01Icon} size={13} strokeWidth={2} />
							{copied === kind ? 'Copied' : 'Copy'}
						</button>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<Advanced {asset} live={working} />
{/if}
