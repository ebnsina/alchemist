<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import type { IconSvgElement } from '@hugeicons/svelte';
	import { Timer02Icon, DatabaseIcon, EyeIcon, InformationCircleIcon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Check from '$lib/components/Check.svelte';
	import { usage, listAssets, ApiError, type UsageLine, type Asset } from '$lib/api';

	// The API takes a period; until now the page never passed one and always showed
	// the calendar month it defaults to.
	const startOfMonth = (back = 0) => {
		const d = new Date();
		return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth() - back, 1));
	};
	const PERIODS = [
		{ id: 'this', label: 'This month', hint: 'From the 1st until now' },
		{ id: 'last', label: 'Last month', hint: 'The whole of the previous month' },
		{ id: '30', label: 'Last 30 days', hint: 'A rolling window, not a calendar month' }
	];

	let period = $state('this');
	let lines = $state<UsageLine[]>([]);
	let assets = $state<Asset[]>([]);
	let range = $state({ from: '', to: '' });
	let loading = $state(true);
	let error = $state('');

	function bounds(id: string) {
		const now = new Date();
		if (id === 'last') {
			return { from: startOfMonth(1).toISOString(), to: startOfMonth(0).toISOString() };
		}
		if (id === '30') {
			return { from: new Date(now.getTime() - 30 * 86400000).toISOString(), to: now.toISOString() };
		}
		return { from: startOfMonth(0).toISOString(), to: now.toISOString() };
	}

	$effect(() => {
		const { from, to } = bounds(period);
		loading = true;
		error = '';
		Promise.all([usage(from, to), listAssets()])
			.then(([u, a]) => {
				lines = u.lines;
				assets = a.assets;
				range = { from: u.from, to: u.to };
			})
			.catch((e) => (error = e instanceof ApiError ? e.message : 'Something went wrong.'))
			.finally(() => (loading = false));
	});

	// Keyed on what the engine actually writes. It emits one kind — `ingest`, in
	// seconds. Storage and delivery are not metered yet, so they are named below as
	// not counted rather than printed as a zero, which would read as "you used none".
	const KINDS: Record<
		string,
		{ label: string; icon: IconSvgElement; what: string; counted: string; lower: string }
	> = {
		ingest: {
			label: 'Video sent in',
			icon: Timer02Icon,
			what: 'The running time of everything you gave us to process.',
			counted:
				'Counted once, when a video finishes processing. A retry on our side is not counted again, and a video that failed is not counted at all.',
			lower: 'Trim before you send. This counts the minutes we process, not the minutes anybody watches.'
		}
	};

	const NOT_YET = [
		{
			label: 'Held for you',
			icon: DatabaseIcon,
			what: 'Storage — every size we keep, plus the master we rebuild from.'
		},
		{
			label: 'Watched',
			icon: EyeIcon,
			what: 'Delivery — what your viewers download while watching.'
		}
	];

	const total = $derived(lines.reduce((sum, l) => sum + l.quantity, 0));
	// The API reports ingest in seconds. Minutes is what a person thinks in, so it is
	// converted for display and the real unit is never guessed at.
	const qty = (n: number, unit: string) => {
		if (unit === 'seconds') {
			const mins = n / 60;
			return mins < 1
				? new Intl.NumberFormat('en', { maximumFractionDigits: 0 }).format(n) + ' sec'
				: new Intl.NumberFormat('en', { maximumFractionDigits: 1 }).format(mins) + ' min';
		}
		if (unit === 'bytes' || unit === 'gb') {
			return new Intl.NumberFormat('en', { maximumFractionDigits: 2 }).format(n) + ' GB';
		}
		return new Intl.NumberFormat('en', { maximumFractionDigits: 2 }).format(n) + ' ' + unit;
	};
	const day = (iso: string) =>
		iso ? new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeZone: 'UTC' }).format(new Date(iso)) : '—';

	// Videos created inside the window, so the quantities have something to sit against.
	const inWindow = $derived(
		assets.filter((a) => {
			const t = new Date(a.created_at).getTime();
			return t >= new Date(range.from || 0).getTime() && t <= new Date(range.to || Date.now()).getTime();
		}).length
	);
</script>

<Seo title="Usage — Alchemist" description="What you have used, and how each number is counted." />

<h1 class="text-2xl font-semibold tracking-tight">Usage</h1>
<p class="sub mt-1 max-w-xl">
	Counted as the work happens rather than worked out afterwards. Rates are not final until
	launch, so this is quantities — what you owe is not calculated here.
</p>

<fieldset class="mt-6">
	<legend class="label mb-2.5">Period</legend>
	<div class="grid gap-2.5 sm:grid-cols-3">
		{#each PERIODS as p (p.id)}
			<Check
				type="radio"
				card
				name="period"
				value={p.id}
				bind:group={period}
				label={p.label}
				hint={p.hint}
			/>
		{/each}
	</div>
</fieldset>

<p class="mono mt-4">{day(range.from)} — {day(range.to)} · {inWindow} videos sent in this window</p>

{#if error}
	<p class="mt-6 text-sm text-red" role="alert">{error}</p>
{:else if loading}
	<div class="mt-6 grid gap-3">
		{#each [0, 1, 2] as i (i)}<div class="sk h-40"></div>{/each}
	</div>
{:else if lines.length === 0}
	<div class="card mt-6 text-center">
		<p class="title">Nothing counted in this period</p>
		<p class="sub mx-auto mt-2 max-w-sm">
			Either nothing was sent, or nothing has finished processing yet. Send a video and this
			fills in as the work happens.
		</p>
		<a href="/app/upload/" class="btn-solid mt-6">Upload a video</a>
	</div>
{:else}
	<div class="mt-6 grid gap-3">
		{#each lines as l (l.kind)}
			{@const spec = KINDS[l.kind]}
			{@const share = total > 0 ? Math.round((l.quantity / total) * 100) : 0}
			<section class="card">
				<div class="flex flex-wrap items-start justify-between gap-4">
					<div class="min-w-0">
						<div class="flex items-center gap-2">
							{#if spec}
								<HugeiconsIcon icon={spec.icon} size={17} strokeWidth={1.7} class="flex-none text-faint" />
							{/if}
							<p class="title">{spec?.label ?? l.kind}</p>
						</div>
						<p class="sub mt-1.5 max-w-md">{spec?.what ?? ''}</p>
					</div>
					<p class="num flex-none text-[26px] leading-none text-accent">
						{qty(l.quantity, l.unit)}
					</p>
				</div>

				<dl class="mt-5 grid gap-4 border-t border-sunk pt-4 sm:grid-cols-2">
					<div class="flex gap-2.5">
						<HugeiconsIcon
							icon={InformationCircleIcon}
							size={15}
							strokeWidth={1.8}
							class="mt-0.5 flex-none text-faint"
						/>
						<div>
							<dt class="label">How it is counted</dt>
							<dd class="sub mt-1">{spec?.counted ?? 'Counted as the work happens.'}</dd>
						</div>
					</div>
					<div>
						<dt class="label">Bringing it down</dt>
						<dd class="sub mt-1">{spec?.lower ?? ''}</dd>
					</div>
				</dl>
			</section>
		{/each}
	</div>

	<div class="card mt-4">
		<p class="title">Not counted yet</p>
		<p class="sub mt-2">
			These will appear here once we meter them. Nothing is being billed for them in the
			meantime — an empty figure would have been easy to mistake for zero usage.
		</p>
		<dl class="mt-4 grid gap-4 sm:grid-cols-2">
			{#each NOT_YET as n (n.label)}
				<div class="flex gap-2.5">
					<HugeiconsIcon icon={n.icon} size={16} strokeWidth={1.7} class="mt-0.5 flex-none text-faint" />
					<div>
						<dt class="text-sm font-medium">{n.label}</dt>
						<dd class="sub mt-0.5">{n.what}</dd>
					</div>
				</div>
			{/each}
		</dl>
	</div>

	<p class="sub mt-6">
		Every figure here is per account, across every API key and everyone on your team.
	</p>
{/if}
