<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import { session, listMembers, ApiError, type Session, type Member } from '$lib/api';

	let me = $state<Session | null>(null);
	let mine = $state<Member | null>(null);
	let loading = $state(true);
	let error = $state('');

	// Who you are comes from the session; what you may do is on your own row in the
	// team, which is the only place the API says it.
	async function load() {
		error = '';
		try {
			me = await session();
			mine = (await listMembers()).members.find((m) => m.you) ?? null;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	const ROLE: Record<string, string> = {
		owner: 'Owner — full control, including other owners',
		admin: 'Admin — everything a member can, plus inviting and removing people',
		member: 'Member — upload, edit and read everything, but not change the team'
	};

	const when = (iso: string | null) =>
		iso ? new Intl.DateTimeFormat('en', { dateStyle: 'medium' }).format(new Date(iso)) : '—';
</script>

<Seo title="Your profile — Alchemist" description="The account you are signed in with." />

<h1 class="text-2xl font-semibold tracking-tight">Your profile</h1>
<p class="sub mt-1 max-w-xl">The account you are signed in with, and what it can do here.</p>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="mt-6 grid gap-3">
		{#each [0, 1] as i (i)}<div class="sk h-24"></div>{/each}
	</div>
{:else}
	<div class="card mt-6">
		<p class="title">You</p>
		<dl class="mt-4 grid gap-4 sm:grid-cols-2">
			<div class="min-w-0">
				<dt class="label">Email</dt>
				<dd class="mt-1 truncate text-sm font-medium">{me?.email ?? '—'}</dd>
			</div>
			<div class="min-w-0">
				<dt class="label">Can do</dt>
				<dd class="mt-1 text-sm">{mine ? (ROLE[mine.role] ?? mine.role) : '—'}</dd>
			</div>
			<div class="min-w-0">
				<dt class="label">Joined</dt>
				<dd class="mt-1 text-sm">{when(mine?.created_at ?? null)}</dd>
			</div>
			<div class="min-w-0">
				<dt class="label">Last signed in</dt>
				<dd class="mt-1 text-sm">{when(mine?.last_login_at ?? null)}</dd>
			</div>
		</dl>
	</div>

	<div class="card mt-4">
		<p class="title">Organization</p>
		<dl class="mt-4 grid gap-4 sm:grid-cols-2">
			<div class="min-w-0">
				<dt class="label">Name</dt>
				<dd class="mt-1 truncate text-sm font-medium">{me?.org ?? '—'}</dd>
			</div>
			<div class="min-w-0">
				<dt class="label">Tenant</dt>
				<dd class="mono mt-1 truncate">{me?.tenant_id ?? '—'}</dd>
			</div>
		</dl>
		<p class="sub mt-4">
			Everyone here shares one organization. <a href="/app/team/" class="link">Team</a> is where
			you add people to it, and <a href="/app/settings/" class="link">Branding</a> is what your
			viewers see.
		</p>
	</div>

	<!-- Said rather than shown as a dead form: there is no endpoint behind either, and
	     a button that cannot work is worse than a sentence that is true. -->
	<div class="card mt-4">
		<p class="title">Changing your email or password</p>
		<p class="sub mt-2">
			Not something you can do here yet. <a href="/contact/" class="link">Ask us</a> and we will
			do it for you.
		</p>
	</div>
{/if}
