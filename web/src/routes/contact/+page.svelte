<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Tick02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import { contact, ApiError } from '$lib/api';

	let name = $state('');
	let email = $state('');
	let org = $state('');
	let message = $state('');
	// Honeypot: shown to nothing with eyes, filled by most things without.
	let website = $state('');
	let busy = $state(false);
	let error = $state('');
	let sent = $state(false);

	const valid = $derived(
		name.trim().length > 0 &&
			/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email.trim()) &&
			message.trim().length >= 10
	);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		if (!valid) return;
		busy = true;
		error = '';
		try {
			await contact({
				name: name.trim(),
				email: email.trim(),
				org: org.trim(),
				message: message.trim(),
				website
			});
			sent = true;
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'That did not send. Try again.';
		} finally {
			busy = false;
		}
	}
</script>

<Seo
	title="Talk to us — Alchemist"
	description="Tell us what you are trying to put online and we will tell you whether we can help."
/>

<main id="main" class="mx-auto max-w-2xl px-4 pt-32 pb-20 sm:px-6 sm:pt-40 sm:pb-28">
	<h1 class="text-3xl leading-[1.32] font-semibold tracking-tight sm:text-4xl">Talk to us</h1>
	<p class="mt-3 max-w-xl text-dim">
		Tell us what you are putting online and roughly how much of it. We answer with what it would
		cost and whether we are the right people for it.
	</p>

	{#if sent}
		<div class="card mt-8 p-8 text-center" in:fly={{ y: 12, duration: 380, easing: cubicOut }}>
			<span
				class="mx-auto flex h-11 w-11 items-center justify-center rounded-full bg-solid text-on-solid"
			>
				<HugeiconsIcon icon={Tick02Icon} size={20} strokeWidth={2.6} />
			</span>
			<p class="mt-4 font-semibold">We have it</p>
			<p class="mx-auto mt-2 max-w-sm text-sm text-dim">
				Someone reads every one of these. Expect a reply at {email} within a working day.
			</p>
			<a href="/" class="btn mt-6">Back to the start</a>
		</div>
	{:else}
		<form class="card mt-8 grid gap-4 p-6 sm:p-8" onsubmit={submit}>
			<div class="grid gap-4 sm:grid-cols-2">
				<label class="block">
					<span class="mb-1.5 block text-xs text-dim">Your name</span>
					<input bind:value={name} class="field" type="text" name="name" autocomplete="name" required />
				</label>
				<label class="block">
					<span class="mb-1.5 block text-xs text-dim">Email</span>
					<input
						bind:value={email}
						class="field"
						type="email"
						name="email"
						autocomplete="email"
						placeholder="you@example.com"
						required
					/>
				</label>
			</div>

			<label class="block">
				<span class="mb-1.5 block text-xs text-dim">Organisation <span class="text-dim">(optional)</span></span>
				<input bind:value={org} class="field" type="text" name="organization" autocomplete="organization" />
			</label>

			<label class="block">
				<span class="mb-1.5 block text-xs text-dim">What are you putting online?</span>
				<textarea
					bind:value={message}
					class="field min-h-32 resize-y"
					name="message"
					placeholder="About 40 recorded lectures a month, mostly an hour long, watched inside Bangladesh."
					required
				></textarea>
			</label>

			<!-- Not display:none — some bots skip hidden fields but fill offscreen ones. -->
			<label class="vh" aria-hidden="true">
				Website
				<input bind:value={website} type="text" name="website" tabindex="-1" autocomplete="off" />
			</label>

			{#if error}
				<p class="text-sm text-red" role="alert">{error}</p>
			{/if}

			<div class="flex items-center gap-4">
				<button type="submit" class="btn-solid" disabled={!valid || busy}>
					{busy ? 'Sending…' : 'Send it'}
				</button>
				<p class="text-xs text-dim">No newsletter. We reply and that is all.</p>
			</div>
		</form>
	{/if}
</main>
