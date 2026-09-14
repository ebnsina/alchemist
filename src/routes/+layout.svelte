<script lang="ts">
	import '../app.css';
	import Nav from '$lib/components/Nav.svelte';
	import Footer from '$lib/components/Footer.svelte';

	let { children } = $props();
</script>

<a
	href="#main"
	class="sr-only focus:not-sr-only focus:absolute focus:top-0 focus:left-0 focus:z-[60] focus:rounded-br-lg focus:bg-emerald focus:px-4 focus:py-2 focus:font-semibold focus:text-body"
>
	Skip to content
</a>

<Nav />
<main id="main">{@render children()}</main>
<Footer />
