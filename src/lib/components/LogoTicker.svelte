<script lang="ts">
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

	const VISIBLE = 7;
	const CENTER = 3;
	const rows = [...kinds, ...kinds];

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

	const fade = (idx: number) => Math.max(0.18, 1 - Math.abs(idx - (i + CENTER)) * 0.3);
</script>

<section class="border-y border-hairline py-14">
	<p class="mb-8 text-center text-xs tracking-widest text-muted uppercase">
		Built for creators, teachers, and small businesses
	</p>

	<div
		class="picker-mask mx-auto overflow-hidden px-5"
		style="height: calc({VISIBLE} * var(--row))"
	>
		<ul
			class="mx-auto max-w-xs"
			style="transform: translateY(calc({-i} * var(--row))); transition: transform {animate
				? '600ms cubic-bezier(0.32, 0.72, 0.3, 1)'
				: '0ms'}"
		>
			{#each rows as kind, idx (idx)}
				{@const active = idx === i + CENTER}
				<li
					class="flex items-center justify-between rounded-lg px-3 text-sm whitespace-nowrap transition-colors duration-300"
					class:bg-brand-mid={active}
					class:text-body={active}
					class:font-semibold={active}
					class:text-muted={!active}
					style="height: var(--row); line-height: var(--row); opacity: {fade(idx)}"
					aria-hidden={idx >= kinds.length ? 'true' : undefined}
				>
					{kind}
					<svg
						class="h-4 w-4 flex-none transition-opacity duration-300"
						style="opacity: {active ? 1 : 0}"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2.2"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<path d="m5 12.5 4.5 4.5L19 7" />
					</svg>
				</li>
			{/each}
		</ul>
	</div>
</section>

<style>
	.picker-mask {
		--row: 2.5rem;
		mask-image: linear-gradient(180deg, transparent, #000 28%, #000 72%, transparent);
	}
</style>
