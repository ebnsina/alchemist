<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Tick02Icon } from '@hugeicons/core-free-icons';
	// Names the kinds of work Alchemist is for. Claims nothing about who uses it.
	const kinds = [
		'Online courses',
		'Recorded lessons',
		'Recipe videos',
		'Wedding films',
		'Podcasts',
		'Product demos',
		'Fitness classes',
		'Client work'
	];

	const VISIBLE = 9;
	const CENTER = 4;
	// Three copies, not two: the window is i..i+VISIBLE, so at the wrap index the
	// second copy alone runs out of rows and the tail of the list goes blank.
	const rows = [...kinds, ...kinds, ...kinds];

	let i = $state(0);
	let animate = $state(true);

	$effect(() => {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		const id = setInterval(() => {
			if (i === kinds.length - 1) {
				// Wrapping: land on the duplicate, then snap back without a visible rewind.
				i = kinds.length;
				setTimeout(() => {
					animate = false;
					i = 0;
					requestAnimationFrame(() => (animate = true));
				}, 600);
			} else {
				i += 1;
			}
		}, 1900);
		return () => clearInterval(id);
	});

	const fade = (idx: number) => Math.max(0.25, 1 - Math.abs(idx - (i + CENTER)) * 0.26);
</script>

<section class="screen border-y border-sunk">
	<p class="label mb-6 text-center text-dim">
		Built for creators, teachers, and small businesses
	</p>

	<div
		class="picker-mask mx-auto w-full overflow-hidden px-5"
		style="height: calc({VISIBLE} * var(--row))"
	>
		<ul
			class="mx-auto max-w-md"
			style="transform: translateY(calc({-i} * var(--row))); transition: transform {animate
				? '600ms cubic-bezier(0.32, 0.72, 0.3, 1)'
				: '0ms'}"
		>
			{#each rows as kind, idx (idx)}
				{@const active = idx === i + CENTER}
				<li
					class="flex items-center justify-between rounded-xl px-5 text-lg whitespace-nowrap transition-colors duration-300 sm:text-xl"
					class:bg-solid={active}
					class:text-on-solid={active}
					class:font-semibold={active}
					class:text-dim={!active}
					style="height: var(--row); line-height: var(--row); opacity: {fade(idx)}"
					aria-hidden={idx >= kinds.length ? 'true' : undefined}
				>
					{kind}
					<HugeiconsIcon
						icon={Tick02Icon}
						size={20}
						strokeWidth={2.2}
						class="flex-none transition-opacity duration-300"
						style="opacity: {active ? 1 : 0}"
					/>
				</li>
			{/each}
		</ul>
	</div>
</section>

<style>
	.picker-mask {
		--row: 3rem;
		mask-image: linear-gradient(180deg, transparent, #000 28%, #000 72%, transparent);
	}
	@media (min-width: 640px) {
		.picker-mask {
			--row: 3.75rem;
		}
	}
</style>
