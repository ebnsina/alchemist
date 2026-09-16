<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Copy01Icon,
		Tick02Icon,
		Delete02Icon,
		UserAdd01Icon,
		Mail01Icon,
		SecurityIcon,
		ArrowLeft01Icon,
		ArrowRight01Icon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Check from '$lib/components/Check.svelte';
	import {
		listMembers,
		inviteMember,
		withdrawInvite,
		setMemberRole,
		removeMember,
		ApiError,
		type Member,
		type Invite
	} from '$lib/api';

	let members: Member[] = $state([]);
	let invites: Invite[] = $state([]);
	let loading = $state(true);
	let error = $state('');
	let email = $state('');
	let role = $state('member');
	let busy = $state(false);
	let fresh = $state<{ email: string; link: string } | null>(null);
	let linkCard = $state<HTMLElement | null>(null);
	let copied = $state(false);
	let confirming = $state('');
	let step = $state(0);
	let open = $state(false);

	// Who, then what they may do, then a link to hand over. Choosing a role in the
	// same breath as typing an address is how people invite an owner by accident.
	const steps = [
		{
			key: 'who',
			icon: Mail01Icon,
			title: 'Who are you inviting?',
			hint: 'The address they will sign in with.'
		},
		{
			key: 'role',
			icon: SecurityIcon,
			title: 'What may they do?',
			hint: 'You can change this afterwards, from the list below.'
		},
		{
			key: 'send',
			icon: UserAdd01Icon,
			title: 'Ready to invite them?',
			hint: 'You get a link to pass on. We do not send the email yet.'
		}
	];

	const ROLES = [
		{ id: 'member', label: 'Member', hint: 'Upload, edit and read everything. Cannot change the team.' },
		{ id: 'admin', label: 'Admin', hint: 'Everything a member can, plus inviting and removing people.' },
		{ id: 'owner', label: 'Owner', hint: 'Full control, including other owners. Only an owner may add one.' }
	];

	const filled = $derived([email.trim().length > 0, role !== '', true][step]);

	async function load() {
		error = '';
		try {
			const r = await listMembers();
			members = r.members;
			invites = r.invites;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	// Whether this person may invite, and at what role, is the API's ruling. The page
	// sends the request and shows the answer rather than deciding in advance.
	async function invite(e: SubmitEvent) {
		e.preventDefault();
		if (!filled) return;
		if (step < steps.length - 1) {
			step += 1;
			return;
		}
		error = '';
		busy = true;
		try {
			const r = await inviteMember(email.trim(), role);
			step = 0;
			open = false;
			fresh = { email: r.email, link: `${location.origin}/invite/?token=${r.token}` };
			email = '';
			// Queued so it lands after the dialog's own close puts focus back on "Invite
			// someone"; the link is shown once and it, not the button, gets the cursor.
			queueMicrotask(() => linkCard?.focus());
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}

	async function act(fn: () => Promise<unknown>) {
		error = '';
		try {
			await fn();
			confirming = '';
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		}
	}

	async function copy(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			copied = false;
		}
	}

	const when = (iso: string | null) =>
		iso ? new Intl.DateTimeFormat('en', { dateStyle: 'medium' }).format(new Date(iso)) : 'never';
	const until = (iso: string) =>
		new Intl.RelativeTimeFormat('en', { numeric: 'auto' }).format(
			Math.round((new Date(iso).getTime() - Date.now()) / 86400000),
			'day'
		);
</script>

<Seo title="Team — Alchemist" description="Who can reach this account, and what they can do." />

<header class="flex flex-wrap items-start justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Team</h1>
		<p class="sub mt-1 max-w-xl">
			Everyone here can sign in to this account. An API key cannot change any of this —
			only a person who is signed in can.
		</p>
	</div>
	<button
		type="button"
		class="btn-solid flex-none"
		onclick={() => {
			step = 0;
			open = true;
		}}
	>
		Invite someone
	</button>
</header>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if fresh}
	<div class="card mt-6" tabindex="-1" bind:this={linkCard}>
		<p class="title">Invite ready for {fresh.email}</p>
		<p class="sub mt-2">
			We do not send the email yet, so pass this link on yourself. It works once, and only for
			the next seven days.
		</p>
		<div class="mt-3 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
			<code class="truncate font-mono text-xs">{fresh.link}</code>
			<button
				type="button"
				onclick={() => copy(fresh!.link)}
				class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
			>
				<HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} size={14} strokeWidth={2} />
				{copied ? 'Copied' : 'Copy'}
			</button>
		</div>
		<button type="button" class="btn btn-sm mt-4" onclick={() => (fresh = null)}>Done</button>
	</div>
{/if}

<Dialog bind:open title="Invite someone">
	<form onsubmit={invite}>
		<Steps {steps} {step}>
		<div class="mt-5">
			{#if step === 0}
				<label class="block">
					<span class="vh">Their email</span>
					<input
						bind:value={email}
						class="field"
						type="email"
						placeholder="name@example.com"
						required
					/>
				</label>
			{:else if step === 1}
				<fieldset>
					<legend class="vh">Role</legend>
					<div class="grid gap-2.5">
						{#each ROLES as r (r.id)}
							<Check
								type="radio"
								card
								name="role"
								value={r.id}
								bind:group={role}
								label={r.label}
								hint={r.hint}
							/>
						{/each}
					</div>
				</fieldset>
			{:else}
				<dl class="grid gap-3 rounded-md border border-sunk p-4">
					<div class="flex items-baseline justify-between gap-4">
						<dt class="label">Inviting</dt>
						<dd class="truncate text-sm font-medium">{email}</dd>
					</div>
					<div class="flex items-baseline justify-between gap-4">
						<dt class="label">As</dt>
						<dd class="text-sm">{ROLES.find((r) => r.id === role)?.label ?? role}</dd>
					</div>
				</dl>
				<p class="sub mt-3">
					Nothing is emailed. You get a link to pass on yourself, good once and for seven
					days, and you can withdraw it until they use it.
				</p>
			{/if}

			{#if error}
				<p class="mt-3 text-sm text-red" role="alert">{error}</p>
			{/if}

			<div class="mt-6 flex items-center gap-2">
				{#if step > 0 && !busy}
					<button
						type="button"
						class="btn flex-none"
						onclick={() => (step -= 1)}
						aria-label="Back"
					>
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
						Inviting…
					{:else if step < steps.length - 1}
						Next
						<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2.2} />
					{:else}
						<HugeiconsIcon icon={UserAdd01Icon} size={15} strokeWidth={2} />
						Make the invite link
					{/if}
				</button>
			</div>
		</div>
		</Steps>
	</form>
</Dialog>

<h2 class="mt-10 text-lg font-semibold tracking-tight">People</h2>
{#if loading}
	<div class="mt-4 grid gap-2">
		{#each [0, 1, 2] as i (i)}<div class="sk h-14"></div>{/each}
	</div>
{:else if members.length === 0}
	<p class="sub mt-4">Nobody here yet, which should not be possible — try reloading.</p>
{:else}
	<ul class="mt-4 divide-y divide-sunk border-y border-sunk">
		{#each members as m (m.id)}
			<li class="flex flex-wrap items-center gap-x-4 gap-y-2 py-3.5">
				<div class="min-w-48 flex-1">
					<p class="text-sm font-medium">
						{m.email}
						{#if m.you}<span class="chip ml-2">You</span>{/if}
					</p>
					<p class="mono mt-0.5">
						Joined {when(m.created_at)} ·
						{m.last_login_at ? `last signed in ${when(m.last_login_at)}` : 'has not signed in yet'}
					</p>
				</div>
				{#if m.you}
					<!-- Both controls only ever refused on your own row: the API will not let
					     anyone remove themselves or leave the account without an owner. -->
					<span class="chip w-40 justify-center">{m.role}</span>
					<span class="sub flex-none">Another owner has to change this</span>
				{:else}
					<select
						class="select w-40"
						value={m.role}
						onchange={(e) => act(() => setMemberRole(m.id, e.currentTarget.value))}
						aria-label="Role for {m.email}"
					>
						<option value="member">Member</option>
						<option value="admin">Admin</option>
						<option value="owner">Owner</option>
					</select>
					<button
						type="button"
						class="icon-btn"
						onclick={() => (confirming = confirming === m.id ? '' : m.id)}
						aria-label="Remove {m.email}"
					>
						<HugeiconsIcon icon={Delete02Icon} size={16} strokeWidth={1.8} />
					</button>
				{/if}
				{#if confirming === m.id}
					<div class="w-full rounded-md border border-sunk p-4">
						<p class="text-sm">Remove {m.email}?</p>
						<p class="sub mt-1.5">
							They lose access to this account immediately and any invite link they used is
							spent. Videos and keys are the account's, so nothing of theirs is deleted. You
							can invite them again afterwards.
						</p>
						<div class="mt-3 flex flex-wrap gap-2">
							<button
								type="button"
								class="btn btn-sm"
								onclick={() => act(() => removeMember(m.id))}
							>
								Yes, remove them
							</button>
							<button type="button" class="btn btn-sm" onclick={() => (confirming = '')}>
								Keep them
							</button>
						</div>
					</div>
				{/if}
			</li>
		{/each}
	</ul>
{/if}

{#if invites.length > 0}
	<h2 class="mt-10 text-lg font-semibold tracking-tight">Waiting to accept</h2>
	<ul class="mt-4 divide-y divide-sunk border-y border-sunk">
		{#each invites as i (i.id)}
			<li class="flex flex-wrap items-center gap-x-4 gap-y-2 py-3.5">
				<div class="min-w-48 flex-1">
					<p class="text-sm font-medium">{i.email}</p>
					<p class="mono mt-0.5">Invited as {i.role} · Expires {until(i.expires_at)}</p>
				</div>
				<button type="button" class="btn btn-sm" onclick={() => act(() => withdrawInvite(i.id))}>
					Withdraw
				</button>
			</li>
		{/each}
	</ul>
{/if}
