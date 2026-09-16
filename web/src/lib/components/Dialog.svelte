<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Cancel01Icon } from '@hugeicons/core-free-icons';

	// Native <dialog>: focus trap, Escape, inert background and a backdrop, none of
	// which we have to write or get wrong.
	let {
		open = $bindable(false),
		title,
		hint,
		id,
		children,
		footer
	}: {
		open?: boolean;
		title: string;
		hint?: string;
		id?: string;
		children: import('svelte').Snippet;
		footer?: import('svelte').Snippet;
	} = $props();

	let el = $state<HTMLDialogElement | null>(null);

	$effect(() => {
		if (!el) return;
		if (open && !el.open) el.showModal();
		if (!open && el.open) el.close();
	});
</script>

<dialog
	bind:this={el}
	{id}
	class="dlg"
	aria-label={title}
	onclose={() => (open = false)}
	onclick={(e) => {
		// The backdrop is the dialog element itself; the panel inside stops the click.
		if (e.target === el) open = false;
	}}
>
	<div class="dlg__panel">
		<div class="dlg__head">
			<div class="min-w-0">
				<h2 class="text-lg font-semibold tracking-tight">{title}</h2>
				{#if hint}<p class="sub mt-1">{hint}</p>{/if}
			</div>
			<button
				type="button"
				class="icon-btn flex-none"
				onclick={() => (open = false)}
				aria-label="Close"
			>
				<HugeiconsIcon icon={Cancel01Icon} size={17} strokeWidth={1.8} />
			</button>
		</div>

		<div class="dlg__body">
			{@render children()}
		</div>

		{#if footer}
			<div class="dlg__foot">
				{@render footer()}
			</div>
		{/if}
	</div>
</dialog>

<style>
	/* One width everywhere, and a height that fits its content up to a cap: a two-field
	   form has no business drawing a box two thirds of the screen tall. */
	.dlg {
		margin: auto;
		padding: 0;
		border: 0;
		background: transparent;
		width: min(560px, calc(100vw - 32px));
		height: auto;
		max-height: min(620px, calc(100dvh - 48px));
		overflow: visible;
		color: var(--color-ink);
	}
	.dlg::backdrop {
		background: rgb(0 0 0 / 0.55);
		backdrop-filter: blur(2px);
	}
	.dlg__panel {
		display: flex;
		flex-direction: column;
		max-height: 100%;
		background: var(--color-card);
		border: 1px solid var(--color-sunk);
		border-radius: var(--radius-lg);
		overflow: hidden;
	}
	.dlg__head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 16px;
		flex: none;
		padding: 20px 24px;
		border-bottom: 1px solid var(--color-sunk);
	}
	.dlg__body {
		min-height: 0;
		overflow-y: auto;
		padding: 24px;
	}
	.dlg__foot {
		flex: none;
		padding: 16px 24px;
		border-top: 1px solid var(--color-sunk);
		background: var(--color-card);
	}
	.dlg[open] {
		animation: dlg-in 180ms cubic-bezier(0.22, 1, 0.36, 1);
	}
	@keyframes dlg-in {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.dlg[open] {
			animation: none;
		}
	}
</style>
