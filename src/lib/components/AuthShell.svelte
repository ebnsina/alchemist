<script lang="ts">
	import Logo from './Logo.svelte';

	let { title, children } = $props();
</script>

<section class="mx-auto flex min-h-svh max-w-[1200px] items-center px-4 py-16 sm:px-6">
	<div class="mx-auto w-full max-w-[420px]">
		<a href="/" class="mb-6 flex items-center justify-center gap-2 text-lg font-semibold">
			<Logo size={24} />
			Alchemist
		</a>
		<h1 class="mb-6 text-center">{title}</h1>
		<div class="card">
			{@render children()}
		</div>
	</div>
</section>
