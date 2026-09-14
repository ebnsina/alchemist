<script lang="ts">
	import { page } from '$app/state';
	import Logo from './Logo.svelte';

	let stuck = $state(false);

	// Scroll position, not an observer. The sentinel this replaced sat at top:0
	// inside the fixed header, so a negative rootMargin put it outside the
	// observer's box on the first frame and stuck latched true before the page had
	// moved — which painted the scrolled backdrop over the hero.
	$effect(() => {
		const read = () => (stuck = window.scrollY > 8);
		read();
		window.addEventListener('scroll', read, { passive: true });
		return () => window.removeEventListener('scroll', read);
	});
	const links = [
		{ href: '/#features', en: 'Features', bn: 'যা যা আছে' },
		{ href: '/#how', en: 'How it works', bn: 'কীভাবে কাজ করে' },
		{ href: '/#pricing', en: 'Pricing', bn: 'দাম' },
		{ href: '/#faq', en: 'Questions', bn: 'প্রশ্ন' }
	];
	const lang = $derived(page.url.pathname);
</script>

<!-- The nav never draws an edge. Over the hero it is fully transparent so the
     shader runs behind it; once scrolled it fades in a background that dissolves
     downward rather than ending on a line, so there is no seam in either state. -->
<header
	class="nav-shell fixed inset-x-0 top-0 z-50 transition-opacity duration-200"
	data-stuck={stuck}
>
	<!-- relative, so the nav content paints above the absolutely-positioned
	     backdrop. Without it the backdrop-filter treats the links and buttons as
	     part of what it blurs. -->
	<div
		class="relative z-10 mx-auto flex max-w-6xl flex-wrap items-center gap-3 px-4 py-3 sm:grid sm:grid-cols-[1fr_auto_1fr] sm:px-6"
	>
		<a href="/" class="flex flex-none items-center gap-2 text-lg font-bold tracking-tight">
			<Logo size={26} />
			Alchemist
		</a>

		<nav aria-label="Sections" class="order-3 w-full sm:order-none sm:w-auto sm:justify-self-center">
			<ul class="flex flex-wrap items-center justify-center gap-x-6 gap-y-1 border-t border-hairline pt-2 sm:border-0 sm:pt-0">
				{#each links as l (l.href)}
					<li>
						<a
							href={l.href}
							class="block py-1 text-sm text-muted transition-colors hover:text-ink"
							>{l.en}</a
						>
					</li>
				{/each}
			</ul>
		</nav>

		<div class="ml-auto flex flex-none items-center gap-2 sm:ml-0 sm:justify-self-end">
			<a href="/#pricing" class="btn-primary hidden text-sm sm:inline-flex">
				Start free
			</a>
		</div>
	</div>
</header>
