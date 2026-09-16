<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Copy01Icon,
		Tick02Icon,
		Delete02Icon,
		PencilEdit02Icon,
		Key01Icon,
		CheckmarkCircle02Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon
	} from '@hugeicons/core-free-icons';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import Seo from '$lib/Seo.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import RowMenu from '$lib/components/RowMenu.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import { when } from '$lib/assets';
	import { listKeys, createKey, renameKey, revokeKey, ApiError, type ApiKey } from '$lib/api';
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';

	let rows = $state<ApiKey[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');
	let name = $state('');
	type Minted = { name: string; api_key: string };
	let fresh = $state<Minted | null>(null);
	let keyCard = $state<HTMLElement | null>(null);
	let copied = $state(false);
	let step = $state(0);
	let busy = $state(false);
	let open = $state(false);

	let page = $state(0);
	let size = $state(10);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');
	let only = $state('');

	// Each dialog owns a plain boolean: passing !!row unbound means Escape closes it
	// and the next render opens it straight back up.
	let renameOpen = $state(false);
	let revokeOpen = $state(false);
	let renaming = $state<ApiKey | null>(null);
	let revoking = $state<ApiKey | null>(null);
	let newName = $state('');
	let busyRow = $state(false);
	// Counted over the whole account, not this page: the last live key is still the
	// last one when it happens to be on page three.
	let liveCount = $state(0);

	// One question a screen, like every other form here. Naming a key and being told
	// what happens when it is made are two different things to take in.
	const steps = [
		{
			key: 'name',
			icon: Key01Icon,
			title: 'What is this key for?',
			hint: 'A name only you see, so you can tell your keys apart later.'
		},
		{
			key: 'make',
			icon: CheckmarkCircle02Icon,
			title: 'Ready to make it?',
			hint: 'The key is shown once, on the next screen, and never again.'
		}
	];

	const FILTERS = [
		{ value: '', label: 'Every key' },
		{ value: 'false', label: 'Live' },
		{ value: 'true', label: 'Switched off' }
	];

	// One read of every control the table owns, so a change to any of them refetches
	// exactly once rather than each firing its own request.
	const query = $derived({
		limit: size,
		offset: page * size,
		q,
		sort: sorting[0]?.id ?? 'created_at',
		order: (sorting[0]?.desc ?? true ? 'desc' : 'asc') as 'asc' | 'desc',
		filters: { revoked: only || undefined }
	});

	async function load() {
		error = '';
		try {
			const [r, live] = await Promise.all([
				listKeys(query),
				listKeys({ limit: 1, filters: { revoked: 'false' } })
			]);
			rows = r.keys;
			total = r.total;
			liveCount = live.total;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		query;
		load();
	});

	async function mint(e: SubmitEvent) {
		e.preventDefault();
		if (step === 0) {
			if (name.trim()) step = 1;
			return;
		}
		error = '';
		busy = true;
		try {
			fresh = await createKey(name.trim() || 'Untitled key');
			name = '';
			step = 0;
			open = false;
			// Queued so it lands after the dialog's own close puts focus back on "Make a key";
			// otherwise the cursor sits on that button, off a secret shown exactly once.
			queueMicrotask(() => keyCard?.focus());
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}

	async function rename(e: SubmitEvent) {
		e.preventDefault();
		if (!renaming) return;
		error = '';
		busyRow = true;
		try {
			await renameKey(renaming.id, newName.trim());
			renameOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busyRow = false;
		}
	}

	async function revoke() {
		if (!revoking) return;
		error = '';
		busyRow = true;
		try {
			await revokeKey(revoking.id);
			revokeOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busyRow = false;
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

	const sample = $derived(
		`curl -H "Authorization: Bearer ${fresh?.api_key ?? 'YOUR_KEY'}" \\\n  ${PUBLIC_ALCHEMIST_API}/v1/whoami`
	);

	const columns: ColumnDef<any, ApiKey>[] = [
		{ accessorKey: 'name', header: 'Name' },
		{ accessorKey: 'created_at', header: 'Made', cell: (c) => when(String(c.getValue())) },
		{
			id: 'state',
			header: 'State',
			enableSorting: false,
			cell: (c) => {
				const k = c.row.original as ApiKey;
				return renderComponent(Badge, {
					label: k.revoked_at ? 'Switched off' : 'Live',
					tone: k.revoked_at ? 'idle' : 'good'
				});
			}
		},
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const k = c.row.original as ApiKey;
				return renderComponent(RowMenu, {
					label: `Actions for ${k.name}`,
					actions: [
						{
							label: 'Rename',
							icon: PencilEdit02Icon,
							onclick: () => {
								renaming = k;
								newName = k.name;
								renameOpen = true;
							}
						},
						{
							label: 'Switch off',
							icon: Delete02Icon,
							danger: true,
							disabled: !!k.revoked_at || liveCount === 1,
							why: k.revoked_at
								? 'This key is already switched off'
								: 'Make another key before switching this one off',
							onclick: () => {
								revoking = k;
								revokeOpen = true;
							}
						}
					]
				});
			}
		}
	];
</script>

<Seo title="API keys — Alchemist" description="Create and revoke the keys your code uses." />

<header class="flex flex-wrap items-start justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">API keys</h1>
		<p class="mt-1 max-w-xl text-sm text-dim">
			A key is how your own code talks to us. We keep only a scrambled copy, so a key is
			shown once — if one goes missing, revoke it and make another.
		</p>
	</div>
	<button
		type="button"
		aria-label="Make a key"
		class="btn-solid flex-none"
		onclick={() => {
			step = 0;
			open = true;
		}}
	>
		Add new
	</button>
</header>

{#if fresh}
	<div
		class="card mt-6 p-6"
		tabindex="-1"
		bind:this={keyCard}
		in:fly={{ y: 12, duration: 340, easing: cubicOut }}
	>
		<p class="text-sm font-semibold">{fresh.name}</p>
		<p class="mt-1 text-xs text-dim">Copy it now. This is the only time it is on screen.</p>
		<div class="mt-3 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
			<code class="truncate font-mono text-xs">{fresh.api_key}</code>
			<button
				type="button"
				class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
				onclick={() => fresh && copy(fresh.api_key)}
			>
				<HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} size={14} strokeWidth={2} />
				{copied ? 'Copied' : 'Copy'}
			</button>
		</div>

		<p class="mt-5 text-xs text-dim">Try it:</p>
		<pre
			class="mt-2 overflow-x-auto rounded-xl border border-sunk bg-bg p-3 font-mono text-xs text-dim">{sample}</pre>
		<button type="button" class="btn mt-4" onclick={() => (fresh = null)}>
			I have saved it
		</button>
	</div>
{/if}

<Dialog bind:open title="Make a key">
	<form id="key-form" onsubmit={mint}>
		<Steps {steps} {step}>
		<div class="mt-5">
			{#if step === 0}
				<label class="block">
					<span class="vh">Name this key</span>
					<input
						bind:value={name}
						class="field"
						type="text"
						placeholder="e.g. Website"
						maxlength="60"
						required
					/>
				</label>
			{:else}
				<dl class="grid gap-3 rounded-md border border-sunk p-4">
					<div class="flex items-baseline justify-between gap-4">
						<dt class="label">Name</dt>
						<dd class="truncate text-sm font-medium">{name}</dd>
					</div>
					<div class="flex items-baseline justify-between gap-4">
						<dt class="label">Can do</dt>
						<dd class="text-sm">Everything an account can, except manage the team</dd>
					</div>
				</dl>
				<p class="sub mt-3">
					Have somewhere to paste it before you press this. We keep only a scrambled copy,
					so there is no reading it back — a lost key is replaced, not recovered.
				</p>
			{/if}

			{#if error}
				<p class="mt-3 text-sm text-red" role="alert">{error}</p>
			{/if}

		</div>
		</Steps>
	</form>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			{#if step > 0 && !busy}
				<button type="button" class="btn" onclick={() => (step = 0)}>
					<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2.2} />
					Back
				</button>
			{/if}
			<!-- The form attribute keeps this bound to a form it is no longer inside. -->
			<button
				type="submit"
				form="key-form"
				class="btn-solid"
				disabled={busy || (step === 0 && !name.trim())}
				aria-disabled={busy || (step === 0 && !name.trim())}
			>
				{#if busy}
					Making it…
				{:else if step === 0}
					Next
					<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2.2} />
				{:else}
					Make the key
				{/if}
			</button>
		</div>
	{/snippet}
</Dialog>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

<div class="mt-6">
	<DataTable
		{columns}
		{rows}
		{total}
		{loading}
		bind:page
		bind:size
		bind:sorting
		bind:q
		searchLabel="Search by name"
	>
		{#snippet toolbar()}
			<label class="flex items-center gap-2">
				<span class="vh">Show</span>
				<select
					class="select w-44"
					value={only}
					onchange={(e) => {
						only = e.currentTarget.value;
						page = 0;
					}}
				>
					{#each FILTERS as f (f.value)}<option value={f.value}>{f.label}</option>{/each}
				</select>
			</label>
		{/snippet}

		{#snippet empty()}
			<p class="title">No keys here</p>
			<p class="sub mx-auto mt-2 max-w-sm">
				{#if q || only}
					No key matches that. Clear the search, or pick Every key above.
				{:else}
					Use <b>Add new</b> to make the key your own code signs in with.
				{/if}
			</p>
		{/snippet}
	</DataTable>
</div>

<Dialog
	bind:open={renameOpen}
	title="Rename this key"
	hint="A name only you see. It does not change the key itself."
>
	<form id="key-rename-form" onsubmit={rename}>
		<label class="block">
			<span class="vh">Name</span>
			<input bind:value={newName} class="field" type="text" maxlength="60" required />
		</label>
	</form>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			<button type="button" class="btn" onclick={() => (renameOpen = false)}>Cancel</button>
			<button
				type="submit"
				form="key-rename-form"
				class="btn-solid"
				disabled={busyRow || !newName.trim()}
			>
				{busyRow ? 'Saving…' : 'Save the name'}
			</button>
		</div>
	{/snippet}
</Dialog>

<!-- Said before it happens, not after: switching a key off is instant and there is
     no putting it back. -->
<Confirm
	bind:open={revokeOpen}
	title="Switch this key off?"
	confirm="Yes, switch it off"
	destructive
	busy={busyRow}
	onconfirm={revoke}
>
	<p class="text-sm">{revoking?.name ?? ''}</p>
	<p class="sub mt-2">
		Anything still using this key stops working straight away, and it cannot be turned back
		on. Videos already uploaded with it are not affected.
	</p>
</Confirm>
