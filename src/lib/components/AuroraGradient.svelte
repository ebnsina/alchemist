<!--
  Hero background: a linear aurora in five greens with grain.

  Pure CSS. The previous WebGL shader is gone with it — a linear blend needs no
  vertex displacement, so a canvas, a render loop and 9 KB of GLSL bought nothing
  the browser cannot paint for free. It also means nothing to lazy-load, nothing to
  pause on scroll, and no fallback path to keep correct.

  Grain is one inline SVG turbulence, which stays sharp at any size and costs no
  extra request.
-->
<div class="aurora" aria-hidden="true">
	<div class="aurora__scrim"></div>
	<div class="aurora__grain"></div>
</div>

<style>
	.aurora {
		position: absolute;
		inset: 0;
		z-index: -10;
		pointer-events: none;
		background: linear-gradient(
			160deg,
			#1b4332 0%,
			#2d6a4f 26%,
			#40916c 50%,
			#52b788 74%,
			#74c69d 100%
		);
		/* Fades into the onyx page rather than ending on a hard edge, so the hero has
		   no seam at its base the way the header no longer has one at its top. */
		-webkit-mask-image: linear-gradient(to bottom, #000 0%, #000 45%, transparent 100%);
		mask-image: linear-gradient(to bottom, #000 0%, #000 45%, transparent 100%);
	}

	/* The gradient's lighter stops reach rgb(64,142,106), where even pure white
	   measures 3.96:1 — so no text colour can pass AA against it and the surface
	   itself has to come down. The scrim darkens toward onyx while keeping the hue,
	   and is weakest at the very top where no text sits. */
	.aurora__scrim {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			to bottom,
			color-mix(in srgb, var(--color-body) 46%, transparent) 0%,
			color-mix(in srgb, var(--color-body) 62%, transparent) 38%,
			color-mix(in srgb, var(--color-body) 68%, transparent) 100%
		);
	}

	.aurora__grain {
		position: absolute;
		inset: 0;
		opacity: 0.25;
		mix-blend-mode: overlay;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3' stitchTiles='stitch'/%3E%3CfeColorMatrix type='saturate' values='0'/%3E%3C/filter%3E%3Crect width='160' height='160' filter='url(%23n)' opacity='0.78'/%3E%3C/svg%3E");
		background-repeat: repeat;
		background-size: 160px 160px;
	}
</style>
