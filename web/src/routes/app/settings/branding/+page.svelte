<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { ImageAdd02Icon, Delete02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';
	import {
		getBranding,
		putLogo,
		deleteLogo,
		whoami,
		ApiError,
		type Branding,
		type Whoami
	} from '$lib/api';

	let brand = $state<Branding | null>(null);
	let who = $state<Whoami | null>(null);
	let loading = $state(true);
	let busy = $state(false);
	let error = $state('');
	let removeOpen = $state(false);
	let input: HTMLInputElement | null = $state(null);

	async function load() {
		error = '';
		try {
			const [b, w] = await Promise.all([getBranding(), whoami()]);
			brand = b;
			who = w;
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
			removeOpen = false;
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}



	const logoPath = $derived(brand?.logo_url ?? null);
	const logoSrc = $derived(logoPath ? PUBLIC_ALCHEMIST_API + logoPath : null);
</script>

<Seo title="Branding — Alchemist" description="The name and logo your viewers see." />

<p class="sub max-w-xl">
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
					<button type="button" class="btn" onclick={() => (removeOpen = true)} disabled={busy}>
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

{/if}

<Confirm
	bind:open={removeOpen}
	title="Remove your logo?"
	confirm="Yes, remove it"
	destructive
	busy={busy}
	onconfirm={remove}
>
	<p class="sub">
		Every player and embed goes back to showing nothing in its place, straight away. You
		can upload it again whenever you like — we do not keep a copy.
	</p>
</Confirm>
