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
	formatToParts rather than format, so the symbol stays addressable. The body face
	now covers U+09F3 itself, so no size correction is needed — but splitting the
	string by hand would still be wrong, because English puts the sign before the
	number and bn-BD after it.
-->
<span class="l-en"
	>{#each en as p (p.type + p.value)}{#if p.type === 'currency'}{p.value}{:else}{p.value}{/if}{/each}</span
><span class="l-bn"
	>{#each bn as p (p.type + p.value)}{#if p.type === 'currency'}{p.value}{:else}{p.value}{/if}{/each}</span
>
