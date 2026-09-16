<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Upload01Icon,
		VideoReplayIcon,
		Timer02Icon,
		DatabaseIcon,
		EyeIcon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AssetList from '$lib/components/AssetList.svelte';
	import StatCard from '$lib/components/StatCard.svelte';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import { listAssets, usage, session, ApiError, type Asset, type UsageLine } from '$lib/api';

	let assets = $state<Asset[]>([]);
	let lines = $state<UsageLine[]>([]);
	let period = $state({ from: '', to: '' });
	let email = $state('');
	let loading = $state(true);
	let error = $state('');
	let now = $state(new Date());

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
		session()
			.then((s) => (email = s.email))
			.catch(() => {});
	});

	$effect(() => {
		const id = setInterval(() => (now = new Date()), 30000);
		return () => clearInterval(id);
	});

	$effect(() => {
		if (!busy) return;
		const id = setInterval(load, 5000);
		return () => clearInterval(id);
	});

	// Local time, not ours: the greeting is about the person reading it.
	const hour = $derived(now.getHours());
	const greeting = $derived(
		hour < 5 ? 'Still up' : hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening'
	);
	const firstName = $derived(email ? email.split('@')[0].replace(/[._-]+/g, ' ') : '');
	const stamp = $derived(
		new Intl.DateTimeFormat('en', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			hour: 'numeric',
			minute: '2-digit'
		}).format(now)
	);

	// The engine writes one usage line: ingest, in seconds. Looking up gb/gb_month
	// returned undefined and printed 0 GB, which read as measured-and-empty.
	const line = (unit: string) => lines.find((l) => l.unit === unit)?.quantity ?? 0;
	const num = (n: number, opts: Intl.NumberFormatOptions = {}) =>
		new Intl.NumberFormat('en', { maximumFractionDigits: 1, ...opts }).format(n);

	const ready = $derived(assets.filter((a) => a.state === 'ready').length);
	const working = $derived(
		assets.filter((a) => !['ready', 'failed', 'partially_ready'].includes(a.state)).length
	);
	const failed = $derived(assets.filter((a) => a.state === 'failed').length);

	const month = $derived(
		period.from
			? new Intl.DateTimeFormat('en', { month: 'long', timeZone: 'UTC' }).format(new Date(period.from))
			: 'this month'
	);

	// Uploads per day across the period, which is the only series the API gives us
	// enough to draw honestly.
	const days = $derived(
		period.from
			? Math.max(
					1,
					Math.round(
						(Math.min(Date.now(), new Date(period.to).getTime()) -
							new Date(period.from).getTime()) /
							86400000
					)
				)
			: 1
	);
	const series = $derived.by(() => {
		const start = period.from ? new Date(period.from).getTime() : Date.now();
		const buckets = new Array(days).fill(0);
		for (const a of assets) {
			const i = Math.floor((new Date(a.created_at).getTime() - start) / 86400000);
			if (i >= 0 && i < buckets.length) buckets[i] += 1;
		}
		return buckets;
	});
	const seriesLabels = $derived(
		series.map((_, i) => {
			const d = new Date(period.from || Date.now());
			d.setUTCDate(d.getUTCDate() + i);
			return new Intl.DateTimeFormat('en', { day: 'numeric', month: 'short', timeZone: 'UTC' }).format(d);
		})
	);

</script>

<Seo title="Overview — Alchemist" description="Your videos and this month's usage." />

<header class="flex flex-wrap items-end justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">
			{greeting}{firstName ? ', ' + firstName : ''}
		</h1>
		<p class="mono mt-1.5">{stamp}</p>
	</div>
	<div class="flex flex-none items-center gap-2">
		<a href="/app/upload/" class="btn-solid btn-sm">
			<HugeiconsIcon icon={Upload01Icon} size={14} strokeWidth={2} />
			Upload
		</a>
	</div>
</header>

{#if error}
	<p class="mt-6 text-sm text-red" role="alert">{error}</p>
{/if}

<div class="mt-7 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
	<StatCard
		label="Videos"
		icon={VideoReplayIcon}
		{loading}
		value={num(assets.length)}
		hint={assets.length === 0
			? 'Nothing sent yet'
			: working > 0
				? `${working} still being made${failed ? `, ${failed} failed` : ''}`
				: failed > 0
					? `${ready} ready, ${failed} failed`
					: `All ${ready} ready to play`}
	/>
	<StatCard
		label="Sent in"
		icon={Timer02Icon}
		{loading}
		accent
		value={num(line('seconds') / 60, { maximumFractionDigits: 1 }) + ' min'}
		hint="Video you gave us to process in {month}"
	/>
	<StatCard
		label="Ready to play"
		icon={EyeIcon}
		{loading}
		accent
		value={num(ready, { maximumFractionDigits: 0 })}
		hint={assets.length ? `of ${assets.length} sent` : 'Nothing sent yet'}
	/>
	<StatCard
		label="Still working"
		icon={DatabaseIcon}
		{loading}
		value={num(working, { maximumFractionDigits: 0 })}
		hint={working ? 'Small sizes finish first' : 'Nothing in the queue'}
	/>
</div>

<div class="mt-4 grid gap-4 xl:grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)]">
	<section class="card min-w-0">
		<div class="flex flex-wrap items-baseline justify-between gap-3">
			<div>
				<p class="title">Uploads through {month}</p>
				<p class="sub mt-1">One bar a day. Hover a bar for the date.</p>
			</div>
			<p class="mono">{days} days</p>
		</div>
		<div class="mt-6">
			{#if loading}
				<div class="sk h-28"></div>
			{:else}
				<Sparkline points={series} labels={seriesLabels} unit="videos" />
			{/if}
		</div>
	</section>

	<section class="card min-w-0">
		<p class="title">Where your videos are</p>
		<p class="sub mt-1">Every one you have sent us, by what it is doing.</p>
		<dl class="mt-5 grid gap-3">
			{#each [['Ready to play', ready, 'Every size made. Nothing left to wait for.'], ['Still being made', working, 'Watchable already if a small size is done.'], ['Did not work', failed, 'Something about the file stopped us.']] as [label, count, why] (label)}
				<div class="flex items-start justify-between gap-4">
					<div class="min-w-0">
						<dt class="text-sm font-medium">{label}</dt>
						<dd class="sub mt-0.5">{why}</dd>
					</div>
					<b class="num flex-none text-[17px]" class:text-accent={label === 'Ready to play'}>
						{count}
					</b>
				</div>
			{/each}
		</dl>
	</section>
</div>

<section class="mt-8">
	<div class="flex items-baseline justify-between gap-4">
		<h2 class="text-lg font-semibold tracking-tight">Your videos</h2>
		{#if assets.length}
			<p class="label">{assets.length} newest first</p>
		{/if}
	</div>
	<div class="card mt-4">
		<AssetList {assets} {loading} />
	</div>
</section>
