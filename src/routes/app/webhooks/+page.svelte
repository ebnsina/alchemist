<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Copy01Icon, Tick02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Check from '$lib/components/Check.svelte';
	import { listWebhooks, createWebhook, WEBHOOK_EVENTS, ApiError, type Webhook } from '$lib/api';

	let hooks = $state<Webhook[]>([]);
	let loading = $state(true);
	let error = $state('');
	let url = $state('');
	let picked = $state<string[]>([...WEBHOOK_EVENTS]);
	let busy = $state(false);
	let fresh = $state<{ url: string; secret: string } | null>(null);
	let secretCard = $state<HTMLElement | null>(null);
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

	async function add(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			const r = await createWebhook(url.trim(), picked);
			fresh = { url: r.url, secret: r.secret };
			url = '';
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

<form class="card mt-6" onsubmit={add}>
	<p class="title">Add an endpoint</p>
	<label class="mt-4 block">
		<span class="label mb-1.5 block">Where we should call</span>
		<input
			bind:value={url}
			class="field"
			type="url"
			placeholder="https://your-app.example/hooks/alchemist"
			required
		/>
	</label>
	<fieldset class="mt-5">
		<legend class="label mb-2.5">Tell me about</legend>
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
	<button type="submit" class="btn-solid mt-5" disabled={busy || !url.trim()}>
		{busy ? 'Adding…' : 'Add endpoint'}
	</button>
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
