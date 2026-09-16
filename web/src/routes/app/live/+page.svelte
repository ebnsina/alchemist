<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Copy01Icon, Tick02Icon, Delete02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AssetPlayer from '$lib/components/AssetPlayer.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
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
	let source = $state<'camera' | 'encoder'>('camera');
	let protocol = $state<'srt' | 'rtmp'>('srt');
	// One decision per screen, even here: two fields is still two decisions, and a
	// customer who has never set up an encoder should meet them one at a time.
	let step = $state(1);
	let open = $state(false);
	// Somebody using their own camera is never asked which encoder protocol to use,
	// so the sequence is the state rather than a count -- a branch on step numbers
	// drifts the moment a step moves.
	const steps = $derived(
		source === 'camera'
			? ['name', 'source', 'confirm']
			: ['name', 'source', 'protocol', 'confirm']
	);
	const at = $derived(steps[step - 1]);
	let fresh = $state<{ name: string; stream_key: string } | null>(null);
	let keyCard = $state<HTMLElement | null>(null);
	let armed = $state<{ id: string; ingest_url: string } | null>(null);
	let watching = $state<{ id: string; name: string; asset: AssetDetail } | null>(null);
	let copied = $state('');
	let viewer = $state<HTMLElement | null>(null);
	let confirming = $state('');
	let unsold = $state(false);
	let justMade = $state('');

	const UNKNOWN = { chip: 'Unknown', means: 'We do not recognise the state this stream is in. Reload the page.' };
	const STATE: Record<LiveStream['state'], { chip: string; means: string }> = {
		idle: { chip: 'Idle', means: 'Made, never started. Open it to go on air.' },
		armed: { chip: 'Waiting', means: 'Holding a slot for your encoder. It goes on air the moment one connects.' },
		live: { chip: 'On air', means: 'Going out now. Viewers can watch it.' },
		// Nothing went out means nothing was kept, and that stream ends here too.
		ended: { chip: 'Ended', means: 'Finished. If anything went out, it is under Recordings.' }
	};

	const SOURCE: Record<LiveStream['protocol'], string> = {
		camera: 'From this browser',
		srt: 'SRT encoder',
		rtmp: 'RTMP encoder'
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
		justMade = '';
		try {
			const s = await createLiveStream(name.trim(), source === 'camera' ? 'camera' : protocol);
			// A camera stream's key is minted again when it goes on air, so there is
			// nothing here for the customer to write down.
			fresh = source === 'camera' ? null : { name: s.name, stream_key: s.stream_key };
			// A camera stream shows no key panel, so without this the form simply emptied
			// itself and the only sign anything happened was a new row further down.
			justMade = fresh ? '' : s.name;
			name = '';
			source = 'camera';
			protocol = 'srt';
			step = 1;
			open = false;
			// Queued so it lands after the dialog's own close puts focus back on "Make a
			// stream"; the key is on screen once and it, not the button, gets the cursor.
			queueMicrotask(() => keyCard?.focus());
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
			confirming = '';
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
			// The player renders below the list, which on a long one is off the screen the
			// button was pressed on.
			queueMicrotask(() => viewer?.scrollIntoView({ block: 'nearest' }));
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

<header class="flex flex-wrap items-start justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Live streams</h1>
		<p class="sub mt-1.5 max-w-xl">
			Go live from the camera in this browser, or from an encoder like OBS. Either way it goes
			out to viewers and is kept afterwards as an ordinary recording.
		</p>
	</div>
	{#if !unsold}
		<button
			type="button"
			class="btn-solid flex-none"
			onclick={() => {
				step = 1;
				open = true;
			}}
		>
			Make a stream
		</button>
	{/if}
</header>

<Dialog bind:open title="Make a stream">
	<form id="stream-form" onsubmit={make}>
		<p class="label">Step {step} of {steps.length}</p>

		{#if at === 'name'}
			<h3 class="mt-4">What is this stream for?</h3>
			<p class="sub mt-1">A name only you see, so you can tell your streams apart later.</p>
			<!-- Focused only when the customer is actually on step 1 of a fresh form, never
			     when step 1 is being re-rendered behind a stream key they must copy. -->
			<input
				bind:value={name}
				class="field mt-4"
				type="text"
				placeholder="e.g. Friday physics class"
				maxlength="60"
				required
			/>
		{:else if at === 'source'}
			<h3 class="mt-4">Where will the picture come from?</h3>
			<p class="sub mt-1">
				You can go live straight from this browser with the camera in your laptop or phone,
				or send it from software like OBS if you already use one.
			</p>
			<div class="mt-4 grid gap-2.5">
				{#each [{ id: 'camera', title: 'This browser', why: 'Your webcam and microphone. Nothing to install.' }, { id: 'encoder', title: 'An encoder', why: 'OBS, vMix or a hardware encoder you already have.' }] as o (o.id)}
					<label class="card flex cursor-pointer items-start gap-3 p-4" class:tier--lead={source === o.id}>
						<input
							type="radio"
							name="source"
							value={o.id}
							checked={source === o.id}
							onchange={() => (source = o.id as 'camera' | 'encoder')}
							class="mt-1"
						/>
						<span>
							<span class="block text-sm font-semibold">{o.title}</span>
							<span class="sub">{o.why}</span>
						</span>
					</label>
				{/each}
			</div>
		{:else if at === 'protocol'}
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
		{:else}
			<h3 class="mt-4">Ready to make it?</h3>
			<p class="sub mt-1">
				{#if source === 'camera'}
					Nothing goes on air until you open the stream and press Go live. We ask for your
					camera then, not before.
				{:else}
					The stream key is shown once, on the next screen. Nothing goes on air until you
					press Start and your encoder connects.
				{/if}
			</p>
			<dl class="mt-4 grid gap-2.5">
				<div class="flex items-baseline justify-between gap-3">
					<dt class="sub">Name</dt>
					<dd class="text-sm font-semibold">{name}</dd>
				</div>
				<div class="flex items-baseline justify-between gap-3">
					<dt class="sub">Picture comes from</dt>
					<dd class="text-sm font-semibold">
						{source === 'camera' ? 'This browser' : protocol.toUpperCase() + ' encoder'}
					</dd>
				</div>
			</dl>
			{#if error}
				<p class="mt-4 text-sm text-red" role="alert">{error}</p>
			{/if}
		{/if}
	</form>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			{#if step > 1}
				<button type="button" class="btn" onclick={() => (step -= 1)}>Back</button>
			{/if}
			{#if at === 'confirm'}
				<!-- The form attribute keeps this bound to a form it is no longer inside. -->
				<button type="submit" form="stream-form" class="btn-solid">Make a stream</button>
			{:else}
				<button
					type="button"
					class="btn-solid"
					disabled={at === 'name' && !name.trim()}
					onclick={() => (step += 1)}
				>
					Next
				</button>
			{/if}
		</div>
	{/snippet}
</Dialog>

{#if fresh}
	<!-- Focused on appearing: closing the dialog puts the cursor back on "Make a stream",
	     which is not where a key shown exactly once should leave it. -->
	<div
		class="card mt-6 p-6"
		tabindex="-1"
		bind:this={keyCard}
		in:fly={{ y: 12, duration: 340, easing: cubicOut }}
	>
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
		<!-- Where it goes, at the moment they are holding it. Finding this out later
		     means starting the stream and reading a placeholder in a URL. -->
		<p class="sub mt-3">
			When you press <b>Start</b> you get a server address containing
			<code class="font-mono">YOUR_STREAM_KEY</code>. Put this key there, and leave your
			encoder's own <b>Stream Key</b> field empty.
		</p>
		<button type="button" class="btn mt-4" onclick={() => (fresh = null)}>I have saved it</button>
	</div>
{/if}

{#if justMade}
	<p class="mt-4 text-sm" role="status">
		“{justMade}” is ready. Open it below to turn your camera on and go live — nothing goes
		out until you do.
	</p>
{/if}

{#if armed}
	<div class="card mt-6 p-6">
		<p class="text-sm font-semibold">Point your encoder here</p>
		<p class="mt-1 text-xs text-dim">
			In OBS pick Custom, paste this as the <b>Server</b> with YOUR_STREAM_KEY swapped for
			your key, and leave <b>Stream Key</b> empty — OBS joins the two with a slash, so a key
			in both fields is sent twice and nothing connects.
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
			Use <b>Make a stream</b>. The camera in this browser, or an encoder pointed at the
			address we give you. Either way the broadcast is kept as an ordinary video afterwards.
		</p>
	</div>
{:else}
	<ul class="mt-6 grid gap-3">
		{#each streams as s (s.id)}
			<li class="card flex flex-wrap items-center gap-3 p-4">
				<a href="/app/live/{s.id}/" class="min-w-0 flex-1 hover:opacity-80">
					<p class="truncate text-sm font-semibold">{s.name}</p>
					<p class="mt-0.5 text-xs text-dim">
						{SOURCE[s.protocol] ?? 'A stream'} · made {when(s.created_at)} · {(STATE[s.state] ?? UNKNOWN).means}
					</p>
				</a>
				<span class="chip {s.state === 'live' ? 'chip-on' : ''}">{(STATE[s.state] ?? UNKNOWN).chip}</span>
				{#if s.protocol === 'camera' && s.state !== 'live'}
					<!-- The camera is asked for on the stream's own page, at the moment it is
					     needed. Arming from here would hand out a key nothing is holding. -->
					<a href="/app/live/{s.id}/" class="btn btn-sm">Go live</a>
				{:else if s.state === 'idle' || s.state === 'ended'}
					<button type="button" class="btn btn-sm" onclick={() => start(s)}>Start</button>
				{:else}
					<button type="button" class="btn btn-sm" onclick={() => watch(s)}>Watch</button>
				{/if}
				<button
					type="button"
					class="flex items-center gap-1.5 text-xs text-dim transition-colors hover:text-red"
					onclick={() => (confirming = confirming === s.id ? '' : s.id)}
				>
					<HugeiconsIcon icon={Delete02Icon} size={14} strokeWidth={1.8} />
					Delete
				</button>
				{#if confirming === s.id}
					<!-- Deleting was one click and no warning, on the row above a stream key
					     that cannot be issued again. -->
					<div class="w-full rounded-md border border-sunk p-4">
						<p class="text-sm">Delete “{s.name}”?</p>
						<p class="sub mt-1.5">
							The stream and its key are gone for good, and an encoder still pointed here
							stops being accepted. Recordings of broadcasts it already made stay in your
							library.
						</p>
						<div class="mt-3 flex flex-wrap gap-2">
							<button type="button" class="btn btn-sm" onclick={() => remove(s)}>
								Yes, delete it
							</button>
							<button type="button" class="btn btn-sm" onclick={() => (confirming = '')}>
								Keep it
							</button>
						</div>
					</div>
				{/if}
			</li>
		{/each}
	</ul>
{/if}

{/if}

{#if watching}
	<section class="mt-8" bind:this={viewer}>
		<div class="flex items-baseline justify-between gap-3">
			<h2 class="text-lg font-semibold tracking-tight">{watching.name}</h2>
			<button type="button" class="btn btn-sm" onclick={() => (watching = null)}>Close</button>
		</div>
		{#if onAir && watching.asset.playback}
			<div class="mt-4">
				<AssetPlayer playback={watching.asset.playback} live={onAir} />
			</div>
		{:else}
			<p class="sub mt-4">
				Nothing is coming in yet. The picture appears here on its own, the moment your encoder
				connects to the address above.
			</p>
		{/if}
	</section>
{/if}
