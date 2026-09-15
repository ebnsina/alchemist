<script lang="ts">
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

	// The whole price list. Three meters, no plans, and the rate is the pitch — a
	// page that hides them is a page that intends to negotiate.
	const rates = [
		{ amount: 2, unit: 'per minute sent in' },
		{ amount: 1.2, unit: 'per GB held, a month' },
		{ amount: 0.35, unit: 'per GB watched in Bangladesh' }
	];
</script>

<Seo
	title="Alchemist — video infrastructure for your product"
	description="An API for upload, encoding and delivery. Your users send video, we make every size their viewers need, and you hand out signed links. Usage pricing, isolated accounts, no seats."
/>

<!-- One screen. Everything that used to be nine sections is here because nine
     sections was nine ways of saying the same thing. -->
<main id="main" class="mx-auto flex min-h-svh max-w-[1100px] flex-col justify-center gap-10 px-4 py-14 sm:px-6">
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

	<div class="rule"></div>

	<div class="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
		<dl class="flex flex-wrap gap-x-10 gap-y-4">
			{#each rates as r (r.unit)}
				<div>
					<dt class="label">{r.unit}</dt>
					<dd class="num mt-1 text-lg"><Money amount={r.amount} /></dd>
				</div>
			{/each}
		</dl>

		<p class="sub sm:text-right">
			No seats, no minimum, nothing to pay to leave.<br />
			Rates are not final until launch.
		</p>
	</div>
</main>
