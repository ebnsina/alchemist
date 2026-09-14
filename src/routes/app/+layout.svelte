<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		DashboardSquare01Icon,
		Upload01Icon,
		KeyframeIcon,
		Logout01Icon
	} from '@hugeicons/core-free-icons';
	import Logo from '$lib/components/Logo.svelte';
	import { session, logout, type Session } from '$lib/api';

	let { children } = $props();
	let me: Session | null = $state(null);
	let checked = $state(false);

	const nav = [
		{ href: '/app/', label: 'Overview', icon: DashboardSquare01Icon },
		{ href: '/app/upload/', label: 'Upload', icon: Upload01Icon },
		{ href: '/app/keys/', label: 'API keys', icon: KeyframeIcon }
	];

	// The session lives in an HttpOnly cookie, so the page cannot read it — asking
	// the API is the only way to know, and the only way to be sure it is still live.
	$effect(() => {
		session()
			.then((s) => (me = s))
			.catch(() => goto('/login/'))
			.finally(() => (checked = true));
	});

	async function signOut() {
		await logout().catch(() => {});
		goto('/login/');
	}
</script>

<div class="mx-auto flex min-h-svh max-w-6xl flex-col gap-8 px-4 py-8 sm:px-6 lg:flex-row lg:gap-10">
	<aside class="lg:w-56 lg:flex-none lg:pt-2">
		<a href="/" class="flex items-center gap-2 text-lg font-semibold">
			<Logo size={24} />
			Alchemist
		</a>

		{#if me}
			<p class="mt-4 truncate text-xs text-muted">{me.org}</p>
		{/if}

		<nav class="mt-6 flex gap-1 overflow-x-auto lg:flex-col lg:overflow-visible" aria-label="Dashboard">
			{#each nav as item (item.href)}
				{@const active = page.url.pathname === item.href}
				<a
					href={item.href}
					class="flex flex-none items-center gap-2.5 rounded-xl px-3 py-2 text-sm whitespace-nowrap transition-colors {active
						? 'bg-card text-ink'
						: 'text-muted hover:text-ink'}"
					aria-current={active ? 'page' : undefined}
				>
					<HugeiconsIcon icon={item.icon} size={17} strokeWidth={1.7} />
					{item.label}
				</a>
			{/each}
		</nav>

		{#if me}
			<button
				type="button"
				onclick={signOut}
				class="mt-6 flex items-center gap-2.5 px-3 py-2 text-sm text-muted transition-colors hover:text-ink"
			>
				<HugeiconsIcon icon={Logout01Icon} size={17} strokeWidth={1.7} />
				Sign out
			</button>
		{/if}
	</aside>

	<main class="min-w-0 flex-1">
		{#if !checked}
			<p class="text-sm text-muted">Checking your session…</p>
		{:else if me}
			{@render children()}
		{/if}
	</main>
</div>
