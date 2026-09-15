<script lang="ts">
	import { page } from '$app/state';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Copy01Icon, Tick02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import { setCrumbs } from '$lib/crumbs.svelte';
	import AssetPlayer from '$lib/components/AssetPlayer.svelte';
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

	$effect(() => {
		setCrumbs([{ label: 'Videos', href: '/app/videos/' }, { label: id.slice(0, 8) }]);
	});

	// Keep asking while work is outstanding, and stop the moment it is not.
	const working = $derived(
		!!asset &&
			(asset.renditions.some((r) => r.state !== 'ready' && !r.lazy) ||
				!['ready', 'partially_ready', 'failed', 'live_ended'].includes(asset.state))
	);
	$effect(() => {
		if (!working) return;
		const t = setInterval(load, 4000);
		return () => clearInterval(t);
	});

	// Each state gets a name and a sentence. A status word on its own leaves the
	// reader guessing whether to wait, retry, or tell somebody.
	const STATE: Record<string, { label: string; means: string }> = {
		created: { label: 'Waiting for the file', means: 'We have made a space for it but the bytes have not arrived yet.' },
		uploading: { label: 'Arriving', means: 'The file is on its way to storage. Nothing for you to do.' },
		uploaded: { label: 'Queued', means: 'We have it. Work starts as soon as a machine is free.' },
		probing: { label: 'Looking it over', means: 'Reading what is actually in the file — length, size, how it was recorded.' },
		mezzanine: { label: 'Preparing', means: 'Making one clean master copy that every size is then built from.' },
		analyzing: { label: 'Preparing', means: 'Working out how hard this particular video can be squeezed without looking worse.' },
		encoding: { label: 'Making the sizes', means: 'Building each size in parallel chunks. Small ones finish first.' },
		packaging: { label: 'Almost there', means: 'Wrapping the sizes up so any player can read them.' },
		partially_ready: { label: 'Watchable now', means: 'Enough sizes are done to play it. The rest are still coming.' },
		live: { label: 'On air', means: 'This is a broadcast going out right now. It plays below as it happens.' },
		live_ended: { label: 'Broadcast finished', means: 'The live broadcast has ended. What went out is kept here to watch back.' },
		ready: { label: 'Ready to watch', means: 'Every size is made. Nothing left to wait for.' },
		failed: { label: 'Did not work', means: 'Something about the file stopped us. Sending it again rarely helps — the detail below says what happened.' }
	};

	const FAILURES: Record<string, string> = {
		source_unreadable: 'We could not read the file. It may be damaged, or only partly uploaded.',
		no_video_stream: 'There is no video track in it — an audio file or a document, most likely.',
		source_too_large: 'It is bigger than this account is allowed to send.',
		source_url_not_allowed: 'That link pointed somewhere we will not fetch from.',
		source_unreachable: 'We could not download it from that link.',
		encode_failed: 'The encode failed on our side. Nothing wrong with your file — worth trying again.',
		processing_failed: 'The work failed on our side. Worth trying again.'
	};

	// What each playback link is actually for.
	const LINKS: Record<string, { label: string; what: string }> = {
		hls: { label: 'HLS', what: 'Hand this to a player. iPhones and most web players want this one.' },
		dash: { label: 'DASH', what: 'The same video, in the format Android and smart TVs tend to prefer.' },
		poster: { label: 'Poster', what: 'A still from the video, for the thumbnail before anyone presses play.' },
		thumbnails: { label: 'Scrub previews', what: 'The little images that appear when a viewer drags along the timeline.' }
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

	// The payoff of the whole product, stated once: what it weighed against what a
	// viewer now downloads.
	const smallest = $derived(
		asset?.renditions.filter((r) => r.state === 'ready' && r.bytes).sort((a, b) => (a.bytes ?? 0) - (b.bytes ?? 0))[0] ?? null
	);
	const saving = $derived(
		asset?.source_bytes && smallest?.bytes
			? Math.round((1 - smallest.bytes / asset.source_bytes) * 100)
			: null
	);

	// Three states per figure, not two: measured, still being measured, and never
	// going to be. The middle one is the common case on this page and the one an
	// em-dash was hiding.
	const stats = $derived.by(() => {
		if (!asset) return [];
		const probed = asset.duration_seconds != null;
		const failed = asset.state === 'failed';
		// Once an asset stops working, a missing number is missing for good. Treating
		// it as pending leaves a skeleton spinning forever on every video ingested
		// before we started recording sizes.
		const settled = ['ready', 'partially_ready', 'failed', 'live_ended'].includes(asset.state);
		const never = 'Not recorded for this video';
		return [
			{
				k: 'Length',
				v: probed ? length(asset.duration_seconds) : '',
				why: 'How long it plays for',
				pending: settled ? '' : 'Measuring it now',
				absent: failed ? 'We never got far enough to read it' : never
			},
			{
				k: 'Recorded at',
				v: asset.width && asset.height ? `${asset.width}\u00d7${asset.height}` : '',
				why: 'The size it came in at',
				pending: settled ? '' : 'Reading the file',
				absent: failed ? 'We could not read the file' : never
			},
			{
				k: 'They sent',
				v: asset.source_bytes ? size(asset.source_bytes) : '',
				why: 'What the original weighed',
				pending: settled ? '' : 'Still arriving',
				absent: failed ? 'Nothing reached us' : never
			},
			{
				k: 'Smallest we made',
				v: smallest ? size(smallest.bytes) : '',
				why: saving ? `${saving}% lighter than the original` : 'The lightest size a viewer gets',
				pending: settled ? '' : 'Nothing finished yet',
				absent: failed ? 'No size was ever made' : never
			}
		];
	});

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

{#if loading}
	<div class="mt-6 grid gap-3">
		<div class="sk h-8 w-64"></div>
		<div class="sk h-4 w-96"></div>
		<div class="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
			{#each [0, 1, 2, 3] as i (i)}<div class="sk h-28"></div>{/each}
		</div>
	</div>
{:else if error}
	<p class="mt-6 text-sm text-red" role="alert">{error}</p>
{:else if asset}
	<header class="mt-4 flex flex-wrap items-start justify-between gap-4">
		<div class="min-w-0">
			<p class="label">{STATE[asset.state]?.label ?? asset.state}</p>
			<h1 class="mt-1.5 truncate text-2xl font-semibold tracking-tight">
				{asset.duration_seconds
					? `${length(asset.duration_seconds)} of video`
					: (STATE[asset.state]?.label ?? 'Your video')}
			</h1>
			<p class="sub mt-1.5 max-w-xl">{STATE[asset.state]?.means ?? ''}</p>
		</div>
		{#if asset.playback}
			<a href="/app/studio/{asset.id}/" class="btn btn-sm flex-none">Open in Studio</a>
		{/if}
		{#if working}
			<span class="flex items-center gap-2 text-xs text-dim">
				<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-brand"></span>
				Still working — this page keeps itself up to date
			</span>
		{/if}
	</header>

	{#if asset.error_code}
		<div class="card mt-5 border border-red" role="alert">
			<p class="title text-red">This one did not work</p>
			<p class="sub mt-2">{FAILURES[asset.error_code] ?? 'Something stopped us processing it.'}</p>
			<p class="mono mt-3">Reference: {asset.error_code}</p>
		</div>
	{/if}

	<!-- What it is, before what we did to it. A figure that has not been measured yet
	     shimmers, because it is coming; one that never will says so in words. -->
	<dl class="mt-6 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
		{#each stats as stat (stat.k)}
			<div class="card">
				<p class="label">{stat.k}</p>
				{#if stat.v}
					<p class="num mt-2.5 text-[22px] leading-none">{stat.v}</p>
					<p class="sub mt-2">{stat.why}</p>
				{:else if stat.pending}
					<div class="sk mt-2.5 h-[22px] w-20"></div>
					<p class="sub mt-2">{stat.pending}</p>
				{:else}
					<p class="mt-2.5 text-[22px] leading-none text-faint">Not known</p>
					<p class="sub mt-2">{stat.absent}</p>
				{/if}
			</div>
		{/each}
	</dl>

	{#if asset.playback}
		<section class="mt-6">
			<AssetPlayer playback={asset.playback} />
		</section>
	{/if}

	<section class="mt-8">
		<h2 class="text-lg font-semibold tracking-tight">The sizes we made</h2>
		<p class="mt-1 text-sm text-dim">
			Each one is encoded in chunks, in parallel. The ones marked on demand are only made when
			somebody actually asks for that size.
		</p>

		{#if asset.renditions.length === 0 && working}
			<div class="mt-4 grid gap-3">
				{#each [0, 1, 2] as i (i)}<div class="sk h-20"></div>{/each}
			</div>
			<p class="sub mt-3">We work out which sizes to make after looking the file over.</p>
		{:else if asset.renditions.length === 0}
			<p class="sub mt-4">No sizes were made for this one.</p>
		{:else}
			<ul class="mt-4 grid gap-3">
				{#each asset.renditions as r (r.height + r.codec)}
					{@const done = pct(r)}
					<li class="card p-5">
						<div class="flex flex-wrap items-center justify-between gap-3">
							<div>
								<p class="font-semibold">
									{r.height}p
									<span class="ml-1 font-normal text-dim">{r.codec.toUpperCase()} · {mbps(r.bitrate_bps)}</span>
								</p>
								<p class="mt-0.5 text-xs text-dim">
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
							<span class="tabular-nums text-sm {r.state === 'ready' ? 'text-ink' : 'text-dim'}">
								{r.state === 'ready' ? 'Ready' : `${done}%`}
							</span>
						</div>
						<div class="mt-3 h-1.5 overflow-hidden rounded-full bg-sunk">
							<div
								class="h-full rounded-full bg-brand transition-[width] duration-500 ease-out {r.state ===
								'ready'
									? ''
									: 'opacity-60'}"
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
			<p class="mt-1 text-sm text-dim">
				Signed, good for four hours. Ask for the asset again when you need fresh ones rather
				than building them yourself.
			</p>
			<div class="card mt-4 divide-y divide-sunk">
				{#each Object.entries(asset.playback).filter((e): e is [string, string] => e[0] in LINKS) as [kind, url] (kind)}
					<div class="flex flex-wrap items-center gap-x-3 gap-y-2 p-4">
						<div class="w-40 flex-none">
							<p class="text-sm font-medium">{LINKS[kind]?.label ?? kind}</p>
							<p class="sub mt-0.5">{LINKS[kind]?.what ?? ''}</p>
						</div>
						<code class="min-w-0 flex-1 truncate font-mono text-xs text-dim">{url}</code>
						<button
							type="button"
							class="flex flex-none items-center gap-1.5 text-xs text-ink"
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
