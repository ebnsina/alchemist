<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Tick02Icon } from '@hugeicons/core-free-icons';
	import { reveal } from '$lib/utils/reveal';
	import Money from '$lib/Money.svelte';

	// Three meters, not three tiers. A platform that charges per seat punishes its
	// customer for growing their own product, and the costs here are per unit
	// anyway: minutes encoded, gigabytes held, gigabytes delivered.
	const meters = [
		{
			title: 'Preparing it',
			amount: 2,
			unit: 'for each minute of video sent in',
			usd: '$0.016',
			body: 'Charged once, on the length of the source. We make several sizes of it and you are not charged per size — an hour of lecture is an hour.'
		},
		{
			title: 'Keeping it',
			amount: 1.2,
			unit: 'per GB, per month',
			usd: '$0.010',
			body: 'Counted on what we actually hold. The larger sizes only exist for videos somebody has watched, so the bill follows real use.'
		},
		{
			title: 'Delivering it',
			amount: 0.35,
			unit: 'per GB watched inside Bangladesh',
			usd: '$0.003',
			body: 'Reaching a viewer in Bangladesh costs us a fraction of reaching one abroad. One flat rate would mean your Bangladeshi users quietly paying the difference.',
			second: { amount: 1.2, unit: 'per GB watched anywhere else', usd: '$0.010' }
		}
	];

	const included = [
		'Every size we make, from the one source',
		'HLS and DASH, signed links, expiring on your terms',
		'Webhooks, so you are told rather than polling',
		'As many API keys and accounts as you need',
		'No seat charge, no minimum, nothing to pay to leave'
	];
</script>

<section id="pricing" class="screen px-4 sm:px-6" use:reveal>
	<div class="mx-auto max-w-2xl text-center">
		<p class="label text-ink">Pricing</p>
		<h2 class="mt-3 text-3xl leading-[1.32] font-semibold tracking-tight sm:text-4xl">
			You pay for three things.
			<span class="text-ink">Nothing else.</span>
		</h2>
		<p class="mt-4 text-sm text-dim">
			Per unit, with no plans to choose between and nobody to negotiate with. Rates are not final
			until launch.
		</p>
	</div>

	<div class="mt-12 grid gap-4 md:grid-cols-3">
		{#each meters as m, i (m.title)}
			<article class="card flex flex-col px-6 py-8">
				<h3 class="label text-dim">{m.title}</h3>
				<p class="mt-4 font-mono text-4xl font-medium tracking-tight tabular-nums">
					<Money amount={m.amount} />
				</p>
				<p class="mt-1 text-sm text-dim">{m.unit} · about {m.usd}</p>

				{#if m.second}
					<p class="mt-4 font-mono text-2xl font-medium tracking-tight tabular-nums">
						<Money amount={m.second.amount} />
					</p>
					<p class="mt-1 text-sm text-dim">{m.second.unit} · about {m.second.usd}</p>
				{/if}

				<p class="mt-5 flex-1 text-sm leading-[1.75] text-dim">{m.body}</p>
			</article>
		{/each}
	</div>

	<div class="card mt-4 p-8">
		<h3 class="text-sm font-semibold">In every account, at no extra charge</h3>
		<ul class="mt-4 grid gap-3 sm:grid-cols-2">
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
		<div class="mt-7 flex flex-wrap items-center gap-3">
			<a href="/signup/" class="btn-solid">Start free</a>
			<a href="/contact/" class="btn">Ask about volume</a>
		</div>
	</div>
</section>
