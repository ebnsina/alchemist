<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import { listProfiles, setProfile, ApiError, type LadderProfile } from '$lib/api';

	let profiles = $state<LadderProfile[]>([]);
	let switching = $state('');
	let loading = $state(true);
	let error = $state('');

	async function load() {
		error = '';
		try {
			profiles = (await listProfiles()).profiles;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	async function choose(name: string) {
		error = '';
		switching = name;
		try {
			await setProfile(name);
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		} finally {
			switching = '';
		}
	}

	const kbps = (n: number) =>
		new Intl.NumberFormat('en', { maximumFractionDigits: 0 }).format(n / 1000) + 'k';
</script>

<Seo title="Encoding — Alchemist" description="Which sizes we make from every video you send." />

<h1 class="text-2xl font-semibold tracking-tight">Encoding</h1>
<p class="sub mt-1 max-w-xl">
	Which sizes we make, and how hard we squeeze them. Changing this affects the next video you
	send — everything already encoded stays exactly as it is, because re-making a whole library on
	a settings change would be a surprise that costs real money.
</p>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="mt-6 grid gap-3">
		{#each [0, 1, 2] as i (i)}<div class="sk h-28"></div>{/each}
	</div>
{:else}
	<div class="mt-6 grid gap-3">
		{#each profiles as p (p.name)}
			<div class="preset" class:preset--on={p.current}>
				<div class="flex flex-wrap items-start justify-between gap-3">
					<div class="min-w-0">
						<p class="text-sm font-semibold">{p.description}</p>
						<p class="mono mt-0.5">{p.name}</p>
					</div>
					{#if p.current}
						<span class="chip chip-on flex-none">In use</span>
					{:else}
						<button
							type="button"
							class="btn btn-sm flex-none"
							onclick={() => choose(p.name)}
							disabled={switching === p.name}
						>
							{switching === p.name ? 'Switching…' : 'Use this'}
						</button>
					{/if}
				</div>

				<ul class="mt-3 flex flex-wrap gap-1.5">
					{#each p.rungs as r (r.height)}
						<li class="chip" title="{r.codec} at about {kbps(r.maxrate_bps)} a second">
							{r.height}p{r.lazy ? ' ·' : ''}
						</li>
					{/each}
				</ul>
				<p class="sub mt-2.5">
					{p.rungs.length} sizes, up to {Math.max(...p.rungs.map((r) => r.height))}p.
					{#if p.rungs.some((r) => r.lazy)}
						The ones marked · are only made when somebody first asks for them, so you are not
						billed for sizes nobody watches.
					{/if}
				</p>
			</div>
		{/each}
	</div>
{/if}

<style>
	.preset {
		padding: 16px;
		border-radius: var(--radius-md);
		border: 1px solid var(--color-sunk);
	}
	.preset--on {
		border-color: var(--color-brand);
	}
</style>
