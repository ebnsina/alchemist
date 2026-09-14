<script lang="ts">
	import AuroraGradient from './AuroraGradient.svelte';
	import ConversionFlow from './ConversionFlow.svelte';
</script>

<section class="relative overflow-hidden pt-32 pb-16 sm:pt-40 sm:pb-24">
	<AuroraGradient />

	<div class="mx-auto grid max-w-6xl items-center gap-12 px-4 sm:px-6 lg:grid-cols-2 lg:gap-14">
		<div class="text-center lg:text-left">
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

		<h1 class="mt-7 text-4xl font-bold tracking-tight sm:text-5xl leading-[1.32]">
			Convert any video,
			<span class="gradient-text block"
				>into something magical.</span
			>
		</h1>

		<p class="mx-auto mt-5 max-w-xl text-base text-muted sm:text-lg lg:mx-0">
			Drop in a video and get back one that plays anywhere, loads fast, and is a fraction of the size. No settings to learn, nothing to install.
		</p>

		<div class="mt-8 flex flex-wrap justify-center gap-3 lg:justify-start">
			<a href="/#pricing" class="btn-primary">Start free</a>
			<a href="/#how" class="btn-ghost">See how it works</a>
		</div>

		<ul class="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-2 text-xs text-muted lg:justify-start">
			<li>Cancel any time</li>
			<li>Files deleted after 24 hours</li>
			<li>Works on any phone</li>
		</ul>

		</div>

		<ConversionFlow />
	</div>
</section>
