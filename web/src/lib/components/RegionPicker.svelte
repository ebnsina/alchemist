<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Search01Icon, Tick02Icon } from '@hugeicons/core-free-icons';
	import { searchRegions, REGIONS } from '$lib/regions';

	let { value = $bindable() }: { value: string } = $props();

	let query = $state('');
	let open = $state(false);
	const results = $derived(searchRegions(query));
	const chosen = $derived(REGIONS.find((r) => r.id === value));
</script>

<div class="relative">
	<button
		type="button"
		class="field flex items-center gap-2 text-left"
		onclick={() => (open = !open)}
		aria-expanded={open}
		aria-haspopup="listbox"
	>
		<span class="min-w-0 flex-1 truncate">
			{#if chosen}
				{chosen.city} <span class="mono">· {chosen.id}</span>
			{:else}
				{value || 'Choose a region'}
			{/if}
		</span>
		<HugeiconsIcon icon={Search01Icon} size={15} strokeWidth={2} class="flex-none text-faint" />
	</button>

	{#if open}
		<div class="picker">
			<input
				bind:value={query}
				class="field rounded-b-none border-0 border-b border-sunk"
				type="search"
				placeholder="Mumbai, ap-south, Europe…"
				aria-label="Search regions"
			/>
			<ul class="max-h-64 overflow-y-auto p-1.5" role="listbox">
				{#each results as r (r.id)}
					<li>
						<button
							type="button"
							class="node w-full"
							onclick={() => {
								value = r.id;
								open = false;
								query = '';
							}}
							role="option"
							aria-selected={value === r.id}
						>
							<span class="min-w-0 flex-1 text-left">
								<span class="block truncate text-sm font-medium">{r.city}</span>
								<span class="mono block truncate">{r.id} · {r.area}</span>
							</span>
							{#if value === r.id}
								<HugeiconsIcon
									icon={Tick02Icon}
									size={15}
									strokeWidth={2.4}
									class="flex-none text-accent"
								/>
							{/if}
						</button>
					</li>
				{:else}
					<li class="sub px-3 py-4 text-center">
						Nothing matches that. Type the code if your provider uses one we do not list.
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</div>

<style>
	.picker {
		position: absolute;
		z-index: 20;
		top: calc(100% + 6px);
		left: 0;
		right: 0;
		border-radius: var(--radius-md);
		background: var(--color-card);
		border: 1px solid var(--color-sunk);
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.28);
		overflow: hidden;
	}
</style>
