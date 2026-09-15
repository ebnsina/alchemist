<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import Footer from '$lib/components/Footer.svelte';

	let { children } = $props();

	// The dashboard brings its own frame, and the landing page carries its own
	// centred mark. Neither wants a header, and there is no longer one to give them.
	const chrome = $derived(!page.url.pathname.startsWith('/app'));
</script>

<a
	href="#main"
	class="sr-only focus:not-sr-only focus:absolute focus:top-0 focus:left-0 focus:z-[60] focus:bg-solid focus:px-4 focus:py-2 focus:font-bold focus:text-on-solid"
>
	Skip to content
</a>

{@render children()}

{#if chrome}
	<Footer />
{/if}
