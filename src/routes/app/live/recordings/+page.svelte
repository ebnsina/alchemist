<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import AssetList from '$lib/components/AssetList.svelte';
	import { listAssets, ApiError, type Asset } from '$lib/api';

	let assets = $state<Asset[]>([]);
	let loading = $state(true);
	let error = $state('');

	// A finished broadcast is an ordinary asset in live_ended, so this page is a
	// filter over the same list and each row opens the video page everything else uses.
	async function load() {
		try {
			assets = (await listAssets()).assets.filter((a) => a.state === 'live_ended');
			error = '';
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});
</script>

<Seo title="Recordings — Alchemist" description="What your live broadcasts left behind." />

<h1 class="text-2xl font-semibold tracking-tight">Recordings</h1>
<p class="sub mt-1.5 max-w-xl">
	Every broadcast is kept as an ordinary video. Open one to watch it, share the link, or edit it.
</p>

{#if error}
	<p class="mt-6 text-sm text-red" role="alert">{error}</p>
{:else if !loading && assets.length === 0}
	<div class="card mt-6 py-8 text-center">
		<p class="title">Nothing recorded yet</p>
		<p class="sub mx-auto mt-2 max-w-sm">
			A recording appears here when a broadcast ends. Start one under
			<a href="/app/live/" class="link">Streams</a>.
		</p>
	</div>
{:else}
	<div class="card mt-6">
		<AssetList {assets} {loading} />
	</div>
{/if}
