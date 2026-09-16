<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Copy01Icon,
		Tick02Icon,
		Delete02Icon,
		UserAdd01Icon,
		UserRemove01Icon,
		Mail01Icon,
		SecurityIcon,
		ArrowLeft01Icon,
		ArrowRight01Icon
	} from '@hugeicons/core-free-icons';
	import { renderComponent, type ColumnDef } from '@tanstack/svelte-table';
	import Seo from '$lib/Seo.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import RowMenu from '$lib/components/RowMenu.svelte';
	import Badge from '$lib/components/Badge.svelte';
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
	let total = $state(0);
	let loading = $state(true);
	let error = $state('');
	let email = $state('');
	let role = $state('member');
	let busy = $state(false);
	let fresh = $state<{ email: string; link: string } | null>(null);
	let linkCard = $state<HTMLElement | null>(null);
	let copied = $state(false);
	let step = $state(0);
	let open = $state(false);

	let page = $state(0);
	let size = $state(10);
	let sorting = $state<{ id: string; desc: boolean }[]>([{ id: 'created_at', desc: true }]);
	let q = $state('');

	// Each dialog owns a plain boolean: passing !!row unbound means Escape closes it
	// and the next render opens it straight back up.
	let roleOpen = $state(false);
	let removeOpen = $state(false);
	let withdrawOpen = $state(false);
	let changing = $state<Member | null>(null);
	let removing = $state<Member | null>(null);
	let withdrawing = $state<Invite | null>(null);
	let newRole = $state('member');
	let busyRow = $state(false);

	// Pending invites always come back whole — the endpoint's total counts members only —
	// so this second table searches and pages the list it already has.
	let invitePage = $state(0);
	let inviteSize = $state(10);
	let inviteQ = $state('');
	const matching = $derived(
		invites.filter((i) => i.email.toLowerCase().includes(inviteQ.trim().toLowerCase()))
	);
	const invitePageRows = $derived(
		matching.slice(invitePage * inviteSize, invitePage * inviteSize + inviteSize)
	);

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

	// One read of every control the table owns, so a change to any of them refetches
	// exactly once rather than each firing its own request.
	const query = $derived({
		limit: size,
		offset: page * size,
		q,
		sort: sorting[0]?.id ?? 'created_at',
		order: (sorting[0]?.desc ?? true ? 'desc' : 'asc') as 'asc' | 'desc'
	});

	async function load() {
		error = '';
		try {
			const r = await listMembers(query);
			members = r.members;
			invites = r.invites;
			total = r.total;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		query;
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

	async function act(fn: () => Promise<unknown>, close?: () => void) {
		error = '';
		busyRow = true;
		try {
			await fn();
			close?.();
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			busyRow = false;
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
	const label = (r: string) => ROLES.find((x) => x.id === r)?.label ?? r;

	const columns: ColumnDef<any, Member>[] = [
		{ accessorKey: 'email', header: 'Email' },
		{
			accessorKey: 'role',
			header: 'Role',
			cell: (c) => {
				const m = c.row.original as Member;
				return renderComponent(Badge, { label: label(m.role), tone: m.you ? 'good' : 'idle' });
			}
		},
		{ accessorKey: 'created_at', header: 'Joined', cell: (c) => when(String(c.getValue())) },
		{
			id: 'last_login_at',
			header: 'Last signed in',
			enableSorting: false,
			cell: (c) => when((c.row.original as Member).last_login_at)
		},
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const m = c.row.original as Member;
				// Both controls are only ever refused on your own row: the API will not let
				// anyone remove themselves or leave the account without an owner.
				if (m.you) return 'You. Another owner has to change this';
				return renderComponent(RowMenu, {
					label: `Actions for ${m.email}`,
					actions: [
						{
							label: 'Role',
							icon: SecurityIcon,
							onclick: () => {
								changing = m;
								newRole = m.role;
								roleOpen = true;
							}
						},
						{
							label: 'Remove',
							icon: UserRemove01Icon,
							danger: true,
							onclick: () => {
								removing = m;
								removeOpen = true;
							}
						}
					]
				});
			}
		}
	];

	// Nothing here sorts: the invite list is whatever the API returned, and a header
	// that reordered only the rows on screen would lie about the ones it did not.
	const inviteColumns: ColumnDef<any, Invite>[] = [
		{ accessorKey: 'email', header: 'Email', enableSorting: false },
		{
			accessorKey: 'role',
			header: 'Invited as',
			enableSorting: false,
			cell: (c) => label(String(c.getValue()))
		},
		{
			accessorKey: 'expires_at',
			header: 'Expires',
			enableSorting: false,
			cell: (c) => until(String(c.getValue()))
		},
		{
			id: 'actions',
			header: '',
			enableSorting: false,
			cell: (c) => {
				const i = c.row.original as Invite;
				return renderComponent(RowMenu, {
					label: `Actions for the invite to ${i.email}`,
					actions: [
						{
							label: 'Withdraw',
							icon: Delete02Icon,
							danger: true,
							onclick: () => {
								withdrawing = i;
								withdrawOpen = true;
							}
						}
					]
				});
			}
		}
	];
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
		aria-label="Invite someone"
		class="btn-solid flex-none"
		onclick={() => {
			step = 0;
			open = true;
		}}
	>
		Add new
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
	<form id="invite-form" onsubmit={invite}>
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

		</div>
		</Steps>
	</form>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			{#if step > 0 && !busy}
				<button type="button" class="btn" onclick={() => (step -= 1)}>
					<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2.2} />
					Back
				</button>
			{/if}
			<!-- The form attribute keeps this bound to a form it is no longer inside. -->
			<button type="submit" form="invite-form" class="btn-solid" disabled={busy || !filled} aria-disabled={busy || !filled}>
				{#if busy}
					Inviting…
				{:else if step < steps.length - 1}
					Next
					<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2.2} />
				{:else}
					Make the invite link
				{/if}
			</button>
		</div>
	{/snippet}
</Dialog>

<h2 class="mt-10 text-lg font-semibold tracking-tight">People</h2>

<div class="mt-4">
	<DataTable
		{columns}
		rows={members}
		{total}
		{loading}
		bind:page
		bind:size
		bind:sorting
		bind:q
		searchLabel="Search by email"
	>
		{#snippet empty()}
			<p class="title">Nobody here</p>
			<p class="sub mx-auto mt-2 max-w-sm">
				{#if q}
					Nobody on the team matches that. Clear the search to see everyone.
				{:else}
					Nobody here yet, which should not be possible — try reloading.
				{/if}
			</p>
		{/snippet}
	</DataTable>
</div>

{#if invites.length > 0}
	<h2 class="mt-10 text-lg font-semibold tracking-tight">Waiting to accept</h2>
	<div class="mt-4">
		<DataTable
			columns={inviteColumns}
			rows={invitePageRows}
			total={matching.length}
			bind:page={invitePage}
			bind:size={inviteSize}
			bind:q={inviteQ}
			searchLabel="Search by email"
		>
			{#snippet empty()}
				<p class="title">No invite matches that</p>
				<p class="sub mx-auto mt-2 max-w-sm">Clear the search to see everyone invited.</p>
			{/snippet}
		</DataTable>
	</div>
{/if}

<Dialog
	bind:open={roleOpen}
	title="What may they do?"
	hint="Changing a role takes effect the next time they load a page."
>
	<fieldset>
		<legend class="vh">Role</legend>
		<div class="grid gap-2.5">
			{#each ROLES as r (r.id)}
				<Check
					type="radio"
					card
					name="new-role"
					value={r.id}
					bind:group={newRole}
					label={r.label}
					hint={r.hint}
				/>
			{/each}
		</div>
	</fieldset>
	<p class="sub mt-3">{changing?.email ?? ''}</p>

	{#snippet footer()}
		<div class="flex items-center justify-end gap-2">
			<button type="button" class="btn" onclick={() => (roleOpen = false)}>Cancel</button>
			<button
				type="button"
				class="btn-solid"
				disabled={busyRow}
				onclick={() =>
					changing &&
					act(() => setMemberRole(changing!.id, newRole), () => (roleOpen = false))}
			>
				{busyRow ? 'Saving…' : 'Save the role'}
			</button>
		</div>
	{/snippet}
</Dialog>

<Confirm
	bind:open={removeOpen}
	title="Remove them from the team?"
	confirm="Yes, remove them"
	destructive
	busy={busyRow}
	onconfirm={() =>
		removing && act(() => removeMember(removing!.id), () => (removeOpen = false))}
>
	<p class="text-sm">{removing?.email ?? ''}</p>
	<p class="sub mt-2">
		They lose access to this account immediately and any invite link they used is spent.
		Videos and keys are the account's, so nothing of theirs is deleted. You can invite them
		again afterwards.
	</p>
</Confirm>

<Confirm
	bind:open={withdrawOpen}
	title="Withdraw this invite?"
	confirm="Yes, withdraw it"
	destructive
	busy={busyRow}
	onconfirm={() =>
		withdrawing && act(() => withdrawInvite(withdrawing!.id), () => (withdrawOpen = false))}
>
	<p class="text-sm">{withdrawing?.email ?? ''}</p>
	<p class="sub mt-2">
		The link you passed on stops working. Nobody has used it yet — invite them again and
		they get a fresh one.
	</p>
</Confirm>
