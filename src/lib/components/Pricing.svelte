<script lang="ts">
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const tiers = [
		{
			ten: 'Free',
			tbn: 'ফ্রি',
			price: '$0',
			pen: 'forever',
			pbn: 'সবসময়',
			den: 'For trying it out and for the occasional video.',
			dbn: 'একটু যাচাই করে দেখতে, আর মাঝেমধ্যে দু-একটা ভিডিওর জন্য।',
			cta_en: 'Start free',
			cta_bn: 'ফ্রি শুরু করুন',
			rows: [
				'100 videos a month',
				'Up to 1080p',
				'Files up to 2 GB',
				'Deleted after 24 hours'
			]
		},
		{
			ten: 'Creator',
			tbn: 'ক্রিয়েটর',
			price: '$12',
			pen: 'a month',
			pbn: 'মাসে',
			featured: true,
			den: 'For anyone posting video as part of the job.',
			dbn: 'যাঁদের কাজের অংশ হিসেবেই নিয়মিত ভিডিও দিতে হয়।',
			cta_en: 'Choose Creator',
			cta_bn: 'ক্রিয়েটর নিন',
			rows: [
				'Unlimited videos',
				'Up to 4K',
				'Files up to 20 GB',
				'Batch up to 10 at once',
				'Shareable links'
			]
		},
		{
			ten: 'Studio',
			tbn: 'স্টুডিও',
			price: '$39',
			pen: 'a month',
			pbn: 'মাসে',
			den: 'For client work, long files and big finals.',
			dbn: 'ক্লায়েন্টের কাজ, লম্বা ফাইল আর বড় ফাইনালের জন্য।',
			cta_en: 'Choose Studio',
			cta_bn: 'স্টুডিও নিন',
			rows: [
				'Everything in Creator',
				'Up to 8K',
				'Files up to 100 GB',
				'Batch up to 50 at once',
				'Priority queue'
			]
		}
	];
</script>

<section id="pricing" class="mx-auto max-w-6xl scroll-mt-24 px-4 py-16 sm:px-6 sm:py-24">
	<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
		<p class="text-xs tracking-widest text-emerald-light uppercase">Pricing</p>
		<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
			Simple pricing.
			<span class="gradient-text">No surprises.</span>
		</h2>
		<p class="mt-4 text-sm text-muted">
			Cancel any time, from your account page, in two clicks. No notice period and no asterisk.
		</p>
	</div>

	<div class="mx-auto mt-12 grid max-w-4xl gap-4 sm:grid-cols-2 lg:grid-cols-3">
		{#each tiers as t, i (t.ten)}
			<article
				class="card relative flex flex-col px-6 py-8 {t.featured ? 'border-emerald/60 bg-card' : ''}"
				use:scrollReveal={{ delay: i * 80 }}
			>
				{#if t.featured}
					<p
						class="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-emerald px-2.5 py-0.5 text-xs font-semibold whitespace-nowrap text-body"
					>
						Most popular
					</p>
				{/if}
				<h3 class="text-sm font-semibold tracking-wide text-muted uppercase">
					{t.ten}
				</h3>
				<p class="mt-4 flex items-baseline gap-1.5">
					<span class="text-4xl font-semibold tracking-tight">{t.price}</span>
					<span class="text-sm text-muted">{t.pen}</span>
				</p>
				<p class="mt-2 text-sm text-muted">{t.den}</p>
				<ul class="mt-7 flex-1 space-y-3">
					{#each t.rows as row (row)}
						<li class="flex items-start gap-2.5 text-sm">
							<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" class="mt-1 flex-none text-emerald-light" aria-hidden="true"><path d="M4.6 12.5 9.4 17.3 19.4 6.9" /></svg>
							<span>{row}</span>
						</li>
					{/each}
				</ul>
				<a href="/#faq" class="mt-8 {t.featured ? 'btn-primary' : 'btn-ghost'}">
					{t.cta_en}
				</a>
			</article>
		{/each}
	</div>
</section>
