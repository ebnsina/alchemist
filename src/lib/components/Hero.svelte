<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Shield01Icon } from '@hugeicons/core-free-icons';
	import AuroraGradient from './AuroraGradient.svelte';
	import ConversionFlow from './ConversionFlow.svelte';
</script>

<section class="relative overflow-hidden pt-32 pb-16 sm:pt-40 sm:pb-24">
	<AuroraGradient />

	<div class="mx-auto grid max-w-6xl items-center gap-12 px-4 sm:px-6 lg:grid-cols-2 lg:gap-14">
		<div class="text-center lg:text-left">
		<p
			class="inline-flex items-center gap-2 rounded-full border border-emerald/35 bg-emerald/10 px-4 py-1.5 text-xs text-ink"
		>
			<HugeiconsIcon icon={Shield01Icon} size={15} strokeWidth={1.8} class="shrink-0 text-emerald-light" />
			Encrypted the whole way, and deleted after 24 hours
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
			<li>100 free videos a month, no card</li>
			<li>Works on any phone</li>
		</ul>

		</div>

		<div class="relative">
			<!-- The right half of the hero is where the aurora has already faded out.
			     These two layers light it: a tilted plate and a soft halo. -->
			<div
				class="pointer-events-none absolute -inset-5 -rotate-3 rounded-3xl border border-emerald/15 bg-gradient-to-br from-emerald/15 via-emerald/5 to-transparent sm:-inset-7"
				aria-hidden="true"
			></div>
			<div
				class="pointer-events-none absolute -inset-16 rounded-full bg-emerald/10 blur-3xl"
				aria-hidden="true"
			></div>
			<ConversionFlow />
		</div>
	</div>
</section>
