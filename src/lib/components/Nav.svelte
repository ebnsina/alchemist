<script lang="ts">
	import { page } from '$app/state';
	import Logo from './Logo.svelte';
	import T from '$lib/T.svelte';

	let open = $state(false);
	const links = [
		{ href: '/#features', en: 'Features', bn: 'যা যা আছে' },
		{ href: '/#how', en: 'How it works', bn: 'কীভাবে কাজ করে' },
		{ href: '/#pricing', en: 'Pricing', bn: 'দাম' },
		{ href: '/#faq', en: 'Questions', bn: 'প্রশ্ন' }
	];
	const lang = $derived(page.url.pathname);
</script>

<header
	class="fixed inset-x-0 top-0 z-50 border-b border-hairline bg-body/70 backdrop-blur-xl backdrop-saturate-150"
>
	<div class="mx-auto flex max-w-6xl flex-wrap items-center gap-3 px-4 py-3 sm:px-6">
		<a href="/" class="flex flex-none items-center gap-2 text-lg font-extrabold tracking-tight">
			<Logo size={26} />
			Alchemist
		</a>

		<nav aria-label="Sections" class="order-3 w-full sm:order-none sm:mx-auto sm:w-auto">
			<ul class="flex flex-wrap items-center gap-x-6 gap-y-1 border-t border-hairline pt-2 sm:border-0 sm:pt-0">
				{#each links as l (l.href)}
					<li>
						<a
							href={l.href}
							class="block py-1 text-sm text-muted transition-colors hover:text-ink"
							onclick={() => (open = false)}><T en={l.en} bn={l.bn} /></a
						>
					</li>
				{/each}
			</ul>
		</nav>

		<div class="ml-auto flex flex-none items-center gap-2">
			<div
				class="flex items-center gap-0.5 rounded-lg border border-hairline bg-card p-0.5"
				role="group"
				aria-label="Language / ভাষা"
			>
				<button
					type="button"
					data-set-lang="en"
					aria-pressed="true"
					lang="en"
					class="rounded-md px-2 py-1 text-xs font-medium text-muted transition-colors hover:text-ink aria-pressed:bg-body aria-pressed:text-ink"
					>EN</button
				>
				<button
					type="button"
					data-set-lang="bn"
					aria-pressed="false"
					lang="bn"
					class="rounded-md px-2 py-1 text-xs font-medium text-muted transition-colors hover:text-ink aria-pressed:bg-body aria-pressed:text-ink"
					>বাংলা</button
				>
			</div>
			<a href="/#pricing" class="btn-primary hidden text-sm sm:inline-flex">
				<T en="Start free" bn="ফ্রি শুরু করুন" />
			</a>
		</div>
	</div>
</header>
