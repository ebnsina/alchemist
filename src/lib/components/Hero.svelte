<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Shield01Icon } from '@hugeicons/core-free-icons';
	import AuroraGradient from './AuroraGradient.svelte';
	import ConversionFlow from './ConversionFlow.svelte';
</script>

<section class="relative flex min-h-svh items-center overflow-hidden pt-32 pb-16 sm:pt-40 sm:pb-24">
	<AuroraGradient />

	<div class="mx-auto w-full grid max-w-6xl items-center gap-12 px-4 sm:px-6 lg:grid-cols-2 lg:gap-14">
		<div class="text-center lg:text-left">
		<p
			class="inline-flex items-center gap-2 rounded-full border border-brand-mid/35 bg-brand-mid/10 px-4 py-1.5 text-xs text-ink"
		>
			<HugeiconsIcon icon={Shield01Icon} size={15} strokeWidth={1.8} class="shrink-0 text-brand-light" />
			One API for upload, encode and delivery
		</p>

		<h1 class="mt-7 text-4xl font-semibold tracking-tight sm:text-5xl leading-[1.32]">
			Video infrastructure
			<span class="gradient-text block">for your product.</span>
		</h1>

		<p class="mx-auto mt-5 max-w-xl text-base text-muted sm:text-lg lg:mx-0">
			Your users upload video. We encode it, hold it, and stream it to their viewers behind links only they can hand out. You get an API and a key, not a project to run.
		</p>

		<div class="mt-8 flex flex-wrap justify-center gap-3 lg:justify-start">
			<a href="/signup/" class="btn-primary">Start free</a>
			<a href="/#how" class="btn-ghost">See how it works</a>
		</div>

		<ul class="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-2 text-xs text-muted lg:justify-start">
			<li>Free while you build</li>
			<li>Usage pricing, no seats</li>
			<li>Every account is isolated</li>
		</ul>

		</div>

		<div class="relative">
			<!-- The right half of the hero is where the aurora has already faded out.
			     These two layers light it: a tilted plate and a soft halo. -->
			<div
				class="pointer-events-none absolute -inset-5 -rotate-3 rounded-3xl border border-brand-mid/15 bg-gradient-to-br from-brand-mid/15 via-brand-mid/5 to-transparent sm:-inset-7"
				aria-hidden="true"
			></div>
			<div
				class="pointer-events-none absolute -inset-16 rounded-full bg-brand-mid/10 blur-3xl"
				aria-hidden="true"
			></div>
			<ConversionFlow />
		</div>
	</div>
</section>
