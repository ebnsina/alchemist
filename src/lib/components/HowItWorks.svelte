<script lang="ts">
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const steps = [
		{
			ten: 'Drop your video in',
			tbn: 'ভিডিওটা রেখে দিন',
			ben: 'Drag it onto the page, or pick it from your phone. Any format you have ever been handed is one we accept.',
			bbn: 'পাতার উপর টেনে আনুন, বা ফোন থেকে বেছে নিন। এ যাবৎ আপনাকে যত রকম ফরম্যাট কেউ পাঠিয়েছে, সবই আমরা নিই।'
		},
		{
			ten: 'We do the work',
			tbn: 'কাজটা আমরা করি',
			ben: 'Pick how you will use it — phone, web, or keep it big — and we handle the rest. There is nothing else to set.',
			bbn: 'শুধু বলুন কোথায় ব্যবহার করবেন — ফোনে, ওয়েবে, নাকি বড় আকারেই রাখবেন — বাকিটা আমরা সামলাই। আর কিছু ঠিক করার নেই।'
		},
		{
			ten: 'Take it away',
			tbn: 'নিয়ে যান',
			ben: 'Download it, or copy a link and send it. A day later we delete our copy, and that is the end of it.',
			bbn: 'নামিয়ে নিন, বা লিংক কপি করে পাঠিয়ে দিন। একদিন পর আমাদের কপিটা আমরা মুছে ফেলি, ব্যস।'
		}
	];
</script>

<section id="how" class="relative scroll-mt-24 overflow-hidden py-16 sm:py-24">
	<div class="glow-2 pointer-events-none absolute top-10 -right-20 -z-10 h-[26rem] w-[26rem]" aria-hidden="true"></div>
	<div class="mx-auto max-w-6xl px-4 sm:px-6">
		<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
			<p class="text-xs tracking-widest text-brand-light uppercase">
				How it works
			</p>
			<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
				Three easy steps.
				<span class="gradient-text">That's it.</span>
			</h2>
		</div>

		<ol class="mt-14 grid gap-5 md:grid-cols-3">
			{#each steps as s, i (s.ten)}
				<li class="card p-8" use:scrollReveal={{ delay: i * 90 }}>
					<div class="flex items-start justify-between gap-5">
						<h3 class="text-lg font-semibold tracking-tight">{s.ten}</h3>
						<span class="flex-none text-2xl leading-none font-semibold text-brand-light" aria-hidden="true"
							>{i + 1}</span
						>
					</div>
					<p class="mt-3.5 text-sm leading-[1.75] text-muted">{s.ben}</p>
				</li>
			{/each}
		</ol>
	</div>
</section>
