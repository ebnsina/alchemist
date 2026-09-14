<script lang="ts">
	import T from '$lib/T.svelte';
	import { scrollReveal } from '$lib/utils/scroll-reveal.js';

	// The reference used customer counts, video counts and a star rating here. These
	// are facts about the product instead, which is both true and more persuasive.
	const stats = [
		{ value: '93%', en: 'Smaller files on average', bn: 'গড়ে এত ছোট ফাইল' },
		{ value: 'Under 60s', vbn: '৬০ সেকেন্ডের কম', en: 'For a 10-minute video', bn: '১০ মিনিটের ভিডিওর জন্য' },
		{ value: '8K', en: 'Highest resolution supported', bn: 'সর্বোচ্চ যত রেজোলিউশন চলে' },
		{ value: 'Any device', vbn: 'যেকোনো ডিভাইস', en: 'Phones, laptops, TVs', bn: 'ফোন, ল্যাপটপ, টিভি' }
	];
</script>

<section class="mx-auto max-w-6xl px-4 py-16 sm:px-6 sm:py-20">
	<dl class="grid grid-cols-2 gap-x-6 gap-y-10 text-center md:grid-cols-4">
		{#each stats as s, i (s.value)}
			<div use:scrollReveal={{ delay: i * 70 }}>
				<dt class="vh"><T as="span" en={s.en} bn={s.bn} /></dt>
				<dd class="gradient-text text-3xl font-extrabold tracking-tight sm:text-4xl">
					{#if s.vbn}<T as="span" en={s.value} bn={s.vbn} />{:else}{s.value}{/if}
				</dd>
				<p class="mt-2 text-sm text-muted"><T as="span" en={s.en} bn={s.bn} /></p>
			</div>
		{/each}
	</dl>
</section>
