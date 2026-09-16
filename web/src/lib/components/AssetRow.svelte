<script lang="ts">
	import { assetName } from '$lib/assets';
	import type { Asset } from '$lib/api';

	let { asset }: { asset: Asset } = $props();
</script>

<a href="/app/videos/{asset.id}/" class="block min-w-0 hover:opacity-80">
	<span class="block truncate font-medium">{assetName(asset)}</span>
	<!-- The id stays visible when a title exists: it is what the API, the logs and the
	     playback URL all call this video. -->
	{#if asset.title}<span class="mono mt-0.5 block">{asset.id.slice(0, 8)}</span>{/if}
</a>
