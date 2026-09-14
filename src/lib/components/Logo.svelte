<script lang="ts">
	// The mark: a flask whose contents are a play triangle. Swap this one file to
	// change the logo everywhere.
	let { size = 28 }: { size?: number } = $props();
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 32 32"
	fill="none"
	aria-hidden="true"
	focusable="false"
>
	<defs>
		<linearGradient id="alc-flask" x1="4" y1="2" x2="28" y2="30" gradientUnits="userSpaceOnUse">
			<stop stop-color="#34d399" />
			<stop offset="1" stop-color="#fbbf24" />
		</linearGradient>
	</defs>
	<path
		d="M12.4 3h7.2M13.3 3v7.7L5.9 23.3A3.1 3.1 0 0 0 8.6 28h14.8a3.1 3.1 0 0 0 2.7-4.7l-7.4-12.6V3"
		stroke="url(#alc-flask)"
		stroke-width="2"
		stroke-linecap="round"
		stroke-linejoin="round"
	/>
	<path
		d="M9.1 19.4h13.8l2.9 5a2.7 2.7 0 0 1-2.4 4.1H8.6a2.7 2.7 0 0 1-2.4-4.1z"
		fill="url(#alc-flask)"
		opacity=".22"
	/>
	<path d="M13.9 21.6v5.2l4.9-2.6z" fill="#34d399" />
</svg>
