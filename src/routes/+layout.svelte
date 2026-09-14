<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import T from '$lib/T.svelte';

	let { children } = $props();

	const nav = [
		{ href: '/edtech/', en: 'Course platforms', bn: 'কোর্স প্ল্যাটফর্ম' },
		{ href: '/media/', en: 'News and media', bn: 'সংবাদ ও মিডিয়া' },
		{ href: '/pricing/', en: 'Pricing', bn: 'মূল্য' },
		{ href: '/docs/', en: 'For developers', bn: 'ডেভেলপারদের জন্য' },
		{ href: '/about/', en: 'About', bn: 'পরিচিতি' }
	];
	const here = $derived(page.url.pathname);
</script>

<a class="skip" href="#main"><T en="Skip to content" bn="মূল অংশে যান" /></a>

<header>
	<div class="shell bar">
		<a class="brand" href="/">
			<svg width="22" height="26" viewBox="0 0 22 26" aria-hidden="true" focusable="false">
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
				<button type="button" data-set-lang="en" aria-pressed="true" lang="en">English</button>
				<button type="button" data-set-lang="bn" aria-pressed="false" lang="bn">বাংলা</button>
			</div>
			<button type="button" class="iconbtn" data-toggle-theme aria-label="Switch colour theme">
				<svg class="i-sun" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" aria-hidden="true" focusable="false">
					<circle cx="12" cy="12" r="4.6" />
					<path d="M12 1.8v2.4M12 19.8v2.4M2.6 12h2.4M19 12h2.4M5.3 5.3l1.7 1.7M17 17l1.7 1.7M18.7 5.3 17 7M7 17l-1.7 1.7" />
				</svg>
				<svg class="i-moon" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
					<path d="M20.4 14.5A8.6 8.6 0 0 1 9.5 3.6a8.8 8.8 0 1 0 10.9 10.9z" />
				</svg>
			</button>
		</div>
	</div>
</header>

<main id="main">{@render children()}</main>

<footer>
	<div class="shell">
		<div class="fgrid">
			<div>
				<p class="fname">Alchemist</p>
				<p class="small">
					<T
						as="span"
						en="Video hosting for apps and websites in Bangladesh. Kept inside the country, and made small enough for mobile data."
						bn="বাংলাদেশের অ্যাপ আর ওয়েবসাইটের জন্য ভিডিও হোস্টিং। দেশের ভেতরে রাখা, আর মোবাইল ডেটার উপযোগী করে ছোট করা।"
					/>
				</p>
			</div>
			<nav aria-label="Footer">
				<ul class="flinks">
					{#each nav as n (n.href)}
						<li><a href={n.href}><T en={n.en} bn={n.bn} /></a></li>
					{/each}
				</ul>
			</nav>
		</div>
		<p class="fine disclaim">
			<T
				as="span"
				en="Alchemist is in development and not yet generally available. Nothing here is measured from production traffic, and the prices shown are indicative — none has been set."
				bn="Alchemist এখনও তৈরির পর্যায়ে, সবার জন্য উন্মুক্ত নয়। এখানকার কোনো সংখ্যা চালু সার্ভিসের মাপা ফলাফল নয়, আর দেখানো দামগুলো সম্ভাব্য — কোনোটিই এখনও নির্ধারিত হয়নি।"
			/>
		</p>
	</div>
</footer>

<style>
	header {
		position: sticky;
		top: 0;
		z-index: 10;
		background: color-mix(in srgb, var(--paper) 92%, transparent);
		backdrop-filter: blur(8px);
		border-bottom: 1px solid var(--rule);
	}
	.bar {
		display: flex;
		align-items: center;
		gap: 1rem;
		min-height: 62px;
		flex-wrap: wrap;
	}
	.brand {
		display: inline-flex;
		align-items: center;
		gap: 0.55rem;
		font-weight: 600;
		font-size: var(--step-2);
		letter-spacing: -0.03em;
		text-decoration: none;
	}
	.brand svg {
		color: var(--brass);
		flex: 0 0 auto;
	}
	nav {
		margin-inline-start: auto;
	}
	ul {
		list-style: none;
		display: flex;
		flex-wrap: wrap;
		margin: 0;
		padding: 0;
		gap: 0 1.35rem;
	}
	nav a {
		position: relative;
		display: block;
		padding: 0.5rem 0;
		color: var(--graphite);
		text-decoration: none;
		font-size: var(--step-0);
	}
	nav a::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0.28rem;
		height: 1px;
		background: currentColor;
		transform: scaleX(0);
		transform-origin: left;
		transition: transform 170ms var(--ease);
	}
	nav a:hover {
		color: var(--ink);
	}
	nav a:hover::after {
		transform: scaleX(1);
	}
	nav a[aria-current='page'] {
		color: var(--ink);
	}
	nav a[aria-current='page']::after {
		background: var(--brass);
		transform: scaleX(1);
	}
	.tools {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.seg {
		display: inline-flex;
		border: 1px solid var(--rule-firm);
		border-radius: 3px;
		overflow: hidden;
	}
	.seg button {
		border: 0;
		background: transparent;
		color: var(--graphite);
		font: inherit;
		font-size: var(--step--1);
		padding: 0.3rem 0.6rem;
		cursor: pointer;
		transition: color 150ms var(--ease);
	}
	.seg button:hover {
		color: var(--ink);
	}
	.seg button[aria-pressed='true'] {
		background: var(--ink);
		color: var(--paper);
	}
	.iconbtn {
		display: inline-grid;
		place-items: center;
		width: 31px;
		height: 31px;
		border: 1px solid var(--rule-firm);
		border-radius: 3px;
		background: transparent;
		color: var(--graphite);
		cursor: pointer;
		transition: color 150ms var(--ease);
	}
	.iconbtn:hover {
		color: var(--ink);
	}
	/* The button shows the theme you are in, so the icon follows the theme. */
	.i-moon {
		display: none;
	}
	:global(html[data-theme='dark']) .i-sun {
		display: none;
	}
	:global(html[data-theme='dark']) .i-moon {
		display: block;
	}
	@media (prefers-color-scheme: dark) {
		:global(html:not([data-theme='light'])) .i-sun {
			display: none;
		}
		:global(html:not([data-theme='light'])) .i-moon {
			display: block;
		}
	}
	@media (max-width: 820px) {
		nav {
			order: 3;
			width: 100%;
			margin-inline-start: 0;
			border-top: 1px solid var(--rule);
		}
		nav ul {
			gap: 0 1.1rem;
			padding-block: 0.15rem 0.4rem;
		}
		.tools {
			margin-inline-start: auto;
		}
	}
	footer {
		border-top: 1px solid var(--rule);
		padding-block: 3rem 3.5rem;
		margin-top: var(--bay);
	}
	.fgrid {
		display: flex;
		gap: 2.5rem;
		flex-wrap: wrap;
		justify-content: space-between;
	}
	.fname {
		font-weight: 600;
		margin-bottom: 0.35rem;
	}
	.flinks {
		flex-direction: column;
		gap: 0.35rem;
	}
	.flinks a {
		color: var(--graphite);
		text-decoration: none;
		font-size: var(--step-0);
	}
	.flinks a:hover {
		color: var(--ink);
	}
	.disclaim {
		margin-top: 2.5rem;
		padding-top: 1.2rem;
		border-top: 1px solid var(--rule);
	}
</style>
