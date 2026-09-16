<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import type { IconSvgElement } from '@hugeicons/svelte';

	// A figure alone tells nobody anything. Every tile carries what it is, what it
	// means, and — where there is one — how it is trending.
	let {
		label,
		value,
		hint,
		icon,
		accent = false,
		loading = false
	}: {
		label: string;
		value: string;
		hint: string;
		icon: IconSvgElement;
		accent?: boolean;
		loading?: boolean;
	} = $props();
</script>

<div class="card">
	<div class="flex items-start justify-between gap-3">
		<p class="label">{label}</p>
		<HugeiconsIcon {icon} size={18} strokeWidth={1.7} class="flex-none text-faint" />
	</div>
	{#if loading}
		<div class="sk mt-3 h-7 w-24"></div>
		<div class="sk mt-2 h-3.5 w-32"></div>
	{:else}
		<p class="num mt-3 text-[26px] leading-none" class:text-accent={accent}>{value}</p>
		<p class="sub mt-2">{hint}</p>
	{/if}
</div>
