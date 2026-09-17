<script lang="ts">
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';

	// Studio only. Previews use AssetPlayer, which is the real player customers embed.
	// This one stays because Studio drives the raw element from its own timeline and
	// draws a crop box on the frame, which a player that owns its chrome cannot give.
	let {
		hls,
		poster,
		currentTime = $bindable(0),
		playable = $bindable(true),
		element = $bindable(null),
		playing = $bindable(false),
		overlay,
		footer = true,
		fit = false
	}: {
		hls: string;
		poster?: string;
		currentTime?: number;
		playable?: boolean;
		element?: HTMLVideoElement | null;
		playing?: boolean;
		fit?: boolean;
		overlay?: import('svelte').Snippet;
		footer?: boolean;
	} = $props();

	let video: HTMLVideoElement | null = $state(null);
	let pausedInner = $state(true);
	$effect(() => {
		playing = !pausedInner;
	});
	let note = $state('The real file, played from the signed link');
	let blocked = $state(false);
	let level = $state('');

	const src = $derived(PUBLIC_ALCHEMIST_API + hls);
	$effect(() => {
		playable = !blocked;
	});
	// Studio drives the video from its own timeline, so it needs the element and to
	// know whether it is running.
	$effect(() => {
		element = video;
	});
	const posterSrc = $derived(poster ? PUBLIC_ALCHEMIST_API + poster : undefined);

	$effect(() => {
		const el = video;
		if (!el || !src || blocked) return;
		let cancelled = false;
		let instance: import('hls.js').default | null = null;

		(async () => {
			// hls.js cannot read DASH, and an encrypted asset is packaged cenc, which
			// only DASH carries. The key itself is served -- what is missing here is a
			// DASH player, which the embed has and this preview does not.
			if (src.includes('.mpd')) {
				blocked = true;
				note = 'Encrypted videos play in the embed, not in this preview. Use the playback link.';
				return;
			}

			// Only "probably" counts as native HLS. Chrome answers "maybe" for the HLS
			// media type and then cannot play it, so trusting anything weaker means
			// handing the stream to a player that will never start.
			const native = el.canPlayType('application/vnd.apple.mpegurl') === 'probably';

			// Media Source Extensions cannot decrypt SAMPLE-AES without a license
			// server. Read the manifest and say so, rather than attaching a player that
			// silently never starts.
			let sampleAes = false;
			try {
				const text = await fetch(src).then((r) => r.text());
				const variant = text.split('\n').find((l) => l.trim() && !l.startsWith('#'));
				if (variant) {
					const child = await fetch(new URL(variant.trim(), src).toString()).then((r) => r.text());
					sampleAes = /METHOD=SAMPLE-AES/.test(child);
				}
			} catch {
				// If the manifest cannot be read, let the player try and report itself.
			}
			if (cancelled) return;

			if (native) {
				el.src = src;
				return;
			}
			if (sampleAes) {
				blocked = true;
				note =
					'This video is encrypted. Safari plays it as it is; Chrome and Firefox need a license server, which is not wired up yet.';
				return;
			}

			const { default: Hls } = await import('hls.js');
			if (cancelled) return;
			if (!Hls.isSupported()) {
				blocked = true;
				note = 'This browser cannot play the stream. Try Safari, Chrome or Firefox.';
				return;
			}
			instance = new Hls({ enableWorker: true });
			instance.on(Hls.Events.LEVEL_SWITCHED, (_e, data) => {
				const l = instance?.levels[data.level];
				level = l ? `${l.height}p` : '';
			});
			instance.on(Hls.Events.ERROR, (_e, data) => {
				// Only fatal errors are worth showing. hls.js recovers from the rest by
				// itself, and surfacing those would make a working player look broken.
				if (data.fatal) {
					blocked = true;
					note = 'The stream stopped. The link may have expired — reload the page.';
				}
			});
			instance.loadSource(src);
			instance.attachMedia(el);
		})();

		return () => {
			cancelled = true;
			instance?.destroy();
		};
	});
</script>

<div class="card overflow-hidden {fit ? 'w-full border-0 bg-transparent p-0' : ''}">
	{#if blocked}
		<!-- The poster is not encrypted, so the frame still proves the file is real. -->
		<div class="relative aspect-video w-full bg-black">
			{#if posterSrc}
				<img src={posterSrc} alt="" class="h-full w-full object-contain opacity-70" />
			{/if}
		</div>
	{:else}
		<div class="relative">
			<!-- svelte-ignore a11y_media_has_caption -->
			<video
				bind:this={video}
				bind:currentTime
				bind:paused={pausedInner}
				class="block aspect-video w-full bg-black object-contain"
				controls
				playsinline
				preload="metadata"
				poster={posterSrc}
			></video>
			{#if overlay}
				<!-- Pointer events off: the overlay is a guide, and swallowing clicks
				     would take the play button away. -->
				<div class="pointer-events-none absolute inset-0">{@render overlay()}</div>
			{/if}
		</div>
	{/if}
	{#if footer}
		<div class="flex flex-wrap items-center justify-between gap-3 px-4 py-2.5 text-xs text-dim">
			<span>{note}</span>
			{#if !blocked}
				<span>{level ? `Playing ${level} · ` : ''}link expires on its own</span>
			{/if}
		</div>
	{/if}
</div>
