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
			var(--aurora-1) 0%,
			var(--aurora-2) 26%,
			var(--aurora-3) 50%,
			var(--aurora-4) 74%,
			var(--aurora-5) 100%
		);
		/* Fades into the onyx page rather than ending on a hard edge, so the hero has
		   no seam at its base the way the header no longer has one at its top. */
		-webkit-mask-image: linear-gradient(to bottom, #000 0%, #000 45%, transparent 100%);
		mask-image: linear-gradient(to bottom, #000 0%, #000 45%, transparent 100%);
	}

	/* On paper the danger runs the other way: the gradient's darkest stop is
	   #a7d9bf, where ink #14171A still measures 11.4:1, so the scrim is not
	   protecting text — it is keeping the wash from shouting. It thins toward the
	   top, where the header glass already sits. */
	.aurora__scrim {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			to bottom,
			color-mix(in srgb, var(--color-body) var(--aurora-scrim-top), transparent) 0%,
			color-mix(in srgb, var(--color-body) var(--aurora-scrim-mid), transparent) 38%,
			color-mix(in srgb, var(--color-body) var(--aurora-scrim-bottom), transparent) 100%
		);
	}

	.aurora__grain {
		position: absolute;
		inset: 0;
		opacity: var(--aurora-grain-opacity);
		mix-blend-mode: var(--aurora-grain-blend);
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3' stitchTiles='stitch'/%3E%3CfeColorMatrix type='saturate' values='0'/%3E%3C/filter%3E%3Crect width='160' height='160' filter='url(%23n)' opacity='0.78'/%3E%3C/svg%3E");
		background-repeat: repeat;
		background-size: 160px 160px;
	}
</style>
