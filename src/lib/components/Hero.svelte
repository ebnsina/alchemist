<script lang="ts">
	import T from '$lib/T.svelte';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';
	import AuroraGradient from './AuroraGradient.svelte';
</script>

<section class="relative overflow-hidden pt-36 pb-20 sm:pt-44 sm:pb-28">
	<AuroraGradient />

	<div class="mx-auto max-w-6xl px-4 text-center sm:px-6">
		<p
			class="inline-flex items-center gap-2 rounded-full border border-hairline bg-card px-4 py-1.5 text-sm text-muted"
		>
			<span class="h-1.5 w-1.5 rounded-full bg-emerald-light"></span>
			<T en="100 free videos every month · no card needed" bn="মাসে ১০০টি ভিডিও ফ্রি · কার্ড লাগবে না" />
		</p>

		<h1 class="mx-auto mt-8 max-w-5xl text-5xl font-bold tracking-tight sm:text-7xl leading-[1.12]">
			<T as="span" en="Convert any video," bn="যেকোনো ভিডিও বদলে ফেলুন," />
			<span class="gradient-text block"
				><T as="span" en="into something magical." bn="জাদুকরি কিছুতে।" /></span
			>
		</h1>

		<p class="mx-auto mt-6 max-w-2xl text-lg text-muted sm:text-xl">
			<T
				as="span"
				en="Drop in a video and get back one that plays anywhere, loads fast, and is a fraction of the size. No settings to learn, nothing to install."
				bn="একটা ভিডিও দিন, ফেরত পান এমন একটা যা সব জায়গায় চলে, দ্রুত লোড হয়, আর আকারে কয়েক গুণ ছোট। শেখার মতো কোনো সেটিং নেই, ইনস্টল করারও কিছু নেই।"
			/>
		</p>

		<div class="mt-8 flex flex-wrap justify-center gap-3">
			<a href="/#pricing" class="btn-primary"><T en="Start free" bn="ফ্রি শুরু করুন" /></a>
			<a href="/#how" class="btn-ghost"><T en="See how it works" bn="কীভাবে কাজ করে দেখুন" /></a>
		</div>

		<ul class="mt-8 flex flex-wrap justify-center gap-x-7 gap-y-2 text-sm text-muted">
			<li><T en="Cancel any time" bn="যেকোনো সময় বন্ধ" /></li>
			<li><T en="Files deleted after 24 hours" bn="২৪ ঘণ্টা পর ফাইল মুছে যায়" /></li>
			<li><T en="Works on any phone" bn="যেকোনো ফোনে চলে" /></li>
		</ul>

		<!-- An illustration of what the tool does, not a screenshot of an app. -->
		<figure class="card shine mx-auto mt-14 max-w-2xl p-4 text-left sm:p-6" use:scrollReveal>
			<figcaption class="mb-5 flex items-center justify-between text-sm text-muted">
				<T as="span" en="What one conversion looks like" bn="একটা ভিডিও রূপান্তরের চেহারা" />
				<T as="span" en="Illustration" bn="দৃষ্টান্ত" />
			</figcaption>
			<div class="flex items-center justify-between gap-4">
				<div>
					<p class="text-sm text-muted"><T as="span" en="Before" bn="আগে" /></p>
					<p class="text-3xl font-bold tracking-tight sm:text-4xl">1.2 GB</p>
				</div>
				<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="flex-none text-emerald-light" aria-hidden="true">
					<path d="M4 12h15M13.5 6.5 20 12l-6.5 5.5" />
				</svg>
				<div class="text-right">
					<p class="text-sm text-muted"><T as="span" en="After" bn="পরে" /></p>
					<p class="gradient-text text-3xl font-bold tracking-tight sm:text-4xl">84 MB</p>
				</div>
			</div>
			<div class="mt-4 h-1.5 overflow-hidden rounded-full bg-white/10">
				<div class="progress-anim h-full w-full rounded-full bg-emerald"></div>
			</div>
			<p class="mt-4 text-sm text-muted">
				<T
					as="span"
					en="A 10-minute clip, 93% smaller, and it still looks right on a big screen."
					bn="১০ মিনিটের একটা ক্লিপ, ৯৩% ছোট — আর বড় পর্দাতেও দেখতে ঠিকঠাক।"
				/>
			</p>
		</figure>
	</div>
</section>
