<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Tick02Icon, ArrowDown01Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import ConversionFlow from '$lib/components/ConversionFlow.svelte';
	import Money from '$lib/Money.svelte';

	// Facts about the product rather than customer counts, which we do not have and
	// would not print if we did.
	const facts = [
		{ value: '93%', label: 'Smaller, on average' },
		{ value: 'Under 60s', label: 'To first playable size' },
		{ value: '8K', label: 'Highest resolution in' },
		{ value: 'Per tenant', label: 'Isolated by the database' }
	];

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
			title: 'Up to 8K in, ladder out',
			body: 'Send whatever your users hand you. We work out the sizes their viewers need and make them, so you are not maintaining an encoding profile.'
		},
		{
			title: 'Playable in under a minute',
			body: 'The low sizes are made on ingest and the rest on first play, so a video is watchable long before every size exists.'
		},
		{
			title: 'Up to 95% smaller',
			body: 'A gigabyte becomes a few dozen megabytes. Your delivery bill drops, and so does the one your users on mobile data are paying.'
		},
		{
			title: 'Links only their viewers can use',
			body: 'Every play runs through a signed link that expires when you say. Pass it around and it stops working; a saved copy will not play.'
		},
		{
			title: 'One tenant cannot see another',
			body: 'Isolation is a database rule, not a WHERE clause somebody has to remember. A query with no filter still returns only that account.'
		},
		{
			title: 'One file, every device',
			body: 'HLS and DASH out of the same source, so an old Android, a new iPhone, a laptop and a TV all get something they can play.'
		}
	];

	const rates = [
		{ amount: 2, unit: 'per minute sent in', usd: '$0.016' },
		{ amount: 1.2, unit: 'per GB held, a month', usd: '$0.010' },
		{ amount: 0.35, unit: 'per GB watched in Bangladesh', usd: '$0.003' },
		{ amount: 1.2, unit: 'per GB watched anywhere else', usd: '$0.010' }
	];

	const included = [
		'Every size we make, from the one source',
		'HLS and DASH, signed links, expiring on your terms',
		'Webhooks, so you are told rather than polling',
		'As many API keys and accounts as you need',
		'No seat charge, no minimum, nothing to pay to leave'
	];

	const questions = [
		{
			q: 'Where does the video live, and who can watch it?',
			a: 'With us, playing from a signed link your product hands out, and you decide how long that link keeps working. The original upload is a separate thing — once the playable sizes exist, we can delete it for you.'
		},
		{
			q: 'How isolated is one account from another?',
			a: 'The database enforces it. Every tenant-scoped table has row-level security forced on, so a query with no filter still returns only that account’s rows.'
		},
		{
			q: 'Which formats can we send?',
			a: 'Whatever your users hand you. If a camera, phone or editor produced it, we will take it, and what comes back plays everywhere.'
		},
		{
			q: 'Do we have to poll?',
			a: 'No. Register a webhook and we tell you when an asset is playable or has failed. Deliveries are signed, so you can check a payload really came from us before trusting it.'
		},
		{
			q: 'What happens when we stop?',
			a: 'You export and leave. No notice period, no exit fee, and nothing charged for taking your own files out.'
		}
	];

	const uses = [
		'Course platforms, where a lecture has to open on a phone on mobile data',
		'Marketplaces and social products, where anybody can upload anything',
		'Agencies, where 4K client work goes out as a link that later stops working'
	];
</script>

<Seo
	title="Alchemist — video infrastructure for your product"
	description="An API for upload, encoding and delivery. Your users send video, we make every size their viewers need, and you hand out signed links. Usage pricing, isolated accounts, no seats."
/>

<main id="main" class="mx-auto max-w-[1100px] px-4 sm:px-6">
	<!-- One screen to open on. Everything a buyer has to check — what it does, how it
	     is integrated, what it costs — is below it, in one column, short. -->
	<section class="flex min-h-svh flex-col justify-center gap-10 py-14">
		<a href="/" class="mx-auto flex items-center gap-2 text-lg font-extrabold tracking-tight">
			<Logo size={26} />
			Alchemist
		</a>

		<div class="grid items-center gap-10 lg:grid-cols-[1.05fr_1fr] lg:gap-14">
			<div>
				<h1 class="big">Video infrastructure for your product.</h1>

				<p class="measure mt-5 text-dim">
					Your users upload video. We encode it, hold it, and stream it to their viewers behind
					links only they can hand out. You get an API and a key, not a project to run.
				</p>

				<div class="mt-7 flex flex-wrap items-center gap-2">
					<a href="/signup/" class="btn-solid">Get started</a>
					<a href="/contact/" class="btn">Talk to us</a>
				</div>

				<dl class="mt-9 grid max-w-md grid-cols-2 gap-x-8 gap-y-5">
					{#each facts as f (f.value)}
						<div>
							<dt class="label">{f.label}</dt>
							<dd class="num mt-1 text-xl">{f.value}</dd>
						</div>
					{/each}
				</dl>
			</div>

			<ConversionFlow />
		</div>

		<a href="#how" class="label mx-auto flex items-center gap-2 pt-2 hover:text-ink">
			What it does
			<HugeiconsIcon icon={ArrowDown01Icon} size={14} strokeWidth={2.4} />
		</a>
	</section>

	<div class="rule"></div>

	<section id="how" class="scroll-mt-6 py-16">
		<h2>Three calls, and you are integrated.</h2>
		<ol class="mt-8 grid gap-6 md:grid-cols-3">
			{#each steps as s (s.n)}
				<li>
					<span class="badge">{s.n}</span>
					<h3 class="mt-3">{s.title}</h3>
					<p class="sub mt-2">{s.body}</p>
				</li>
			{/each}
		</ol>
	</section>

	<div class="rule"></div>

	<section id="features" class="scroll-mt-6 py-16">
		<h2>What you get</h2>
		<div class="mt-8 grid gap-x-10 gap-y-7 md:grid-cols-2 lg:grid-cols-3">
			{#each features as f (f.title)}
				<div>
					<h3>{f.title}</h3>
					<p class="sub mt-2">{f.body}</p>
				</div>
			{/each}
		</div>
		<p class="sub mt-8">
			Built for: {uses.join(' · ')}
		</p>
	</section>

	<div class="rule"></div>

	<section id="pricing" class="scroll-mt-6 py-16">
		<h2>You pay for three things.</h2>
		<p class="sub mt-2">
			Per unit, with no plans to choose between and nobody to negotiate with. Rates are not final
			until launch.
		</p>

		<dl class="mt-8 grid gap-x-10 gap-y-6 sm:grid-cols-2 lg:grid-cols-4">
			{#each rates as r (r.unit)}
				<div>
					<dt class="label">{r.unit}</dt>
					<dd class="num mt-1 text-2xl"><Money amount={r.amount} /></dd>
					<p class="sub mt-1">about {r.usd}</p>
				</div>
			{/each}
		</dl>

		<ul class="mt-10 grid gap-3 sm:grid-cols-2">
			{#each included as row (row)}
				<li class="flex items-start gap-2.5 text-sm">
					<HugeiconsIcon
						icon={Tick02Icon}
						size={15}
						strokeWidth={2.4}
						class="mt-1 flex-none text-ink"
					/>
					<span>{row}</span>
				</li>
			{/each}
		</ul>
	</section>

	<div class="rule"></div>

	<section id="faq" class="scroll-mt-6 py-16">
		<h2>Questions</h2>
		<div class="mt-6 divide-y divide-sunk border-y border-sunk">
			{#each questions as q (q.q)}
				<details class="group">
					<summary
						class="title flex cursor-pointer list-none items-center justify-between gap-4 py-4 transition-colors hover:text-dim"
					>
						{q.q}
						<HugeiconsIcon
							icon={ArrowDown01Icon}
							size={16}
							strokeWidth={2}
							class="chevron flex-none text-faint"
						/>
					</summary>
					<p class="sub measure pb-5">{q.a}</p>
				</details>
			{/each}
		</div>

		<div class="mt-12 flex flex-wrap items-center gap-3">
			<a href="/signup/" class="btn-solid">Get started</a>
			<a href="/contact/" class="btn">Talk to us</a>
			<p class="sub">Free while you build. No card.</p>
		</div>
	</section>
</main>
