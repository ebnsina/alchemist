<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import Nav from '$lib/components/Nav.svelte';
	import Footer from '$lib/components/Footer.svelte';

	let { children } = $props();

	// The dashboard brings its own frame. Without this it inherits the marketing
	// header and footer as well, and the page opens with two Alchemist logos.
	const chrome = $derived(!page.url.pathname.startsWith('/app'));
</script>

<a
	href="#main"
	class="sr-only focus:not-sr-only focus:absolute focus:top-0 focus:left-0 focus:z-[60] focus:rounded-br-lg focus:bg-brand-mid focus:px-4 focus:py-2 focus:font-semibold focus:text-body"
>
	Skip to content
</a>

{#if chrome}
	<Nav />
{/if}
<main id="main">{@render children()}</main>
{#if chrome}
	<Footer />
{/if}
