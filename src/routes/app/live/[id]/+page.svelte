<script lang="ts">
	import { page } from '$app/state';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Tick02Icon, Copy01Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import { setCrumbs } from '$lib/crumbs.svelte';
	import {
		getLiveStream,
		startLiveStream,
		replaceLiveKey,
		ApiError,
		type LiveStream
	} from '$lib/api';

	const id = $derived(page.params.id ?? '');

	let stream = $state<LiveStream | null>(null);
	let loading = $state(true);
	let error = $state('');
	let ingest = $state('');
	// Shown once, then gone from memory as well as from the screen.
	let freshKey = $state('');
	let replacing = $state(false);
	let confirming = $state(false);
	let copied = $state('');

	const STATE: Record<LiveStream['state'], string> = {
		idle: 'Made, never started.',
		armed: 'Waiting for your encoder.',
		live: 'On air now.',
		ended: 'Finished. The recording is under Recordings.'
	};

	const said = (e: unknown) => (e instanceof ApiError ? e.message : 'Something went wrong.');

	$effect(() => {
		if (!id) return;
		setCrumbs([{ label: 'Live', href: '/app/live/' }, { label: id.slice(0, 8) }]);
		load();
	});

	async function load() {
		loading = true;
		try {
			stream = await getLiveStream(id);
			error = '';
		} catch (e) {
			error = said(e);
		} finally {
			loading = false;
		}
	}

	async function start() {
		try {
			ingest = (await startLiveStream(id)).ingest_url;
			await load();
		} catch (e) {
			error = said(e);
		}
	}

	async function replace() {
		replacing = true;
		try {
			freshKey = (await replaceLiveKey(id)).stream_key;
			confirming = false;
			await load();
		} catch (e) {
			error = said(e);
		} finally {
			replacing = false;
		}
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
		new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(
			new Date(iso)
		);
</script>

<Seo title="Stream — Alchemist" description="One live stream and its settings." />

{#if loading && !stream}
	<div class="sk h-8 w-64"></div>
{:else if error && !stream}
	<p class="text-sm text-red" role="alert">{error}</p>
{:else if stream}
	<header>
		<h1 class="text-2xl font-semibold tracking-tight">{stream.name}</h1>
		<p class="sub mt-1">
			{stream.protocol.toUpperCase()} · made {when(stream.created_at)} · {STATE[stream.state]}
		</p>
	</header>

	{#if freshKey}
		<div class="card mt-6 p-6" in:fly={{ y: 12, duration: 340, easing: cubicOut }}>
			<p class="text-sm font-semibold">Your new stream key</p>
			<p class="mt-1 text-xs text-dim">
				Copy it now — this is the only time it is on screen. The previous key stopped working
				the moment this one was made.
			</p>
			<div class="mt-3 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
				<code class="truncate font-mono text-xs">{freshKey}</code>
				<button
					type="button"
					class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
					onclick={() => copy(freshKey, 'key')}
				>
					<HugeiconsIcon
						icon={copied === 'key' ? Tick02Icon : Copy01Icon}
						size={13}
						strokeWidth={2}
					/>
					{copied === 'key' ? 'Copied' : 'Copy'}
				</button>
			</div>
			<button type="button" class="btn btn-sm mt-4" onclick={() => (freshKey = '')}>
				I have saved it
			</button>
		</div>
	{/if}

	<section class="card mt-6 p-6">
		<h2 class="text-lg font-semibold tracking-tight">Where your encoder connects</h2>
		{#if ingest}
			<div class="mt-3 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
				<code class="truncate font-mono text-xs">{ingest}</code>
				<button
					type="button"
					class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
					onclick={() => copy(ingest, 'url')}
				>
					<HugeiconsIcon
						icon={copied === 'url' ? Tick02Icon : Copy01Icon}
						size={13}
						strokeWidth={2}
					/>
					{copied === 'url' ? 'Copied' : 'Copy'}
				</button>
			</div>
			<p class="sub mt-2">
				Replace YOUR_STREAM_KEY with your key. We cannot put it in for you — we only keep a
				fingerprint of it.
			</p>
		{:else}
			<p class="sub mt-2">
				The address appears when you start the stream. Starting holds a slot for your encoder;
				nothing goes out until one connects.
			</p>
			<button type="button" class="btn-solid mt-4" onclick={start}>Start this stream</button>
		{/if}
	</section>

	<section class="card mt-4 p-6">
		<h2 class="text-lg font-semibold tracking-tight">Stream key</h2>
		<p class="sub mt-2">
			We keep only a fingerprint of your key, so there is no way to show it to you again — not
			to us either. If you did not save it, or it has been somewhere it should not, make a new
			one.
		</p>
		{#if confirming}
			<p class="mt-4 text-sm">
				The old key stops working immediately. An encoder already sending keeps going until it
				disconnects; anything that reconnects needs the new key.
			</p>
			<div class="mt-4 flex flex-wrap gap-2">
				<button type="button" class="btn-solid" onclick={replace} disabled={replacing}>
					{replacing ? 'Making one…' : 'Yes, replace it'}
				</button>
				<button type="button" class="btn" onclick={() => (confirming = false)}>Keep it</button>
			</div>
		{:else}
			<button type="button" class="btn mt-4" onclick={() => (confirming = true)}>
				Replace the stream key
			</button>
		{/if}
	</section>

	{#if stream.asset_id}
		<section class="card mt-4 p-6">
			<h2 class="text-lg font-semibold tracking-tight">The recording</h2>
			<p class="sub mt-2">Everything sent on this stream is kept as an ordinary video.</p>
			<a href="/app/videos/{stream.asset_id}/" class="btn mt-4">Open the recording</a>
		</section>
	{/if}

	{#if error}
		<p class="mt-4 text-sm text-red" role="alert">{error}</p>
	{/if}
{/if}
