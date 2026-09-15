<script lang="ts">
	import Logo from './Logo.svelte';

	const cols = [
		{
			hen: 'Product',
			links: [
				{ href: '/#how', en: 'How it works' },
				{ href: '/#features', en: 'What you get' },
				{ href: '/#pricing', en: 'Pricing' },
				{ href: '/#faq', en: 'Questions' }
			]
		},
		{
			hen: 'Account',
			links: [
				{ href: '/signup/', en: 'Get started' },
				{ href: '/login/', en: 'Sign in' },
				{ href: '/app/docs/', en: 'API reference' },
				{ href: '/contact/', en: 'Talk to us' }
			]
		}
	];
</script>

<footer class="border-t border-sunk">
	<div class="mx-auto grid max-w-6xl gap-10 px-4 py-14 sm:px-6 lg:grid-cols-[5fr_7fr]">
		<div>
			<p class="flex items-center gap-2 text-lg font-semibold tracking-tight">
				<Logo size={24} /> Alchemist
			</p>
			<p class="mt-3 max-w-sm text-sm text-dim">
				Heavy video in, light video out. Made for creators, teachers and small businesses who would rather not think about file sizes.
			</p>
		</div>
		<nav class="grid gap-8 sm:grid-cols-2" aria-label="Footer">
			{#each cols as c (c.hen)}
				<div>
					<h2 class="text-xs font-semibold tracking-widest text-dim uppercase">
						{c.hen}
					</h2>
					<ul class="mt-4 space-y-2.5">
						{#each c.links as l (l.href)}
							<li>
								<a href={l.href} class="text-sm text-dim transition-colors hover:text-ink"
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
		class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-x-8 gap-y-2 border-t border-sunk px-4 py-6 text-xs text-dim sm:px-6"
	>
		<p>© 2026 Alchemist</p>
		<p>Made in Bangladesh</p>
	</div>
</footer>
