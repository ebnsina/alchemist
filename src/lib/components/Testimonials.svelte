<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Mortarboard01Icon, PlaySquareIcon, StarIcon } from '@hugeicons/core-free-icons';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	// The reference put three invented people with names, roles, avatars and star
	// ratings here. Same grid, same cards, same reveal — but they describe the work
	// rather than pretending someone said something.
	const uses = [
		{
			icon: Mortarboard01Icon,
			ten: 'You teach, and you record every class',
			tbn: 'আপনি পড়ান, আর প্রতিটি ক্লাস রেকর্ড করেন',
			ben: 'An hour of screen recording is enormous and nobody wants to wait for it. Send the week’s lessons through in one batch and hand out links your students can open on any phone.',
			bbn: 'এক ঘণ্টার স্ক্রিন রেকর্ডিং আকারে বিশাল, আর কেউ সেটার জন্য অপেক্ষা করতে চান না। সপ্তাহের ক্লাসগুলো একবারে পাঠিয়ে দিন, আর এমন লিংক দিন যা শিক্ষার্থীরা যেকোনো ফোনে খুলতে পারেন।'
		},
		{
			icon: PlaySquareIcon,
			ten: 'You run a shop, and you post videos',
			tbn: 'আপনার দোকান আছে, আর আপনি ভিডিও দেন',
			ben: 'Recipe clips, product demos, the thing you filmed on the counter this morning. Shrink it before it goes out so the page loads before the customer loses interest.',
			bbn: 'রেসিপির ক্লিপ, পণ্যের ডেমো, আজ সকালে কাউন্টারে দাঁড়িয়ে যেটা তুলেছেন। ছাড়ার আগে ছোট করে নিন, যাতে খদ্দেরের আগ্রহ ফুরোনোর আগেই পাতাটা লোড হয়।'
		},
		{
			icon: StarIcon,
			ten: 'You shoot for clients, in 4K',
			tbn: 'আপনি ক্লায়েন্টের জন্য শুট করেন, ৪কে-তে',
			ben: 'The final cut is beautiful and completely unsendable. Get it down to something that fits in an email, without the client noticing anything has been taken away.',
			bbn: 'ফাইনাল কাটটা দারুণ, কিন্তু কোনোভাবেই পাঠানো যায় না। এমন আকারে নামিয়ে আনুন যা ইমেইলেই আঁটে — আর ক্লায়েন্ট টেরও পাবেন না কিছু কমেছে।'
		}
	];
</script>

<section class="mx-auto max-w-6xl px-4 py-16 sm:px-6 sm:py-24">
	<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
		<p class="text-xs tracking-widest text-brand-light uppercase">
			Who it's for
		</p>
		<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
			Made for the way you work.
		</h2>
	</div>

	<div class="mt-14 grid gap-5 md:grid-cols-3">
		{#each uses as u, i (u.ten)}
			<article class="card card-hover p-8" use:scrollReveal={{ delay: i * 90 }}>
				<div class="flex items-start justify-between gap-5">
					<h3 class="text-lg font-semibold tracking-tight">{u.ten}</h3>
					<HugeiconsIcon
						icon={u.icon}
						size={42}
						strokeWidth={1.6}
						class="tilt flex-none text-brand-light"
					/>
				</div>
				<p class="mt-3.5 text-sm leading-[1.75] text-muted">{u.ben}</p>
			</article>
		{/each}
	</div>
</section>
