<script lang="ts">
	import Dialog from './Dialog.svelte';

	// Asking before, not explaining after. Same size and same three regions as every
	// other dialog, so a question never arrives looking like a different kind of thing.
	let {
		open = $bindable(false),
		title,
		confirm,
		destructive = false,
		busy = false,
		onconfirm,
		children
	}: {
		open?: boolean;
		title: string;
		confirm: string;
		destructive?: boolean;
		busy?: boolean;
		onconfirm: () => void;
		children: import('svelte').Snippet;
	} = $props();
</script>

<Dialog bind:open {title}>
	{@render children()}

	{#snippet footer()}
		<div class="flex flex-wrap items-center justify-end gap-2">
			<button type="button" class="btn" disabled={busy} onclick={() => (open = false)}>
				Keep it
			</button>
			<button
				type="button"
				class={destructive ? 'btn-danger' : 'btn-solid'}
				disabled={busy}
				onclick={onconfirm}
			>
				{busy ? 'Working…' : confirm}
			</button>
		</div>
	{/snippet}
</Dialog>
