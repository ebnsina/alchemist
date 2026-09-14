<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Copy01Icon, Tick02Icon, Delete02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import { listKeys, createKey, revokeKey, ApiError, type ApiKey } from '$lib/api';
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';

	let keys: ApiKey[] = $state([]);
	let loading = $state(true);
	let error = $state('');
	let name = $state('');
	type Minted = { name: string; api_key: string };
	let fresh = $state<Minted | null>(null);
	let copied = $state(false);

	async function load() {
		try {
			keys = (await listKeys()).keys;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	async function mint(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		try {
			fresh = await createKey(name.trim() || 'Untitled key');
			name = '';
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		}
	}

	async function revoke(k: ApiKey) {
		error = '';
		try {
			await revokeKey(k.id);
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		}
	}

	async function copy(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			copied = false;
		}
	}

	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium' }).format(new Date(iso));

	const live = $derived(keys.filter((k) => !k.revoked_at));
	const sample = $derived(
		`curl -H "Authorization: Bearer ${fresh?.api_key ?? 'YOUR_KEY'}" \\\n  ${PUBLIC_ALCHEMIST_API}/v1/whoami`
	);
</script>

<Seo title="API keys — Alchemist" description="Create and revoke the keys your code uses." />

<h1 class="text-2xl font-semibold tracking-tight">API keys</h1>
<p class="mt-1 max-w-xl text-sm text-muted">
	A key is how your own code talks to us. We keep only a scrambled copy, so a key is shown
	once — if one goes missing, revoke it and make another.
</p>

{#if fresh}
	<div class="card mt-6 p-6" in:fly={{ y: 12, duration: 340, easing: cubicOut }}>
		<p class="text-sm font-semibold">{fresh.name}</p>
		<p class="mt-1 text-xs text-muted">Copy it now. This is the only time it is on screen.</p>
		<div class="mt-3 flex items-center gap-2 rounded-xl border border-hairline bg-body px-3 py-2">
			<code class="truncate font-mono text-xs">{fresh.api_key}</code>
			<button
				type="button"
				class="ml-auto flex flex-none items-center gap-1.5 text-xs text-brand-light"
				onclick={() => fresh && copy(fresh.api_key)}
			>
				<HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} size={14} strokeWidth={2} />
				{copied ? 'Copied' : 'Copy'}
			</button>
		</div>

		<p class="mt-5 text-xs text-muted">Try it:</p>
		<pre
			class="mt-2 overflow-x-auto rounded-xl border border-hairline bg-body p-3 font-mono text-xs text-muted">{sample}</pre>
		<button type="button" class="btn-ghost mt-4" onclick={() => (fresh = null)}>
			I have saved it
		</button>
	</div>
{/if}

<form class="mt-6 flex flex-col gap-3 sm:flex-row" onsubmit={mint}>
	<label class="flex-1">
		<span class="vh">Name this key</span>
		<input bind:value={name} class="field" type="text" placeholder="What is it for? e.g. Website" maxlength="60" />
	</label>
	<button type="submit" class="btn-primary flex-none">Make a key</button>
</form>

{#if error}
	<p class="mt-4 text-sm text-[#fca5a5]" role="alert">{error}</p>
{/if}

{#if loading}
	<p class="mt-6 text-sm text-muted">Loading…</p>
{:else}
	<ul class="mt-6 grid gap-3">
		{#each keys as k (k.id)}
			<li class="card flex flex-wrap items-center gap-3 p-4">
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm font-semibold">{k.name}</p>
					<p class="mt-0.5 text-xs text-muted">
						Made {when(k.created_at)}{#if k.revoked_at} · switched off {when(k.revoked_at)}{/if}
					</p>
				</div>
				{#if k.revoked_at}
					<span class="text-xs text-muted">Off</span>
				{:else}
					<button
						type="button"
						class="flex items-center gap-1.5 text-xs text-muted transition-colors hover:text-[#fca5a5]"
						onclick={() => revoke(k)}
						disabled={live.length === 1}
						title={live.length === 1 ? 'Make another key before switching this one off' : undefined}
					>
						<HugeiconsIcon icon={Delete02Icon} size={14} strokeWidth={1.8} />
						Switch off
					</button>
				{/if}
			</li>
		{/each}
	</ul>
{/if}
