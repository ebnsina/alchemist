<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { RefreshIcon, Upload01Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AssetList from '$lib/components/AssetList.svelte';
	import { listAssets, usage, ApiError, type Asset, type UsageLine } from '$lib/api';

	let assets: Asset[] = $state([]);
	let lines: UsageLine[] = $state([]);
	let period = $state({ from: '', to: '' });
	let loading = $state(true);
	let error = $state('');

	// Anything still moving is worth another look without the customer asking.
	const busy = $derived(
		assets.some((a) => !['ready', 'failed', 'partially_ready'].includes(a.state))
	);

	async function load() {
		error = '';
		try {
			const [a, u] = await Promise.all([listAssets(), usage()]);
			assets = a.assets;
			lines = u.lines;
			period = { from: u.from, to: u.to };
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	$effect(() => {
		if (!busy) return;
		const id = setInterval(load, 5000);
		return () => clearInterval(id);
	});

	const month = $derived(
		period.from
			? new Intl.DateTimeFormat('en', { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(
					new Date(period.from)
				)
			: ''
	);

	const UNITS: Record<string, { label: string; format: (q: number) => string }> = {
		minutes: {
			label: 'Video sent in',
			format: (q) =>
				new Intl.NumberFormat('en', {
					style: 'unit',
					unit: 'minute',
					unitDisplay: 'long',
					maximumFractionDigits: 0
				}).format(q)
		},
		gb_month: {
			label: 'Held for you',
			format: (q) =>
				new Intl.NumberFormat('en', {
					style: 'unit',
					unit: 'gigabyte',
					maximumFractionDigits: 1
				}).format(q)
		},
		gb: {
			label: 'Watched',
			format: (q) =>
				new Intl.NumberFormat('en', {
					style: 'unit',
					unit: 'gigabyte',
					maximumFractionDigits: 1
				}).format(q)
		}
	};
</script>

<Seo title="Overview — Alchemist" description="Your videos and this month's usage." />

<header class="flex flex-wrap items-center justify-between gap-4">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">Overview</h1>
		{#if month}
			<p class="mt-1 text-sm text-muted">Usage so far in {month}</p>
		{/if}
	</div>
	<div class="flex items-center gap-2">
		<button type="button" class="btn-ghost" onclick={load} disabled={loading}>
			<HugeiconsIcon icon={RefreshIcon} size={15} strokeWidth={1.8} />
			Refresh
		</button>
		<a href="/app/upload/" class="btn-primary">
			<HugeiconsIcon icon={Upload01Icon} size={15} strokeWidth={1.8} />
			Upload
		</a>
	</div>
</header>

{#if error}
	<p class="mt-6 text-sm text-[#fca5a5]" role="alert">{error}</p>
{/if}

<section class="mt-6 grid gap-4 sm:grid-cols-3">
	{#each Object.entries(UNITS) as [unit, spec] (unit)}
		{@const line = lines.find((l) => l.unit === unit)}
		<div class="card p-5">
			<p class="text-xs text-muted">{spec.label}</p>
			<p class="mt-2 text-2xl font-semibold tabular-nums tracking-tight">
				{line ? spec.format(line.quantity) : '—'}
			</p>
		</div>
	{/each}
</section>

<section class="mt-8">
	<h2 class="text-lg font-semibold tracking-tight">Your videos</h2>
	<AssetList {assets} {loading} />
</section>
