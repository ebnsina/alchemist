<script lang="ts">
	// Both languages render into the HTML; CSS hides one by html[lang]. No flash,
	// no JS needed, and display:none keeps the hidden copy out of screen readers.
	let { en, bn, as = 'span' }: { en: string; bn: string; as?: string } = $props();
</script>

<svelte:element this={as} class="l-en">{en}</svelte:element><svelte:element this={as} class="l-bn"
	>{bn}</svelte:element
>
