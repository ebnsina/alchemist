<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import Hero from '$lib/components/Hero.svelte';
	import LogoTicker from '$lib/components/LogoTicker.svelte';
	import Stats from '$lib/components/Stats.svelte';
	import Features from '$lib/components/Features.svelte';
	import HowItWorks from '$lib/components/HowItWorks.svelte';
	import Pricing from '$lib/components/Pricing.svelte';
	import Testimonials from '$lib/components/Testimonials.svelte';
	import FAQ from '$lib/components/FAQ.svelte';
	import CTA from '$lib/components/CTA.svelte';
	import { veil } from '$lib/utils/veil';

	let stack: HTMLElement | null = $state(null);

	// The stack finds its own sections. They are components, and a second list of
	// them here would be one more thing to keep in step with this file.
	$effect(() => {
		if (!stack) return;
		return veil(stack);
	});
</script>

<Seo
	title="Alchemist — video infrastructure for your product"
	description="An API for upload, encoding and delivery. Your users send video, we make every size their viewers need, and you hand out signed links. Usage pricing, isolated accounts, no seats."
/>

<div class="stack" bind:this={stack}>
	<Hero />
	<LogoTicker />
	<Stats />
	<Features />
	<HowItWorks />
	<Pricing />
	<Testimonials />
	<FAQ />
	<CTA />
</div>
