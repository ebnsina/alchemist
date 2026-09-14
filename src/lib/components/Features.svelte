<script lang="ts">
	import T from '$lib/T.svelte';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const features = [
		{
			d: 'M4.5 6.5h11a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2h-11a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2zM17.5 10.5l4-2.5v8l-4-2.5z',
			ten: 'Up to 8K, kept sharp',
			tbn: '৮কে পর্যন্ত, ঝকঝকে',
			ben: 'Send us anything up to 8K and it comes back looking like the original. We keep the detail and throw away the weight.',
			bbn: '৮কে পর্যন্ত যা খুশি পাঠান, ফেরত আসবে আসলটার মতোই দেখতে। খুঁটিনাটি থেকে যায়, শুধু ওজনটা চলে যায়।'
		},
		{
			d: 'M12 2.8a9.2 9.2 0 1 0 0 18.4 9.2 9.2 0 0 0 0-18.4zM12 6.8V12l3.4 2',
			ten: 'Done in under a minute',
			tbn: 'এক মিনিটের মধ্যেই শেষ',
			ben: 'A ten-minute video is usually ready before you have finished writing the caption. Longer files take a little more, but not much.',
			bbn: 'দশ মিনিটের ভিডিও সাধারণত ক্যাপশন লেখা শেষ হওয়ার আগেই তৈরি। বড় ফাইলে একটু বেশি লাগে, তবে খুব বেশি নয়।'
		},
		{
			d: 'M20.5 8.5 12 3 3.5 8.5v7L12 21l8.5-5.5zM12 12l8.5-3.5M12 12v9M12 12 3.5 8.5',
			ten: 'Up to 95% smaller',
			tbn: '৯৫% পর্যন্ত ছোট',
			ben: 'A gigabyte becomes a few dozen megabytes. Your viewers stop waiting, and the ones paying for their own data stop paying so much.',
			bbn: 'এক গিগাবাইট নেমে আসে কয়েক ডজন মেগাবাইটে। দর্শকের অপেক্ষা কমে, আর যাঁরা নিজের ডেটা কিনে দেখেন তাঁদের খরচও কমে।'
		},
		{
			d: 'M6.6 10.4h10.8a1 1 0 0 1 1 1v8.1a1 1 0 0 1-1 1H6.6a1 1 0 0 1-1-1v-8.1a1 1 0 0 1 1-1zM8.6 10.4V7a3.4 3.4 0 1 1 6.8 0v3.4',
			ten: 'Encrypted the whole way',
			tbn: 'পুরো পথেই এনক্রিপ্ট করা',
			ben: 'Your video is encrypted on the way up, while it sits with us, and on the way back down. Nobody else can open it.',
			bbn: 'আপনার ভিডিও এনক্রিপ্ট করা থাকে ওঠার সময়, আমাদের কাছে থাকার সময়, আর ফেরার সময়ও। আর কেউ এটি খুলতে পারে না।'
		},
		{
			d: 'M4.5 6.5h15M9.5 6.5V4.8a1.3 1.3 0 0 1 1.3-1.3h2.4a1.3 1.3 0 0 1 1.3 1.3v1.7M6.5 6.5l.9 12.2a1.6 1.6 0 0 0 1.6 1.5h6a1.6 1.6 0 0 0 1.6-1.5l.9-12.2',
			ten: 'Gone in 24 hours',
			tbn: '২৪ ঘণ্টায় মুছে যায়',
			ben: 'We delete your files a day after you download them. We are a workshop, not a warehouse — nothing of yours stays with us.',
			bbn: 'আপনি নামিয়ে নেওয়ার একদিন পরেই আমরা ফাইল মুছে ফেলি। আমরা কারখানা, গুদাম নই — আপনার কিছুই আমাদের কাছে থেকে যায় না।'
		},
		{
			d: 'M12 2.8a9.2 9.2 0 1 0 0 18.4 9.2 9.2 0 0 0 0-18.4zM2.8 12h18.4M12 2.8c2.3 2.5 3.5 5.7 3.5 9.2s-1.2 6.7-3.5 9.2c-2.3-2.5-3.5-5.7-3.5-9.2S9.7 5.3 12 2.8z',
			ten: 'Plays on anything',
			tbn: 'যেকোনো কিছুতে চলে',
			ben: 'One file that works on an old Android, a new iPhone, a laptop and a smart TV. No more sending three versions to three people.',
			bbn: 'একটাই ফাইল — পুরোনো অ্যান্ড্রয়েড, নতুন আইফোন, ল্যাপটপ আর স্মার্ট টিভি, সবখানেই চলে। তিনজনকে তিনটা আলাদা ভার্সন পাঠানোর দিন শেষ।'
		}
	];
</script>

<section id="features" class="mx-auto max-w-6xl scroll-mt-24 px-4 py-16 sm:px-6 sm:py-24">
	<div class="mx-auto max-w-2xl text-center" use:scrollReveal>
		<p class="text-xs tracking-widest text-emerald-light uppercase">
			<T en="What you get" bn="যা যা পাবেন" />
		</p>
		<h2 class="mt-3 text-3xl font-bold tracking-tight sm:text-4xl leading-[1.18]">
			<T as="span" en="Video made simple." bn="ভিডিও, একেবারে সহজ।" />
			<span class="gradient-text"><T as="span" en="Really simple." bn="সত্যিই সহজ।" /></span>
		</h2>
	</div>

	<div class="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		{#each features as f, i (f.ten)}
			<article class="card card-hover p-6" use:scrollReveal={{ delay: (i % 3) * 80 }}>
				<span
					class="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-xl border border-hairline bg-body text-emerald-light"
				>
					<svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d={f.d} /></svg>
				</span>
				<h3 class="text-lg font-bold tracking-tight"><T as="span" en={f.ten} bn={f.tbn} /></h3>
				<p class="mt-2 text-sm leading-relaxed text-muted"><T as="span" en={f.ben} bn={f.bbn} /></p>
			</article>
		{/each}
	</div>
</section>
