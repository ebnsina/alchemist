<script lang="ts">
	import { fly, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Mail01Icon,
		SquareLock01Icon,
		Tick02Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { goto } from '$app/navigation';
	import { login, ApiError } from '$lib/api';

	// Two steps, like signing up. One thing on screen at a time reads as less to do,
	// and the address is asked for first because that is the half people remember.
	const steps = [
		{ key: 'email', icon: Mail01Icon, title: 'Welcome back', hint: 'The address you signed up with.' },
		{ key: 'password', icon: SquareLock01Icon, title: 'Your password', hint: 'The one you picked when you started.' }
	];

	let step = $state(0);
	let email = $state('');
	let password = $state('');
	let busy = $state(false);
	let error = $state('');
	let done = $state(false);
	let field: HTMLInputElement | null = $state(null);

	// The API rules on the address and the password. This only checks there is
	// something to send, so no rule lives in two places.
	const filled = $derived([email.trim(), password][step].length > 0);

	$effect(() => {
		if (step >= 0 && !done) field?.focus();
	});

	function back() {
		error = '';
		if (step > 0) step -= 1;
	}

	async function next() {
		error = '';
		// Safari's password autofill writes straight into the DOM without firing the
		// events bind:value listens for, so the bound state can still be empty while
		// the field visibly holds a value. Reading the element back is what makes
		// autofilled credentials work at all.
		if (field?.value) {
			if (step === 0) email = field.value;
			else password = field.value;
		}
		if (!filled) return;
		if (step < steps.length - 1) {
			step += 1;
			return;
		}
		busy = true;
		try {
			await login(email.trim(), password);
			done = true;
			// Straight to the dashboard. The success card is what they see on the way.
			goto('/app/');
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong. Please try again.';
			// Wrong credentials are wrong as a pair, and the address is rarely the
			// half that is wrong. Clear the password and stay put.
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
				class="mx-auto flex h-11 w-11 items-center justify-center rounded-full bg-solid text-on-solid"
				in:scale={{ start: 0.5, duration: 420, easing: cubicOut }}
			>
				<HugeiconsIcon icon={Tick02Icon} size={20} strokeWidth={2.6} />
			</span>
			<p class="sub mt-4">You are signed in as {email}.</p>
			<a href="/app/" class="btn-solid mt-6 w-full">Go to your dashboard</a>
		</div>
	{:else}
		<ol class="mb-6 flex items-center gap-1.5" aria-label="Progress">
			{#each steps as s, i (s.key)}
				<li
					class="step-dot"
					data-state={i < step ? 'done' : i === step ? 'current' : 'todo'}
					aria-current={i === step ? 'step' : undefined}
				>
					<span class="vh">{s.title}</span>
				</li>
			{/each}
		</ol>

		{#key step}
			<div in:fly={{ y: 14, duration: 340, delay: 80, easing: cubicOut }}>
				<div class="flex items-start justify-between gap-4">
					<div>
						<h2 class="title">{steps[step].title}</h2>
						<p class="sub mt-1.5">{steps[step].hint}</p>
					</div>
					<HugeiconsIcon
						icon={steps[step].icon}
						size={30}
						strokeWidth={1.6}
						class="flex-none text-accent"
					/>
				</div>

				<form
					class="mt-5"
					onsubmit={(e) => {
						e.preventDefault();
						next();
					}}
				>
					{#if step === 0}
						<label class="block">
							<span class="vh">Email address</span>
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
					{:else}
						<label class="block">
							<!-- The address stays in the DOM as a hidden field so a password
							     manager still sees the pair it needs to save or fill. -->
							<span class="vh">Password</span>
							<input type="hidden" name="username" autocomplete="username" value={email} />
							<input
								bind:this={field}
								bind:value={password}
								class="field"
								type="password"
								name="current-password"
								autocomplete="current-password"
								aria-invalid={error ? 'true' : undefined}
								required
							/>
						</label>
						<p class="mono mt-2.5">Signing in as {email}</p>
					{/if}

					{#if error}
						<p class="mt-3 text-sm text-red" role="alert">{error}</p>
					{/if}

					<div class="mt-6 flex items-center gap-2">
						{#if step > 0}
							<button type="button" class="btn flex-none" onclick={back} aria-label="Back">
								<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2.2} />
							</button>
						{/if}
						<button
							type="submit"
							class="btn-solid flex-1"
							disabled={busy || !filled}
							aria-disabled={busy || !filled}
						>
							{#if busy}
								Signing you in…
							{:else if step < steps.length - 1}
								Next
								<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2.2} />
							{:else}
								Sign in
							{/if}
						</button>
					</div>
				</form>
			</div>
		{/key}

		<p class="mt-6 text-center text-xs text-dim">
			New here? <a href="/signup/" class="link">Start free</a>
		</p>
	{/if}
</AuthShell>
