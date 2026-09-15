<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { fly, fade } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		DashboardSquare01Icon,
		Upload01Icon,
		Key01Icon,
		Book02Icon,
		Logout01Icon,
		Menu01Icon,
		Cancel01Icon,
		ArrowUpRight01Icon,
		ArrowDown01Icon,
		UserGroupIcon,
		PaintBoardIcon,
		ConnectIcon,
		CloudServerIcon,
		ChartHistogramIcon,
		LiveStreaming01Icon,
		ArrowDataTransferHorizontalIcon,
		Scissor01Icon,
		RecordIcon
	} from '@hugeicons/core-free-icons';
	import Logo from '$lib/components/Logo.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { session, logout, getBranding, ApiError, type Session, type Branding } from '$lib/api';
	import { crumbs, setCrumbs } from '$lib/crumbs.svelte';
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';

	let { children } = $props();
	let me = $state<Session | null>(null);
	let checked = $state(false);
	let trouble = $state('');
	let drawer = $state(false);
	let brand = $state<Branding | null>(null);
	let menu = $state(false);

	const initial = $derived((me?.org ?? '?').trim().charAt(0).toUpperCase());

	const groups = [
		{
			label: 'Videos',
			items: [
				{ href: '/app/videos/', label: 'Videos', icon: DashboardSquare01Icon },
				{ href: '/app/upload/', label: 'Upload', icon: Upload01Icon },
				{ href: '/app/studio/', label: 'Studio', icon: Scissor01Icon },
				{ href: '/app/sources/', label: 'Connected buckets', icon: CloudServerIcon },
				{ href: '/app/migrate/', label: 'Move a library', icon: ArrowDataTransferHorizontalIcon }
			]
		},
		{
			label: 'Develop',
			items: [
				{ href: '/app/keys/', label: 'API keys', icon: Key01Icon },
				{ href: '/app/webhooks/', label: 'Webhooks', icon: ConnectIcon },
				{ href: '/app/docs/', label: 'API reference', icon: Book02Icon }
			]
		},
		{
			label: 'Live',
			soon: true,
			items: [
				{ href: '/app/live/', label: 'Streams', icon: LiveStreaming01Icon },
				{ href: '/app/live/recordings/', label: 'Recordings', icon: RecordIcon }
			]
		},
		{
			label: 'Account',
			items: [
				{ href: '/app/team/', label: 'Team', icon: UserGroupIcon },
				{ href: '/app/settings/', label: 'Branding', icon: PaintBoardIcon },
				{ href: '/app/usage/', label: 'Usage', icon: ChartHistogramIcon }
			]
		}
	];

	const flat = groups.flatMap((g) => g.items);

	// A nested route keeps its section lit. Exact match alone left every deeper page
	// with nothing selected; a bare startsWith would light Overview on all of them,
	// since /app/ prefixes the lot.
	function isActive(item: { href: string; owns?: string[] }, path: string) {
		if (path === item.href) return true;
		if (item.href !== '/app/' && path.startsWith(item.href)) return true;
		return (item.owns ?? []).some((p) => path === p || path.startsWith(p + '/'));
	}

	const active = $derived(flat.find((i) => isActive(i, page.url.pathname)));
	const current = $derived(
		page.url.pathname.startsWith('/app/studio/')
			? 'Studio'
			: page.url.pathname.startsWith('/app/videos/')
				? 'Video'
				: (active?.label ?? 'Videos')
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

	// The logo is the account's own, not ours. It loads separately because a missing
	// one is not a reason to hold up the whole shell.
	$effect(() => {
		if (me) getBranding().then((b) => (brand = b)).catch(() => {});
	});

	// A route change closes the drawer, or it stays open over the page it just
	// navigated to, and clears a trail belonging to the page being left.
	$effect(() => {
		page.url.pathname;
		drawer = false;
		menu = false;
		setCrumbs([]);
	});

	const trail = $derived(crumbs().length ? crumbs() : [{ label: current }]);

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

{#if !checked}
	<!-- No shell until there is an account behind it. Rendering the navigation first
	     and filling it in later shows a signed-out visitor the shape of someone's
	     dashboard, and flashes it at everyone else on every load. -->
	<div class="flex min-h-svh items-center justify-center">
		<p class="label" role="status">Checking your session</p>
	</div>
{:else if !me}
	<div class="flex min-h-svh items-center justify-center px-5">
		<div class="card w-full max-w-sm text-center">
			<p class="title">We could not load your account</p>
			<p class="sub mt-2">{trouble || 'You are not signed in.'}</p>
			<div class="mt-6 flex justify-center gap-2">
				<button type="button" class="btn-solid" onclick={() => location.reload()}>Try again</button>
				<a href="/login/" class="btn">Sign in</a>
			</div>
		</div>
	</div>
{:else}
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

		<nav class="mt-5 flex-1 overflow-y-auto px-3" aria-label="Dashboard">
			{#each groups as group (group.label)}
				<p class="flex items-center gap-2 px-2 pt-4 pb-2">
					<span class="label">{group.label}</span>
					{#if group.soon}
						<span class="chip">Soon</span>
					{/if}
				</p>
				<ul class="grid gap-0.5">
					{#each group.items as item (item.href)}
						{@const on = isActive(item, page.url.pathname)}
						<li>
							{#if group.soon}
								<span class="side-link opacity-45" aria-disabled="true">
									<HugeiconsIcon icon={item.icon} size={17} strokeWidth={1.7} />
									{item.label}
								</span>
							{:else}
								<a
									href={item.href}
									class="side-link"
									class:side-link--on={on}
									aria-current={on ? 'page' : undefined}
								>
									<HugeiconsIcon icon={item.icon} size={17} strokeWidth={1.7} />
									{item.label}
								</a>
							{/if}
						</li>
					{/each}
				</ul>
			{/each}
		</nav>

		<!-- One control at the bottom for everything about the person using it. -->
		<div class="relative p-3">
			{#if menu}
				<div
					class="account-menu"
					transition:fly={{ y: 6, duration: 160, easing: cubicOut }}
				>
					<a href="/app/settings/" class="side-link">
						<HugeiconsIcon icon={PaintBoardIcon} size={16} strokeWidth={1.7} />
						Account settings
					</a>
					<a href="/" class="side-link">
						<HugeiconsIcon icon={ArrowUpRight01Icon} size={16} strokeWidth={1.7} />
						Back to the site
					</a>
					<div class="my-1 h-px bg-sunk"></div>
					<button type="button" class="side-link w-full" onclick={signOut}>
						<HugeiconsIcon icon={Logout01Icon} size={16} strokeWidth={1.7} />
						Sign out
					</button>
				</div>
			{/if}

			<button
				type="button"
				class="account"
				onclick={() => (menu = !menu)}
				aria-expanded={menu}
				aria-haspopup="menu"
			>
				{#if brand?.logo_url}
					<img src={PUBLIC_ALCHEMIST_API + brand.logo_url} alt="" class="account__mark" />
				{:else}
					<span class="account__mark account__mark--initial">{initial}</span>
				{/if}
				<span class="min-w-0 flex-1 text-left">
					<span class="block truncate text-sm font-medium">{me?.org}</span>
					<span class="mono block truncate">{me?.email}</span>
				</span>
				<HugeiconsIcon
					icon={ArrowDown01Icon}
					size={15}
					strokeWidth={2}
					class="chevron flex-none text-faint {menu ? 'rotate-180' : ''}"
				/>
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
				<nav class="flex min-w-0 items-center gap-2 text-sm" aria-label="Breadcrumb">
					<a href="/app/videos/" class="flex-none text-dim transition-colors hover:text-ink">Dashboard</a>
					{#each trail as c, i (c.label + i)}
						<span class="flex-none text-dim" aria-hidden="true">/</span>
						{#if c.href && i < trail.length - 1}
							<a href={c.href} class="flex-none truncate text-dim transition-colors hover:text-ink">
								{c.label}
							</a>
						{:else}
							<span class="truncate font-medium" aria-current={i === trail.length - 1 ? 'page' : undefined}>
								{c.label}
							</span>
						{/if}
					{/each}
				</nav>
				<div class="ml-auto">
					<ThemeToggle />
				</div>
			</header>

			<div id="main" class="panel__body">
				<div in:fly={{ y: 8, duration: 220, easing: cubicOut }}>
					{@render children()}
				</div>
			</div>
		</div>
	</div>
</div>
{/if}

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
		background: var(--color-bg);
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
		background: var(--color-card);
		border: 1px solid var(--color-sunk);
		border-radius: var(--radius-lg);
		overflow: hidden;
	}
	.panel__bar {
		display: flex;
		flex: none;
		align-items: center;
		gap: 12px;
		padding: 12px 16px;
		border-bottom: 1px solid var(--color-sunk);
	}
	.account {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		padding: 8px 10px;
		border-radius: var(--radius-md);
		text-align: left;
		transition: background 120ms ease;
	}
	.account:hover {
		background: var(--color-sunk);
	}
	.account__mark {
		height: 32px;
		width: 32px;
		flex: none;
		border-radius: var(--radius-sm);
		object-fit: contain;
		background: var(--color-sunk);
	}
	.account__mark--initial {
		display: grid;
		place-items: center;
		background: var(--color-brand);
		color: var(--color-on-brand);
		font-weight: 700;
		font-size: 14px;
	}
	.account-menu {
		position: absolute;
		bottom: calc(100% - 4px);
		left: 12px;
		right: 12px;
		z-index: 10;
		padding: 6px;
		border-radius: var(--radius-md);
		background: var(--color-card);
		border: 1px solid var(--color-sunk);
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.28);
	}

	.panel__body {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding: 24px 20px 64px;
	}
	.account {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		padding: 8px 10px;
		border-radius: var(--radius-md);
		text-align: left;
		transition: background 120ms ease;
	}
	.account:hover {
		background: var(--color-sunk);
	}
	.account__mark {
		height: 32px;
		width: 32px;
		flex: none;
		border-radius: var(--radius-sm);
		object-fit: contain;
		background: var(--color-sunk);
	}
	.account__mark--initial {
		display: grid;
		place-items: center;
		background: var(--color-brand);
		color: var(--color-on-brand);
		font-weight: 700;
		font-size: 14px;
	}
	.account-menu {
		position: absolute;
		bottom: calc(100% - 4px);
		left: 12px;
		right: 12px;
		z-index: 10;
		padding: 6px;
		border-radius: var(--radius-md);
		background: var(--color-card);
		border: 1px solid var(--color-sunk);
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.28);
	}

	/* One gutter, and it grows with the viewport rather than disappearing at the
	   narrow end. The account rules above had been swallowed by this media query, so
	   they only applied over 640px. */
	@media (min-width: 640px) {
		.panel__body {
			padding: 32px 28px 64px;
		}
	}
	@media (min-width: 1024px) {
		.panel__body {
			padding: 40px 40px 64px;
		}
	}
</style>
