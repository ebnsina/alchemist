<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Copy01Icon, Tick02Icon, Delete02Icon, RefreshIcon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Player from '$lib/components/Player.svelte';
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

	// cenc is a DASH thing; HLS cannot carry it, so the API names the URL to use.
	const url = $derived(
		watching?.asset.playback
			? watching.asset.playback.preferred === 'dash'
				? watching.asset.playback.dash
				: watching.asset.playback.hls
			: ''
	);
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
<form class="mt-6 flex flex-col gap-3 sm:flex-row" onsubmit={make}>
	<label class="flex-1">
		<span class="vh">Name this stream</span>
		<input
			bind:value={name}
			class="field"
			type="text"
			placeholder="What is it for? e.g. Friday class"
			maxlength="60"
			required
		/>
	</label>
	<label class="flex-none">
		<span class="vh">How your encoder sends it</span>
		<select bind:value={protocol} class="select">
			<option value="srt">SRT</option>
			<option value="rtmp">RTMP</option>
		</select>
	</label>
	<button type="submit" class="btn-solid flex-none">Make a stream</button>
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
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm font-semibold">{s.name}</p>
					<p class="mt-0.5 text-xs text-dim">
						{s.protocol.toUpperCase()} · made {when(s.created_at)} · {STATE[s.state].means}
					</p>
				</div>
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
		{#if onAir && url}
			<div class="mt-4">
				<Player hls={url} poster={watching.asset.playback?.poster} />
			</div>
		{:else}
			<p class="sub mt-4">
				Nothing is coming in yet. The picture appears here on its own, the moment your encoder
				connects to the address above.
			</p>
		{/if}
	</section>
{/if}
