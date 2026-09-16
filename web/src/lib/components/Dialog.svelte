<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Cancel01Icon } from '@hugeicons/core-free-icons';

	// Native <dialog>: focus trap, Escape, inert background and a backdrop, none of
	// which we have to write or get wrong.
	let {
		open = $bindable(false),
		title,
		children
	}: { open?: boolean; title: string; children: import('svelte').Snippet } = $props();

	let el = $state<HTMLDialogElement | null>(null);

	$effect(() => {
		if (!el) return;
		if (open && !el.open) el.showModal();
		if (!open && el.open) el.close();
	});
</script>

<dialog
	bind:this={el}
	class="dlg"
	aria-label={title}
	onclose={() => (open = false)}
	onclick={(e) => {
		// The backdrop is the dialog element itself; the panel inside stops the click.
		if (e.target === el) open = false;
	}}
>
	<div class="dlg__panel">
		<div class="flex items-start justify-between gap-4">
			<h2 class="text-lg font-semibold tracking-tight">{title}</h2>
			<button type="button" class="icon-btn flex-none" onclick={() => (open = false)} aria-label="Close">
				<HugeiconsIcon icon={Cancel01Icon} size={17} strokeWidth={1.8} />
			</button>
		</div>
		<div class="mt-5">
			{@render children()}
		</div>
	</div>
</dialog>

<style>
	.dlg {
		margin: auto;
		padding: 0;
		border: 0;
		background: transparent;
		max-width: min(640px, calc(100vw - 32px));
		width: 100%;
		max-height: calc(100dvh - 48px);
		overflow: visible;
		color: var(--color-ink);
	}
	.dlg::backdrop {
		background: rgb(0 0 0 / 0.55);
		backdrop-filter: blur(2px);
	}
	.dlg__panel {
		background: var(--color-card);
		border: 1px solid var(--color-sunk);
		border-radius: var(--radius-lg);
		padding: 24px;
		max-height: calc(100dvh - 48px);
		overflow-y: auto;
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
