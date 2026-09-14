<script lang="ts">
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const features = [
		{
			d: 'M4.5 6.5h11a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2h-11a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2zM17.5 10.5l4-2.5v8l-4-2.5z',
			ten: 'Up to 8K, kept sharp',
			ben: 'Send us anything up to 8K and it comes back looking like the original. We keep the detail and throw away the weight.',
		},
		{
			d: 'M12 2.8a9.2 9.2 0 1 0 0 18.4 9.2 9.2 0 0 0 0-18.4zM12 6.8V12l3.4 2',
			ten: 'Done in under a minute',
			ben: 'A ten-minute video is usually ready before you have finished writing the caption. Longer files take a little more, but not much.',
		},
		{
			d: 'M20.5 8.5 12 3 3.5 8.5v7L12 21l8.5-5.5zM12 12l8.5-3.5M12 12v9M12 12 3.5 8.5',
			ten: 'Up to 95% smaller',
			ben: 'A gigabyte becomes a few dozen megabytes. Your viewers stop waiting, and the ones paying for their own data stop paying so much.',
		},
		{
			d: 'M6.6 10.4h10.8a1 1 0 0 1 1 1v8.1a1 1 0 0 1-1 1H6.6a1 1 0 0 1-1-1v-8.1a1 1 0 0 1 1-1zM8.6 10.4V7a3.4 3.4 0 1 1 6.8 0v3.4',
			ten: 'Only the people you sold it to',
			ben: 'Each viewer gets a link that stops working when you decide. Pass it around and it does nothing, and a downloaded copy will not play for anyone else.',
		},
		{
			d: 'M4.5 6.5h15M9.5 6.5V4.8a1.3 1.3 0 0 1 1.3-1.3h2.4a1.3 1.3 0 0 1 1.3 1.3v1.7M6.5 6.5l.9 12.2a1.6 1.6 0 0 0 1.6 1.5h6a1.6 1.6 0 0 0 1.6-1.5l.9-12.2',
			ten: 'We keep it, not your big file',
			ben: 'Once your video is ready to watch, the huge file you sent can go. Your course keeps playing exactly as before, and nothing sits around that nobody needs.',
		},
		{
			d: 'M12 2.8a9.2 9.2 0 1 0 0 18.4 9.2 9.2 0 0 0 0-18.4zM2.8 12h18.4M12 2.8c2.3 2.5 3.5 5.7 3.5 9.2s-1.2 6.7-3.5 9.2c-2.3-2.5-3.5-5.7-3.5-9.2S9.7 5.3 12 2.8z',
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
					<svg class="tilt flex-none text-brand-light" width="42" height="42" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d={f.d} /></svg>
				</div>
				<p class="mt-3.5 text-sm leading-[1.75] text-muted">{f.ben}</p>
			</article>
		{/each}
	</div>
</section>
