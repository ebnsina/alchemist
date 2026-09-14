<script lang="ts">
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';
	import AuroraGradient from './AuroraGradient.svelte';

	/* Illustrative rates, not a promise: a phone recording in, a web-ready file out. */
	const SOURCE_MBPS = 16;
	const QUALITIES = [
		{ label: '1080p', mbps: 1.12 },
		{ label: '720p', mbps: 0.7 },
		{ label: '480p', mbps: 0.4 }
	];

	let minutes = $state(10);
	let quality = $state(QUALITIES[0]);

	const bytes = (mbps: number) => (mbps * 1_000_000 * minutes * 60) / 8;
	const before = $derived(bytes(SOURCE_MBPS));
	const after = $derived(bytes(quality.mbps));
	const saved = $derived(Math.round((1 - after / before) * 100));

	const size = (n: number) =>
		n >= 1_000_000_000
			? new Intl.NumberFormat('en', {
					style: 'unit',
					unit: 'gigabyte',
					maximumFractionDigits: 1
				}).format(n / 1_000_000_000)
			: new Intl.NumberFormat('en', {
					style: 'unit',
					unit: 'megabyte',
					maximumFractionDigits: 0
				}).format(n / 1_000_000);

	const duration = $derived(
		new Intl.NumberFormat('en', { style: 'unit', unit: 'minute', unitDisplay: 'long' }).format(
			minutes
		)
	);
	const percent = $derived(new Intl.NumberFormat('en', { style: 'percent' }).format(saved / 100));
</script>

<section class="relative overflow-hidden pt-36 pb-20 sm:pt-44 sm:pb-28">
	<AuroraGradient />

	<div class="mx-auto max-w-6xl px-4 text-center sm:px-6">
		<p
			class="inline-flex items-center gap-2 rounded-lg border border-hairline bg-transparent px-3 py-1.5 text-xs text-muted"
		>
			<svg
				class="h-3.5 w-3.5 shrink-0 text-emerald-light"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="1.5"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d="M12 3v3m0 12v3m9-9h-3M6 12H3m12.4-5.4 2.1-2.1M6.5 17.5l-2.1 2.1m13 0-2.1-2.1M6.5 6.5 4.4 4.4" />
				<circle cx="12" cy="12" r="3.25" />
			</svg>
			100 free videos every month · no card needed
		</p>

		<h1 class="mx-auto mt-8 max-w-5xl text-4xl font-bold tracking-tight sm:text-6xl leading-[1.32]">
			Convert any video,
			<span class="gradient-text block"
				>into something magical.</span
			>
		</h1>

		<p class="mx-auto mt-5 max-w-xl text-base text-muted sm:text-lg">
			Drop in a video and get back one that plays anywhere, loads fast, and is a fraction of the size. No settings to learn, nothing to install.
		</p>

		<div class="mt-8 flex flex-wrap justify-center gap-3">
			<a href="/#pricing" class="btn-primary">Start free</a>
			<a href="/#how" class="btn-ghost">See how it works</a>
		</div>

		<ul class="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-2 text-xs text-muted">
			<li>Cancel any time</li>
			<li>Files deleted after 24 hours</li>
			<li>Works on any phone</li>
		</ul>

		<!-- An illustration of what the tool does, not a screenshot of an app. -->
		<figure class="card shine mx-auto mt-14 max-w-2xl p-4 text-left sm:p-6" use:scrollReveal>
			<figcaption class="mb-4 flex items-center justify-between text-xs text-muted">
				<span>Try it with your own numbers</span>
				<span>Illustration</span>
			</figcaption>
			<div class="flex items-center justify-between gap-4">
				<div>
					<p class="text-xs text-muted">Before</p>
					<p class="text-2xl font-bold tracking-tight tabular-nums sm:text-3xl">{size(before)}</p>
				</div>
				<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="flex-none text-emerald-light" aria-hidden="true">
					<path d="M4 12h15M13.5 6.5 20 12l-6.5 5.5" />
				</svg>
				<div class="text-right">
					<p class="text-xs text-muted">After</p>
					<p class="gradient-text text-2xl font-bold tracking-tight tabular-nums sm:text-3xl">
						{size(after)}
					</p>
				</div>
			</div>

			<div class="mt-4 h-1.5 overflow-hidden rounded-full bg-white/10">
				<div
					class="h-full rounded-full bg-emerald transition-[width] duration-500 ease-out"
					style="width: {100 - saved}%"
				></div>
			</div>

			<p class="mt-3 text-xs text-muted" aria-live="polite">
				{duration} of video comes out {percent} smaller, and still plays anywhere.
			</p>

			<div class="mt-5 grid gap-4 border-t border-hairline pt-5 sm:grid-cols-[1fr_auto] sm:items-end">
				<label class="block text-xs text-muted">
					<span class="flex items-center justify-between">
						How long is it? <span class="tabular-nums text-ink">{duration}</span>
					</span>
					<input
						type="range"
						min="1"
						max="120"
						step="1"
						bind:value={minutes}
						class="range mt-2 w-full"
					/>
				</label>
				<div class="flex gap-1.5" role="group" aria-label="Quality">
					{#each QUALITIES as q (q.label)}
						<button
							type="button"
							class="chip"
							aria-pressed={quality.label === q.label}
							onclick={() => (quality = q)}
						>
							{q.label}
						</button>
					{/each}
				</div>
			</div>
		</figure>
	</div>
</section>
