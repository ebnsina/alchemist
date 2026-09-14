<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import Icon from '$lib/Icon.svelte';
	import T from '$lib/T.svelte';

	let { children } = $props();

	const nav = [
		{ href: '/edtech/', en: 'For edtech', bn: 'এডটেকের জন্য' },
		{ href: '/media/', en: 'For media', bn: 'মিডিয়ার জন্য' },
		{ href: '/pricing/', en: 'Pricing', bn: 'মূল্য' },
		{ href: '/docs/', en: 'Docs', bn: 'ডকুমেন্টেশন' },
		{ href: '/about/', en: 'About', bn: 'পরিচিতি' }
	];
	const here = $derived(page.url.pathname);
</script>

<a class="skip" href="#main"><T en="Skip to content" bn="মূল অংশে যান" /></a>

<header>
	<div class="wrap bar">
		<a class="brand" href="/" aria-label="Alchemist — home">
			<svg width="26" height="26" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
				<path
					d="M9 2.8h6M10 2.8v5.4L4.9 17.6a2.6 2.6 0 0 0 2.3 3.9h9.6a2.6 2.6 0 0 0 2.3-3.9L14 8.2V2.8"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
				<path d="M7.1 14.4h9.8l1.9 3.4a2.6 2.6 0 0 1-2.3 3.7H7.5a2.6 2.6 0 0 1-2.3-3.7z" fill="currentColor" opacity=".85" />
			</svg>
			<span>Alchemist</span>
		</a>

		<details class="nav">
			<summary><T en="Menu" bn="মেনু" /></summary>
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
		</details>

		<div class="tools">
			<div class="seg" role="group" aria-label="Language / ভাষা">
				<button type="button" data-set-lang="en" aria-pressed="true" lang="en">EN</button>
				<button type="button" data-set-lang="bn" aria-pressed="false" lang="bn">বাংলা</button>
			</div>
			<button
				type="button"
				class="iconbtn"
				data-toggle-theme
				aria-label="Switch colour theme"
				title="Switch colour theme"
			>
				<span class="t-dark"><Icon name="sun" size={19} /></span>
				<span class="t-light"><Icon name="moon" size={19} /></span>
			</button>
		</div>
	</div>
</header>

<main id="main">{@render children()}</main>

<footer>
	<div class="wrap">
		<div class="fgrid">
			<div>
				<p class="fbrand">Alchemist</p>
				<p class="fsmall">
					<T
						as="span"
						en="Video transcoding and delivery, sold as an API. Hosted inside BDIX, built for metered mobile."
						bn="ভিডিও ট্রান্সকোডিং ও ডেলিভারি, একটি API হিসেবে। BDIX-এর ভেতরে হোস্ট করা, মিটারড মোবাইলের কথা মাথায় রেখে তৈরি।"
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
		<p class="fsmall disclaim">
			<T
				as="span"
				en="Alchemist is in development and not yet generally available. Nothing on this site is a measurement of production traffic, and prices shown are placeholders."
				bn="Alchemist এখনও তৈরির পর্যায়ে আছে, সবার জন্য উন্মুক্ত নয়। এই সাইটের কোনো সংখ্যা চালু সার্ভিসের মাপা ফলাফল নয়, আর দেখানো দামগুলো অনুমান মাত্র।"
			/>
		</p>
	</div>
</footer>

<style>
	header {
		position: sticky;
		top: 0;
		z-index: 10;
		background: color-mix(in srgb, var(--bg) 88%, transparent);
		backdrop-filter: saturate(140%) blur(10px);
		border-bottom: 1px solid var(--line);
	}
	.bar {
		display: flex;
		align-items: center;
		gap: 1rem;
		min-height: 60px;
		flex-wrap: wrap;
	}
	.brand {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 600;
		font-size: 1.05rem;
		color: var(--ink);
		text-decoration: none;
		letter-spacing: -0.02em;
	}
	.brand svg {
		color: var(--brass);
		flex: 0 0 auto;
	}
	.nav {
		margin-inline-start: auto;
	}
	.nav summary {
		display: none;
		list-style: none;
		cursor: pointer;
		border: 1px solid var(--line);
		border-radius: 8px;
		padding: 0.35rem 0.7rem;
		font-size: 0.85rem;
	}
	.nav summary::-webkit-details-marker {
		display: none;
	}
	ul {
		list-style: none;
		display: flex;
		gap: 0.25rem;
		margin: 0;
		padding: 0;
	}
	nav a {
		display: block;
		padding: 0.4rem 0.7rem;
		border-radius: 8px;
		color: var(--muted);
		text-decoration: none;
		font-size: 0.93rem;
		font-weight: 500;
	}
	nav a:hover {
		color: var(--ink);
		background: var(--raise2);
	}
	nav a[aria-current='page'] {
		color: var(--brass);
	}
	.tools {
		display: flex;
		align-items: center;
		gap: 0.4rem;
	}
	.seg {
		display: inline-flex;
		border: 1px solid var(--line);
		border-radius: 999px;
		overflow: hidden;
	}
	.seg button {
		border: 0;
		background: transparent;
		color: var(--muted);
		font: inherit;
		font-size: 0.8rem;
		font-weight: 600;
		padding: 0.3rem 0.65rem;
		cursor: pointer;
	}
	.seg button[aria-pressed='true'] {
		background: var(--brass-soft);
		color: var(--brass);
	}
	.iconbtn {
		display: inline-grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border: 1px solid var(--line);
		border-radius: 999px;
		background: transparent;
		color: var(--muted);
		cursor: pointer;
	}
	.iconbtn:hover {
		color: var(--ink);
	}
	/* The theme button shows the icon for the theme you would switch to. */
	:global(html[data-theme='light']) .t-dark,
	.t-light {
		display: none;
	}
	:global(html[data-theme='light']) .t-light {
		display: block;
	}
	@media (prefers-color-scheme: light) {
		:global(html:not([data-theme='dark'])) .t-dark {
			display: none;
		}
		:global(html:not([data-theme='dark'])) .t-light {
			display: block;
		}
	}
	@media (max-width: 760px) {
		.nav {
			order: 3;
			width: 100%;
			margin-inline-start: 0;
		}
		.nav summary {
			display: inline-block;
			margin-block: 0.1rem 0.4rem;
		}
		.nav ul {
			flex-direction: column;
			gap: 0;
			padding-bottom: 0.5rem;
		}
		nav a {
			padding: 0.6rem 0.2rem;
			font-size: 1rem;
		}
		.tools {
			margin-inline-start: auto;
		}
	}
	footer {
		border-top: 1px solid var(--line);
		padding-block: 2.5rem 3rem;
		margin-top: 2rem;
	}
	.fgrid {
		display: flex;
		gap: 2rem;
		flex-wrap: wrap;
		justify-content: space-between;
	}
	.fbrand {
		font-weight: 600;
		margin: 0 0 0.3rem;
	}
	.fsmall {
		color: var(--muted);
		font-size: 0.88rem;
		max-width: 46ch;
	}
	.flinks {
		flex-direction: column;
		gap: 0.3rem;
	}
	.flinks a {
		color: var(--muted);
		text-decoration: none;
		font-size: 0.9rem;
	}
	.flinks a:hover {
		color: var(--ink);
	}
	.disclaim {
		margin-top: 2rem;
		padding-top: 1.2rem;
		border-top: 1px solid var(--line);
		max-width: 72ch;
	}
</style>
