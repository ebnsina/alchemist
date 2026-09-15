<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import type { IconSvgElement } from '@hugeicons/svelte';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	// The step pattern, defined once. Signup, sign-in, upload and connecting a bucket
	// are the same shape: a progress rail, one question at a time, back and next.
	export type Step = { key: string; icon: IconSvgElement; title: string; hint: string };

	let {
		steps,
		step,
		children
	}: { steps: Step[]; step: number; children: import('svelte').Snippet } = $props();
</script>

<ol class="mb-6 flex items-center gap-1.5" aria-label="Progress">
	{#each steps as s, i (s.key)}
		<li
			class="step-dot"
			data-state={i < step ? 'done' : i === step ? 'current' : 'todo'}
			aria-current={i === step ? 'step' : undefined}
		>
			<span class="vh">{s.title}</span>
		</li>
	{/each}
</ol>

{#key step}
	<div in:fly={{ y: 14, duration: 340, delay: 80, easing: cubicOut }}>
		<div class="flex items-start justify-between gap-4">
			<div class="min-w-0">
				<h2 class="title">{steps[step].title}</h2>
				<p class="sub mt-1.5">{steps[step].hint}</p>
			</div>
			<HugeiconsIcon
				icon={steps[step].icon}
				size={30}
				strokeWidth={1.6}
				class="flex-none text-accent"
			/>
		</div>
		{@render children()}
	</div>
{/key}
