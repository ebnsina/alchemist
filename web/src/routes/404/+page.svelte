<script lang="ts">
</script>

<svelte:head>
	<title>Page not found — Alchemist</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<main id="main" class="shell bay">
	<div class="flow">
		<p class="marker">404</p>
		<h1>That page is not here.</h1>
		<p class="lede quiet">
			Either it moved or it never existed. Neither is your fault. Everything this site has is one click away.
		</p>
		<ul class="ways">
			<li><a class="tlink" href="/">Home</a></li>
			<li><a class="tlink" href="/contact/">Talk to us</a></li>
			<li><a class="tlink" href="/app/">Dashboard</a></li>
			</ul>
	</div>
</main>

<style>
	.marker {
		font-family: var(--mono);
		color: var(--brass);
		max-width: 5rem;
	}
	h1 {
		max-width: 13ch;
	}
	.ways {
		list-style: none;
		padding: 0;
		margin: 2rem 0 0;
		display: flex;
		flex-wrap: wrap;
		gap: 0.6rem 1.6rem;
	}
</style>
