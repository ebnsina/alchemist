<script lang="ts">
	import { untrack } from 'svelte';
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';
	import type { Playback } from '$lib/api';

	// The real player, the same bundle a customer embeds. Using it here means an
	// encrypted asset previews properly -- it carries the DASH build and the Clear Key
	// licence path, which the dashboard's own hls.js never could.
	// live is the asset state, not shaka's guess: the live playlist is EXT-X-PLAYLIST-TYPE:EVENT
	// so a viewer can seek back to the start, and shaka reads EVENT as a growing VOD.
	let {
		playback,
		live = false,
		class: cls = ''
	}: { playback: Playback; live?: boolean; class?: string } = $props();

	// preferred names the URL that will actually play: HLS for a clear asset, DASH for
	// an encrypted one, because cenc cannot ride on HLS.
	const src = $derived(
		PUBLIC_ALCHEMIST_API + (playback.preferred === 'dash' ? playback.dash : playback.hls)
	);
	const poster = $derived(playback.poster ? PUBLIC_ALCHEMIST_API + playback.poster : undefined);

	let host: HTMLDivElement | null = $state(null);
	let failed = $state('');

	$effect(() => {
		const node = host;
		if (!node || !src) return;
		let player: { destroy(): Promise<void> } | null = null;
		let cancelled = false;

		(async () => {
			try {
				const { AlchemistPlayer } = await import('@alchemist/player');
				if (cancelled) return;
				player = new AlchemistPlayer(node, {
					src,
					poster,
					// Read untracked: a broadcast ending must not tear the player down mid-play.
					live: untrack(() => live),
					lang: 'en',
					// Telemetry belongs to real viewers, not to an operator checking a file.
					beacon: false
				});
			} catch {
				failed = "We couldn't load the player. Open the playback link directly.";
			}
		})();

		return () => {
			cancelled = true;
			void player?.destroy();
			node.innerHTML = '';
		};
	});
</script>

{#if failed}
	<p class="sub">{failed}</p>
{:else}
	<div bind:this={host} class={cls}></div>
{/if}
