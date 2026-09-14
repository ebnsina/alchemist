<script lang="ts">
	import T from './T.svelte';

	// A diagram, not a screenshot. One uploaded video, the five sizes we send it at,
	// and what an hour of each costs the person watching. Bar length is the value
	// against the largest, so it is a single-series magnitude chart: every value is
	// directly labelled, which is why it needs no axis, no gridlines and no legend.
	const TOP = 810;
	const sizes = [
		{ en: 'Lowest', bn: 'সবচেয়ে কম', mb: 68 },
		{ en: 'Low', bn: 'কম', mb: 135 },
		{ en: 'Default', bn: 'ডিফল্ট', mb: 270, dflt: true },
		{ en: 'Higher', bn: 'বেশি', mb: 450 },
		{ en: 'Highest', bn: 'সবচেয়ে বেশি', mb: 810 }
	];
</script>

<figure class="frame">
	<figcaption class="frame__bar">
		<span class="frame__title">
			<T
				as="span"
				en="One video you send us, and the sizes it goes out at"
				bn="আপনার পাঠানো একটি ভিডিও, আর যে যে আকারে সেটি যায়"
			/>
		</span>
		<span class="frame__note">
			<T
				as="span"
				en="A diagram, not a screen from an app"
				bn="এটি একটি চিত্র, কোনো অ্যাপের পর্দা নয়"
			/>
		</span>
	</figcaption>
	<div class="frame__body">
		<p class="capt">
			<T
				as="span"
				en="Data an hour of watching costs the viewer"
				bn="এক ঘণ্টা দেখতে দর্শকের যত ডেটা লাগে"
			/>
		</p>
		<ul>
			{#each sizes as s (s.en)}
				<li class:is-default={s.dflt}>
					<span class="name"><T as="span" en={s.en} bn={s.bn} /></span>
					<span class="track"><span class="bar" style="--w:{((s.mb / TOP) * 100).toFixed(1)}%"></span></span>
					<span class="val">{s.mb}<span> MB</span></span>
				</li>
			{/each}
		</ul>
		<p class="foot">
			<T
				as="span"
				en="The player picks a size to suit the connection and starts at the default. Worked out from the size we send at each step; a real video varies a little either way."
				bn="সংযোগ বুঝে প্লেয়ার নিজেই একটি আকার বেছে নেয়, শুরু করে ডিফল্ট থেকে। প্রতিটি ধাপে আমরা যে আকারে পাঠাই তা থেকে কষা; আসল ভিডিওতে এদিক-ওদিক একটু হয়।"
			/>
		</p>
	</div>
</figure>

<style>
	figure {
		margin: 0;
	}
	.capt {
		margin: 0;
		font-size: var(--step--1);
		letter-spacing: 0.04em;
		color: var(--ink-3);
		margin-bottom: 0.9rem;
	}
	:global(html[lang='bn']) .capt {
		letter-spacing: 0;
		font-size: var(--step-0);
	}
	ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		/* The 2px surface gap that separates touching marks. */
		gap: 6px;
	}
	li {
		display: grid;
		grid-template-columns: 6.5rem 1fr auto;
		align-items: center;
		gap: 0.8rem;
	}
	.name {
		font-size: var(--step-0);
		color: var(--graphite);
	}
	:global(html[lang='bn']) .name {
		font-size: var(--step-1);
	}
	.track {
		display: block;
		height: 14px;
	}
	.bar {
		display: block;
		height: 14px;
		width: var(--w);
		min-width: 4px;
		background: var(--rule-firm);
		border-radius: 0 4px 4px 0;
	}
	.is-default .bar {
		background: var(--brass);
	}
	.val {
		font-family: var(--mono);
		font-variant-numeric: tabular-nums;
		font-size: var(--step-0);
		white-space: nowrap;
		color: var(--graphite);
	}
	.val span {
		color: var(--ink-3);
	}
	.is-default .name,
	.is-default .val {
		color: var(--ink);
	}
	.foot {
		margin: 1.1rem 0 0;
		padding-top: 0.9rem;
		border-top: 1px solid var(--rule);
		font-size: var(--step--1);
		line-height: 1.6;
		color: var(--ink-3);
		max-width: none;
	}
	:global(html[lang='bn']) .foot {
		font-size: var(--step-0);
	}
	@media (max-width: 460px) {
		li {
			grid-template-columns: 5rem 1fr auto;
			gap: 0.5rem;
		}
	}
</style>
