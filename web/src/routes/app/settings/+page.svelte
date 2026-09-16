<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { ImageAdd02Icon, Delete02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';
	import {
		getBranding,
		putLogo,
		deleteLogo,
		whoami,
		listProfiles,
		setProfile,
		ApiError,
		type Branding,
		type Whoami,
		type LadderProfile
	} from '$lib/api';

	let brand = $state<Branding | null>(null);
	let who = $state<Whoami | null>(null);
	let profiles = $state<LadderProfile[]>([]);
	let switching = $state('');
	let loading = $state(true);
	let busy = $state(false);
	let error = $state('');
	let input: HTMLInputElement | null = $state(null);

	async function load() {
		error = '';
		try {
			const [b, w, p] = await Promise.all([getBranding(), whoami(), listProfiles()]);
			brand = b;
			who = w;
			profiles = p.profiles;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	// Which types and sizes are allowed is the API's ruling; sending the file and
	// showing what it says back keeps one copy of that rule.
	async function upload(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		error = '';
		busy = true;
		try {
			await putLogo(file);
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busy = false;
			if (input) input.value = '';
		}
	}

	async function remove() {
		error = '';
		busy = true;
		try {
			await deleteLogo();
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}

	async function choose(name: string) {
		error = '';
		switching = name;
		try {
			await setProfile(name);
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			switching = '';
		}
	}

	const kbps = (n: number) =>
		new Intl.NumberFormat('en', { maximumFractionDigits: 0 }).format(n / 1000) + 'k';

	const logoPath = $derived(brand?.logo_url ?? null);
	const logoSrc = $derived(logoPath ? PUBLIC_ALCHEMIST_API + logoPath : null);
</script>

<Seo title="Branding — Alchemist" description="The name and logo your viewers see." />

<h1 class="text-2xl font-semibold tracking-tight">Branding</h1>
<p class="sub mt-1 max-w-xl">
	What your viewers see when they watch something of yours. Your logo is public by design —
	it is shown to people who have no account with us.
</p>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="sk mt-6 h-44"></div>
{:else}
	<div class="card mt-6">
		<p class="title">Logo</p>
		<p class="sub mt-2">PNG, JPEG or WebP, up to 1 MB.</p>

		<div class="mt-5 flex flex-wrap items-center gap-5">
			<div
				class="grid h-24 w-24 flex-none place-items-center rounded-md border border-sunk bg-bg"
			>
				{#if logoSrc}
					<img src={logoSrc} alt="Your logo" class="max-h-20 max-w-20 object-contain" />
				{:else}
					<HugeiconsIcon icon={ImageAdd02Icon} size={24} strokeWidth={1.6} class="text-faint" />
				{/if}
			</div>

			<div class="flex flex-wrap items-center gap-2">
				<label class="btn-solid cursor-pointer">
					{busy ? 'Saving…' : logoSrc ? 'Replace' : 'Upload a logo'}
					<input
						bind:this={input}
						type="file"
						accept="image/png,image/jpeg,image/webp"
						class="vh"
						onchange={upload}
						disabled={busy}
					/>
				</label>
				{#if logoSrc}
					<button type="button" class="btn" onclick={remove} disabled={busy}>
						<HugeiconsIcon icon={Delete02Icon} size={15} strokeWidth={1.8} />
						Remove
					</button>
				{/if}
			</div>
		</div>
	</div>

	<div class="card mt-4">
		<p class="title">Account</p>
		<dl class="mt-4 grid gap-4 sm:grid-cols-2">
			<div>
				<dt class="label">Name</dt>
				<dd class="mt-1 text-sm font-medium">{brand?.name ?? '—'}</dd>
			</div>
			<div>
				<dt class="label">Tenant</dt>
				<dd class="mono mt-1 truncate">{who?.tenant_id ?? '—'}</dd>
			</div>
		</dl>
	</div>

	<div class="card mt-4">
		<p class="title">Encoding preset</p>
		<p class="sub mt-2">
			Which sizes we make, and how hard we squeeze them. Changing this affects the next video
			you send — everything already encoded stays exactly as it is, because re-making a whole
			library on a settings change would be a surprise that costs real money.
		</p>

		<div class="mt-5 grid gap-3">
			{#each profiles as p (p.name)}
				<div class="preset" class:preset--on={p.current}>
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div class="min-w-0">
							<p class="text-sm font-semibold">{p.description}</p>
							<p class="mono mt-0.5">{p.name}</p>
						</div>
						{#if p.current}
							<span class="chip chip-on flex-none">In use</span>
						{:else}
							<button
								type="button"
								class="btn btn-sm flex-none"
								onclick={() => choose(p.name)}
								disabled={switching === p.name}
							>
								{switching === p.name ? 'Switching…' : 'Use this'}
							</button>
						{/if}
					</div>

					<ul class="mt-3 flex flex-wrap gap-1.5">
						{#each p.rungs as r (r.height)}
							<li class="chip" title="{r.codec} at about {kbps(r.maxrate_bps)} a second">
								{r.height}p{r.lazy ? ' ·' : ''}
							</li>
						{/each}
					</ul>
					<p class="sub mt-2.5">
						{p.rungs.length} sizes, up to {Math.max(...p.rungs.map((r) => r.height))}p.
						{#if p.rungs.some((r) => r.lazy)}
							The ones marked · are only made when somebody first asks for them, so you are not
							billed for sizes nobody watches.
						{/if}
					</p>
				</div>
			{/each}
		</div>
	</div>
{/if}

<style>
	.preset {
		padding: 16px;
		border-radius: var(--radius-md);
		border: 1px solid var(--color-sunk);
	}
	.preset--on {
		border-color: var(--color-brand);
	}
</style>
