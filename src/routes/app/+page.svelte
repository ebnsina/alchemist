<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { RefreshIcon, Upload01Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AssetList from '$lib/components/AssetList.svelte';
	import Ring from '$lib/components/Ring.svelte';
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

	const ready = $derived(
		assets.filter((a) => a.state === 'ready' || a.state === 'partially_ready').length
	);

	const month = $derived(
		period.from
			? new Intl.DateTimeFormat('en', { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(
					new Date(period.from)
				)
			: ''
	);

	const UNITS: Record<string, { label: string; format: (q: number) => string }> = {
		minutes: {
			label: 'Sent in',
			format: (q) =>
				new Intl.NumberFormat('en', {
					style: 'unit',
					unit: 'minute',
					unitDisplay: 'long',
					maximumFractionDigits: 0
				}).format(q)
		},
		gb_month: {
			label: 'Held',
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
		<h1>Your videos</h1>
		{#if month}
			<p class="sub mt-1">Usage so far in {month}</p>
		{/if}
	</div>
	<div class="flex items-center gap-2">
		<button type="button" class="btn btn-sm" onclick={load} disabled={loading}>
			<HugeiconsIcon icon={RefreshIcon} size={14} strokeWidth={2} />
			Refresh
		</button>
		<a href="/app/upload/" class="btn-solid btn-sm">
			<HugeiconsIcon icon={Upload01Icon} size={14} strokeWidth={2} />
			Upload
		</a>
	</div>
</header>

{#if error}
	<p class="mt-6 text-sm font-extrabold text-red" role="alert">{error}</p>
{/if}

<!-- The list on the left, the numbers on the right: the rail is everything that is
     a figure rather than an action. -->
<div class="mt-6 grid gap-8 lg:grid-cols-[minmax(0,1fr)_240px]">
	<section class="min-w-0">
		<AssetList {assets} {loading} />
	</section>

	<aside class="lg:pt-1">
		<div class="mx-auto w-[108px]">
			<Ring value={ready} total={assets.length} caption="Ready" />
		</div>

		<div class="mt-6 grid gap-3">
			{#each Object.entries(UNITS) as [unit, spec] (unit)}
				{@const line = lines.find((l) => l.unit === unit)}
				<div class="flex items-baseline justify-between gap-3">
					<span class="label">{spec.label}</span>
					<b class="num text-[17px]">{line ? spec.format(line.quantity) : '—'}</b>
				</div>
			{/each}
			<div class="rule my-1"></div>
			<div class="flex items-baseline justify-between gap-3">
				<span class="label">Videos</span>
				<b class="num text-[17px]">{assets.length}</b>
			</div>
		</div>
	</aside>
</div>
