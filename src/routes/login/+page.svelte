<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Tick02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AuthShell from '$lib/components/AuthShell.svelte';
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
				class="mx-auto flex h-11 w-11 items-center justify-center rounded-full bg-brand-mid text-on-brand"
			>
				<HugeiconsIcon icon={Tick02Icon} size={20} strokeWidth={2.6} />
			</span>
			<p class="mt-4 text-sm text-muted">You are signed in as {email}.</p>
			<a href="/docs/" class="btn-primary mt-6 w-full">Pick up where you left off</a>
		</div>
	{:else}
		<form onsubmit={submit}>
			<label class="block">
				<span class="mb-1.5 block text-xs text-muted">Email</span>
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
				<span class="mb-1.5 block text-xs text-muted">Password</span>
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
				<p class="mt-3 text-sm text-[#fca5a5]" role="alert">{error}</p>
			{/if}

			<button type="submit" class="btn-primary mt-6 w-full" disabled={busy}>
				{busy ? 'Signing you in…' : 'Sign in'}
			</button>
		</form>

		<p class="mt-6 text-center text-xs text-muted">
			New here? <a href="/signup/" class="text-brand-light">Start free</a>
		</p>
	{/if}
</AuthShell>
