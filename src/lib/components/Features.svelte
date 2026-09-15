<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Video01Icon,
		Clock01Icon,
		Package01Icon,
		SquareLock01Icon,
		Database01Icon,
		GlobalIcon
	} from '@hugeicons/core-free-icons';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const features = [
		{
			icon: Video01Icon,
			ten: 'Up to 8K in, ladder out',
			ben: 'Send whatever your users hand you. We work out the sizes their viewers need and make them, so you are not maintaining an encoding profile.'
		},
		{
			icon: Clock01Icon,
			ten: 'Playable in under a minute',
			ben: 'The low sizes are made on ingest and the rest on first play, so a video is watchable long before every size exists.'
		},
		{
			icon: Package01Icon,
			ten: 'Up to 95% smaller',
			ben: 'A gigabyte becomes a few dozen megabytes. Your delivery bill drops, and so does the one your users on mobile data are paying.'
		},
		{
			icon: SquareLock01Icon,
			ten: 'Links only their viewers can use',
			ben: 'Every play runs through a signed link that expires when you say. Pass it around and it stops working, and a saved copy will not play.'
		},
		{
			icon: Database01Icon,
			ten: 'One tenant cannot see another',
			ben: 'Isolation is a database rule, not a WHERE clause somebody has to remember. A query with no filter still returns only that account.'
		},
		{
			icon: GlobalIcon,
			ten: 'One file, every device',
			ben: 'HLS and DASH out of the same source, so an old Android, a new iPhone, a laptop and a TV all get something they can play.'
		}
	];
</script>

<section id="features" class="mx-auto max-w-6xl scroll-mt-24 px-4 py-16 sm:px-6 sm:py-24">
	<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
		<p class="label text-ink">
			What you get
		</p>
		<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
			The video parts you would rather not build.
			<span class="text-ink">All of them.</span>
		</h2>
	</div>

	<div class="mt-14 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
		{#each features as f, i (f.ten)}
			<article class="card p-8" use:scrollReveal={{ delay: (i % 3) * 80 }}>
				<div class="flex items-start justify-between gap-5">
					<h3 class="text-lg font-semibold tracking-tight">{f.ten}</h3>
					<HugeiconsIcon
						icon={f.icon}
						size={42}
						strokeWidth={1.6}
						class="flex-none text-ink"
					/>
				</div>
				<p class="mt-3.5 text-sm leading-[1.75] text-dim">{f.ben}</p>
			</article>
		{/each}
	</div>
</section>
