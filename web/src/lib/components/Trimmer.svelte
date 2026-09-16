<script lang="ts">
	// In and out points on one rail. Two handles rather than two number inputs: a trim
	// is a thing you feel against the video, not a pair of numbers you calculate.
	let {
		duration,
		start = $bindable(),
		end = $bindable(),
		current = 0
	}: { duration: number; start: number; end: number; current?: number } = $props();

	let rail: HTMLDivElement | null = $state(null);
	let dragging = $state<'start' | 'end' | null>(null);

	const pct = (t: number) => (duration > 0 ? (t / duration) * 100 : 0);

	function at(e: PointerEvent): number {
		if (!rail || duration <= 0) return 0;
		const r = rail.getBoundingClientRect();
		return Math.min(duration, Math.max(0, ((e.clientX - r.left) / r.width) * duration));
	}

	function move(e: PointerEvent) {
		if (!dragging) return;
		const t = at(e);
		// A quarter second of daylight, so the handles can never cross and produce a
		// clip that ends before it starts.
		if (dragging === 'start') start = Math.min(t, end - 0.25);
		else end = Math.max(t, start + 0.25);
	}

	const clock = (t: number) => {
		const m = Math.floor(t / 60);
		const s = t % 60;
		return `${m}:${s.toFixed(1).padStart(4, '0')}`;
	};
</script>

<svelte:window
	onpointermove={move}
	onpointerup={() => (dragging = null)}
	onpointercancel={() => (dragging = null)}
/>

<div class="grid gap-2">
	<div class="flex items-baseline justify-between">
		<span class="label">Keep</span>
		<span class="mono">{clock(start)} — {clock(end)} · {(end - start).toFixed(1)}s</span>
	</div>

	<div bind:this={rail} class="rail">
		<div class="rail__keep" style="left: {pct(start)}%; width: {pct(end - start)}%"></div>
		{#if current > 0}
			<div class="rail__play" style="left: {pct(current)}%"></div>
		{/if}
		<button
			type="button"
			class="rail__grip"
			style="left: {pct(start)}%"
			onpointerdown={() => (dragging = 'start')}
			aria-label="Start, {clock(start)}"
		></button>
		<button
			type="button"
			class="rail__grip"
			style="left: {pct(end)}%"
			onpointerdown={() => (dragging = 'end')}
			aria-label="End, {clock(end)}"
		></button>
	</div>

	<!-- The rail is a pointer affordance; these are how it is used without one. -->
	<div class="flex flex-wrap gap-3">
		<label class="flex items-center gap-2">
			<span class="mono">Start</span>
			<input
				type="number"
				class="field w-24 py-1.5 text-xs"
				min="0"
				max={Math.max(0, end - 0.25)}
				step="0.1"
				bind:value={start}
			/>
		</label>
		<label class="flex items-center gap-2">
			<span class="mono">End</span>
			<input
				type="number"
				class="field w-24 py-1.5 text-xs"
				min={start + 0.25}
				max={duration}
				step="0.1"
				bind:value={end}
			/>
		</label>
	</div>
</div>

<style>
	.rail {
		position: relative;
		height: 36px;
		border-radius: var(--radius-sm);
		background: var(--color-sunk);
		touch-action: none;
	}
	.rail__keep {
		position: absolute;
		top: 0;
		bottom: 0;
		background: color-mix(in oklab, var(--color-brand) 28%, transparent);
		border-top: 2px solid var(--color-brand);
		border-bottom: 2px solid var(--color-brand);
	}
	.rail__play {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 2px;
		background: var(--color-ink);
	}
	.rail__grip {
		position: absolute;
		top: -4px;
		bottom: -4px;
		width: 12px;
		margin-left: -6px;
		border-radius: var(--radius-sk);
		background: var(--color-brand);
		border: 1px solid var(--color-on-brand);
		cursor: ew-resize;
		touch-action: none;
	}
</style>
