<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import Seo from '$lib/Seo.svelte';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { previewInvite, acceptInvite, ApiError } from '$lib/api';

	let invite = $state<{ email: string; role: string; org: string } | null>(null);
	let loading = $state(true);
	let error = $state('');
	let password = $state('');
	let busy = $state(false);
	let field: HTMLInputElement | null = $state(null);

	const token = $derived(page.url.searchParams.get('token') ?? '');

	$effect(() => {
		if (!token) {
			error = 'That link is missing its invite code. Ask for a new one.';
			loading = false;
			return;
		}
		previewInvite(token)
			.then((i) => (invite = i))
			.catch((e) => (error = e instanceof ApiError ? e.message : 'Something went wrong.'))
			.finally(() => (loading = false));
	});

	$effect(() => {
		if (invite) field?.focus();
	});

	async function accept(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			await acceptInvite(token, password);
			goto('/app/');
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
			password = '';
		} finally {
			busy = false;
		}
	}
</script>

<Seo title="Join the team — Alchemist" description="Accept your invitation to an Alchemist account." />

<AuthShell title={invite ? `Join ${invite.org}` : 'Your invitation'}>
	{#if loading}
		<div class="grid gap-3">
			<div class="sk h-5 w-2/3"></div>
			<div class="sk h-11"></div>
		</div>
	{:else if !invite}
		<div class="text-center">
			<p class="sub">{error}</p>
			<a href="/login/" class="btn mt-6 w-full">Go to sign in</a>
		</div>
	{:else}
		<div in:fly={{ y: 12, duration: 340, easing: cubicOut }}>
			<p class="sub">
				You have been invited to <strong class="text-ink">{invite.org}</strong> as
				{invite.role}. Pick a password and you are in.
			</p>
			<p class="mono mt-2">Your account will be {invite.email}</p>

			<form class="mt-5" onsubmit={accept}>
				<label class="block">
					<span class="vh">Password</span>
					<input type="hidden" name="username" autocomplete="username" value={invite.email} />
					<input
						bind:this={field}
						bind:value={password}
						class="field"
						type="password"
						name="new-password"
						autocomplete="new-password"
						placeholder="Pick a password"
						aria-invalid={error ? 'true' : undefined}
						required
					/>
				</label>

				{#if error}
					<p class="mt-3 text-sm text-red" role="alert">{error}</p>
				{/if}

				<button type="submit" class="btn-solid mt-6 w-full" disabled={busy || !password}>
					{busy ? 'Setting you up…' : 'Join ' + invite.org}
				</button>
			</form>
		</div>
	{/if}
</AuthShell>
