<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import type { IconSvgElement } from '@hugeicons/svelte';
	import { MoreVerticalIcon } from '@hugeicons/core-free-icons';

	export type RowAction = {
		label: string;
		icon?: IconSvgElement;
		href?: string;
		onclick?: () => void;
		danger?: boolean;
		disabled?: boolean;
		why?: string;
	};

	let { actions, label = 'Actions' }: { actions: RowAction[]; label?: string } = $props();

	let open = $state(false);
	let host = $state<HTMLElement | null>(null);
</script>

<svelte:window
	onclick={(e) => {
		// Anywhere else closes it, including another row's menu opening.
		if (open && host && !host.contains(e.target as Node)) open = false;
	}}
	onkeydown={(e) => {
		if (e.key === 'Escape') open = false;
	}}
/>

<div class="rowmenu" bind:this={host}>
	<button
		type="button"
		class="icon-btn"
		aria-label={label}
		aria-expanded={open}
		aria-haspopup="menu"
		onclick={() => (open = !open)}
	>
		<HugeiconsIcon icon={MoreVerticalIcon} size={16} strokeWidth={1.9} />
	</button>

	{#if open}
		<div class="rowmenu__list" role="menu">
			{#each actions as a (a.label)}
				{#if a.href}
					<a href={a.href} class="rowmenu__item" role="menuitem" onclick={() => (open = false)}>
						{#if a.icon}<HugeiconsIcon icon={a.icon} size={15} strokeWidth={1.8} />{/if}
						{a.label}
					</a>
				{:else}
					<button
						type="button"
						class="rowmenu__item"
						class:rowmenu__item--danger={a.danger}
						role="menuitem"
						disabled={a.disabled}
						title={a.disabled ? a.why : undefined}
						onclick={() => {
							open = false;
							a.onclick?.();
						}}
					>
						{#if a.icon}<HugeiconsIcon icon={a.icon} size={15} strokeWidth={1.8} />{/if}
						{a.label}
					</button>
				{/if}
			{/each}
		</div>
	{/if}
</div>

<style>
	.rowmenu {
		position: relative;
		display: inline-flex;
	}
	.rowmenu__list {
		position: absolute;
		right: 0;
		top: calc(100% + 4px);
		z-index: 20;
		min-width: 170px;
		padding: 5px;
		border-radius: var(--radius-md);
		border: 1px solid var(--color-sunk);
		background: var(--color-card);
		box-shadow: 0 12px 28px rgb(0 0 0 / 0.28);
	}
	.rowmenu__item {
		display: flex;
		width: 100%;
		align-items: center;
		gap: 9px;
		padding: 8px 10px;
		border-radius: var(--radius-sm);
		font-size: 13px;
		text-align: left;
		white-space: nowrap;
		color: var(--color-ink);
		cursor: pointer;
	}
	.rowmenu__item:hover:not(:disabled) {
		background: var(--color-sunk);
	}
	.rowmenu__item:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
	.rowmenu__item--danger {
		color: var(--color-red);
	}
</style>
