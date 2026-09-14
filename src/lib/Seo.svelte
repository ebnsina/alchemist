<script lang="ts">
	import { page } from '$app/state';
	import { SITE } from './site';

	// English is the canonical language of the site, so it is what goes in <head>.
	// Page content is bilingual; metadata is not, because a document has one title.
	let { title, description }: { title: string; description: string } = $props();
	const url = $derived(SITE.origin + page.url.pathname);
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="description" content={description} />
	<link rel="canonical" href={url} />
	<link rel="alternate" hreflang="en" href={url} />
	<link rel="alternate" hreflang="bn" href={url} />
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content="Alchemist" />
	<meta property="og:title" content={title} />
	<meta property="og:description" content={description} />
	<meta property="og:url" content={url} />
	<meta property="og:locale" content="en_US" />
	<meta property="og:locale:alternate" content="bn_BD" />
	<meta property="og:image" content="{SITE.origin}/og.png" />
	<meta property="og:image:width" content="1200" />
	<meta property="og:image:height" content="630" />
	<meta property="og:image:alt" content="Alchemist — video hosting inside BDIX" />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={title} />
	<meta name="twitter:description" content={description} />
	<meta name="twitter:image" content="{SITE.origin}/og.png" />
</svelte:head>
