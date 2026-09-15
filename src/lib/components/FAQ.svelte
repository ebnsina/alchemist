<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { ArrowDown01Icon } from '@hugeicons/core-free-icons';

	const qs = [
		{
			qen: 'Where does the video live, and who can watch it?',
			aen: 'With us, playing from a signed link your product hands out, and you decide how long that link keeps working. The original upload is a separate thing — once the playable sizes exist, we can delete it for you.'
		},
		{
			qen: 'How isolated is one account from another?',
			aen: 'The database enforces it. Every tenant-scoped table has row-level security forced on, so a query with no filter still returns only that account’s rows. It is not a WHERE clause somebody has to remember to write.'
		},
		{
			qen: 'Will it look worse?',
			aen: 'Not in any way a viewer would notice. We take out the parts of a video file that eyes were never going to use, which is where nearly all the weight is hiding.'
		},
		{
			qen: 'Which formats can we send?',
			aen: 'Whatever your users hand you. If a camera, phone or editor produced it, we will take it, and what comes back plays everywhere.'
		},
		{
			qen: 'Do we have to poll?',
			aen: 'No. Register a webhook and we tell you when an asset is playable or has failed. Deliveries are signed, so you can check a payload really came from us before trusting it.'
		},
		{
			qen: 'What happens when we stop?',
			aen: 'You export and leave. No notice period, no exit fee, and nothing charged for taking your own files out.'
		}
	];
</script>

<section id="faq" class="screen px-4 sm:px-6">
	<div class="text-center">
		<p class="label text-ink">
			Questions
		</p>
		<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
			Got questions?
			<span class="text-ink">We've got answers.</span>
		</h2>
	</div>

	<div class="mt-10 divide-y divide-sunk border-y border-sunk">
		{#each qs as q (q.qen)}
			<details class="group">
				<summary
					class="flex cursor-pointer list-none items-center justify-between gap-4 py-4 font-medium transition-colors hover:text-ink"
				>
					{q.qen}
					<HugeiconsIcon
						icon={ArrowDown01Icon}
						size={17}
						strokeWidth={2}
						class="chevron flex-none text-dim"
					/>
				</summary>
				<p class="pb-5 text-sm leading-relaxed text-dim">{q.aen}</p>
			</details>
		{/each}
	</div>
</section>
