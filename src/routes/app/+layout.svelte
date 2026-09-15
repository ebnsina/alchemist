<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { fly, fade } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		DashboardSquare01Icon,
		Upload01Icon,
		KeyframeIcon,
		Book02Icon,
		Logout01Icon,
		Menu01Icon,
		Cancel01Icon,
		ArrowUpRight01Icon
	} from '@hugeicons/core-free-icons';
	import Logo from '$lib/components/Logo.svelte';
	import { session, logout, ApiError, type Session } from '$lib/api';

	let { children } = $props();
	let me = $state<Session | null>(null);
	let checked = $state(false);
	let trouble = $state('');
	let drawer = $state(false);

	const groups = [
		{
			label: 'Your videos',
			items: [
				{ href: '/app/', label: 'Overview', icon: DashboardSquare01Icon },
				{ href: '/app/upload/', label: 'Upload', icon: Upload01Icon }
			]
		},
		{
			label: 'Build with it',
			items: [
				{ href: '/app/keys/', label: 'API keys', icon: KeyframeIcon },
				{ href: '/app/docs/', label: 'API reference', icon: Book02Icon }
			]
		}
	];

	const flat = groups.flatMap((g) => g.items);
	const current = $derived(
		page.url.pathname.startsWith('/app/videos/')
			? 'Video'
			: (flat.find((i) => i.href === page.url.pathname)?.label ?? 'Overview')
	);

	// The session lives in an HttpOnly cookie, so the page cannot read it — asking
	// the API is the only way to know, and the only way to be sure it is still live.
	//
	// Only a refused session sends someone to the login page. A network blip or a
	// rate limit is not a signed-out state, and treating it as one throws away
	// whatever they were in the middle of.
	$effect(() => {
		session()
			.then((s) => {
				me = s;
				trouble = '';
			})
			.catch((e) => {
				if (e instanceof ApiError && e.code === 'no_session') goto('/login/');
				else trouble = e instanceof ApiError ? e.message : 'We could not reach Alchemist.';
			})
			.finally(() => (checked = true));
	});

	// A route change closes the drawer, or it stays open over the page it just
	// navigated to.
	$effect(() => {
		page.url.pathname;
		drawer = false;
	});

	async function signOut() {
		await logout().catch(() => {});
		goto('/login/');
	}
</script>

<svelte:head>
	<!-- The shell owns the viewport; only the panel scrolls. -->
	<style>
		body {
			overflow: hidden;
		}
	</style>
</svelte:head>

<div class="shell-grid">
	<!-- The sidebar sits on the page ground. That is what makes the panel inset. -->
	<aside class="sidebar" class:sidebar--open={drawer}>
		<div class="flex items-center justify-between px-4 pt-4">
			<a href="/" class="flex items-center gap-2 text-base font-semibold tracking-tight">
				<Logo size={22} />
				Alchemist
			</a>
			<button
				type="button"
				class="icon-btn lg:hidden"
				onclick={() => (drawer = false)}
				aria-label="Close menu"
			>
				<HugeiconsIcon icon={Cancel01Icon} size={17} strokeWidth={1.8} />
			</button>
		</div>

		{#if me}
			<div class="mx-4 mt-4 rounded-md border border-outline bg-surface px-3 py-2.5">
				<p class="truncate text-sm font-medium">{me.org}</p>
				<p class="mt-0.5 truncate text-xs text-secondary">{me.email}</p>
			</div>
		{/if}

		<nav class="mt-4 flex-1 overflow-y-auto px-3" aria-label="Dashboard">
			{#each groups as group (group.label)}
				<p class="label-caps px-2 pt-4 pb-2 text-secondary">
					{group.label}
				</p>
				<ul class="grid gap-0.5">
					{#each group.items as item (item.href)}
						{@const active = page.url.pathname === item.href}
						<li>
							<a
								href={item.href}
								class="side-link"
								class:side-link--on={active}
								aria-current={active ? 'page' : undefined}
							>
								<HugeiconsIcon icon={item.icon} size={17} strokeWidth={1.7} />
								{item.label}
							</a>
						</li>
					{/each}
				</ul>
			{/each}
		</nav>

		<div class="border-t border-outline p-3">
			<a href="/" class="side-link text-secondary">
				<HugeiconsIcon icon={ArrowUpRight01Icon} size={17} strokeWidth={1.7} />
				Back to the site
			</a>
			<button type="button" class="side-link w-full text-secondary" onclick={signOut}>
				<HugeiconsIcon icon={Logout01Icon} size={17} strokeWidth={1.7} />
				Sign out
			</button>
		</div>
	</aside>

	{#if drawer}
		<button
			class="scrim lg:hidden"
			onclick={() => (drawer = false)}
			aria-label="Close menu"
			transition:fade={{ duration: 150 }}
		></button>
	{/if}

	<!-- The inset panel: its own surface, its own scroll, floating on the ground. -->
	<div class="panel-wrap">
		<div class="panel">
			<header class="panel__bar">
				<button
					type="button"
					class="icon-btn lg:hidden"
					onclick={() => (drawer = true)}
					aria-label="Open menu"
				>
					<HugeiconsIcon icon={Menu01Icon} size={18} strokeWidth={1.8} />
				</button>
				<nav class="flex items-center gap-2 text-sm" aria-label="Breadcrumb">
					<span class="text-secondary">Dashboard</span>
					<span class="text-secondary" aria-hidden="true">/</span>
					<span class="font-medium">{current}</span>
				</nav>
			</header>

			<div class="panel__body">
				{#if !checked}
					<p class="text-sm text-secondary">Checking your session…</p>
				{:else if me}
					<div in:fly={{ y: 8, duration: 220, easing: cubicOut }}>
						{@render children()}
					</div>
				{:else if trouble}
					<div class="card text-center">
						<p class="font-semibold">We could not load your account</p>
						<p class="mx-auto mt-2 max-w-sm text-sm text-secondary">{trouble}</p>
						<button type="button" class="btn-secondary mt-5" onclick={() => location.reload()}>
							Try again
						</button>
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.shell-grid {
		display: grid;
		height: 100svh;
		grid-template-columns: 1fr;
	}
	@media (min-width: 1024px) {
		.shell-grid {
			grid-template-columns: 15rem 1fr;
		}
	}

	.sidebar {
		display: flex;
		flex-direction: column;
		height: 100svh;
		width: 15rem;
		position: fixed;
		inset: 0 auto 0 0;
		z-index: 40;
		background: var(--color-neutral);
		border-right: 1px solid var(--color-outline);
		transform: translateX(-100%);
		transition: transform 200ms ease;
	}
	.sidebar--open {
		transform: none;
		box-shadow: var(--shadow-float);
	}
	@media (min-width: 1024px) {
		.sidebar {
			position: sticky;
			top: 0;
			transform: none;
			width: auto;
			box-shadow: none;
		}
	}

	.scrim {
		position: fixed;
		inset: 0;
		z-index: 30;
		background: rgba(21, 24, 27, 0.4);
	}

	.panel-wrap {
		min-width: 0;
		height: 100svh;
		padding: 8px;
	}
	@media (min-width: 1024px) {
		.panel-wrap {
			padding: 16px 16px 16px 0;
		}
	}

	/* Depth is the surface colour against paper plus the hairline. No shadow here:
	   this is static content, and the system has one shadow, for things that float. */
	.panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		min-width: 0;
		background: var(--color-surface);
		border: 1px solid var(--color-outline);
		border-radius: var(--radius-lg);
		overflow: hidden;
	}
	.panel__bar {
		display: flex;
		flex: none;
		align-items: center;
		gap: 12px;
		padding: 12px 16px;
		border-bottom: 1px solid var(--color-outline);
	}
	.panel__body {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding: 24px 16px 64px;
	}
	@media (min-width: 640px) {
		.panel__body {
			padding: 40px 40px 64px;
		}
	}
</style>
