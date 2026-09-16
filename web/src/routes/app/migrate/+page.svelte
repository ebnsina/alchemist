<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		ArrowDataTransferHorizontalIcon,
		SquareLock01Icon,
		Settings02Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon,
		Alert02Icon,
		CustomerSupportIcon,
		CheckmarkCircle02Icon,
		PauseIcon,
		PlayIcon,
		Delete02Icon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Check from '$lib/components/Check.svelte';
	import {
		listMigrationProviders,
		listMigrations,
		listMigrationItems,
		startMigration,
		confirmMigration,
		pauseMigration,
		resumeMigration,
		deleteMigration,
		ApiError,
		type MigrationProvider,
		type Migration,
		type MigrationItem
	} from '$lib/api';

	let providers = $state<MigrationProvider[]>([]);
	let migrations = $state<Migration[]>([]);
	let items = $state<Record<string, MigrationItem[]>>({});
	let itemsFailed = $state<Record<string, boolean>>({});
	let loading = $state(true);
	let error = $state('');
	let busy = $state('');
	let confirming = $state('');
	let open = $state(false);
	let step = $state(0);
	let provider = $state('');
	let secret = $state('');
	let config = $state<Record<string, string>>({});

	const chosen = $derived(providers.find((p) => p.name === provider) ?? null);

	const steps = [
		{
			key: 'who',
			icon: ArrowDataTransferHorizontalIcon,
			title: 'Where is it now?',
			hint: 'We only list hosts that will hand back a file. Some never do.'
		},
		{
			key: 'settings',
			icon: Settings02Icon,
			title: 'Which library?',
			hint: 'The bits that are not secret, so we know where to look.'
		},
		{
			key: 'key',
			icon: SquareLock01Icon,
			title: 'How do we read it?',
			hint: 'An API key with read access. Encrypted before it is stored.'
		}
	];

	// A provider with nothing extra to ask skips the middle step rather than showing
	// an empty page with a Next button on it.
	const visible = $derived(chosen && chosen.config.length === 0 ? [steps[0], steps[2]] : steps);
	const stepKey = $derived(visible[step]?.key ?? 'who');

	const filled = $derived(
		stepKey === 'who'
			? provider !== ''
			: stepKey === 'settings'
				? (chosen?.config ?? []).every((f) => (config[f.key] ?? '').trim().length > 0)
				: secret.trim().length > 0
	);

	async function load() {
		error = '';
		try {
			const [p, m] = await Promise.all([listMigrationProviders(), listMigrations()]);
			providers = p.providers;
			migrations = m.migrations;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	// Anything still moving is worth another look without the customer asking.
	$effect(() => {
		if (!migrations.some((m) => m.state === 'scanning' || m.state === 'previewing')) return;
		const id = setInterval(load, 5000);
		return () => clearInterval(id);
	});

	// A finished preview is the whole point of the screen, so its list is fetched
	// rather than waiting behind a button.
	$effect(() => {
		for (const m of migrations) {
			if (m.state === 'previewing' && m.preview_done && !items[m.id] && !itemsFailed[m.id])
				void fetchItems(m.id);
		}
	});

	async function fetchItems(id: string) {
		try {
			items = { ...items, [id]: (await listMigrationItems(id)).items };
			itemsFailed = { ...itemsFailed, [id]: false };
		} catch (e) {
			// Remembered, or the skeleton below waits for a list that is never coming.
			itemsFailed = { ...itemsFailed, [id]: true };
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		}
	}

	function back() {
		error = '';
		if (step > 0) step -= 1;
	}

	async function next() {
		error = '';
		if (!filled) return;
		if (step < visible.length - 1) {
			step += 1;
			return;
		}
		busy = 'new';
		try {
			await startMigration(provider, secret.trim(), config);
			open = false;
			step = 0;
			provider = '';
			secret = '';
			config = {};
			await load();
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			busy = '';
		}
	}

	async function act(id: string, fn: (id: string) => Promise<unknown>) {
		error = '';
		busy = id;
		try {
			await fn(id);
			confirming = '';
			const { [id]: _drop, ...rest } = items;
			items = rest;
			await load();
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			busy = '';
		}
	}

	// Counting client-side is a courtesy, not a rule: the API parses the list again
	// and its count is the one that matters.
	const counted = $derived(secret.split('\n').filter((l) => /https?:\/\//.test(l)).length);

	async function readFile(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		secret = await file.text();
	}

	async function toggleItems(id: string) {
		if (items[id]) {
			const { [id]: _drop, ...rest } = items;
			items = rest;
			return;
		}
		await fetchItems(id);
	}

	const STATE: Record<string, { label: string; means: string }> = {
		listing: {
			label: 'Having a look',
			means: 'We are reading your library to see what is there. Nothing has been brought across.'
		},
		previewing: {
			label: 'Waiting on you',
			means:
				'This is everything we found. Nothing has been brought across yet, and nothing will be until you say so.'
		},
		scanning: {
			label: 'Bringing them across',
			means: 'We are working through the list, taking each video as we reach it.'
		},
		done: { label: 'Finished', means: 'Every video we could get has been brought across.' },
		failed: {
			label: 'Stopped',
			means: 'Something went wrong that we cannot get past on our own.'
		},
		paused: {
			label: 'On hold',
			means: 'Stopped for now. Everything already brought across is still here.'
		}
	};

	// The two halves of 'previewing' are different screens, so they get different copy.
	const phase = (m: Migration) =>
		m.state === 'previewing' && !m.preview_done ? 'listing' : m.state;

	const label = (name: string) => providers.find((p) => p.name === name)?.label ?? name;
	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(
			new Date(iso)
		);
	const pct = (m: Migration) => (m.total === 0 ? 0 : Math.round((m.handled / m.total) * 100));
	const portion = new Intl.NumberFormat('en');

</script>

<Seo title="Move a library — Alchemist" description="Bring your videos across from another host." />

<header class="flex flex-wrap items-start justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Move a library</h1>
		<p class="sub mt-1 max-w-xl">
			Give us read access to your account somewhere else and we will show you what is there.
			You decide what happens next — nothing is brought across until you say so, and nothing
			is deleted on the far side.
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
		Start a migration
	</button>
</header>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

<Dialog bind:open title="Start a migration">
	<Steps steps={visible} {step}>
			<div class="mt-5 grid gap-4">
				{#if stepKey === 'who'}
					<div class="grid gap-2.5 sm:grid-cols-2">
						{#each providers as p (p.name)}
							<Check
								type="radio"
								card
								name="provider"
								value={p.name}
								bind:group={provider}
								label={p.label}
								hint={p.secret.kind === 'text'
									? 'For anywhere else — you bring the links'
									: p.config.length
										? 'Needs a library ID as well'
										: 'Just an access token'}
							/>
						{/each}
					</div>
					<div class="flex gap-2.5 rounded-md border border-sunk p-4">
						<HugeiconsIcon
							icon={Alert02Icon}
							size={16}
							strokeWidth={1.8}
							class="mt-0.5 flex-none text-faint"
						/>
						<div>
							<p class="sub">
								Not every host is on this list. Some never release the original file at all —
								that is how they are built, so there is no way to fetch it. Export your links
								from their dashboard and pick
								<strong class="text-ink">A list of links</strong> instead.
							</p>
							<a href="/contact/" class="btn btn-sm mt-3">
								<HugeiconsIcon icon={CustomerSupportIcon} size={15} strokeWidth={1.8} />
								Not seeing your host? Talk to us
							</a>
						</div>
					</div>
				{:else if stepKey === 'settings'}
					{#each chosen?.config ?? [] as f (f.key)}
						<label class="block">
							<span class="label mb-1.5 block">{f.label}</span>
							<input
								class="field"
								type="text"
								value={config[f.key] ?? ''}
								oninput={(e) => (config = { ...config, [f.key]: e.currentTarget.value })}
								required
							/>
							<span class="sub mt-1.5 block">{f.hint}</span>
						</label>
					{/each}
				{:else}
					<label class="block">
						<span class="label mb-1.5 block">{chosen?.secret.label ?? 'API key'}</span>
						{#if chosen?.secret.kind === 'text'}
							<textarea
								bind:value={secret}
								class="field min-h-44 resize-y font-mono text-xs"
								placeholder={'https://example.com/lectures/week-1.mp4\nhttps://example.com/lectures/week-2.mp4'}
								spellcheck="false"
								required
							></textarea>
						{:else}
							<input
								bind:value={secret}
								class="field"
								type="password"
								autocomplete="off"
								spellcheck="false"
								required
							/>
						{/if}
						<span class="sub mt-1.5 block">{chosen?.secret.hint ?? ''}</span>
					</label>

					{#if chosen?.secret.kind === 'text'}
						<div class="flex flex-wrap items-center gap-3">
							<label class="btn btn-sm cursor-pointer">
								Load a .csv file
								<input
									type="file"
									accept=".csv,.txt,text/csv,text/plain"
									class="vh"
									onchange={readFile}
								/>
							</label>
							{#if counted > 0}
								<p class="mono">{counted} links found</p>
							{/if}
						</div>
						<p class="sub">
							Nothing is deleted anywhere. Each link is fetched the same way a single URL
							import is, so anything pointing at a private address is refused.
						</p>
					{:else}
						<p class="sub">
							Read access is enough — we never write to your account there, and nothing is
							deleted. The key is encrypted before it is stored, tied to this one migration,
							and never shown again.
						</p>
					{/if}
					{#if provider === 'vimeo'}
						<p class="sub">
							Vimeo only exposes files on a paid plan, and the token needs the
							<code class="font-mono text-xs">video_files</code> scope. Without it every video
							comes back with nothing to download.
						</p>
					{:else if provider === 'bunny'}
						<p class="sub">
							Bunny only makes a downloadable file when MP4 Fallback is on for the library,
							and only for videos uploaded after you turned it on. Its fallback also caps at
							1080p, so anything sharper comes across at 1080p.
						</p>
					{/if}
					<p class="sub">
						We will read the library and show you the list first. Nothing is brought across
						until you have seen it and said yes.
					</p>
				{/if}

				<div class="flex items-center gap-2">
					{#if step > 0 && busy !== 'new'}
						<button type="button" class="btn flex-none" onclick={back} aria-label="Back">
							<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2.2} />
						</button>
					{/if}
					<button
						type="button"
						class="btn-solid flex-1"
						onclick={next}
						disabled={busy === 'new' || !filled}
						aria-disabled={busy === 'new' || !filled}
					>
						{#if busy === 'new'}
							Having a look…
						{:else if step < visible.length - 1}
							Next
							<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2.2} />
						{:else}
							Show me what is there
						{/if}
					</button>
				</div>
			</div>
	</Steps>
</Dialog>

{#if loading}
	<div class="mt-6 grid gap-3">
		{#each [0, 1] as i (i)}
			<section class="card">
				<div class="flex items-start justify-between gap-4">
					<div class="min-w-0 flex-1">
						<div class="sk h-4 w-32"></div>
						<div class="sk mt-2.5 h-3 w-full max-w-md"></div>
						<div class="sk mt-2 h-3 w-40"></div>
					</div>
					<div class="sk h-6 w-24 flex-none"></div>
				</div>
				<div class="sk mt-5 h-1.5 w-full"></div>
			</section>
		{/each}
	</div>
{:else if migrations.length === 0}
	<div class="card mt-6 py-8 text-center">
		<p class="title">Nothing moved yet</p>
		<p class="sub mx-auto mt-2 max-w-md">
			This is a one-off for when you are arriving from somewhere else. Use
			<b>Start a migration</b> and we will show you what is there before anything moves. Day
			to day, videos come in through the API or a
			<a href="/app/sources/" class="link">connected bucket</a>.
		</p>
	</div>
{:else}
	<div class="mt-6 grid gap-3">
		{#each migrations as m (m.id)}
			{@const p = phase(m)}
			<section class="card">
				<div class="flex flex-wrap items-start justify-between gap-4">
					<div class="min-w-0">
						<p class="title">{label(m.provider)}</p>
						<p class="sub mt-1 max-w-lg">{STATE[p].means}</p>
						<p class="mono mt-1.5">Started {when(m.created_at)}</p>
					</div>
					<span class="chip flex-none" class:chip-on={m.state === 'done'}>
						{STATE[p].label}
					</span>
				</div>

				{#if m.last_error}
					<p class="mt-3 text-sm text-red" role="alert">{m.last_error}</p>
				{/if}

				<div class="mt-4 border-t border-sunk pt-4">
					{#if p === 'listing'}
						<div class="sk h-1.5 w-full"></div>
						<p class="sub mt-2">
							{portion.format(m.total)} found so far.
						</p>
					{:else if p === 'previewing'}
						<p class="text-sm">
							<strong class="num">{portion.format(m.total)}</strong> videos ready to come
							across. Nothing here is yours yet.
						</p>
					{:else}
						<div
							class="h-1.5 w-full overflow-hidden rounded-full bg-sunk"
							role="progressbar"
							aria-valuenow={pct(m)}
							aria-valuemin={0}
							aria-valuemax={100}
							aria-label="How far this migration has got"
						>
							<div class="h-full rounded-full bg-brand" style="width: {pct(m)}%"></div>
						</div>
						<p class="sub mt-2">
							{portion.format(m.handled)} of {portion.format(m.total)} done —
							{portion.format(m.imported)} brought across,
							{portion.format(m.skipped)} left behind.
						</p>
					{/if}
				</div>

				<div class="mt-4 flex flex-wrap items-center gap-2">
					{#if p === 'previewing'}
						<button
							type="button"
							class="btn-solid btn-sm"
							disabled={busy === m.id}
							onclick={() => act(m.id, confirmMigration)}
						>
							<HugeiconsIcon icon={CheckmarkCircle02Icon} size={15} strokeWidth={1.9} />
							{busy === m.id ? 'Starting…' : 'Yes, bring these across'}
						</button>
					{:else if p === 'scanning'}
						<button
							type="button"
							class="btn btn-sm"
							disabled={busy === m.id}
							onclick={() => act(m.id, pauseMigration)}
						>
							<HugeiconsIcon icon={PauseIcon} size={15} strokeWidth={1.9} />
							Stop for now
						</button>
					{:else if p === 'paused' || p === 'failed'}
						<button
							type="button"
							class="btn-solid btn-sm"
							disabled={busy === m.id}
							onclick={() => act(m.id, resumeMigration)}
						>
							<HugeiconsIcon icon={PlayIcon} size={15} strokeWidth={1.9} />
							Carry on
						</button>
					{/if}

					{#if m.total > 0 && p !== 'previewing'}
						<button type="button" class="btn btn-sm" onclick={() => toggleItems(m.id)}>
							{items[m.id] ? 'Hide the list' : 'See every video'}
						</button>
					{/if}

					<button
						type="button"
						class="btn btn-sm ml-auto"
						disabled={busy === m.id}
						onclick={() => (confirming = confirming === m.id ? '' : m.id)}
					>
						<HugeiconsIcon icon={Delete02Icon} size={15} strokeWidth={1.9} />
						{p === 'done' ? 'Forget this' : 'Cancel'}
					</button>
				</div>

				{#if confirming === m.id}
					<div class="mt-3 rounded-md border border-sunk p-4">
						<p class="text-sm">Cancel this migration?</p>
						<p class="sub mt-1.5">
							We forget the list and the key. Any videos already brought across stay in your
							library, and nothing at {label(m.provider)} is touched.
						</p>
						<div class="mt-3 flex flex-wrap gap-2">
							<button
								type="button"
								class="btn btn-sm"
								disabled={busy === m.id}
								onclick={() => act(m.id, deleteMigration)}
							>
								{busy === m.id ? 'Cancelling…' : 'Yes, cancel it'}
							</button>
							<button type="button" class="btn btn-sm" onclick={() => (confirming = '')}>
								Keep it
							</button>
						</div>
					</div>
				{/if}

				{#if p === 'previewing' && itemsFailed[m.id]}
					<div class="mt-4 border-t border-sunk pt-4">
						<p class="sub">We could not load the list this time. Nothing has moved.</p>
						<button type="button" class="btn btn-sm mt-3" onclick={() => fetchItems(m.id)}>
							Try again
						</button>
					</div>
				{:else if p === 'previewing' && !items[m.id]}
					<ul class="mt-4 grid gap-2 border-t border-sunk pt-4">
						{#each [0, 1, 2] as i (i)}<li class="sk h-4 w-full"></li>{/each}
					</ul>
				{:else if items[m.id]}
					{#if items[m.id].length === 0}
						<p class="sub mt-4 border-t border-sunk pt-4">
							There was nothing in that library to list.
						</p>
					{:else}
						<ul class="mt-3 divide-y divide-sunk border-t border-sunk">
							{#each items[m.id] as it (it.remote_id)}
								<li class="flex flex-wrap items-center gap-x-3 gap-y-1 py-2.5">
									<span class="min-w-0 flex-1 truncate text-sm">
										{it.title || it.remote_id}
									</span>
									{#if it.reason}
										<span class="sub flex-none">{it.reason}</span>
									{/if}
									{#if it.asset_id}
										<a href="/app/videos/{it.asset_id}/" class="link flex-none text-xs">Open</a>
									{:else if it.state === 'pending'}
										<span class="sub flex-none">Not brought across yet</span>
									{:else}
										<span class="chip flex-none">{it.state}</span>
									{/if}
								</li>
							{/each}
						</ul>
					{/if}
				{/if}
			</section>
		{/each}
	</div>
{/if}
