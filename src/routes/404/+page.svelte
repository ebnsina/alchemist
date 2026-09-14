<script lang="ts">
	import T from '$lib/T.svelte';
</script>

<svelte:head>
	<title>Page not found — Alchemist</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<section class="shell bay">
	<div class="flow">
		<p class="marker">404</p>
		<h1><T as="span" en="That page is not here." bn="এই পাতাটি এখানে নেই।" /></h1>
		<p class="lede quiet">
			<T
				as="span"
				en="Either it moved or it never existed. Neither is your fault. Everything this site has is one click away."
				bn="হয় পাতাটি সরে গেছে, নয়তো কখনও ছিলই না। দুটোর কোনোটাই আপনার দোষ নয়। এই সাইটে যা যা আছে, সবই এক ক্লিক দূরে।"
			/>
		</p>
		<ul class="ways">
			<li><a class="tlink" href="/"><T en="Home" bn="হোম" /></a></li>
			<li><a class="tlink" href="/edtech/"><T en="Course platforms" bn="কোর্স প্ল্যাটফর্ম" /></a></li>
			<li><a class="tlink" href="/media/"><T en="News and media" bn="সংবাদ ও মিডিয়া" /></a></li>
			<li><a class="tlink" href="/pricing/"><T en="Pricing" bn="মূল্য" /></a></li>
			<li><a class="tlink" href="/docs/"><T en="Docs" bn="ডকুমেন্টেশন" /></a></li>
			<li><a class="tlink" href="/about/"><T en="About" bn="পরিচিতি" /></a></li>
		</ul>
	</div>
</section>

<style>
	.marker {
		font-family: var(--mono);
		color: var(--brass);
		max-width: 5rem;
	}
	h1 {
		max-width: 13ch;
	}
	.ways {
		list-style: none;
		padding: 0;
		margin: 2rem 0 0;
		display: flex;
		flex-wrap: wrap;
		gap: 0.6rem 1.6rem;
	}
</style>
