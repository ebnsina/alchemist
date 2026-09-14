<script lang="ts">
	import T from '$lib/T.svelte';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	const qs = [
		{
			qen: 'What happens to my video after I download it?',
			qbn: 'নামিয়ে নেওয়ার পর আমার ভিডিওর কী হয়?',
			aen: 'We delete it 24 hours later, along with the original you sent. We are a workshop, not a warehouse — nothing of yours stays with us, and nothing of yours is ever used for anything else.',
			abn: 'ঠিক ২৪ ঘণ্টা পর আমরা সেটি মুছে ফেলি, সঙ্গে আপনার পাঠানো আসল ফাইলটিও। আমরা কারখানা, গুদাম নই — আপনার কিছুই আমাদের কাছে থেকে যায় না, আর আপনার কিছু অন্য কোনো কাজেও লাগানো হয় না।'
		},
		{
			qen: 'Will it look worse?',
			qbn: 'দেখতে কি খারাপ হয়ে যাবে?',
			aen: 'Not in any way you would notice. We take out the parts of a video file that your eyes were never going to use, which is where nearly all the weight is hiding.',
			abn: 'চোখে পড়ার মতো কোনোভাবেই নয়। ভিডিও ফাইলের যে অংশগুলো আপনার চোখ কোনোদিনই কাজে লাগাত না, আমরা সেগুলোই বাদ দিই — আর ওজনের প্রায় পুরোটা ওখানেই লুকিয়ে থাকে।'
		},
		{
			qen: 'Which formats can I send?',
			qbn: 'কোন কোন ফরম্যাট পাঠাতে পারি?',
			aen: 'Whatever your camera, phone or editor produced. If somebody has handed you the file, we will take it, and you will get back something that plays everywhere.',
			abn: 'আপনার ক্যামেরা, ফোন বা এডিটর যা-ই বানিয়েছে। কেউ যদি আপনাকে ফাইলটা দিয়ে থাকেন, আমরা সেটা নেব — আর ফেরত পাবেন এমন কিছু যা সব জায়গায় চলে।'
		},
		{
			qen: 'Do I need to install anything?',
			qbn: 'কিছু ইনস্টল করতে হবে?',
			aen: 'No. It runs in the browser you already have, on a laptop or on a phone. There is nothing to download and nothing to keep updated.',
			abn: 'না। আপনার এখনকার ব্রাউজারেই চলে, ল্যাপটপে বা ফোনে। নামানোরও কিছু নেই, আপডেট রাখারও কিছু নেই।'
		},
		{
			qen: 'Can I stop paying whenever I want?',
			qbn: 'যখন খুশি টাকা দেওয়া বন্ধ করতে পারব?',
			aen: 'Yes. Two clicks on your account page, no notice period, no exit fee and nobody on the phone trying to talk you out of it. The free tier is still there afterwards.',
			abn: 'হ্যাঁ। নিজের অ্যাকাউন্ট পাতায় দুটি ক্লিক — আগে জানানোর দরকার নেই, বেরোনোর ফি নেই, আর ফোন করে কেউ আপনাকে আটকানোর চেষ্টাও করবে না। এরপরও ফ্রি প্ল্যানটা থেকেই যায়।'
		},
		{
			qen: 'Is 100 free videos a month really free?',
			qbn: 'মাসে ১০০টি ভিডিও কি সত্যিই ফ্রি?',
			aen: 'Yes, with no card and no trial clock. If you never go past a hundred, you never pay, and we will not email you to ask why.',
			abn: 'হ্যাঁ — কার্ড লাগে না, ট্রায়ালের ঘড়িও নেই। একশর বেশি কখনও না লাগলে আপনাকে কখনও টাকা দিতে হবে না, আর কেন লাগছে না তা জানতে চেয়ে আমরা ইমেইলও করব না।'
		}
	];
</script>

<section id="faq" class="mx-auto max-w-3xl scroll-mt-24 px-4 py-16 sm:px-6 sm:py-24">
	<div class="text-center" use:scrollReveal>
		<p class="text-xs tracking-widest text-emerald-light uppercase">
			<T en="Questions" bn="প্রশ্ন" />
		</p>
		<h2 class="mt-3 text-3xl font-bold tracking-tight sm:text-4xl">
			<T as="span" en="Got questions?" bn="প্রশ্ন আছে?" />
			<span class="gradient-text"><T as="span" en="We've got answers." bn="উত্তরও আছে।" /></span>
		</h2>
	</div>

	<div class="mt-10 divide-y divide-hairline border-y border-hairline" use:scrollReveal>
		{#each qs as q (q.qen)}
			<details class="group">
				<summary
					class="flex cursor-pointer list-none items-center justify-between gap-4 py-4 font-medium transition-colors hover:text-emerald-light"
				>
					<T as="span" en={q.qen} bn={q.qbn} />
					<svg
						class="chevron flex-none text-muted"
						width="17"
						height="17"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"><path d="m6 9 6 6 6-6" /></svg
					>
				</summary>
				<p class="pb-5 text-sm leading-relaxed text-muted"><T as="span" en={q.aen} bn={q.abn} /></p>
			</details>
		{/each}
	</div>
</section>
