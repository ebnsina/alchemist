<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Scissor01Icon, ArrowRight01Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import { listAssets, listEdits, ApiError, type Asset, type Edit } from '$lib/api';

	let assets = $state<Asset[]>([]);
	let edits = $state<Edit[]>([]);
	let loading = $state(true);
	let error = $state('');

	async function load() {
		error = '';
		try {
			const [a, e] = await Promise.all([listAssets(), listEdits()]);
			assets = a.assets;
			edits = e.edits;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	// Keep asking while something is rendering, and stop the moment nothing is.
	$effect(() => {
		if (!edits.some((e) => e.state === 'queued' || e.state === 'rendering')) return;
		const t = setInterval(load, 4000);
		return () => clearInterval(t);
	});

	// Only finished videos can be edited — there is nothing to cut before the
	// mezzanine exists, and offering it would only produce a refusal.
	const editable = $derived(
		assets.filter((a) => a.state === 'ready' || a.state === 'partially_ready')
	);
	const waiting = $derived(
		assets.filter((a) => !['ready', 'partially_ready', 'failed'].includes(a.state)).length
	);

	const editCount = (id: string) => edits.filter((e) => e.source_asset_id === id).length;

	const length = (secs: number | null) => {
		if (secs == null) return '—';
		const m = Math.floor(secs / 60);
		return `${m}:${String(Math.round(secs % 60)).padStart(2, '0')}`;
	};
	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium' }).format(new Date(iso));
	// The list endpoint reports height only; the full shape lives on the asset itself.
	const shape = (a: Asset) => (a.height ? `${a.height}p` : '—');

	const STATE: Record<string, string> = {
		queued: 'Waiting to start',
		rendering: 'Making it',
		done: 'Ready',
		failed: 'Did not work'
	};
	const describe = (o: Edit['ops']) => {
		const bits: string[] = [];
		if (o.start_sec || o.end_sec) bits.push('trimmed');
		if (o.crop_w) bits.push('cropped');
		if (o.aspect) bits.push(o.aspect);
		if (o.width) bits.push(`${o.width}px`);
		if (o.mute) bits.push('muted');
		return bits.length ? bits.join(' · ') : 'a copy';
	};
</script>

<Seo title="Studio — Alchemist" description="Cut, crop and reshape your videos." />

<h1 class="text-2xl font-semibold tracking-tight">Studio</h1>
<p class="sub mt-1 max-w-xl">
	Cut a video down, crop it, or turn it upright for phones. Everything you make here is a
	<strong class="text-ink">new</strong> video — the one you started from never changes, so a link
	you have already handed out keeps working.
</p>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="mt-6 grid gap-2">
		{#each [0, 1, 2] as i (i)}<div class="sk h-16"></div>{/each}
	</div>
{:else if editable.length === 0}
	<div class="card mt-6 text-center">
		<p class="title">Nothing to edit yet</p>
		<p class="sub mx-auto mt-2 max-w-sm">
			{#if waiting > 0}
				{waiting} video{waiting === 1 ? ' is' : 's are'} still being processed. Editing opens as
				soon as one is ready.
			{:else}
				Send us a video first. Once it has finished processing you can cut it up here.
			{/if}
		</p>
		<a href="/app/upload/" class="btn-solid mt-6">Upload a video</a>
	</div>
{:else}
	<h2 class="mt-8 text-lg font-semibold tracking-tight">Pick one to work on</h2>
	<ul class="card mt-4 divide-y divide-sunk p-0">
		{#each editable as a (a.id)}
			<li>
				<a href="/app/studio/{a.id}/" class="flex items-center gap-4 px-4 py-3.5 hover:bg-sunk">
					<span class="grid h-9 w-9 flex-none place-items-center rounded-md bg-sunk text-faint">
						<HugeiconsIcon icon={Scissor01Icon} size={16} strokeWidth={1.8} />
					</span>
					<span class="min-w-0 flex-1">
						<span class="block truncate font-mono text-sm">{a.id.slice(0, 8)}</span>
						<span class="mono mt-0.5 block">
							{length(a.duration_sec)} · {shape(a)} · added {when(a.created_at)}
							{#if editCount(a.id) > 0}
								· {editCount(a.id)} made from it
							{/if}
						</span>
					</span>
					<HugeiconsIcon
						icon={ArrowRight01Icon}
						size={15}
						strokeWidth={2.2}
						class="flex-none text-faint"
					/>
				</a>
			</li>
		{/each}
	</ul>

	{#if edits.length > 0}
		<h2 class="mt-10 text-lg font-semibold tracking-tight">What you have made</h2>
		<ul class="card mt-4 divide-y divide-sunk p-0">
			{#each edits.slice(0, 10) as e (e.id)}
				<li class="flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3.5">
					<span class="min-w-0 flex-1">
						<span class="block text-sm font-medium">{describe(e.ops)}</span>
						<span class="mono mt-0.5 block">
							from {e.source_asset_id.slice(0, 8)} · {when(e.created_at)}
						</span>
					</span>
					{#if e.state === 'done' && e.output_asset_id}
						<a href="/app/videos/{e.output_asset_id}/" class="btn btn-sm flex-none">Open it</a>
					{:else}
						<span class="chip flex-none" class:text-red={e.state === 'failed'}>
							{STATE[e.state] ?? e.state}
						</span>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
{/if}
