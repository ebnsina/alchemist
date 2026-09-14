<script lang="ts">
	import { fly, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Building06Icon,
		Mail01Icon,
		SquareLock01Icon,
		Tick02Icon,
		Copy01Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { signup, ApiError, type SignupResult } from '$lib/api';

	const steps = [
		{ key: 'org', icon: Building06Icon, title: 'Who is this for?', hint: 'The name your viewers will see. You can change it later.' },
		{ key: 'email', icon: Mail01Icon, title: 'Where do we reach you?', hint: 'One address. No newsletter, no drip campaign.' },
		{ key: 'password', icon: SquareLock01Icon, title: 'Pick a password', hint: 'Ten characters or more. A short sentence beats a clever word.' }
	];

	let step = $state(0);
	let org = $state('');
	let email = $state('');
	let password = $state('');
	let busy = $state(false);
	let error = $state('');
	let done: SignupResult | null = $state(null);
	let copied = $state(false);
	let field: HTMLInputElement | null = $state(null);

	// Length is the only rule worth showing: composition rules push people toward
	// Passw0rd! and nothing else.
	const strength = $derived(Math.min(3, Math.floor(password.length / 6)));
	const strengthWord = $derived(['Too short', 'Getting there', 'Good', 'Strong'][strength]);

	const valid = $derived(
		[org.trim().length > 0, /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email.trim()), password.length >= 10][
			step
		]
	);

	// Autofocus on arrival and on every step change, so the whole flow is typeable
	// without reaching for the mouse.
	$effect(() => {
		if (step >= 0 && !done) field?.focus();
	});

	async function next() {
		error = '';
		if (!valid) return;
		if (step < steps.length - 1) {
			step += 1;
			return;
		}
		busy = true;
		try {
			done = await signup(org.trim(), email.trim(), password);
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong. Please try again.';
			// Send them back to the field the API objected to, rather than leaving them
			// on the password step wondering which answer was wrong.
			if (e instanceof ApiError) {
				if (e.code === 'email_taken' || e.code === 'invalid_email') step = 1;
				if (e.code === 'invalid_org') step = 0;
			}
		} finally {
			busy = false;
		}
	}

	function back() {
		error = '';
		if (step > 0) step -= 1;
	}

	async function copyKey() {
		if (!done) return;
		try {
			await navigator.clipboard.writeText(done.api_key);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			// Clipboard can be blocked. The key is on screen and selectable either way.
			copied = false;
		}
	}
</script>

<Seo
	title="Start free — Alchemist"
	description="Create your organisation and start putting video online. 100 free videos a month, no card needed."
/>

<AuthShell title={done ? 'You are in' : 'Start free'}>
	{#if done}
		<div in:fly={{ y: 12, duration: 400, easing: cubicOut }}>
			<div class="flex items-center gap-3">
				<span
					class="flex h-10 w-10 flex-none items-center justify-center rounded-full bg-brand-mid text-on-brand"
					in:scale={{ start: 0.5, duration: 420, easing: cubicOut }}
				>
					<HugeiconsIcon icon={Tick02Icon} size={18} strokeWidth={2.6} />
				</span>
				<div>
					<p class="font-semibold">{done.org}</p>
					<p class="text-xs text-muted">{done.email}</p>
				</div>
			</div>

			<p class="mt-6 text-sm text-muted">
				Here is your key. It is how your own code talks to Alchemist. We keep only a
				scrambled copy, so this is the one time we can show it to you.
			</p>

			<div class="mt-3 flex items-center gap-2 rounded-xl border border-hairline bg-body px-3 py-2">
				<code class="truncate font-mono text-xs">{done.api_key}</code>
				<button
					type="button"
					onclick={copyKey}
					class="ml-auto flex flex-none items-center gap-1.5 text-xs text-brand-light"
				>
					<HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} size={14} strokeWidth={2} />
					{copied ? 'Copied' : 'Copy'}
				</button>
			</div>

			<a href="/docs/" class="btn-primary mt-6 w-full">Show me what to do with it</a>
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
						<h2 class="text-lg font-semibold tracking-tight">{steps[step].title}</h2>
						<p class="mt-1.5 text-sm text-muted">{steps[step].hint}</p>
					</div>
					<HugeiconsIcon
						icon={steps[step].icon}
						size={30}
						strokeWidth={1.6}
						class="flex-none text-brand-light"
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
							<span class="vh">Organisation name</span>
							<input
								bind:this={field}
								bind:value={org}
								class="field"
								type="text"
								name="organization"
								autocomplete="organization"
								placeholder="Nodi Academy"
								maxlength="120"
								required
							/>
						</label>
					{:else if step === 1}
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
							<span class="vh">Password</span>
							<input
								bind:this={field}
								bind:value={password}
								class="field"
								type="password"
								name="new-password"
								autocomplete="new-password"
								placeholder="At least 10 characters"
								minlength="10"
								required
							/>
						</label>
						<div class="mt-2.5 flex items-center gap-2">
							<div class="flex flex-1 gap-1">
								{#each [0, 1, 2] as bar (bar)}
									<span
										class="h-1 flex-1 rounded-full transition-colors duration-300 {bar < strength
											? 'bg-brand-mid'
											: 'bg-white/12'}"
									></span>
								{/each}
							</div>
							<span class="text-xs text-muted">{strengthWord}</span>
						</div>
					{/if}

					{#if error}
						<p class="mt-3 text-sm text-[#fca5a5]" role="alert">{error}</p>
					{/if}

					<div class="mt-6 flex items-center gap-3">
						{#if step > 0}
							<button type="button" class="btn-ghost flex-none" onclick={back}>
								<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2} />
								<span class="vh">Back</span>
							</button>
						{/if}
						<button type="submit" class="btn-primary flex-1" disabled={!valid || busy}>
							{#if busy}
								Setting things up…
							{:else if step < steps.length - 1}
								Next
								<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2} />
							{:else}
								Create my account
							{/if}
						</button>
					</div>
				</form>
			</div>
		{/key}

		<p class="mt-6 text-center text-xs text-muted">
			Already have an account? <a href="/login/" class="text-brand-light">Sign in</a>
		</p>
	{/if}
</AuthShell>
