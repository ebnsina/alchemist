<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Link01Icon } from '@hugeicons/core-free-icons';

	// One file's journey, drawn honestly: no fake footage, no invented customer. The
	// numbers are a single real example of what a lecture recording turns into.
	const source = { name: 'lecture-week-4.mov', mb: 1240, label: '4K ProRes' };

	const rungs = [
		{ height: 1080, mb: 84, note: 'Laptops and TVs', at: 1500 },
		{ height: 720, mb: 41, note: 'Phones on wifi', at: 2600 },
		{ height: 480, mb: 19, note: 'Mobile data', at: 3500 },
		{ height: 360, mb: 11, note: 'A slow connection', at: 4200 }
	];

	const CYCLE = 6400;
	let elapsed = $state(0);

	$effect(() => {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
			elapsed = CYCLE;
			return;
		}
		const id = setInterval(() => (elapsed = (elapsed + 100) % CYCLE), 100);
		return () => clearInterval(id);
	});

	const ready = $derived(rungs.filter((r) => elapsed >= r.at));
	const done = $derived(ready.length === rungs.length);

	const mb = (n: number) =>
		new Intl.NumberFormat('en', { maximumFractionDigits: 0 }).format(n) + ' MB';
	// Bars are proportional to the source, so the shrink is the shape of the thing.
	const width = (n: number) => Math.max(4, (n / source.mb) * 100);
</script>

<div class="card">
	<div>
		<div class="flex items-baseline justify-between gap-3">
			<p class="label">One file in, every size out</p>
			<p class="mono text-dim">{ready.length} of {rungs.length}</p>
		</div>
		<p class="mono mt-2 truncate text-dim">{source.name} · {source.label}</p>

		<!-- The source, at full width, is the thing everything below is measured against. -->
		<div class="mt-4">
			<div class="flex items-baseline justify-between gap-3">
				<span class="text-[13px] font-medium">What they sent</span>
				<span class="num text-[13px]">{mb(source.mb)}</span>
			</div>
			<div class="mt-1.5 h-2 w-full rounded-full bg-muted"></div>
		</div>

		<ul class="mt-5 grid gap-3">
			{#each rungs as r (r.height)}
				{@const made = elapsed >= r.at}
				<li>
					<div class="flex items-baseline justify-between gap-3">
						<span class="text-[13px]">
							<span class="num">{r.height}p</span>
							<span class="text-dim"> · {r.note}</span>
						</span>
						<span class="num text-[13px] tabular-nums {made ? 'text-accent' : 'text-faint'}">
							{made ? mb(r.mb) : '—'}
						</span>
					</div>
					<div class="mt-1.5 h-2 w-full rounded-full bg-sunk">
						<div
							class="h-full rounded-full bg-brand transition-[width] duration-700 ease-out"
							style="width: {made ? width(r.mb) : 0}%"
						></div>
					</div>
				</li>
			{/each}
		</ul>

		<div class="mt-5 flex items-center gap-2 rounded-md border border-sunk px-3 py-2.5">
			<HugeiconsIcon icon={Link01Icon} size={14} strokeWidth={2} class="flex-none text-dim" />
			<code class="truncate font-mono text-[11px]">alchemist.video/w/8fc21a</code>
			<span class="chip chip-on ml-auto flex-none">{done ? 'Ready' : 'Working'}</span>
		</div>
	</div>
</div>
