<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import type { IconSvgElement } from '@hugeicons/svelte';

	// An icon button that says what it does on hover and on focus. title= would do the
	// first but not the second, and it cannot be styled — a gray OS tooltip in the
	// middle of this is exactly the native-element problem everywhere else.
	let {
		icon,
		label,
		on = false,
		disabled = false,
		onclick
	}: {
		icon: IconSvgElement;
		label: string;
		on?: boolean;
		disabled?: boolean;
		onclick?: () => void;
	} = $props();
</script>

<span class="tip">
	<button
		type="button"
		class="tip__btn"
		class:tip__btn--on={on}
		{disabled}
		{onclick}
		aria-label={label}
	>
		<HugeiconsIcon {icon} size={16} strokeWidth={1.9} />
	</button>
	<!-- aria-hidden: the button already carries the same words as its accessible name,
	     so a reader would otherwise say them twice. -->
	<span class="tip__text" aria-hidden="true">{label}</span>
</span>

<style>
	.tip {
		position: relative;
		display: inline-flex;
	}
	.tip__btn {
		display: inline-grid;
		place-items: center;
		height: 32px;
		width: 32px;
		border-radius: var(--radius-sm);
		background: rgba(255, 255, 255, 0.1);
		color: #fff;
		transition: background 120ms ease;
	}
	.tip__btn:hover:not(:disabled) {
		background: rgba(255, 255, 255, 0.2);
	}
	.tip__btn:disabled {
		opacity: 0.35;
		cursor: not-allowed;
	}
	.tip__btn--on {
		background: var(--color-brand);
		color: var(--color-on-brand);
	}
	.tip__text {
		position: absolute;
		top: calc(100% + 6px);
		left: 50%;
		transform: translateX(-50%);
		z-index: 20;
		padding: 3px 8px;
		border-radius: var(--radius-sk);
		background: var(--color-ink);
		color: var(--color-bg);
		font-family: var(--font-mono);
		font-size: 10px;
		white-space: nowrap;
		opacity: 0;
		pointer-events: none;
		transition: opacity 120ms ease;
	}
	.tip:hover .tip__text,
	.tip__btn:focus-visible ~ .tip__text {
		opacity: 1;
	}
</style>
