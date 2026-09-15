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
		background: var(--color-sunk);
		border: 1px solid var(--color-muted);
		overflow: hidden;
	}
	.code__bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 12px;
		border-bottom: 1px solid var(--color-muted);
	}
	.code__body {
		margin: 0;
		padding: 12px 16px;
		overflow-x: auto;
		font-family: var(--font-mono);
		font-size: 0.8125rem;
		line-height: 1.7;
	}
	/* Syntax colour inside the system, and no new hues: keys carry weight, values
	   carry the one accent, and the punctuation between them recedes. Every one of
	   these four is a defined token — the previous set named colours that did not
	   exist anywhere, so the highlighting rendered as flat inherited text. */
	.tok--key {
		color: var(--color-ink);
		font-weight: 600;
	}
	.tok--string {
		color: var(--color-accent);
	}
	.tok--number,
	.tok--literal {
		color: var(--color-ink);
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}
	.tok--plain {
		color: var(--color-faint);
	}
</style>
