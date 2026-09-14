<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { ArrowDown01Icon } from '@hugeicons/core-free-icons';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const qs = [
		{
			qen: 'Where does my video live, and who can watch it?',
			aen: 'It stays with us and plays from a link you hand out, and you decide how long that link keeps working. The big file you first sent us is a separate thing — once your video is ready to watch, we can get rid of that for you. Nothing of yours is ever used for anything else.',
		},
		{
			qen: 'Will it look worse?',
			aen: 'Not in any way you would notice. We take out the parts of a video file that your eyes were never going to use, which is where nearly all the weight is hiding.',
		},
		{
			qen: 'Which formats can I send?',
			aen: 'Whatever your camera, phone or editor produced. If somebody has handed you the file, we will take it, and you will get back something that plays everywhere.',
		},
		{
			qen: 'Do I need to install anything?',
			aen: 'No. It runs in the browser you already have, on a laptop or on a phone. There is nothing to download and nothing to keep updated.',
		},
		{
			qen: 'Can I stop paying whenever I want?',
			aen: 'Yes. Two clicks on your account page, no notice period, no exit fee and nobody on the phone trying to talk you out of it. The free tier is still there afterwards.',
		},
		{
			qen: 'Is 100 free videos a month really free?',
			aen: 'Yes, with no card and no trial clock. If you never go past a hundred, you never pay, and we will not email you to ask why.',
		}
	];
</script>

<section id="faq" class="mx-auto max-w-3xl scroll-mt-24 px-4 py-16 sm:px-6 sm:py-24">
	<div class="text-center" use:scrollReveal>
		<p class="text-xs tracking-widest text-brand-light uppercase">
			Questions
		</p>
		<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
			Got questions?
			<span class="gradient-text">We've got answers.</span>
		</h2>
	</div>

	<div class="mt-10 divide-y divide-hairline border-y border-hairline" use:scrollReveal>
		{#each qs as q (q.qen)}
			<details class="group">
				<summary
					class="flex cursor-pointer list-none items-center justify-between gap-4 py-4 font-medium transition-colors hover:text-brand-light"
				>
					{q.qen}
					<HugeiconsIcon
						icon={ArrowDown01Icon}
						size={17}
						strokeWidth={2}
						class="chevron flex-none text-muted"
					/>
				</summary>
				<p class="pb-5 text-sm leading-relaxed text-muted">{q.aen}</p>
			</details>
		{/each}
	</div>
</section>
