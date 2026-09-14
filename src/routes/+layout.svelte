<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import T from '$lib/T.svelte';
	import { SITE } from '$lib/site';

	let { children } = $props();

	const nav = [
		{ href: '/edtech/', en: 'Course platforms', bn: 'কোর্স প্ল্যাটফর্ম' },
		{ href: '/media/', en: 'News and media', bn: 'সংবাদ ও মিডিয়া' },
		{ href: '/pricing/', en: 'Pricing', bn: 'মূল্য' },
		{ href: '/docs/', en: 'For developers', bn: 'ডেভেলপারদের জন্য' }
	];
	const here = $derived(page.url.pathname);
</script>

<a class="skip" href="#main"><T en="Skip to content" bn="মূল অংশে যান" /></a>

<header>
	<div class="shell bar">
		<a class="brand" href="/">
			<svg width="20" height="24" viewBox="0 0 22 26" aria-hidden="true" focusable="false">
				<path
					d="M7.5 1.5h7M8.6 1.5v7.2L2.9 19.4a3 3 0 0 0 2.6 4.5h11a3 3 0 0 0 2.6-4.5L13.4 8.7V1.5"
					fill="none"
					stroke="currentColor"
					stroke-width="1.6"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
				<path d="M5.4 15.6h11.2l2.3 4.1a2.6 2.6 0 0 1-2.3 3.9H5.4a2.6 2.6 0 0 1-2.3-3.9z" fill="currentColor" />
			</svg>
			<span>Alchemist</span>
		</a>

		<nav aria-label="Main">
			<ul>
				{#each nav as n (n.href)}
					<li>
						<a href={n.href} aria-current={here === n.href ? 'page' : undefined}>
							<T en={n.en} bn={n.bn} />
						</a>
					</li>
				{/each}
			</ul>
		</nav>

		<div class="tools">
			<div class="seg" role="group" aria-label="Language / ভাষা">
				<button type="button" data-set-lang="en" aria-pressed="true" lang="en">EN</button>
				<button type="button" data-set-lang="bn" aria-pressed="false" lang="bn">বাংলা</button>
			</div>
			<button type="button" class="iconbtn" data-toggle-theme aria-label="Switch colour theme">
				<svg class="i-sun" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" aria-hidden="true" focusable="false">
					<circle cx="12" cy="12" r="4.6" />
					<path d="M12 1.8v2.4M12 19.8v2.4M2.6 12h2.4M19 12h2.4M5.3 5.3l1.7 1.7M17 17l1.7 1.7M18.7 5.3 17 7M7 17l-1.7 1.7" />
				</svg>
				<svg class="i-moon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
					<path d="M20.4 14.5A8.6 8.6 0 0 1 9.5 3.6a8.8 8.8 0 1 0 10.9 10.9z" />
				</svg>
			</button>
			<a class="btn btn--solid btn--sm cta" href="/about/">
				<T en="Why this exists" bn="কেন এটি আছে" />
			</a>
		</div>
	</div>
</header>

<main id="main">{@render children()}</main>

<footer>
	<div class="shell">
		<div class="fgrid">
			<div class="fabout">
				<p class="fname">Alchemist</p>
				<p class="small">
					<T
						as="span"
						en="Video hosting for apps and websites in Bangladesh. Kept inside the country, and made small enough for mobile data."
						bn="বাংলাদেশের অ্যাপ আর ওয়েবসাইটের জন্য ভিডিও হোস্টিং। দেশের ভেতরে রাখা, আর মোবাইল ডেটার উপযোগী করে ছোট করা।"
					/>
				</p>
			</div>

			<nav class="fcols" aria-label="Footer">
				<div>
					<h2><T as="span" en="Who it is for" bn="কাদের জন্য" /></h2>
					<ul>
						<li><a href="/edtech/"><T en="Course platforms" bn="কোর্স প্ল্যাটফর্ম" /></a></li>
						<li><a href="/media/"><T en="News and media" bn="সংবাদ ও মিডিয়া" /></a></li>
						<li><a href="/pricing/"><T en="Pricing" bn="মূল্য" /></a></li>
					</ul>
				</div>
				<div>
					<h2><T as="span" en="Developers" bn="ডেভেলপার" /></h2>
					<ul>
						<li><a href="/docs/"><T en="Start here" bn="এখান থেকে শুরু" /></a></li>
						<li><a href="{SITE.repo}/blob/main/api/openapi.yaml"><T en="API reference" bn="API রেফারেন্স" /></a></li>
						<li><a href="{SITE.repo}/blob/main/llms.txt">llms.txt</a></li>
					</ul>
				</div>
				<div>
					<h2><T as="span" en="Project" bn="প্রকল্প" /></h2>
					<ul>
						<li><a href="/about/"><T en="Why this exists" bn="কেন এটি আছে" /></a></li>
						<li><a href="/#faq"><T en="Questions" bn="প্রশ্ন" /></a></li>
						<li><a href="/pricing/"><T en="Pricing" bn="মূল্য" /></a></li>
					</ul>
				</div>
			</nav>
		</div>

		<div class="legal">
			<p class="fine">© 2026 Alchemist</p>
			<p class="fine">
				<T
					as="span"
					en="Pricing is not final until launch."
					bn="চালু হওয়ার আগপর্যন্ত দাম চূড়ান্ত নয়।"
				/>
			</p>
		</div>
	</div>
</footer>

<style>
	header {
		position: sticky;
		top: 0;
		z-index: 10;
		background: color-mix(in srgb, var(--paper) 78%, transparent);
		backdrop-filter: blur(14px) saturate(160%);
		border-bottom: 1px solid var(--rule);
	}
	.bar {
		display: flex;
		align-items: center;
		gap: 0.9rem;
		min-height: 54px;
		flex-wrap: wrap;
	}
	.brand {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 600;
		font-size: var(--step-2);
		letter-spacing: -0.03em;
		text-decoration: none;
		flex: 0 0 auto;
	}
	.brand svg {
		color: var(--brass);
		flex: 0 0 auto;
	}
	.bar nav {
		margin-inline: auto;
	}
	.bar ul {
		list-style: none;
		display: flex;
		flex-wrap: wrap;
		margin: 0;
		padding: 0;
		gap: 0 1.2rem;
	}
	.bar nav a {
		display: block;
		padding: 0.45rem 0;
		color: var(--graphite);
		text-decoration: none;
		font-size: var(--step-0);
		transition: color 140ms var(--ease);
	}
	:global(html[lang='bn']) .bar nav a {
		font-size: var(--step-1);
	}
	.bar nav a:hover,
	.bar nav a[aria-current='page'] {
		color: var(--ink);
	}
	.tools {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		flex: 0 0 auto;
	}
	.seg {
		display: inline-flex;
		border: 1px solid var(--rule);
		border-radius: var(--r-sm);
		background: var(--leaf);
		padding: 2px;
		gap: 2px;
	}
	.seg button {
		border: 0;
		border-radius: 4px;
		background: transparent;
		color: var(--ink-3);
		font: inherit;
		font-size: var(--step--1);
		font-weight: 500;
		padding: 0.22rem 0.45rem;
		cursor: pointer;
		transition:
			color 140ms var(--ease),
			background 140ms var(--ease);
	}
	:global(html[lang='bn']) .seg button {
		font-size: var(--step-0);
	}
	.seg button:hover {
		color: var(--ink);
	}
	.seg button[aria-pressed='true'] {
		background: var(--sink);
		color: var(--ink);
		box-shadow: var(--lift);
	}
	.iconbtn {
		display: inline-grid;
		place-items: center;
		width: 28px;
		height: 28px;
		border: 1px solid var(--rule);
		border-radius: var(--r-sm);
		background: var(--leaf);
		color: var(--graphite);
		cursor: pointer;
		transition: color 140ms var(--ease);
	}
	.iconbtn:hover {
		color: var(--ink);
	}
	.i-sun {
		display: none;
	}
	:global(html[data-theme='light']) .i-sun {
		display: block;
	}
	:global(html[data-theme='light']) .i-moon {
		display: none;
	}
	@media (max-width: 900px) {
		.cta {
			display: none;
		}
	}
	@media (max-width: 820px) {
		.bar nav {
			order: 3;
			width: 100%;
			margin-inline: 0;
			border-top: 1px solid var(--rule);
		}
		.bar nav ul {
			gap: 0 1rem;
			padding-block: 0.1rem 0.35rem;
		}
		.tools {
			margin-inline-start: auto;
		}
	}

	footer {
		border-top: 1px solid var(--rule);
		padding-block: 3rem 2rem;
		margin-top: var(--bay);
	}
	.fgrid {
		display: grid;
		gap: 2.5rem clamp(2rem, 5vw, 4rem);
		grid-template-columns: minmax(0, 1fr);
	}
	@media (min-width: 860px) {
		.fgrid {
			grid-template-columns: minmax(0, 5fr) minmax(0, 7fr);
		}
	}
	.fname {
		font-weight: 600;
		letter-spacing: -0.02em;
		margin-bottom: 0.35rem;
	}
	.fabout .small {
		max-width: 42ch;
	}
	.fcols {
		display: grid;
		gap: 1.75rem 1.5rem;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
	}
	.fcols h2 {
		font-size: var(--step--1);
		font-weight: 500;
		letter-spacing: 0.04em;
		color: var(--ink-3);
		margin-bottom: 0.7rem;
	}
	:global(html[lang='bn']) .fcols h2 {
		letter-spacing: 0;
		font-size: var(--step-0);
	}
	.fcols ul {
		list-style: none;
		display: flex;
		margin: 0;
		padding: 0;
		flex-direction: column;
		gap: 0.45rem;
	}
	.fcols a {
		color: var(--graphite);
		text-decoration: none;
		font-size: var(--step-0);
		transition: color 140ms var(--ease);
	}
	:global(html[lang='bn']) .fcols a {
		font-size: var(--step-1);
	}
	.fcols a:hover {
		color: var(--ink);
	}
	.legal {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem 2rem;
		margin-top: 2.5rem;
		padding-top: 1.1rem;
		border-top: 1px solid var(--rule);
	}
	.legal .fine {
		max-width: 76ch;
	}
</style>
