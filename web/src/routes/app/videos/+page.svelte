<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Upload01Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import AssetList from '$lib/components/AssetList.svelte';
	import { listAssets, ApiError, type Asset } from '$lib/api';

	let assets = $state<Asset[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Anything still moving is worth another look without the customer asking.
	const busy = $derived(
		assets.some((a) => !['ready', 'failed', 'partially_ready'].includes(a.state))
	);

	async function load() {
		error = '';
		try {
			assets = (await listAssets()).assets;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	$effect(() => {
		if (!busy) return;
		const id = setInterval(load, 5000);
		return () => clearInterval(id);
	});
</script>

<Seo title="Videos — Alchemist" description="Every video you have sent us." />

<header class="flex flex-wrap items-end justify-between gap-4">
	<div class="min-w-0">
		<h1 class="text-2xl font-semibold tracking-tight">Videos</h1>
		<p class="sub mt-1.5 max-w-xl">
			Everything you have sent us, newest first. Open one to watch it, get its playback
			links, or see what is still being made.
		</p>
	</div>
	<a href="/app/upload/" class="btn-solid btn-sm flex-none">
		<HugeiconsIcon icon={Upload01Icon} size={14} strokeWidth={2} />
		Upload
	</a>
</header>

{#if error}
	<p class="mt-6 text-sm text-red" role="alert">{error}</p>
{/if}

{#if !loading && assets.length}
	<p class="label mt-6">{assets.length} in all</p>
{/if}

<div class="card mt-4">
	<AssetList {assets} {loading} />
</div>
