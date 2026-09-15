<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Tick02Icon,
		ArrowDown01Icon,
		Layers01Icon,
		Timer02Icon,
		ArrowShrinkIcon,
		SecurityLockIcon,
		DatabaseLockedIcon,
		MonitorSmartphoneIcon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import LadderDemo from '$lib/components/LadderDemo.svelte';
	import Money from '$lib/Money.svelte';

	const steps = [
		{
			n: '1',
			title: 'Your user sends a file',
			body: 'Ask us for an upload target and let the browser put the bytes straight into storage. The file never passes through your servers or ours.'
		},
		{
			n: '2',
			title: 'We do the work',
			body: 'Probe, ladder, encode in parallel chunks, package. A webhook tells you when it is playable, so nothing in your code sits and polls.'
		},
		{
			n: '3',
			title: 'You hand out a link',
			body: 'Ask for the asset and play the links it returns. Signed, expiring on your terms, and they work on anything.'
		}
	];

	const features = [
		{
			icon: Layers01Icon,
			title: 'Up to 8K in, ladder out',
			body: 'Send whatever your users hand you. We work out the sizes their viewers need and make them, so you are not maintaining an encoding profile.'
		},
		{
			icon: Timer02Icon,
			title: 'Playable in under a minute',
			body: 'The low sizes are made on ingest and the rest on first play, so a video is watchable long before every size exists.'
		},
		{
			icon: ArrowShrinkIcon,
			title: 'Up to 95% smaller',
			body: 'A gigabyte becomes a few dozen megabytes. Your delivery bill drops, and so does the one your users on mobile data are paying.'
		},
		{
			icon: SecurityLockIcon,
			title: 'Links only their viewers can use',
			body: 'Every play runs through a signed link that expires when you say. Pass it around and it stops working; a saved copy will not play.'
		},
		{
			icon: DatabaseLockedIcon,
			title: 'One tenant cannot see another',
			body: 'Isolation is a database rule, not a WHERE clause somebody has to remember. A query with no filter still returns only that account.'
		},
		{
			icon: MonitorSmartphoneIcon,
			title: 'One file, every device',
			body: 'HLS and DASH out of the same source, so an old Android, a new iPhone, a laptop and a TV all get something they can play.'
		}
	];

	const tiers = [
		{
			name: 'Starter',
			who: 'Trying it, or running something small',
			price: 'Free',
			unit: 'while you build',
			features: [
				'5 hours of video in, a month',
				'25 GB held for you',
				'50 GB watched',
				'Every size, signed links, webhooks'
			],
			rates: false,
			cta: 'Start free',
			href: '/signup/',
			lead: false
		},
		{
			name: 'Studio',
			who: 'A product with real viewers',
			price: 'Pay per unit',
			unit: 'no plan to outgrow',
			features: [
				'Nothing capped, nothing to upgrade to',
				'As many keys and teammates as you need',
				'Connect a bucket you already own'
			],
			rates: true,
			cta: 'Start free, then pay per unit',
			href: '/signup/',
			lead: true
		},
		{
			name: 'Volume',
			who: 'Thousands of hours, or a private deployment',
			price: 'Talk to us',
			unit: 'rates by arrangement',
			features: [
				'Lower rates at volume',
				'Your own encoding profile',
				'Run it on your own machines',
				'A person to talk to, not a queue'
			],
			rates: false,
			cta: 'Talk to us',
			href: '/contact/',
			lead: false
		}
	];

	const rates = [
		{ amount: 2, unit: 'A minute sent in', usd: '$0.016' },
		{ amount: 1.2, unit: 'A GB held, a month', usd: '$0.010' },
		{ amount: 0.35, unit: 'A GB watched in BD', usd: '$0.003' },
		{ amount: 1.2, unit: 'A GB watched elsewhere', usd: '$0.010' }
	];

	const included = [
		'Every size we make, from the one source',
		'HLS and DASH, signed links, expiring on your terms',
		'Webhooks, so you are told rather than polling',
		'As many API keys and accounts as you need',
		'No seat charge, no minimum, nothing to pay to leave'
	];

	// Plain language, and answers to what a person actually worries about. See
	// CLAUDE.md: no jargon here, and nothing that reads like release notes.
	const questions = [
		{
			q: 'Do I need to know anything about video?',
			a: 'No. Send us the file your user uploaded and ask for a link back. Everything in between — the sizes, the formats, which one a given phone should get — is ours to worry about.'
		},
		{
			q: 'How long until my user can watch it?',
			a: 'Usually under a minute. We make the smaller sizes first so there is something watchable straight away, and the sharper ones finish in the background while people are already watching.'
		},
		{
			q: 'Will it play on my viewers\u2019 phones?',
			a: 'Yes — old Androids, new iPhones, laptops, TVs. We make several sizes from the one file, and the player picks whichever suits the connection at that moment. Nobody has to choose a quality.'
		},
		{
			q: 'Can someone share a link and let the whole world watch?',
			a: 'Not for long. Links stop working after a time you set, so a forwarded one is dead by the time it spreads. A saved copy of the link will not play either.'
		},
		{
			q: 'Could another customer ever see my videos?',
			a: 'No. The separation is enforced by the database itself, not by our code remembering to filter. Even a query written wrongly returns only your own account\u2019s videos.'
		},
		{
			q: 'What does it actually cost me?',
			a: 'You pay for three things: video sent in, video held for you, and video watched. Nothing for seats, nothing as a minimum. Start free while you are building and only start paying when real people are watching.'
		},
		{
			q: 'What if I want to leave?',
			a: 'Export everything and go. No notice period, no exit fee, and we do not charge you to take your own files out.'
		}
	];

	const uses = [
		{
			who: 'Course platforms',
			why: 'A lecture has to open on a phone on mobile data, in a village, on a plan that charges by the megabyte.'
		},
		{
			who: 'Marketplaces and social products',
			why: 'Anybody can upload anything, in any format, at any size, and it still has to play for everybody else.'
		},
		{
			who: 'Agencies',
			why: '4K client work goes out as a link that stops working when the review window closes.'
		}
	];
</script>

<Seo
	title="Alchemist — video infrastructure for your product"
	description="An API for upload, encoding and delivery. Your users send video, we make every size their viewers need, and you hand out signed links. Usage pricing, isolated accounts, no seats."
/>

<!-- overflow-x: clip, because the hero glow is a blurred pseudo-element that bleeds
     past its box by design. Decoration must never be able to scroll the page. -->
<main id="main" class="mx-auto max-w-[1160px] overflow-x-clip px-5 sm:px-8">
	<!-- One screen to open on, then an ordinary column. Everything a buyer has to check
	     — what it does, how it is integrated, what it costs — is below it, short. -->
	<section class="flex min-h-svh flex-col justify-center gap-10 py-20">
		<a href="/" class="flex items-center gap-2.5 text-[17px] font-semibold tracking-tight">
			<Logo size={26} />
			Alchemist
		</a>

		<div class="grid items-center gap-12 lg:grid-cols-[1.05fr_1fr] lg:gap-16">
			<div>
				<h1 class="big">Heavy video in.<br />Light video out.</h1>

				<p class="lead measure mt-6">
					Your users upload video. We encode it, hold it, and stream it to their viewers behind
					links only they can hand out. You get an API and a key, not a project to run.
				</p>

				<div class="mt-8 flex flex-wrap items-center gap-3">
					<a href="/signup/" class="btn-solid">Get started</a>
					<a href="/contact/" class="btn">Talk to us</a>
				</div>
			</div>

			<!-- Tilted and lifted off a lime wash, so the right side reads as an object
			     on the page rather than a second column of text. -->
			<div class="hero-art">
				<div class="hero-art__card">
					<LadderDemo />
				</div>
			</div>
		</div>
	</section>

	<div class="rule"></div>

	<section id="how" class="scroll-mt-8 py-24">
		<p class="label">How it works</p>
		<h2 class="mt-5 max-w-[16ch]">Three calls, and you are integrated.</h2>

		<ol class="mt-14 grid gap-12 md:grid-cols-3 md:gap-10">
			{#each steps as s (s.n)}
				<li>
					<span class="badge">{s.n}</span>
					<h3 class="mt-5">{s.title}</h3>
					<p class="sub measure mt-3">{s.body}</p>
				</li>
			{/each}
		</ol>
	</section>

	<div class="rule"></div>

	<section id="features" class="scroll-mt-8 py-24">
		<p class="label">What you get</p>
		<h2 class="mt-5 max-w-[18ch]">Everything a video product needs, already running.</h2>

		<div class="mt-14 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each features as f (f.title)}
				<div class="card">
					<div class="flex items-start justify-between gap-4">
						<h3>{f.title}</h3>
						<HugeiconsIcon
							icon={f.icon}
							size={22}
							strokeWidth={1.8}
							class="mt-0.5 flex-none text-accent"
						/>
					</div>
					<p class="sub mt-3">{f.body}</p>
				</div>
			{/each}
		</div>
	</section>

	<div class="rule"></div>

	<section id="built-for" class="scroll-mt-8 py-24">
		<p class="label">Built for</p>
		<h2 class="mt-5 max-w-[18ch]">Products where video is not the product.</h2>

		<div class="mt-14 grid gap-4 md:grid-cols-3">
			{#each uses as u (u.who)}
				<div class="card">
					<h3>{u.who}</h3>
					<p class="sub mt-3">{u.why}</p>
				</div>
			{/each}
		</div>
	</section>

	<div class="rule"></div>

	<section id="pricing" class="scroll-mt-8 py-24">
		<p class="label">Pricing</p>
		<h2 class="mt-5 max-w-[18ch]">Start free. Then pay for what you use.</h2>
		<p class="sub measure mt-4">
			No seats, no minimum, and nothing to pay to leave. Rates are not final until launch.
		</p>

		<div class="mt-14 grid gap-4 lg:grid-cols-3">
			{#each tiers as t (t.name)}
				<div class="card flex flex-col" class:tier--lead={t.lead}>
					{#if t.lead}
						<span class="chip chip-on self-start">Most take this</span>
					{/if}
					<h3 class="mt-3">{t.name}</h3>
					<p class="sub mt-1">{t.who}</p>

					<p class="num mt-6 text-[28px] leading-none" class:text-accent={t.lead}>{t.price}</p>
					<p class="mono mt-1.5">{t.unit}</p>

					<ul class="mt-6 grid gap-2.5">
						{#each t.features as f (f)}
							<li class="flex items-start gap-2.5 text-[15px] font-normal">
								<HugeiconsIcon
									icon={Tick02Icon}
									size={15}
									strokeWidth={2.4}
									class="mt-1 flex-none text-accent"
								/>
								<span>{f}</span>
							</li>
						{/each}
					</ul>

					{#if t.rates}
						<!-- The rates belong to the tier that charges them, not to a table
						     underneath that the reader has to carry back up here. -->
						<dl class="mt-6 grid gap-2.5 border-t border-sunk pt-5">
							{#each rates as r (r.unit)}
								<div class="flex items-baseline justify-between gap-3">
									<dt class="sub">{r.unit}</dt>
									<dd class="num flex-none text-[15px]">
										<Money amount={r.amount} />
										<span class="mono ml-1">≈{r.usd}</span>
									</dd>
								</div>
							{/each}
						</dl>
					{/if}

					<div class="mt-auto pt-7">
						<a href={t.href} class="w-full {t.lead ? 'btn-solid' : 'btn'}">{t.cta}</a>
					</div>
				</div>
			{/each}
		</div>

		<ul class="mt-12 grid gap-3.5 sm:grid-cols-2">
			{#each included as row (row)}
				<li class="flex items-start gap-3 text-[15px] font-normal">
					<HugeiconsIcon
						icon={Tick02Icon}
						size={16}
						strokeWidth={2.4}
						class="mt-1 flex-none text-accent"
					/>
					<span>{row}</span>
				</li>
			{/each}
		</ul>
	</section>

	<div class="rule"></div>

	<section id="faq" class="scroll-mt-8 py-24">
		<p class="label">Questions</p>
		<h2 class="mt-5 max-w-[16ch]">The things people ask first.</h2>

		<div class="mt-12 divide-y divide-sunk border-y border-sunk">
			{#each questions as q (q.q)}
				<details class="group">
					<summary
						class="title flex cursor-pointer list-none items-center justify-between gap-6 py-5 transition-colors hover:text-accent"
					>
						{q.q}
						<HugeiconsIcon
							icon={ArrowDown01Icon}
							size={16}
							strokeWidth={2}
							class="chevron flex-none text-faint"
						/>
					</summary>
					<p class="sub measure pb-6">{q.a}</p>
				</details>
			{/each}
		</div>

		<div class="mt-14 flex flex-wrap items-center gap-4">
			<a href="/signup/" class="btn-solid">Get started</a>
			<a href="/contact/" class="btn">Talk to us</a>
			<p class="mono">Free while you build. No card.</p>
		</div>
	</section>
</main>
