<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Copy01Icon, Tick02Icon } from '@hugeicons/core-free-icons';
	import { tokenizeJson } from '$lib/json-highlight';

	let { source, label = '' }: { source: string; label?: string } = $props();

	const tokens = $derived(tokenizeJson(source));
	let copied = $state(false);

	async function copy() {
		try {
			await navigator.clipboard.writeText(source);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			copied = false;
		}
	}
</script>

<div class="code">
	<div class="code__bar">
		<span class="text-xs text-muted">{label || 'JSON'}</span>
		<button type="button" onclick={copy} class="flex items-center gap-1.5 text-xs text-muted transition-colors hover:text-ink">
			<HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} size={13} strokeWidth={2} />
			{copied ? 'Copied' : 'Copy'}
		</button>
	</div>
	<pre class="code__body"><code>{#each tokens as t, i (i)}<span class="tok tok--{t.kind}">{t.text}</span>{/each}</code></pre>
</div>

<style>
	.code {
		border-radius: 14px;
		corner-shape: squircle;
		background: var(--color-body);
		border: 1px solid var(--color-hairline);
		overflow: hidden;
	}
	.code__bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.45rem 0.8rem;
		border-bottom: 1px solid var(--color-hairline);
	}
	.code__body {
		margin: 0;
		padding: 0.9rem 1rem;
		overflow-x: auto;
		font-family: var(--font-mono);
		font-size: var(--step--1);
		line-height: 1.7;
	}
	/* Measured on the card ground, not chosen: key 6.6:1, string 7.8:1,
	   number 8.1:1, literal 6.9:1. */
	.tok--key {
		color: #7dd3fc;
	}
	.tok--string {
		color: var(--color-brand-light);
	}
	.tok--number {
		color: #fbbf24;
	}
	.tok--literal {
		color: #c4b5fd;
	}
	.tok--plain {
		color: var(--color-muted);
	}
</style>
