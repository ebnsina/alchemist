<script lang="ts">
	let { value, total, caption }: { value: number; total: number; caption: string } = $props();

	const pct = $derived(total > 0 ? Math.round((value / total) * 100) : 0);

	// A 108px donut, drawn with one stroke-dasharray rather than two arcs.
	const R = 48;
	const C = 2 * Math.PI * R;
	const dash = $derived(`${(pct / 100) * C} ${C}`);
</script>

<div class="relative h-[108px] w-[108px]">
	<svg viewBox="0 0 108 108" class="h-full w-full -rotate-90" aria-hidden="true">
		<circle cx="54" cy="54" r={R} fill="none" stroke="var(--color-sunk)" stroke-width="8" />
		<circle
			cx="54"
			cy="54"
			r={R}
			fill="none"
			stroke="var(--color-ink)"
			stroke-width="8"
			stroke-linecap="round"
			stroke-dasharray={dash}
			class="transition-[stroke-dasharray] duration-500"
		/>
	</svg>
	<b class="num absolute inset-0 grid place-items-center text-2xl">{pct}%</b>
	<i class="label absolute inset-x-0 bottom-4 text-center not-italic">{caption}</i>
	<span class="vh">{value} of {total} {caption}</span>
</div>
