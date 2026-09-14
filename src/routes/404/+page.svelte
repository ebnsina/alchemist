<script lang="ts">
	import T from '$lib/T.svelte';
</script>

<svelte:head>
	<title>Page not found — Alchemist</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<section class="wrap">
	<p class="code num">404</p>
	<h1>
		<T
			as="span"
			en="That page is not here."
			bn="এই পাতাটি এখানে নেই।"
		/>
	</h1>
	<p class="lede">
		<T
			as="span"
			en="Either it moved or it never existed. Neither is your fault. Everything this site has is one click away."
			bn="হয় পাতাটি সরে গেছে, নয়তো কখনও ছিলই না। দুটোর কোনোটাই আপনার দোষ নয়। এই সাইটে যা যা আছে, সবই এক ক্লিক দূরে।"
		/>
	</p>
	<ul>
		<li><a href="/"><T en="Home" bn="হোম" /></a></li>
		<li><a href="/edtech/"><T en="For edtech" bn="এডটেকের জন্য" /></a></li>
		<li><a href="/media/"><T en="For media" bn="মিডিয়ার জন্য" /></a></li>
		<li><a href="/pricing/"><T en="Pricing" bn="মূল্য" /></a></li>
		<li><a href="/docs/"><T en="Docs" bn="ডকুমেন্টেশন" /></a></li>
		<li><a href="/about/"><T en="About" bn="পরিচিতি" /></a></li>
	</ul>
</section>

<style>
	section {
		padding-block: clamp(3rem, 10vw, 6rem);
	}
	.code {
		font-size: 3rem;
		color: var(--brass);
		margin: 0 0 0.5rem;
		line-height: 1;
	}
	ul {
		list-style: none;
		padding: 0;
		margin: 1.5rem 0 0;
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem 1.2rem;
	}
</style>
