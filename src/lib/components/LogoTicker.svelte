<script lang="ts">
	import T from '$lib/T.svelte';

	// The reference put invented company names here under "Trusted by 50,000+".
	// Same ticker, same animation — but it names the kinds of work Alchemist is for,
	// which claims nothing about who uses it.
	const kinds = [
		{ en: 'Online courses', bn: 'অনলাইন কোর্স' },
		{ en: 'Recorded lessons', bn: 'রেকর্ড করা ক্লাস' },
		{ en: 'Recipe videos', bn: 'রেসিপি ভিডিও' },
		{ en: 'Wedding films', bn: 'বিয়ের ভিডিও' },
		{ en: 'Podcasts', bn: 'পডকাস্ট' },
		{ en: 'Product demos', bn: 'পণ্যের ডেমো' },
		{ en: 'Fitness classes', bn: 'ফিটনেস ক্লাস' },
		{ en: 'Client work', bn: 'ক্লায়েন্টের কাজ' }
	];
</script>

<section class="border-y border-hairline py-10">
	<p class="mb-6 text-center text-xs tracking-widest text-muted uppercase">
		<T en="Built for creators, teachers, and small businesses" bn="ক্রিয়েটর, শিক্ষক আর ছোট ব্যবসার জন্য তৈরি" />
	</p>
	<div class="ticker-mask overflow-hidden">
		<ul class="ticker gap-10 px-5">
			{#each [...kinds, ...kinds] as k, i (i)}
				<li
					class="flex flex-none items-center gap-2 text-sm whitespace-nowrap text-muted"
					aria-hidden={i >= kinds.length ? 'true' : undefined}
				>
					<span class="h-1 w-1 rounded-full bg-emerald-light" aria-hidden="true"></span>
					<T en={k.en} bn={k.bn} />
				</li>
			{/each}
		</ul>
	</div>
</section>
