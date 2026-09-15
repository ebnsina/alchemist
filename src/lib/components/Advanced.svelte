<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { ArrowDown01Icon } from '@hugeicons/core-free-icons';
	import Json from '$lib/components/Json.svelte';
	import {
		getChunks,
		getActivity,
		ApiError,
		type Chunk,
		type Activity,
		type AssetDetail
	} from '$lib/api';

	let { asset, live }: { asset: AssetDetail; live: boolean } = $props();

	let open = $state(false);
	let chunks: Chunk[] = $state([]);
	let activity: Activity[] = $state([]);
	let error = $state('');
	let loaded = $state(false);

	async function load() {
		try {
			const [c, a] = await Promise.all([getChunks(asset.id), getActivity(asset.id)]);
			chunks = c.chunks;
			activity = a.activity;
			error = '';
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not load the details.';
		} finally {
			loaded = true;
		}
	}

	// Nothing is fetched until somebody opens it: this is the deep view, and most
	// visits to the page do not want it.
	$effect(() => {
		if (open && !loaded) load();
	});
	$effect(() => {
		if (!open || !live) return;
		const t = setInterval(load, 4000);
		return () => clearInterval(t);
	});

	const byRendition = $derived(
		[...new Set(chunks.map((c) => c.rendition))].map((name) => ({
			name,
			items: chunks.filter((c) => c.rendition === name)
		}))
	);

	const clock = (iso: string | null) =>
		iso ? new Intl.DateTimeFormat('en', { timeStyle: 'medium' }).format(new Date(iso)) : 'not started';

	const took = (a: Activity) =>
		a.started_at && a.finished_at
			? `${((new Date(a.finished_at).getTime() - new Date(a.started_at).getTime()) / 1000).toFixed(1)}s`
			: a.started_at
				? 'running'
				: 'waiting';

	// Queue words, said plainly. "discarded" is not a failure — it is a job that was
	// superseded, usually because the work was already done.
	const STEP: Record<string, string> = {
		transcode: 'Encode',
		package: 'Package',
		probe: 'Probe',
		mezzanine: 'Normalise',
		analyze: 'Analyse',
		jit_rendition: 'Make a size on demand',
		retention_sweep: 'Clean up',
		webhook_deliver: 'Send webhook',
		republish: 'Republish'
	};
	const STATE: Record<string, string> = {
		available: 'Queued',
		running: 'Running',
		completed: 'Done',
		retryable: 'Will retry',
		discarded: 'Given up',
		cancelled: 'Cancelled',
		scheduled: 'Scheduled',
		pending: 'Pending'
	};
</script>

<section class="mt-8 overflow-hidden rounded-md border border-sunk">
	<button
		type="button"
		class="flex w-full items-center justify-between gap-4 px-4 py-3.5 text-left transition-colors hover:bg-sunk"
		onclick={() => (open = !open)}
		aria-expanded={open}
	>
		<span class="min-w-0">
			<span class="block text-sm font-semibold">Advanced</span>
			<span class="sub mt-0.5 block">Chunks, pipeline steps, and the raw response</span>
		</span>
		<HugeiconsIcon
			icon={ArrowDown01Icon}
			size={17}
			strokeWidth={2}
			class="chevron flex-none text-dim {open ? 'rotate-180' : ''}"
		/>
	</button>

	{#if open}
		<div class="grid gap-7 border-t border-sunk bg-sunk/40 p-4 sm:p-5">
			{#if error}
				<p class="text-sm text-red" role="alert">{error}</p>
			{:else if !loaded}
				<p class="text-sm text-dim">Loading…</p>
			{:else}
				<div>
					<h3 class="text-sm font-semibold">Chunk map</h3>
					<p class="sub mt-1">
						Each square is a slice of the video, encoded on its own. Filled means the slice
						exists in storage.
					</p>
					{#if byRendition.length === 0}
						<p class="mt-3 text-sm text-dim">No chunks yet.</p>
					{:else}
						<div class="mt-3 grid gap-4">
							{#each byRendition as group (group.name)}
								{@const done = group.items.filter((c) => c.done).length}
								<div class="card p-4">
									<div class="flex items-baseline justify-between text-xs">
										<span class="font-medium">{group.name}</span>
										<span class="text-dim tabular-nums">{done} of {group.items.length}</span>
									</div>
									<div class="mt-2.5 flex flex-wrap gap-1">
										{#each group.items as c (c.index)}
											<span
												class="h-3.5 w-3.5 rounded-[3px] {c.done ? 'bg-brand' : 'bg-muted'}"
												title="#{c.index} · {c.start_sec.toFixed(1)}s to {c.end_sec.toFixed(1)}s{c.done
													? ''
													: ' · not yet'}"
											></span>
										{/each}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<div>
					<h3 class="text-sm font-semibold">What happened</h3>
					<p class="sub mt-1">
						Every step the queue ran for this video. Error text stays in our logs — what you
						can act on is the code on the video itself.
					</p>
					{#if activity.length === 0}
						<p class="mt-3 text-sm text-dim">Nothing recorded yet.</p>
					{:else}
						<ul class="card mt-3 divide-y divide-sunk p-0">
							{#each activity as a, i (a.step + a.queued_at + i)}
								<li class="grid gap-x-4 gap-y-1 px-4 py-3 sm:grid-cols-[1fr_auto]">
									<div class="flex min-w-0 items-center gap-2.5">
										<span class="truncate text-sm font-medium">{STEP[a.step] ?? a.step}</span>
										<span
											class="chip flex-none"
											class:chip-on={a.state === 'completed'}
											class:text-red={a.state === 'discarded'}
										>
											{STATE[a.state] ?? a.state}
										</span>
									</div>
									<div class="mono flex flex-wrap items-center gap-x-3 sm:justify-end">
										<span>{clock(a.started_at)}</span>
										<span>{a.finished_at ? `took ${took(a)}` : took(a)}</span>
										<span>
											try {a.attempt}{#if a.failures > 0}<span class="text-red">
													· {a.failures} failed
												</span>{/if}
										</span>
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</div>

				<div>
					<h3 class="text-sm font-semibold">Raw response</h3>
					<p class="sub mt-1">The same call your own code would make.</p>
					<div class="mt-3">
						<Json source={JSON.stringify(asset, null, 2)} label="GET /v1/assets/{asset.id}" />
					</div>
				</div>
			{/if}
		</div>
	{/if}
</section>
