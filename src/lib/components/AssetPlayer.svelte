<script lang="ts">
	import { untrack } from 'svelte';
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';
	import type { Playback } from '$lib/api';
	import type { AlchemistPlayer as Player } from '@alchemist/player';

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
	// Every poll re-signs the URL with a fresh exp, so only the path says "different
	// video". Tracking the whole URL tore the player down every few seconds, which is
	// what stopped a live broadcast mid-play and left the play button showing.
	const srcKey = $derived(src.split('?')[0]);

	let host: HTMLDivElement | null = $state(null);
	let failed = $state('');

	$effect(() => {
		srcKey;
		const node = host;
		if (!node || !srcKey) return;
		let player: Player | null = null;
		let cancelled = false;

		(async () => {
			try {
				const { AlchemistPlayer } = await import('@alchemist/player');
				if (cancelled) return;
				player = new AlchemistPlayer(node, {
					src: untrack(() => src),
					poster: untrack(() => poster),
					// Read untracked: a broadcast ending must not tear the player down mid-play.
					live: untrack(() => live),
					lang: 'en',
					// Telemetry belongs to real viewers, not to an operator checking a file.
					beacon: false
				});
				// The signature outlives a short clip but not a long broadcast; hand the
				// player the latest one rather than letting it die on the next segment.
				player.addEventListener('needs-refresh', () => void player?.refresh(src));
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
