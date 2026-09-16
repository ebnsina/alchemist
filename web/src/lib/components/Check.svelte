<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import type { IconSvgElement } from '@hugeicons/svelte';

	// One component for both, because a checkbox and a radio differ by exactly two
	// things: whether the box is round, and whether choosing one clears the others.
	// The native input does all of that; we only draw the box.
	let {
		checked = $bindable(false),
		group = $bindable(),
		value,
		type = 'checkbox',
		label,
		hint = '',
		name = '',
		card = false,
		disabled = false,
		icon,
		onchange
	}: {
		checked?: boolean;
		group?: string;
		value?: string;
		type?: 'checkbox' | 'radio';
		label: string;
		hint?: string;
		name?: string;
		card?: boolean;
		disabled?: boolean;
		icon?: IconSvgElement;
		onchange?: (e: Event) => void;
	} = $props();
</script>

<label class={card ? 'option' : 'control'}>
	{#if type === 'radio'}
		<input type="radio" {name} {value} {disabled} {onchange} bind:group />
	{:else}
		<input type="checkbox" {disabled} {onchange} bind:checked />
	{/if}
	<span class="control__box" class:control__box--round={type === 'radio'}></span>
	{#if icon}
		<!-- Inside the hit area, so the icon belongs to the option rather than sitting
		     beside it looking like a separate control. -->
		<HugeiconsIcon
			{icon}
			size={18}
			strokeWidth={1.7}
			class="mt-0.5 flex-none {checked || group === value ? 'text-accent' : 'text-faint'}"
		/>
	{/if}
	<span class="min-w-0">
		<span class="block text-sm font-medium">{label}</span>
		{#if hint}<span class="sub mt-0.5 block">{hint}</span>{/if}
	</span>
</label>
