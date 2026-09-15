<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Copy01Icon, Tick02Icon, Delete02Icon, UserAdd01Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
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
	let copied = $state(false);

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
		error = '';
		busy = true;
		try {
			const r = await inviteMember(email.trim(), role);
			fresh = { email: r.email, link: `${location.origin}/invite/?token=${r.token}` };
			email = '';
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

<h1 class="text-2xl font-semibold tracking-tight">Team</h1>
<p class="sub mt-1 max-w-xl">
	Everyone here can sign in to this account. An API key cannot change any of this — only a
	person who is signed in can.
</p>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if fresh}
	<div class="card mt-6">
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

<form class="card mt-6" onsubmit={invite}>
	<p class="title">Invite someone</p>
	<div class="mt-4 flex flex-wrap items-end gap-3">
		<label class="min-w-56 flex-1">
			<span class="label mb-1.5 block">Their email</span>
			<input bind:value={email} class="field" type="email" placeholder="name@example.com" required />
		</label>
		<label class="w-44">
			<span class="label mb-1.5 block">Role</span>
			<select bind:value={role} class="select">
				<option value="member">Member — can use it</option>
				<option value="admin">Admin — can manage the team</option>
				<option value="owner">Owner — full control</option>
			</select>
		</label>
		<button type="submit" class="btn-solid" disabled={busy || !email.trim()}>
			<HugeiconsIcon icon={UserAdd01Icon} size={15} strokeWidth={2} />
			{busy ? 'Inviting…' : 'Invite'}
		</button>
	</div>
</form>

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
					onclick={() => act(() => removeMember(m.id))}
					aria-label="Remove {m.email}"
				>
					<HugeiconsIcon icon={Delete02Icon} size={16} strokeWidth={1.8} />
				</button>
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
