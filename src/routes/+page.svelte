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
</script>

<Seo
	title="Alchemist — big videos in, small videos out"
	description="Drop in a video and get back one that plays anywhere, loads fast, and is up to 95% smaller. 100 free videos a month, no card needed, files deleted after 24 hours."
/>

<Hero />
<LogoTicker />
<Stats />
<Features />
<HowItWorks />
<Pricing />
<Testimonials />
<FAQ />
<CTA />
