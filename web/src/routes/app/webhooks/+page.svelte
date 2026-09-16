<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Copy01Icon,
		Tick02Icon,
		Link01Icon,
		Notification01Icon,
		CheckmarkCircle02Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon,
		PencilEdit02Icon,
		PlayIcon,
		PauseIcon,
		Delete02Icon
	} from '@hugeicons/core-free-icons';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import Seo from '$lib/Seo.svelte';
	import Check from '$lib/components/Check.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import RowMenu from '$lib/components/RowMenu.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import {
		listWebhooks,
		createWebhook,
		patchWebhook,
		deleteWebhook,
		WEBHOOK_EVENTS,
		ApiError,
		type Webhook
	} from '$lib/api';

	let rows = $state<Webhook[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');
	let url = $state('');
	let picked = $state<string[]>([...WEBHOOK_EVENTS]);
	let busy = $state(false);
	let fresh = $state<{ url: string; secret: string } | null>(null);
	let secretCard = $state<HTMLElement | null>(null);
	let step = $state(0);
	let open = $state(false);

	let page = $state(0);
	let size = $state(10);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');
	let only = $state('');

	// Each dialog owns a plain boolean: passing !!row unbound means Escape closes it
	// and the next render opens it straight back up.
	let editOpen = $state(false);
	let removeOpen = $state(false);
	let editing = $state<Webhook | null>(null);
	let removing = $state<Webhook | null>(null);
	let newUrl = $state('');
	let busyRow = $state(false);

	const FILTERS = [
		{ value: '', label: 'Every endpoint' },
		{ value: 'true', label: 'Live' },
		{ value: 'false', label: 'Paused' }
	];

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

	// One read of every control the table owns, so a change to any of them refetches
	// exactly once rather than each firing its own request.
	const query = $derived({
		limit: size,
		offset: page * size,
		q,
		sort: sorting[0]?.id ?? 'created_at',
		order: (sorting[0]?.desc ?? true ? 'desc' : 'asc') as 'asc' | 'desc',
		filters: { active: only || undefined }
	});

	async function load() {
		error = '';
		try {
			const r = await listWebhooks(query);
			rows = r.webhooks;
			total = r.total;
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

	async function saveUrl(e: SubmitEvent) {
		e.preventDefault();
		if (!editing) return;
		error = '';
		busyRow = true;
		try {
			await patchWebhook(editing.id, { url: newUrl.trim() });
			editOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busyRow = false;
		}
	}

	async function toggleActive(h: Webhook) {
		error = '';
		try {
			await patchWebhook(h.id, { active: !h.active });
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		}
	}

	async function remove() {
		if (!removing) return;
		error = '';
		busyRow = true;
		try {
			await deleteWebhook(removing.id);
			removeOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busyRow = false;
		}
	}

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
			open = false;
			// Queued so it lands after the dialog's own close puts focus back on "Add an
			// endpoint"; the secret is on screen once and it, not the button, gets the cursor.
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
	const columns: ColumnDef<any, Webhook>[] = [
		{ accessorKey: 'url', header: 'We call' },
		{
			id: 'events',
			header: 'About',
			enableSorting: false,
			cell: (c) =>
				(c.row.original as Webhook).events.map((e) => EVENTS[e]?.label ?? e).join(', ')
		},
		{
			id: 'state',
			header: 'State',
			enableSorting: false,
			cell: (c) => {
				const h = c.row.original as Webhook;
				return renderComponent(Badge, {
					label: h.active ? 'Live' : 'Paused',
					tone: h.active ? 'good' : 'idle'
				});
			}
		},
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const h = c.row.original as Webhook;
				return renderComponent(RowMenu, {
					label: `Actions for ${h.url}`,
					actions: [
						{
							label: 'Edit URL',
							icon: PencilEdit02Icon,
							onclick: () => {
								editing = h;
								newUrl = h.url;
								editOpen = true;
							}
						},
						{
							label: h.active ? 'Pause' : 'Resume',
							icon: h.active ? PauseIcon : PlayIcon,
							onclick: () => toggleActive(h)
						},
						{
							label: 'Delete',
							icon: Delete02Icon,
							danger: true,
							onclick: () => {
								removing = h;
								removeOpen = true;
							}
						}
					]
				});
			}
		}
	];
</script>

<Seo title="Webhooks — Alchemist" description="Be told when a video is ready instead of asking." />

<header class="flex flex-wrap items-start justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Webhooks</h1>
		<p class="sub mt-1 max-w-xl">
			We call you when something finishes, so your code never has to sit and poll. Every
			delivery is signed — check the signature before you trust the body.
		</p>
	</div>
	<button
		type="button"
		aria-label="Add an endpoint"
		class="btn-solid flex-none"
		onclick={() => {
			step = 0;
			open = true;
		}}
	>
		Add new
	</button>
</header>

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

<Dialog bind:open title="Add an endpoint">
	<form id="webhook-form" onsubmit={add}>
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

		</div>
		</Steps>
	</form>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			{#if step > 0 && !busy}
				<button type="button" class="btn" onclick={() => (step -= 1)}>
					<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2.2} />
					Back
				</button>
			{/if}
			<!-- The form attribute keeps this bound to a form it is no longer inside. -->
			<button type="submit" form="webhook-form" class="btn-solid" disabled={busy || !filled} aria-disabled={busy || !filled}>
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
	{/snippet}
</Dialog>

<h2 class="mt-10 text-lg font-semibold tracking-tight">Your endpoints</h2>

<div class="mt-4">
	<DataTable
		{columns}
		{rows}
		{total}
		{loading}
		bind:page
		bind:size
		bind:sorting
		bind:q
		searchLabel="Search by address"
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
			<p class="title">No endpoints here</p>
			<p class="sub mx-auto mt-2 max-w-sm">
				{#if q || only}
					No endpoint matches that. Clear the search, or pick Every endpoint above.
				{:else}
					Without one, your code has to ask us whether a video is ready — use
					<b>Add new</b> and we will tell you instead.
				{/if}
			</p>
		{/snippet}
	</DataTable>
</div>

<Dialog
	bind:open={editOpen}
	title="Where should we call?"
	hint="Changing the address does not change the signing secret."
>
	<form id="hook-url-form" onsubmit={saveUrl}>
		<label class="block">
			<span class="vh">Address</span>
			<input
				bind:value={newUrl}
				class="field"
				type="url"
				placeholder="https://your-app.example/hooks/alchemist"
				required
			/>
		</label>
		<p class="sub mt-3">
			Deliveries switch to the new address straight away. Anything already on its way to
			the old one is not sent again.
		</p>
	</form>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			<button type="button" class="btn" onclick={() => (editOpen = false)}>Cancel</button>
			<button
				type="submit"
				form="hook-url-form"
				class="btn-solid"
				disabled={busyRow || !newUrl.trim()}
			>
				{busyRow ? 'Saving…' : 'Save the address'}
			</button>
		</div>
	{/snippet}
</Dialog>

<Confirm
	bind:open={removeOpen}
	title="Delete this endpoint?"
	confirm="Yes, delete it"
	destructive
	busy={busyRow}
	onconfirm={remove}
>
	<p class="font-mono text-sm">{removing?.url ?? ''}</p>
	<p class="sub mt-2">
		We stop calling it and its signing secret is gone for good — adding the same address
		again gets a new one. Pause it instead if you only want the calls to stop for a while.
	</p>
</Confirm>
