<script lang="ts">
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const steps = [
		{
			ten: 'Your user sends a file',
			ben: 'Ask us for an upload target and let the browser put the bytes straight into storage. The file never passes through your servers or ours.'
		},
		{
			ten: 'We do the work',
			ben: 'Probe, ladder, encode in parallel chunks, package. You get a webhook when it is playable, so nothing in your code has to sit and poll.'
		},
		{
			ten: 'You hand out a link',
			ben: 'Ask for the asset and play the links it returns. They are signed, they expire when you say, and they work on anything.'
		}
	];
</script>

<section id="how" class="relative scroll-mt-24 overflow-hidden py-16 sm:py-24">
	<div class="glow-2 pointer-events-none absolute top-10 -right-20 -z-10 h-[26rem] w-[26rem]" aria-hidden="true"></div>
	<div class="mx-auto max-w-6xl px-4 sm:px-6">
		<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
			<p class="label text-ink">
				How it works
			</p>
			<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
				Three easy steps.
				<span class="text-ink">That's it.</span>
			</h2>
		</div>

		<ol class="mt-14 grid gap-5 md:grid-cols-3">
			{#each steps as s, i (s.ten)}
				<li class="card p-8" use:scrollReveal={{ delay: i * 90 }}>
					<div class="flex items-start justify-between gap-5">
						<h3 class="text-lg font-semibold tracking-tight">{s.ten}</h3>
						<span class="flex-none text-3xl leading-none font-semibold text-ink" aria-hidden="true"
							>{i + 1}</span
						>
					</div>
					<p class="mt-3.5 text-sm leading-[1.75] text-dim">{s.ben}</p>
				</li>
			{/each}
		</ol>
	</div>
</section>
