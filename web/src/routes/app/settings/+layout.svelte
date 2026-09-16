<script lang="ts">
	import { page } from '$app/state';
	import { setCrumbs } from '$lib/crumbs.svelte';

	let { children } = $props();

	// Things you set once and rarely open again. One page, one row of tabs, rather
	// than three sidebar rows competing with the work.
	const TABS = [
		{ href: '/app/settings/profile/', label: 'Profile' },
		{ href: '/app/settings/branding/', label: 'Branding' },
		{ href: '/app/settings/encoding/', label: 'Encoding' },
		{ href: '/app/settings/playback/', label: 'Playback' }
	];

	const here = $derived(TABS.find((t) => page.url.pathname.startsWith(t.href)));

	$effect(() => {
		setCrumbs([{ label: 'Settings' }, { label: here?.label ?? 'Settings' }]);
	});
</script>

<h1 class="text-2xl font-semibold tracking-tight">Settings</h1>

<nav class="tabs mt-5" aria-label="Settings">
	{#each TABS as t (t.href)}
		<a
			href={t.href}
			class="tab"
			class:tab--on={here?.href === t.href}
			aria-current={here?.href === t.href ? 'page' : undefined}
		>
			{t.label}
		</a>
	{/each}
</nav>

<div class="mt-6">
	{@render children()}
</div>

<style>
	.tabs {
		display: flex;
		gap: 2px;
		overflow-x: auto;
		border-bottom: 1px solid var(--color-sunk);
	}
	.tab {
		flex: none;
		padding: 10px 14px;
		font-size: 13px;
		font-weight: 500;
		color: var(--color-dim);
		/* On the border, not above it: a tab that floats free reads as a button. */
		margin-bottom: -1px;
		border-bottom: 2px solid transparent;
		transition: color 120ms ease;
	}
	.tab:hover {
		color: var(--color-ink);
	}
	.tab--on {
		color: var(--color-ink);
		border-bottom-color: var(--color-brand);
	}
</style>
