<script lang="ts">
	import T from '$lib/T.svelte';
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
				['100 videos a month', 'মাসে ১০০টি ভিডিও'],
				['Up to 1080p', '১০৮০পি পর্যন্ত'],
				['Files up to 2 GB', '২ জিবি পর্যন্ত ফাইল'],
				['Deleted after 24 hours', '২৪ ঘণ্টা পর মুছে যায়']
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
				['Unlimited videos', 'যত খুশি ভিডিও'],
				['Up to 4K', '৪কে পর্যন্ত'],
				['Files up to 20 GB', '২০ জিবি পর্যন্ত ফাইল'],
				['Batch up to 10 at once', 'একসঙ্গে ১০টি পর্যন্ত'],
				['Shareable links', 'শেয়ার করার মতো লিংক']
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
				['Everything in Creator', 'ক্রিয়েটরের সবকিছু'],
				['Up to 8K', '৮কে পর্যন্ত'],
				['Files up to 100 GB', '১০০ জিবি পর্যন্ত ফাইল'],
				['Batch up to 50 at once', 'একসঙ্গে ৫০টি পর্যন্ত'],
				['Priority queue', 'অগ্রাধিকারের সারি']
			]
		}
	];
</script>

<section id="pricing" class="mx-auto max-w-6xl scroll-mt-24 px-4 py-16 sm:px-6 sm:py-24">
	<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
		<p class="text-xs tracking-widest text-emerald-light uppercase"><T en="Pricing" bn="দাম" /></p>
		<h2 class="mt-3 text-3xl font-bold tracking-tight sm:text-4xl">
			<T as="span" en="Simple pricing." bn="সহজ দাম।" />
			<span class="gradient-text"><T as="span" en="No surprises." bn="কোনো চমক নেই।" /></span>
		</h2>
		<p class="mt-4 text-sm text-muted">
			<T
				as="span"
				en="Cancel any time, from your account page, in two clicks. No notice period and no asterisk."
				bn="যেকোনো সময় বন্ধ করতে পারেন, নিজের অ্যাকাউন্ট পাতা থেকে, দুটি ক্লিকে। আগে জানানোর বাধ্যবাধকতা নেই, তারকাচিহ্নও নেই।"
			/>
		</p>
	</div>

	<div class="mt-12 grid gap-4 lg:grid-cols-3">
		{#each tiers as t, i (t.ten)}
			<article
				class="card relative flex flex-col p-6 {t.featured ? 'border-emerald/60 bg-card' : ''}"
				use:scrollReveal={{ delay: i * 80 }}
			>
				{#if t.featured}
					<p
						class="absolute -top-3 left-6 rounded-full bg-emerald px-2.5 py-0.5 text-xs font-semibold text-body"
					>
						<T en="Most popular" bn="সবচেয়ে বেশি নেওয়া হয়" />
					</p>
				{/if}
				<h3 class="text-sm font-semibold tracking-wide text-muted uppercase">
					<T as="span" en={t.ten} bn={t.tbn} />
				</h3>
				<p class="mt-3 flex items-baseline gap-1.5">
					<span class="text-4xl font-bold tracking-tight">{t.price}</span>
					<span class="text-sm text-muted"><T as="span" en={t.pen} bn={t.pbn} /></span>
				</p>
				<p class="mt-2 text-sm text-muted"><T as="span" en={t.den} bn={t.dbn} /></p>
				<ul class="mt-6 flex-1 space-y-2.5">
					{#each t.rows as [en, bn] (en)}
						<li class="flex items-start gap-2.5 text-sm">
							<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" class="mt-1 flex-none text-emerald-light" aria-hidden="true"><path d="M4.6 12.5 9.4 17.3 19.4 6.9" /></svg>
							<T as="span" {en} {bn} />
						</li>
					{/each}
				</ul>
				<a href="/#faq" class="mt-6 {t.featured ? 'btn-primary' : 'btn-ghost'}">
					<T en={t.cta_en} bn={t.cta_bn} />
				</a>
			</article>
		{/each}
	</div>
</section>
