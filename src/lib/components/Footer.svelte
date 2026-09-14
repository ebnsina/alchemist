<script lang="ts">
	import Logo from './Logo.svelte';

	const cols = [
		{
			hen: 'Product',
			links: [
				{ href: '/#features', en: 'Features' },
				{ href: '/#how', en: 'How it works' },
				{ href: '/#pricing', en: 'Pricing' }
			]
		},
		{
			hen: 'Help',
			links: [
				{ href: '/#faq', en: 'Questions' },
				{ href: '/contact/', en: 'Talk to us' },
				{ href: '/app/docs/', en: 'API reference' }
			]
		},
		{
			hen: 'Account',
			links: [
				{ href: '/signup/', en: 'Start free' },
				{ href: '/login/', en: 'Sign in' },
				{ href: '/app/', en: 'Dashboard' }
			]
		}
	];
</script>

<footer class="border-t border-hairline">
	<div class="mx-auto grid max-w-6xl gap-10 px-4 py-14 sm:px-6 lg:grid-cols-[5fr_7fr]">
		<div>
			<p class="flex items-center gap-2 text-lg font-semibold tracking-tight">
				<Logo size={24} /> Alchemist
			</p>
			<p class="mt-3 max-w-sm text-sm text-muted">
				Heavy video in, light video out. Made for creators, teachers and small businesses who would rather not think about file sizes.
			</p>
		</div>
		<nav class="grid gap-8 sm:grid-cols-3" aria-label="Footer">
			{#each cols as c (c.hen)}
				<div>
					<h2 class="text-xs font-semibold tracking-widest text-muted uppercase">
						{c.hen}
					</h2>
					<ul class="mt-4 space-y-2.5">
						{#each c.links as l (l.href)}
							<li>
								<a href={l.href} class="text-sm text-muted transition-colors hover:text-ink"
									>{l.en}</a
								>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		</nav>
	</div>
	<div
		class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-x-8 gap-y-2 border-t border-hairline px-4 py-6 text-xs text-muted sm:px-6"
	>
		<p>© 2026 Alchemist</p>
		<p>Made in Bangladesh</p>
	</div>
</footer>
