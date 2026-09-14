<script lang="ts">
	import { parts } from '$lib/money';

	// Both languages render, as with T; CSS hides one by html[lang].
	let {
		amount,
		decimals = 2,
		currency = 'BDT'
	}: { amount: number; decimals?: number; currency?: string } = $props();

	const en = $derived(parts(amount, 'en', decimals, currency));
	const bn = $derived(parts(amount, 'bn', decimals, currency));
</script>

<!--
	The symbol is wrapped separately because Intl returns one string but the two
	halves come from different fonts: U+09F3 is in the Bengali block, which the body
	face does not cover, so it falls back to Noto Sans Bengali and lands visually
	undersized beside the numerals. formatToParts lets the symbol be corrected
	without hand-splitting the string, which would break for any locale that puts it
	after the number — as bn-BD does.
-->
<span class="l-en"
	>{#each en as p (p.type + p.value)}{#if p.type === 'currency'}<span class="sym">{p.value}</span
			>{:else}{p.value}{/if}{/each}</span
><span class="l-bn"
	>{#each bn as p (p.type + p.value)}{#if p.type === 'currency'}<span class="sym">{p.value}</span
			>{:else}{p.value}{/if}{/each}</span
>

<style>
	.sym {
		font-size: 0.88em;
		/* Noto's taka sits low against Latin numerals; nudge it onto the same optical
		   baseline rather than leaving it hanging. */
		line-height: 1;
		vertical-align: 0.02em;
		padding-inline-end: 0.04em;
	}
</style>
