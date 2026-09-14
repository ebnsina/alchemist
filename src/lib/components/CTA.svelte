<script lang="ts">
	import T from '$lib/T.svelte';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';
</script>

<section class="mx-auto max-w-6xl px-4 pb-20 sm:px-6 sm:pb-28">
	<div class="card relative overflow-hidden px-6 py-14 text-center sm:px-12" use:scrollReveal>
		<div class="glow pointer-events-none absolute -top-24 left-1/2 h-[26rem] w-[26rem] -translate-x-1/2" aria-hidden="true"></div>
		<div class="glow-gold pointer-events-none absolute -bottom-24 right-0 h-[20rem] w-[20rem]" aria-hidden="true"></div>
		<h2 class="relative mx-auto max-w-xl text-3xl font-extrabold tracking-tight sm:text-4xl">
			<T as="span" en="Your next video is about to get a lot lighter." bn="আপনার পরের ভিডিওটা এখনই অনেক হালকা হতে চলেছে।" />
		</h2>
		<p class="relative mx-auto mt-4 max-w-md text-sm text-muted">
			<T
				as="span"
				en="A hundred videos a month, free, with no card and no clock running. See what comes back before you decide anything."
				bn="মাসে একশটি ভিডিও, ফ্রি — কার্ড লাগবে না, কোনো ঘড়িও চলবে না। কিছু ঠিক করার আগে দেখে নিন কী ফেরত আসে।"
			/>
		</p>
		<div class="relative mt-8 flex flex-wrap justify-center gap-3">
			<a href="/#pricing" class="btn-primary"><T en="Start free" bn="ফ্রি শুরু করুন" /></a>
			<a href="/#faq" class="btn-ghost"><T en="Read the questions" bn="প্রশ্নগুলো পড়ুন" /></a>
		</div>
	</div>
</section>
