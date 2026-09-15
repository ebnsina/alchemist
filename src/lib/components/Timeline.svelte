<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { PlayIcon, PauseIcon, VolumeOffIcon, Scissor01Icon } from '@hugeicons/core-free-icons';

	// The timeline an editor actually has: a ruler, one track holding the clip, the
	// kept region drawn on it, draggable in and out handles, and a playhead you can
	// scrub. Times are seconds throughout — pixels never leave this file.
	let {
		duration,
		start = $bindable(),
		end = $bindable(),
		current = $bindable(),
		playing = false,
		muted = false,
		onplay,
		onseek
	}: {
		duration: number;
		start: number;
		end: number;
		current: number;
		playing?: boolean;
		muted?: boolean;
		onplay?: () => void;
		onseek?: (t: number) => void;
	} = $props();

	let rail: HTMLDivElement | null = $state(null);
	let dragging = $state<'start' | 'end' | 'play' | null>(null);

	const pct = (t: number) => (duration > 0 ? Math.min(100, Math.max(0, (t / duration) * 100)) : 0);

	function at(e: PointerEvent): number {
		if (!rail || duration <= 0) return 0;
		const r = rail.getBoundingClientRect();
		return Math.min(duration, Math.max(0, ((e.clientX - r.left) / r.width) * duration));
	}

	function down(e: PointerEvent, what: 'start' | 'end' | 'play') {
		dragging = what;
		if (what === 'play') onseek?.(at(e));
	}

	function move(e: PointerEvent) {
		if (!dragging) return;
		const t = at(e);
		// A quarter second of daylight, so the handles cannot cross and make a clip
		// that ends before it starts.
		if (dragging === 'start') start = Math.min(t, end - 0.25);
		else if (dragging === 'end') end = Math.max(t, start + 0.25);
		else onseek?.(t);
	}

	const clock = (t: number) => {
		const m = Math.floor(t / 60);
		const s = t % 60;
		return `${m}:${s.toFixed(1).padStart(4, '0')}`;
	};

	// A tick every few seconds, chosen so the ruler never turns into a smear.
	const ticks = $derived.by(() => {
		if (duration <= 0) return [];
		const target = 10;
		const raw = duration / target;
		const step = [1, 2, 5, 10, 15, 30, 60, 120, 300, 600].find((s) => s >= raw) ?? 600;
		const out: number[] = [];
		for (let t = 0; t <= duration; t += step) out.push(t);
		return out;
	});
</script>

<svelte:window
	onpointermove={move}
	onpointerup={() => (dragging = null)}
	onpointercancel={() => (dragging = null)}
/>

<div class="tl">
	<div class="tl__bar">
		<button type="button" class="icon-btn" onclick={() => onplay?.()} aria-label={playing ? 'Pause' : 'Play'}>
			<HugeiconsIcon icon={playing ? PauseIcon : PlayIcon} size={16} strokeWidth={2} />
		</button>
		<span class="mono tabular-nums">{clock(current)} / {clock(duration)}</span>
		{#if muted}
			<span class="chip"><HugeiconsIcon icon={VolumeOffIcon} size={11} strokeWidth={2} /></span>
		{/if}
		<span class="mono ml-auto">Keep {clock(start)} — {clock(end)} · {(end - start).toFixed(1)}s</span>
	</div>

	<div class="tl__ruler">
		{#each ticks as t (t)}
			<span
				class="tl__tick"
				style="left: {pct(t)}%; transform: translateX({pct(t) < 4 ? '0' : pct(t) > 96 ? '-100%' : '-50%'})"
			>{clock(t)}</span>
		{/each}
	</div>

	<!-- A slider is what this is: a position in the video, chosen by pointer or by
	     arrow key. The grips and buttons below cover keyboard use for the cut points. -->
	<div
		bind:this={rail}
		class="tl__track"
		role="slider"
		tabindex="0"
		aria-label="Playhead"
		aria-valuemin={0}
		aria-valuemax={duration}
		aria-valuenow={current}
		aria-valuetext={clock(current)}
		onpointerdown={(e) => down(e, 'play')}
		onkeydown={(e) => {
			const step = e.shiftKey ? 5 : 1;
			if (e.key === 'ArrowRight') onseek?.(Math.min(duration, current + step));
			else if (e.key === 'ArrowLeft') onseek?.(Math.max(0, current - step));
			else return;
			e.preventDefault();
		}}
	>
		<!-- Everything outside the kept region is dimmed, so what survives the cut is
		     the thing you see rather than something you work out. -->
		<div class="tl__cut" style="left: 0; width: {pct(start)}%"></div>
		<div class="tl__cut" style="left: {pct(end)}%; right: 0"></div>
		<div class="tl__keep" style="left: {pct(start)}%; width: {pct(end) - pct(start)}%">
			<span class="tl__label">
				<HugeiconsIcon icon={Scissor01Icon} size={11} strokeWidth={2} />
				{(end - start).toFixed(1)}s
			</span>
		</div>

		<button
			type="button"
			class="tl__grip"
			style="left: {pct(start)}%"
			onpointerdown={(e) => {
				e.stopPropagation();
				down(e, 'start');
			}}
			aria-label="Start, {clock(start)}"
		></button>
		<button
			type="button"
			class="tl__grip"
			style="left: {pct(end)}%"
			onpointerdown={(e) => {
				e.stopPropagation();
				down(e, 'end');
			}}
			aria-label="End, {clock(end)}"
		></button>

		<div class="tl__head" style="left: {pct(current)}%"></div>
	</div>

	<!-- The rail is a pointer affordance; these are how it is used without one. -->
	<div class="tl__bar tl__bar--actions">
		<button type="button" class="btn btn-sm" onclick={() => (start = current)}>Cut start here</button>
		<button type="button" class="btn btn-sm" onclick={() => (end = current)}>Cut end here</button>
		<label class="mono ml-auto flex items-center gap-2">
			Start
			<input type="number" class="field w-20 py-1 text-xs" min="0" max={Math.max(0, end - 0.25)} step="0.1" bind:value={start} />
		</label>
		<label class="mono flex items-center gap-2">
			End
			<input type="number" class="field w-20 py-1 text-xs" min={start + 0.25} max={duration} step="0.1" bind:value={end} />
		</label>
	</div>
</div>

<style>
	.tl {
		border-top: 1px solid var(--color-sunk);
		background: var(--color-card);
		padding: 10px 14px 12px;
	}
	.tl__bar {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}
	/* Off the track above it, which it was sitting flush against. */
	.tl__bar--actions {
		margin-top: 12px;
	}
	.tl__ruler {
		position: relative;
		height: 16px;
		margin-top: 8px;
	}
	.tl__tick {
		position: absolute;
		top: 0;
		/* The first and last labels would hang off the ends, so they tuck inward. */
		transform: translateX(-50%);
		font-family: var(--font-mono);
		font-size: 9px;
		color: var(--color-faint);
		white-space: nowrap;
	}
	.tl__track {
		position: relative;
		height: 56px;
		border-radius: var(--radius-sm);
		background: repeating-linear-gradient(
			90deg,
			var(--color-sunk) 0 28px,
			color-mix(in oklab, var(--color-sunk) 70%, var(--color-bg)) 28px 56px
		);
		cursor: pointer;
		touch-action: none;
		overflow: hidden;
	}
	.tl__cut {
		position: absolute;
		top: 0;
		bottom: 0;
		background: color-mix(in oklab, var(--color-bg) 72%, transparent);
	}
	.tl__keep {
		position: absolute;
		top: 0;
		bottom: 0;
		border: 2px solid var(--color-brand);
		border-radius: var(--radius-sk);
		background: color-mix(in oklab, var(--color-brand) 14%, transparent);
	}
	.tl__label {
		position: absolute;
		left: 6px;
		bottom: 4px;
		display: inline-flex;
		align-items: center;
		gap: 3px;
		padding: 1px 5px;
		border-radius: var(--radius-full);
		background: var(--color-brand);
		color: var(--color-on-brand);
		font-family: var(--font-mono);
		font-size: 9px;
	}
	.tl__grip {
		position: absolute;
		top: -2px;
		bottom: -2px;
		width: 10px;
		margin-left: -5px;
		border-radius: var(--radius-sk);
		background: var(--color-brand);
		cursor: ew-resize;
		touch-action: none;
	}
	.tl__head {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 2px;
		margin-left: -1px;
		background: var(--color-ink);
		pointer-events: none;
	}
</style>
