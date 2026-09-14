<script lang="ts">
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	// The reference put three invented people with names, roles, avatars and star
	// ratings here. Same grid, same cards, same reveal — but they describe the work
	// rather than pretending someone said something.
	const uses = [
		{
			d: 'M3.4 7.4 12 3.2l8.6 4.2L12 11.6zM6.6 9v5.2c0 1.6 2.4 2.9 5.4 2.9s5.4-1.3 5.4-2.9V9M20.6 7.4v5.4',
			ten: 'You teach, and you record every class',
			tbn: 'আপনি পড়ান, আর প্রতিটি ক্লাস রেকর্ড করেন',
			ben: 'An hour of screen recording is enormous and nobody wants to wait for it. Send the week’s lessons through in one batch and hand out links your students can open on any phone.',
			bbn: 'এক ঘণ্টার স্ক্রিন রেকর্ডিং আকারে বিশাল, আর কেউ সেটার জন্য অপেক্ষা করতে চান না। সপ্তাহের ক্লাসগুলো একবারে পাঠিয়ে দিন, আর এমন লিংক দিন যা শিক্ষার্থীরা যেকোনো ফোনে খুলতে পারেন।'
		},
		{
			d: 'M4.6 5.5h14.8a1.6 1.6 0 0 1 1.6 1.6v9.8a1.6 1.6 0 0 1-1.6 1.6H4.6A1.6 1.6 0 0 1 3 16.9V7.1a1.6 1.6 0 0 1 1.6-1.6zM10 9.2l4.8 2.8L10 14.8z',
			ten: 'You run a shop, and you post videos',
			tbn: 'আপনার দোকান আছে, আর আপনি ভিডিও দেন',
			ben: 'Recipe clips, product demos, the thing you filmed on the counter this morning. Shrink it before it goes out so the page loads before the customer loses interest.',
			bbn: 'রেসিপির ক্লিপ, পণ্যের ডেমো, আজ সকালে কাউন্টারে দাঁড়িয়ে যেটা তুলেছেন। ছাড়ার আগে ছোট করে নিন, যাতে খদ্দেরের আগ্রহ ফুরোনোর আগেই পাতাটা লোড হয়।'
		},
		{
			d: 'M12 3.2 14.8 9l6.4.9-4.6 4.5 1.1 6.4L12 17.8l-5.7 3 1.1-6.4L2.8 9.9 9.2 9z',
			ten: 'You shoot for clients, in 4K',
			tbn: 'আপনি ক্লায়েন্টের জন্য শুট করেন, ৪কে-তে',
			ben: 'The final cut is beautiful and completely unsendable. Get it down to something that fits in an email, without the client noticing anything has been taken away.',
			bbn: 'ফাইনাল কাটটা দারুণ, কিন্তু কোনোভাবেই পাঠানো যায় না। এমন আকারে নামিয়ে আনুন যা ইমেইলেই আঁটে — আর ক্লায়েন্ট টেরও পাবেন না কিছু কমেছে।'
		}
	];
</script>

<section class="mx-auto max-w-6xl px-4 py-16 sm:px-6 sm:py-24">
	<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
		<p class="text-xs tracking-widest text-emerald-light uppercase">
			Who it's for
		</p>
		<h2 class="mt-3 text-3xl font-bold tracking-tight sm:text-4xl leading-[1.32]">
			Made for the way you work.
		</h2>
	</div>

	<div class="mt-12 grid gap-4 md:grid-cols-3">
		{#each uses as u, i (u.ten)}
			<article class="card card-hover p-6" use:scrollReveal={{ delay: i * 90 }}>
				<span
					class="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-full border border-hairline bg-body text-emerald-light"
					aria-hidden="true"
				>
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"><path d={u.d} /></svg>
				</span>
				<h3 class="text-lg font-bold tracking-tight">{u.ten}</h3>
				<p class="mt-2 text-sm leading-relaxed text-muted">{u.ben}</p>
			</article>
		{/each}
	</div>
</section>
