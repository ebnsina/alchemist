<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Copy01Icon,
		Tick02Icon,
		Link01Icon,
		Notification01Icon,
		CheckmarkCircle02Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Check from '$lib/components/Check.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import { listWebhooks, createWebhook, WEBHOOK_EVENTS, ApiError, type Webhook } from '$lib/api';

	let hooks = $state<Webhook[]>([]);
	let loading = $state(true);
	let error = $state('');
	let url = $state('');
	let picked = $state<string[]>([...WEBHOOK_EVENTS]);
	let busy = $state(false);
	let fresh = $state<{ url: string; secret: string } | null>(null);
	let secretCard = $state<HTMLElement | null>(null);
	let step = $state(0);

	// An address, a choice of events and a signing secret to save are three separate
	// things to get right, so they arrive one at a time.
	const steps = [
		{
			key: 'where',
			icon: Link01Icon,
			title: 'Where should we call?',
			hint: 'An https address of yours that answers a POST.'
		},
		{
			key: 'what',
			icon: Notification01Icon,
			title: 'What should we call about?',
			hint: 'Pick at least one. Most people only need the first.'
		},
		{
			key: 'add',
			icon: CheckmarkCircle02Icon,
			title: 'Ready to add it?',
			hint: 'The signing secret is shown once, on the next screen.'
		}
	];
	let copied = $state(false);

	async function load() {
		error = '';
		try {
			hooks = (await listWebhooks()).webhooks;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	const filled = $derived([url.trim().length > 0, picked.length > 0, true][step]);

	async function add(e: SubmitEvent) {
		e.preventDefault();
		if (!filled) return;
		if (step < steps.length - 1) {
			step += 1;
			return;
		}
		error = '';
		busy = true;
		try {
			const r = await createWebhook(url.trim(), picked);
			fresh = { url: r.url, secret: r.secret };
			url = '';
			step = 0;
			// Same as a stream key and an API key: the secret gets the attention, because
			// this is the only time it is on screen.
			queueMicrotask(() => secretCard?.focus());
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}

	function toggle(ev: string) {
		picked = picked.includes(ev) ? picked.filter((p) => p !== ev) : [...picked, ev];
	}

	// The event names are ours; what they mean to somebody integrating is not obvious
	// from the name alone.
	const EVENTS: Record<string, { label: string; what: string }> = {
		'asset.ready': {
			label: 'Ready to watch',
			what: 'Every size is made. Safe to publish the link.'
		},
		'asset.failed': {
			label: 'Did not work',
			what: 'Something stopped us. Worth telling whoever uploaded it.'
		},
		'rendition.ready': {
			label: 'A size finished',
			what: 'One quality is playable. Fires several times per video.'
		}
	};

	async function copy(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			copied = false;
		}
	}
</script>

<Seo title="Webhooks — Alchemist" description="Be told when a video is ready instead of asking." />

<h1 class="text-2xl font-semibold tracking-tight">Webhooks</h1>
<p class="sub mt-1 max-w-xl">
	We call you when something finishes, so your code never has to sit and poll. Every delivery
	is signed — check the signature before you trust the body.
</p>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if fresh}
	<div class="card mt-6" tabindex="-1" bind:this={secretCard}>
		<p class="title">Signing secret for {fresh.url}</p>
		<p class="sub mt-2">
			Shown once. Deliveries carry <code>X-Alchemist-Signature: sha256=&lt;hmac&gt;</code> over the
			raw body, computed with this.
		</p>
		<div class="mt-3 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
			<code class="truncate font-mono text-xs">{fresh.secret}</code>
			<button
				type="button"
				onclick={() => copy(fresh!.secret)}
				class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
			>
				<HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} size={14} strokeWidth={2} />
				{copied ? 'Copied' : 'Copy'}
			</button>
		</div>
		<button type="button" class="btn btn-sm mt-4" onclick={() => (fresh = null)}>
			I have saved it
		</button>
	</div>
{/if}

<form class="card mt-6 max-w-2xl" onsubmit={add}>
	<Steps {steps} {step}>
		<div class="mt-5">
			{#if step === 0}
				<label class="block">
					<span class="vh">Where we should call</span>
					<input
						bind:value={url}
						class="field"
						type="url"
						placeholder="https://your-app.example/hooks/alchemist"
						required
					/>
				</label>
				<p class="sub mt-2.5">
					It has to be reachable from the public internet, and answer within a few
					seconds. We retry a call that fails.
				</p>
			{:else if step === 1}
				<fieldset>
					<legend class="vh">Tell me about</legend>
					<div class="grid gap-2.5 sm:grid-cols-3">
						{#each WEBHOOK_EVENTS as ev (ev)}
							<Check
								card
								label={EVENTS[ev].label}
								hint={EVENTS[ev].what}
								checked={picked.includes(ev)}
								onchange={() => toggle(ev)}
							/>
						{/each}
					</div>
					{#if picked.length === 0}
						<p class="sub mt-2.5">
							Pick at least one, or we will have nothing to call you about.
						</p>
					{/if}
				</fieldset>
			{:else}
				<dl class="grid gap-3 rounded-md border border-sunk p-4">
					<div class="flex items-baseline justify-between gap-4">
						<dt class="label">We call</dt>
						<dd class="truncate font-mono text-xs">{url}</dd>
					</div>
					<div class="flex items-baseline justify-between gap-4">
						<dt class="label">About</dt>
						<dd class="text-sm">{picked.map((e) => EVENTS[e].label).join(', ')}</dd>
					</div>
				</dl>
				<p class="sub mt-3">
					Every delivery is signed with a secret we show you once, on the next screen.
					Check that signature before you trust anything in the body.
				</p>
			{/if}

			{#if error}
				<p class="mt-3 text-sm text-red" role="alert">{error}</p>
			{/if}

			<div class="mt-6 flex items-center gap-2">
				{#if step > 0 && !busy}
					<button
						type="button"
						class="btn flex-none"
						onclick={() => (step -= 1)}
						aria-label="Back"
					>
						<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2.2} />
					</button>
				{/if}
				<button
					type="submit"
					class="btn-solid flex-1"
					disabled={busy || !filled}
					aria-disabled={busy || !filled}
				>
					{#if busy}
						Adding…
					{:else if step < steps.length - 1}
						Next
						<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2.2} />
					{:else}
						Add the endpoint
					{/if}
				</button>
			</div>
		</div>
	</Steps>
</form>

<h2 class="mt-10 text-lg font-semibold tracking-tight">Your endpoints</h2>
{#if loading}
	<div class="mt-4 grid gap-2">
		{#each [0, 1] as i (i)}<div class="sk h-14"></div>{/each}
	</div>
{:else if hooks.length === 0}
	<p class="sub mt-4">
		None yet. Without one, your code has to ask us whether a video is ready — add an endpoint
		and we will tell you instead.
	</p>
{:else}
	<ul class="mt-4 divide-y divide-sunk border-y border-sunk">
		{#each hooks as h (h.id)}
			<li class="flex flex-wrap items-center gap-x-4 gap-y-2 py-3.5">
				<div class="min-w-48 flex-1">
					<p class="truncate font-mono text-sm">{h.url}</p>
					<p class="mono mt-0.5">{h.events.join(' · ')}</p>
				</div>
				<span class="chip" class:chip-on={h.active}>{h.active ? 'Live' : 'Paused'}</span>
			</li>
		{/each}
	</ul>
{/if}
