<script lang="ts">
	import { parts } from '$lib/money';

	let {
		amount,
		decimals = 2,
		currency = 'BDT'
	}: { amount: number; decimals?: number; currency?: string } = $props();

	const segments = $derived(parts(amount, decimals, currency));
</script>

<!--
	formatToParts rather than format, so grouping, decimal mark and symbol placement
	all come from the locale rather than from a string typed by hand.
-->
<span>{#each segments as p (p.type + p.value)}{p.value}{/each}</span>
