<script lang="ts">
	import { reveal } from '$lib/utils/reveal';

	// The reference used customer counts, video counts and a star rating here. These
	// are facts about the product instead, which is both true and more persuasive.
	const stats = [
		{ value: '93%', en: 'Smaller, on average' },
		{ value: 'Under 60s', en: 'To first playable size' },
		{ value: '8K', en: 'Highest resolution in' },
		{ value: 'Per tenant', en: 'Isolated by the database' }
	];
</script>

<section class="screen px-4 sm:px-6" use:reveal>
	<dl class="grid grid-cols-2 gap-x-6 gap-y-10 text-center md:grid-cols-4">
		{#each stats as s, i (s.value)}
			<div>
				<dt class="vh">{s.en}</dt>
				<dd class="text-ink text-3xl font-semibold tracking-tight sm:text-4xl">
					{s.value}
				</dd>
				<p class="mt-2 text-sm text-dim">{s.en}</p>
			</div>
		{/each}
	</dl>
</section>
