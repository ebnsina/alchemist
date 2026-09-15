<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import Footer from '$lib/components/Footer.svelte';

	let { children } = $props();

	// The dashboard brings its own frame, the landing page carries its own mark, and
	// signing in is one card on an empty screen. None of them wants a footer, and
	// there is no header left to give anything.
	const bare = ['/app', '/login', '/signup', '/invite'];
	const chrome = $derived(!bare.some((p) => page.url.pathname.startsWith(p)));

	// A link to a section scrolls to it without stamping the fragment into the address
	// bar — including from another page, which navigates first and then scrolls. Focus
	// moves too, or a keyboard reader is left where it started.
	function reveal(hash: string) {
		const el = document.querySelector(hash);
		if (!el) return;
		el.scrollIntoView({ behavior: 'smooth', block: 'start' });
		if (!el.hasAttribute('tabindex')) el.setAttribute('tabindex', '-1');
		(el as HTMLElement).focus({ preventScroll: true });
	}

	async function jump(e: MouseEvent) {
		if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey)
			return;
		const a = (e.target as HTMLElement).closest?.('a[href*="#"]') as HTMLAnchorElement | null;
		if (!a || a.target === '_blank') return;

		const url = new URL(a.href);
		if (url.origin !== location.origin || !url.hash) return;

		e.preventDefault();
		if (url.pathname !== location.pathname) {
			await goto(url.pathname, { noScroll: true });
		}
		reveal(url.hash);
	}
</script>

<!-- Capture phase: SvelteKit's own router claims the click first otherwise, and
	 defaultPrevented is already true by the time this would see it. -->
<svelte:window onclickcapture={jump} />

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
