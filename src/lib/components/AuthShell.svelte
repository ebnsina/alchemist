<script lang="ts">
	import AuroraGradient from './AuroraGradient.svelte';
	import Logo from './Logo.svelte';

	let { title, children } = $props();
</script>

<section class="relative flex min-h-svh items-center overflow-hidden px-4 py-20 sm:px-6">
	<AuroraGradient />

	<div class="relative z-10 mx-auto w-full max-w-md">
		<a href="/" class="mb-8 flex items-center justify-center gap-2 text-lg font-semibold">
			<Logo />
			Alchemist
		</a>
		<h1 class="mb-7 text-center text-3xl leading-[1.32] font-semibold tracking-tight">
			{title}
		</h1>
		<div class="card p-6 sm:p-8">
			{@render children()}
		</div>
	</div>
</section>
