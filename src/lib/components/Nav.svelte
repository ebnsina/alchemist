<script lang="ts">
	import { page } from '$app/state';
	import Logo from './Logo.svelte';
	import ThemeToggle from './ThemeToggle.svelte';

	const links = [
		{ href: '/#features', en: 'Features' },
		{ href: '/#how', en: 'How it works' },
		{ href: '/#pricing', en: 'Pricing' },
		{ href: '/#faq', en: 'Questions' },
		{ href: '/contact/', en: 'Talk to us' }
	];

	const here = $derived(page.url.pathname + page.url.hash);
</script>

<!-- A dock rather than a header: the page is one screen, so a bar across the top
     spends the most valuable strip on navigation nobody is using yet. At the bottom
     it is in reach and out of the way, and it needs no scrolled state because it
     never sits over the content it would otherwise have to hide behind. -->
<nav class="dock" aria-label="Main">
	<a href="/" class="dock__mark" aria-label="Alchemist home">
		<Logo size={22} />
	</a>

	<span class="dock__rule" aria-hidden="true"></span>

	<ul class="dock__links">
		{#each links as l (l.href)}
			<li>
				<a href={l.href} class="dock__link" aria-current={here === l.href ? 'page' : undefined}>
					{l.en}
				</a>
			</li>
		{/each}
	</ul>

	<span class="dock__rule" aria-hidden="true"></span>

	<ThemeToggle />
	<a href="/signup/" class="btn-solid btn-sm">Get started</a>
</nav>

<style>
	.dock {
		position: fixed;
		z-index: 50;
		bottom: 16px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		align-items: center;
		gap: 10px;
		max-width: calc(100vw - 24px);
		padding: 8px 10px;
		background: var(--color-card);
		border: 1px solid var(--color-sunk);
		border-radius: 18px;
		corner-shape: squircle;
		box-shadow: var(--shadow-bulk);
	}

	.dock__mark {
		display: grid;
		place-items: center;
		height: 32px;
		width: 32px;
		border-radius: 10px;
		corner-shape: squircle;
	}
	.dock__mark:hover {
		background: var(--color-sunk);
	}

	.dock__rule {
		height: 20px;
		width: 1px;
		flex: none;
		background: var(--color-sunk);
	}

	.dock__links {
		display: flex;
		align-items: center;
		gap: 2px;
		min-width: 0;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.dock__links::-webkit-scrollbar {
		display: none;
	}

	.dock__link {
		display: block;
		padding: 6px 10px;
		border-radius: 10px;
		corner-shape: squircle;
		font-size: 13px;
		font-weight: 700;
		white-space: nowrap;
		color: var(--color-dim);
		transition:
			background 120ms ease,
			color 120ms ease;
	}
	.dock__link:hover,
	.dock__link[aria-current='page'] {
		background: var(--color-sunk);
		color: var(--color-ink);
	}

	/* On a phone the dock is the whole width and the links scroll inside it. */
	@media (max-width: 640px) {
		.dock {
			left: 12px;
			right: 12px;
			transform: none;
			max-width: none;
		}
		.dock__mark,
		.dock__rule {
			display: none;
		}
	}
</style>
