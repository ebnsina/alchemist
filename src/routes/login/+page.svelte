<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Tick02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { goto } from '$app/navigation';
	import { login, ApiError } from '$lib/api';

	let email = $state('');
	let password = $state('');
	let busy = $state(false);
	let error = $state('');
	let done = $state(false);
	let field: HTMLInputElement | null = $state(null);

	$effect(() => {
		field?.focus();
	});

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			await login(email.trim(), password);
			done = true;
			// Straight to the dashboard. The success card is what they see on the way.
			goto('/app/');
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong. Please try again.';
			password = '';
		} finally {
			busy = false;
		}
	}
</script>

<Seo title="Sign in — Alchemist" description="Sign in to your Alchemist account." />

<AuthShell title={done ? 'Welcome back' : 'Sign in'}>
	{#if done}
		<div class="text-center" in:fly={{ y: 12, duration: 360, easing: cubicOut }}>
			<span
				class="mx-auto flex h-11 w-11 items-center justify-center rounded-full bg-tertiary text-on-primary"
			>
				<HugeiconsIcon icon={Tick02Icon} size={20} strokeWidth={2.6} />
			</span>
			<p class="mt-4 text-sm text-secondary">You are signed in as {email}.</p>
			<a href="/app/" class="btn-primary mt-6 w-full">Go to your dashboard</a>
		</div>
	{:else}
		<form onsubmit={submit}>
			<label class="block">
				<span class="mb-1.5 block text-xs text-secondary">Email</span>
				<input
					bind:this={field}
					bind:value={email}
					class="field"
					type="email"
					name="email"
					autocomplete="email"
					placeholder="you@example.com"
					aria-invalid={error ? 'true' : undefined}
					required
				/>
			</label>

			<label class="mt-4 block">
				<span class="mb-1.5 block text-xs text-secondary">Password</span>
				<input
					bind:value={password}
					class="field"
					type="password"
					name="password"
					autocomplete="current-password"
					aria-invalid={error ? 'true' : undefined}
					required
				/>
			</label>

			{#if error}
				<p class="mt-3 text-sm text-danger" role="alert">{error}</p>
			{/if}

			<button type="submit" class="btn-primary mt-6 w-full" disabled={busy}>
				{busy ? 'Signing you in…' : 'Sign in'}
			</button>
		</form>

		<p class="mt-6 text-center text-xs text-secondary">
			New here? <a href="/signup/" class="text-tertiary">Start free</a>
		</p>
	{/if}
</AuthShell>
