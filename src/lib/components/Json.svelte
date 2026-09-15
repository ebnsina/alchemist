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
		<span class="text-xs text-dim">{label || 'JSON'}</span>
		<button type="button" onclick={copy} class="flex items-center gap-1.5 text-xs text-dim transition-colors hover:text-ink">
			<HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} size={13} strokeWidth={2} />
			{copied ? 'Copied' : 'Copy'}
		</button>
	</div>
	<pre class="code__body"><code>{#each tokens as t, i (i)}<span class="tok tok--{t.kind}">{t.text}</span>{/each}</code></pre>
</div>

<style>
	.code {
		border-radius: var(--radius-md);
		background: var(--color-neutral);
		border: 1px solid var(--color-outline);
		overflow: hidden;
	}
	.code__bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 12px;
		border-bottom: 1px solid var(--color-outline);
	}
	.code__body {
		margin: 0;
		padding: 12px 16px;
		overflow-x: auto;
		font-family: var(--font-mono);
		font-size: 0.8125rem;
		line-height: 1.7;
	}
	/* Syntax colour inside the system: ink for keys, teal for values, slate for the
	   punctuation between them. No new hues — weight and the one accent do the work.
	   Measured on the paper ground: key 16.95:1, string 7.02:1, slate 4.93:1. */
	.tok--key {
		color: var(--color-primary);
		font-weight: 500;
	}
	.tok--string {
		color: var(--color-tertiary);
	}
	.tok--number,
	.tok--literal {
		color: var(--color-tertiary-container);
	}
	.tok--plain {
		color: var(--color-secondary);
	}
</style>
