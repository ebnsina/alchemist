<script lang="ts">
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';

	let { hls, poster }: { hls: string; poster?: string } = $props();

	let video: HTMLVideoElement | null = $state(null);
	let note = $state('The real file, played from the signed link');
	let blocked = $state(false);
	let level = $state('');

	const src = $derived(PUBLIC_ALCHEMIST_API + hls);
	const posterSrc = $derived(poster ? PUBLIC_ALCHEMIST_API + poster : undefined);

	$effect(() => {
		const el = video;
		if (!el || !src || blocked) return;
		let cancelled = false;
		let instance: import('hls.js').default | null = null;

		(async () => {
			// Only "probably" counts as native HLS. Chrome answers "maybe" for the HLS
			// media type and then cannot play it, so trusting anything weaker means
			// handing the stream to a player that will never start.
			const native = el.canPlayType('application/vnd.apple.mpegurl') === 'probably';

			// Media Source Extensions cannot decrypt SAMPLE-AES without a licence
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
					'This video is encrypted. Safari plays it as it is; Chrome and Firefox need a licence server, which is not wired up yet.';
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

<div class="card overflow-hidden">
	{#if blocked}
		<!-- The poster is not encrypted, so the frame still proves the file is real. -->
		<div class="relative aspect-video w-full bg-black">
			{#if posterSrc}
				<img src={posterSrc} alt="" class="h-full w-full object-contain opacity-70" />
			{/if}
		</div>
	{:else}
		<!-- svelte-ignore a11y_media_has_caption -->
		<video
			bind:this={video}
			class="aspect-video w-full bg-black"
			controls
			playsinline
			preload="metadata"
			poster={posterSrc}
		></video>
	{/if}
	<div class="flex flex-wrap items-center justify-between gap-3 px-4 py-2.5 text-xs text-muted">
		<span>{note}</span>
		{#if !blocked}
			<span>{level ? `Playing ${level} · ` : ''}link expires in four hours</span>
		{/if}
	</div>
</div>
