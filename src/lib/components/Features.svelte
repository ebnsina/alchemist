<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Video01Icon,
		Clock01Icon,
		Package01Icon,
		SquareLock01Icon,
		Delete02Icon,
		GlobalIcon
	} from '@hugeicons/core-free-icons';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const features = [
		{
			icon: Video01Icon,
			ten: 'Up to 8K, kept sharp',
			ben: 'Send us anything up to 8K and it comes back looking like the original. We keep the detail and throw away the weight.',
		},
		{
			icon: Clock01Icon,
			ten: 'Done in under a minute',
			ben: 'A ten-minute video is usually ready before you have finished writing the caption. Longer files take a little more, but not much.',
		},
		{
			icon: Package01Icon,
			ten: 'Up to 95% smaller',
			ben: 'A gigabyte becomes a few dozen megabytes. Your viewers stop waiting, and the ones paying for their own data stop paying so much.',
		},
		{
			icon: SquareLock01Icon,
			ten: 'Only the people you sold it to',
			ben: 'Each viewer gets a link that stops working when you decide. Pass it around and it does nothing, and a downloaded copy will not play for anyone else.',
		},
		{
			icon: Delete02Icon,
			ten: 'We keep it, not your big file',
			ben: 'Once your video is ready to watch, the huge file you sent can go. Your course keeps playing exactly as before, and nothing sits around that nobody needs.',
		},
		{
			icon: GlobalIcon,
			ten: 'Plays on anything',
			ben: 'One file that works on an old Android, a new iPhone, a laptop and a smart TV. No more sending three versions to three people.',
		}
	];
</script>

<section id="features" class="mx-auto max-w-6xl scroll-mt-24 px-4 py-16 sm:px-6 sm:py-24">
	<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
		<p class="text-xs tracking-widest text-brand-light uppercase">
			What you get
		</p>
		<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
			Video made simple.
			<span class="gradient-text">Really simple.</span>
		</h2>
	</div>

	<div class="mt-14 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
		{#each features as f, i (f.ten)}
			<article class="card card-hover p-8" use:scrollReveal={{ delay: (i % 3) * 80 }}>
				<div class="flex items-start justify-between gap-5">
					<h3 class="text-lg font-semibold tracking-tight">{f.ten}</h3>
					<HugeiconsIcon
						icon={f.icon}
						size={42}
						strokeWidth={1.6}
						class="tilt flex-none text-brand-light"
					/>
				</div>
				<p class="mt-3.5 text-sm leading-[1.75] text-muted">{f.ben}</p>
			</article>
		{/each}
	</div>
</section>
