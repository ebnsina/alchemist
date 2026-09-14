<script lang="ts">
	import { onMount } from 'svelte';

	// The configuration as given. Faithful to the linked ShaderGradient preset.
	const CFG = {
		color1: '#ff5005',
		color2: '#dbba95',
		color3: '#d0bce1',
		uAmplitude: 1,
		uDensity: 1.3,
		uFrequency: 5.5,
		uSpeed: 0.4,
		uStrength: 4,
		cDistance: 3.6,
		cPolarAngle: 90,
		cAzimuthAngle: 180,
		fov: 45,
		positionX: -1.4,
		positionY: 0,
		positionZ: 0,
		rotationX: 0,
		rotationY: 10,
		rotationZ: 50,
		brightness: 1.2,
		reflection: 0.1,
		grain: true
	};

	let host = $state<HTMLElement | null>(null);
	let canvas = $state<HTMLCanvasElement | null>(null);
	let live = $state(false);

	onMount(() => {
		const still = window.matchMedia('(prefers-reduced-motion: reduce)');
		// Never on reduced motion, never on Data Saver: the CSS approximation below is
		// already painted and is a perfectly good background.
		const conn = (navigator as { connection?: { saveData?: boolean } }).connection;
		if (still.matches || conn?.saveData) return;

		let gradient: { start(): void; stop(): void; destroy(): void } | null = null;
		let cancelled = false;
		let onVis: (() => void) | null = null;
		let io: IntersectionObserver | null = null;

		// Lazy: the module is fetched after first paint, never before it.
		const idle = 'requestIdleCallback' in window ? window.requestIdleCallback : setTimeout;
		const handle = idle(async () => {
			if (cancelled || !canvas) return;
			const { createShaderGradient } = await import('$lib/utils/shader-gradient.js');
			if (cancelled || !canvas) return;
			gradient = createShaderGradient(canvas, CFG);
			if (!gradient) return; // no WebGL: the CSS version stays
			live = true;

			// Stop when the hero is off screen, and when the tab is not being looked at.
			let onScreen = true;
			let visible = !document.hidden;
			const sync = () => (onScreen && visible ? gradient?.start() : gradient?.stop());
			io = new IntersectionObserver(
				([e]) => {
					onScreen = e.isIntersecting;
					sync();
				},
				{ threshold: 0 }
			);
			if (host) io.observe(host);
			onVis = () => {
				visible = !document.hidden;
				sync();
			};
			document.addEventListener('visibilitychange', onVis);
			sync();
		});

		return () => {
			cancelled = true;
			if ('cancelIdleCallback' in window) window.cancelIdleCallback(handle as number);
			io?.disconnect();
			if (onVis) document.removeEventListener('visibilitychange', onVis);
			gradient?.destroy();
		};
	});
</script>

<div bind:this={host} class="pointer-events-none absolute inset-0 -z-20 overflow-hidden" aria-hidden="true">
	<!-- Painted immediately, from the same three colours. The canvas fades in over it
	     once WebGL is ready, and this is what stays on reduced motion, Data Saver and
	     anywhere WebGL is unavailable. -->
	<div class="static-gradient absolute inset-0"></div>
	<canvas
		bind:this={canvas}
		class="absolute inset-0 h-full w-full transition-opacity duration-700 {live
			? 'opacity-100'
			: 'opacity-0'}"
	></canvas>
	<!-- Scrim: the preset is warm and light, the page is deep sea green. This lets it
	     read as light behind the content instead of competing with it. -->
	<div class="scrim absolute inset-0"></div>
</div>

<style>
	.static-gradient {
		background:
			radial-gradient(60% 70% at 18% 28%, #ff5005 0%, transparent 62%),
			radial-gradient(55% 60% at 62% 18%, #dbba95 0%, transparent 60%),
			radial-gradient(70% 75% at 78% 72%, #d0bce1 0%, transparent 64%),
			#2a1408;
	}
	/* The preset is warm and light; the page is deep sea green. The scrim is what
	   lets it read as light behind the content rather than competing with it, and
	   it is sized by measurement: the headline bottomed out at 3.70:1 without it. */
	.scrim {
		background:
			radial-gradient(72% 46% at 50% 40%, rgba(6, 16, 13, 0.74) 0%, rgba(6, 16, 13, 0.5) 55%, transparent 78%),
			radial-gradient(120% 95% at 50% 0%, rgba(6, 16, 13, 0.5) 0%, rgba(6, 16, 13, 0.86) 60%, #06100d 100%),
			linear-gradient(to bottom, rgba(6, 16, 13, 0.35) 0%, transparent 30%, transparent 52%, #06100d 100%);
	}
</style>
