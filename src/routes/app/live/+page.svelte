<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Copy01Icon, Tick02Icon, Delete02Icon, RefreshIcon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AssetPlayer from '$lib/components/AssetPlayer.svelte';
	import {
		listLiveStreams,
		getLiveStream,
		createLiveStream,
		startLiveStream,
		deleteLiveStream,
		getAsset,
		ApiError,
		type LiveStream,
		type AssetDetail
	} from '$lib/api';

	let streams = $state<LiveStream[]>([]);
	let loading = $state(true);
	let error = $state('');
	let name = $state('');
	let protocol = $state<'srt' | 'rtmp'>('srt');
	// One decision per screen, even here: two fields is still two decisions, and a
	// customer who has never set up an encoder should meet them one at a time.
	let step = $state(1);
	const STEPS = 3;
	let fresh = $state<{ name: string; stream_key: string } | null>(null);
	let armed = $state<{ id: string; ingest_url: string } | null>(null);
	let watching = $state<{ id: string; name: string; asset: AssetDetail } | null>(null);
	let copied = $state('');
	let unsold = $state(false);

	const STATE: Record<LiveStream['state'], { chip: string; means: string }> = {
		idle: { chip: 'Idle', means: 'Made, never started. Start it to get the address for your encoder.' },
		armed: { chip: 'Waiting', means: 'Holding a slot for your encoder. It goes on air the moment one connects.' },
		live: { chip: 'On air', means: 'Going out now. Viewers can watch it.' },
		ended: { chip: 'Ended', means: 'Finished. The recording is under Recordings.' }
	};

	const said = (e: unknown) => (e instanceof ApiError ? e.message : 'Something went wrong.');

	async function load() {
		try {
			streams = (await listLiveStreams()).live_streams;
			error = '';
		} catch (e) {
			// Live is sold apart from the VOD engine, so a 403 here is a plan, not a fault.
			unsold = e instanceof ApiError && e.code === 'live_not_enabled';
			error = unsold ? '' : said(e);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	// An armed stream changes on its own, when an encoder connects rather than when
	// anybody presses anything.
	const moving = $derived(streams.some((s) => s.state === 'armed' || s.state === 'live'));
	$effect(() => {
		if (!moving) return;
		const t = setInterval(() => {
			load();
			if (watching) refreshWatched();
		}, 5000);
		return () => clearInterval(t);
	});

	async function make(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		try {
			const s = await createLiveStream(name.trim(), protocol);
			fresh = { name: s.name, stream_key: s.stream_key };
			name = '';
			protocol = 'srt';
			step = 1;
			await load();
		} catch (err) {
			error = said(err);
		}
	}

	async function start(s: LiveStream) {
		error = '';
		try {
			const out = await startLiveStream(s.id);
			armed = { id: s.id, ingest_url: out.ingest_url };
			await load();
		} catch (err) {
			error = said(err);
		}
	}

	async function remove(s: LiveStream) {
		error = '';
		try {
			await deleteLiveStream(s.id);
			if (armed?.id === s.id) armed = null;
			if (watching?.id === s.id) watching = null;
			await load();
		} catch (err) {
			error = said(err);
		}
	}

	// The list does not carry the asset, so watching costs one more call for the
	// broadcast this stream is currently going out on.
	async function watch(s: LiveStream) {
		error = '';
		try {
			const full = await getLiveStream(s.id);
			if (!full.asset_id) {
				error = 'That stream has not been started yet.';
				return;
			}
			watching = { id: s.id, name: s.name, asset: await getAsset(full.asset_id) };
		} catch (err) {
			error = said(err);
		}
	}

	async function refreshWatched() {
		if (!watching) return;
		const asset = await getAsset(watching.asset.id).catch(() => null);
		if (asset && watching) watching = { ...watching, asset };
	}

	async function copy(text: string, what: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied = what;
			setTimeout(() => (copied = ''), 2000);
		} catch {
			copied = '';
		}
	}

	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso));

	// The asset carries playback links the moment the stream is armed, but nothing is
	// published to them until an encoder connects — so the stream's own state decides.
	const onAir = $derived(streams.find((s) => s.id === watching?.id)?.state === 'live');

</script>

<Seo title="Live streams — Alchemist" description="Create a stream, point your encoder at it, and watch it go out." />

<header class="flex flex-wrap items-end justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Live streams</h1>
		<p class="sub mt-1.5 max-w-xl">
			A stream is a standing address for your encoder. Start one, paste the address into OBS, and
			what you send goes out to viewers and is kept as a recording.
		</p>
	</div>
	<button type="button" class="btn btn-sm flex-none" onclick={load} disabled={loading}>
		<HugeiconsIcon icon={RefreshIcon} size={14} strokeWidth={2} />
		Refresh
	</button>
</header>

{#if fresh}
	<div class="card mt-6 p-6" in:fly={{ y: 12, duration: 340, easing: cubicOut }}>
		<p class="text-sm font-semibold">{fresh.name}</p>
		<p class="mt-1 text-xs text-dim">
			This is the stream key. Copy it now — it is the only time it is on screen, and we cannot
			show it again.
		</p>
		<div class="mt-3 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
			<code class="truncate font-mono text-xs">{fresh.stream_key}</code>
			<button
				type="button"
				class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
				onclick={() => fresh && copy(fresh.stream_key, 'key')}
			>
				<HugeiconsIcon icon={copied === 'key' ? Tick02Icon : Copy01Icon} size={14} strokeWidth={2} />
				{copied === 'key' ? 'Copied' : 'Copy'}
			</button>
		</div>
		<button type="button" class="btn mt-4" onclick={() => (fresh = null)}>I have saved it</button>
	</div>
{/if}

{#if armed}
	<div class="card mt-6 p-6">
		<p class="text-sm font-semibold">Point your encoder here</p>
		<p class="mt-1 text-xs text-dim">
			Paste this into OBS as a custom server, with the stream key from when you made it.
		</p>
		<div class="mt-3 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
			<code class="truncate font-mono text-xs">{armed.ingest_url}</code>
			<button
				type="button"
				class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
				onclick={() => armed && copy(armed.ingest_url, 'ingest')}
			>
				<HugeiconsIcon icon={copied === 'ingest' ? Tick02Icon : Copy01Icon} size={14} strokeWidth={2} />
				{copied === 'ingest' ? 'Copied' : 'Copy'}
			</button>
		</div>
	</div>
{/if}

{#if unsold}
	<div class="card mt-6 p-6">
		<p class="title">Live is not on this plan</p>
		<p class="sub mt-2 max-w-md">
			This account has the video engine but not live streaming. They are sold separately, so
			nothing here is switched on yet.
		</p>
		<a href="/contact/" class="btn mt-5">Talk to us about live</a>
	</div>
{:else}
<form class="card mt-6" onsubmit={make}>
	<div class="flex items-baseline justify-between gap-3">
		<p class="label">Step {step} of {STEPS}</p>
		{#if step > 1}
			<button type="button" class="label hover:text-ink" onclick={() => (step -= 1)}>Back</button>
		{/if}
	</div>

	{#if step === 1}
		<h3 class="mt-4">What is this stream for?</h3>
		<p class="sub mt-1">A name only you see, so you can tell your streams apart later.</p>
		<!-- svelte-ignore a11y_autofocus -->
		<input
			bind:value={name}
			class="field mt-4"
			type="text"
			placeholder="e.g. Friday physics class"
			maxlength="60"
			autofocus
			required
		/>
		<button
			type="button"
			class="btn-solid mt-5"
			disabled={!name.trim()}
			onclick={() => (step = 2)}
		>
			Next
		</button>
	{:else if step === 2}
		<h3 class="mt-4">How will your encoder send it?</h3>
		<p class="sub mt-1">
			If you are not sure, keep SRT. It holds a picture together on a weak uplink, which
			is most uplinks.
		</p>
		<div class="mt-4 grid gap-2.5">
			{#each [{ id: 'srt', title: 'SRT', why: 'Survives a lossy connection. The right answer almost always.' }, { id: 'rtmp', title: 'RTMP', why: 'For older hardware encoders that speak nothing else.' }] as o (o.id)}
				<label class="card flex cursor-pointer items-start gap-3 p-4" class:tier--lead={protocol === o.id}>
					<input
						type="radio"
						name="protocol"
						value={o.id}
						checked={protocol === o.id}
						onchange={() => (protocol = o.id as 'srt' | 'rtmp')}
						class="mt-1"
					/>
					<span>
						<span class="block text-sm font-semibold">{o.title}</span>
						<span class="sub">{o.why}</span>
					</span>
				</label>
			{/each}
		</div>
		<button type="button" class="btn-solid mt-5" onclick={() => (step = 3)}>Next</button>
	{:else}
		<h3 class="mt-4">Ready to make it?</h3>
		<p class="sub mt-1">
			The stream key is shown once, on the next screen. Nothing goes on air until you
			press Start and your encoder connects.
		</p>
		<dl class="mt-4 grid gap-2.5">
			<div class="flex items-baseline justify-between gap-3">
				<dt class="sub">Name</dt>
				<dd class="text-sm font-semibold">{name}</dd>
			</div>
			<div class="flex items-baseline justify-between gap-3">
				<dt class="sub">Encoder sends over</dt>
				<dd class="text-sm font-semibold">{protocol.toUpperCase()}</dd>
			</div>
		</dl>
		<button type="submit" class="btn-solid mt-5">Make a stream</button>
	{/if}
</form>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="mt-6 grid gap-3">
		{#each [0, 1, 2] as i (i)}<div class="sk h-20"></div>{/each}
	</div>
{:else if streams.length === 0}
	<div class="card mt-6 py-8 text-center">
		<p class="title">No streams yet</p>
		<p class="sub mx-auto mt-2 max-w-sm">
			Make one above. You get an address and a key for your encoder, and the broadcast is kept
			as an ordinary video afterwards.
		</p>
	</div>
{:else}
	<ul class="mt-6 grid gap-3">
		{#each streams as s (s.id)}
			<li class="card flex flex-wrap items-center gap-3 p-4">
				<a href="/app/live/{s.id}/" class="min-w-0 flex-1 hover:opacity-80">
					<p class="truncate text-sm font-semibold">{s.name}</p>
					<p class="mt-0.5 text-xs text-dim">
						{s.protocol.toUpperCase()} · made {when(s.created_at)} · {STATE[s.state].means}
					</p>
				</a>
				<span class="chip {s.state === 'live' ? 'chip-on' : ''}">{STATE[s.state].chip}</span>
				{#if s.state === 'idle' || s.state === 'ended'}
					<button type="button" class="btn btn-sm" onclick={() => start(s)}>Start</button>
				{:else}
					<button type="button" class="btn btn-sm" onclick={() => watch(s)}>Watch</button>
				{/if}
				<button
					type="button"
					class="flex items-center gap-1.5 text-xs text-dim transition-colors hover:text-red"
					onclick={() => remove(s)}
				>
					<HugeiconsIcon icon={Delete02Icon} size={14} strokeWidth={1.8} />
					Delete
				</button>
			</li>
		{/each}
	</ul>
{/if}

{/if}

{#if watching}
	<section class="mt-8">
		<div class="flex items-baseline justify-between gap-3">
			<h2 class="text-lg font-semibold tracking-tight">{watching.name}</h2>
			<button type="button" class="btn btn-sm" onclick={() => (watching = null)}>Close</button>
		</div>
		{#if onAir && watching.asset.playback}
			<div class="mt-4">
				<AssetPlayer playback={watching.asset.playback} />
			</div>
		{:else}
			<p class="sub mt-4">
				Nothing is coming in yet. The picture appears here on its own, the moment your encoder
				connects to the address above.
			</p>
		{/if}
	</section>
{/if}
