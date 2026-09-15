<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Mortarboard01Icon, PlaySquareIcon, StarIcon } from '@hugeicons/core-free-icons';

	// The reference put three invented people with names, roles, avatars and star
	// ratings here. Same grid, same cards, same reveal — but they describe the work
	// rather than pretending someone said something.
	const uses = [
		{
			icon: Mortarboard01Icon,
			ten: 'Course platforms',
			ben: 'Your teachers upload an hour of screen recording and your students open it on a phone, on mobile data, without waiting. Each school is its own account.'
		},
		{
			icon: PlaySquareIcon,
			ten: 'Marketplaces and social products',
			ben: 'Anybody can upload, so nothing can be assumed about the file. We take what your users hand you and return something that plays.'
		},
		{
			icon: StarIcon,
			ten: 'Agencies and internal tools',
			ben: 'Client work in 4K, delivered as a link that stops working when the engagement does, without anybody downloading a 40 GB master.'
		}
	];
</script>

<section class="screen px-4 sm:px-6">
	<div class="mx-auto max-w-2xl text-center">
		<p class="label text-ink">
			Who it's for
		</p>
		<h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl leading-[1.32]">
			Whatever your users upload.
		</h2>
	</div>

	<div class="mt-14 grid gap-5 md:grid-cols-3">
		{#each uses as u, i (u.ten)}
			<article class="card p-8">
				<div class="flex items-start justify-between gap-5">
					<h3 class="text-lg font-semibold tracking-tight">{u.ten}</h3>
					<HugeiconsIcon
						icon={u.icon}
						size={42}
						strokeWidth={1.6}
						class="flex-none text-ink"
					/>
				</div>
				<p class="mt-3.5 text-sm leading-[1.75] text-dim">{u.ben}</p>
			</article>
		{/each}
	</div>
</section>
