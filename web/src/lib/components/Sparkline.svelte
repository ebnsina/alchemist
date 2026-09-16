<script lang="ts">
	// A bar per day of the period. Bars rather than a line: the quantity on a given
	// day is a discrete amount, and a line between them implies a value at 3am that
	// nobody measured.
	let {
		points,
		labels,
		unit
	}: { points: number[]; labels: string[]; unit: string } = $props();

	const max = $derived(Math.max(1, ...points));
	const total = $derived(points.reduce((a, b) => a + b, 0));
	const fmt = (n: number) => new Intl.NumberFormat('en', { maximumFractionDigits: 1 }).format(n);
</script>

{#if total === 0}
	<p class="sub">Nothing yet this period. This fills in as the work happens.</p>
{:else}
	<div class="flex h-28 items-end gap-[3px]" role="img" aria-label="{fmt(total)} {unit} over {points.length} days">
		{#each points as p, i (i)}
			<div class="group relative flex-1">
				<div
					class="w-full rounded-sm bg-brand transition-[height] duration-500"
					class:opacity-25={p === 0}
					style="height: {Math.max(2, (p / max) * 112)}px"
				></div>
				<span class="bar-tip">{labels[i]} · {fmt(p)} {unit}</span>
			</div>
		{/each}
	</div>
{/if}

<style>
	.bar-tip {
		position: absolute;
		bottom: calc(100% + 6px);
		left: 50%;
		transform: translateX(-50%);
		z-index: 5;
		padding: 3px 7px;
		border-radius: var(--radius-sm);
		background: var(--color-ink);
		color: var(--color-bg);
		font-family: var(--font-mono);
		font-size: 10px;
		white-space: nowrap;
		opacity: 0;
		pointer-events: none;
		transition: opacity 120ms ease;
	}
	.group:hover .bar-tip {
		opacity: 1;
	}
</style>
