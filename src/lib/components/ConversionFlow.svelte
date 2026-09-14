<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Video01Icon, Tick02Icon, Link01Icon } from '@hugeicons/core-free-icons';
	import { fly, fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	const steps = [
		{ label: 'Drop it in', hold: 2600 },
		{ label: 'We work on it', hold: 4200 },
		{ label: 'Take the link', hold: 3400 }
	];

	const outputs = [
		{ name: '1080p', at: 900 },
		{ name: '720p', at: 1900 },
		{ name: '480p', at: 2800 }
	];

	let stage = $state(0);
	let elapsed = $state(0);

	$effect(() => {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
			stage = 2;
			elapsed = 9999;
			return;
		}
		let timer: ReturnType<typeof setTimeout>;
		let ticker: ReturnType<typeof setInterval>;
		const run = () => {
			elapsed = 0;
			ticker = setInterval(() => (elapsed += 100), 100);
			timer = setTimeout(() => {
				clearInterval(ticker);
				stage = (stage + 1) % steps.length;
				run();
			}, steps[stage].hold);
		};
		run();
		return () => {
			clearTimeout(timer);
			clearInterval(ticker);
		};
	});

	const progress = $derived(Math.min(100, Math.round((elapsed / 3200) * 100)));
</script>

<div class="card shine relative w-full p-4 text-left sm:p-6">
	<ol class="mb-5 flex items-center justify-center gap-2.5 text-[11px] sm:gap-2 sm:text-xs">
		{#each steps as s, i (s.label)}
			<li
				class="flex items-center gap-2 whitespace-nowrap transition-colors duration-500"
				class:text-ink={i === stage}
				class:text-muted={i !== stage}
			>
				<span
					class="h-1.5 w-1.5 rounded-full transition-all duration-500 {i <= stage
						? 'bg-brand-mid'
						: 'bg-black/15'}"
					class:scale-150={i === stage}
				></span>
				{s.label}
				{#if i < steps.length - 1}
					<span class="ml-1 hidden h-px w-5 bg-black/10 sm:inline-block sm:w-8" aria-hidden="true"></span>
				{/if}
			</li>
		{/each}
	</ol>

	<div class="relative h-[168px] sm:h-[152px]">
		{#key stage}
			<div
				class="absolute inset-0"
				in:fly={{ y: 14, duration: 420, delay: 130, easing: cubicOut }}
				out:fade={{ duration: 120 }}
			>
				{#if stage === 0}
					<div
						class="flex h-full flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-hairline"
					>
						<div
							class="flex items-center gap-3 rounded-lg border border-hairline bg-card px-3 py-2"
							in:fly={{ y: -26, duration: 560, delay: 260, easing: cubicOut }}
						>
							<HugeiconsIcon icon={Video01Icon} size={20} strokeWidth={1.6} class="flex-none text-brand-light" />
							<span class="text-sm">lecture-week-4.mov</span>
							<span class="text-xs text-muted">1.2 GB</span>
						</div>
						<p class="text-xs text-muted" in:fade={{ duration: 400, delay: 620 }}>
							Drag it in, or pick it from your phone
						</p>
					</div>
				{:else if stage === 1}
					<div class="flex h-full flex-col justify-center gap-4">
						<div class="flex items-baseline justify-between">
							<p class="text-sm">Making the sizes your viewers need</p>
							<p class="tabular-nums text-xs text-muted">{progress}%</p>
						</div>
						<div class="h-1.5 overflow-hidden rounded-full bg-black/10">
							<div
								class="h-full rounded-full bg-brand-mid transition-[width] duration-100 ease-linear"
								style="width: {progress}%"
							></div>
						</div>
						<ul class="flex flex-wrap gap-2">
							{#each outputs as o (o.name)}
								{#if elapsed >= o.at}
									<li
										class="flex items-center gap-1.5 rounded-lg border border-hairline px-2.5 py-1 text-xs"
										in:scale={{ start: 0.86, duration: 320, easing: cubicOut }}
									>
										<HugeiconsIcon icon={Tick02Icon} size={12} strokeWidth={2.6} class="flex-none text-brand-light" />
										{o.name}
									</li>
								{/if}
							{/each}
						</ul>
					</div>
				{:else}
					<div class="flex h-full flex-col justify-center gap-4">
						<div class="flex items-center gap-3">
							<span
								class="flex h-9 w-9 flex-none items-center justify-center rounded-full bg-brand-mid text-ink"
								in:scale={{ start: 0.5, duration: 420, easing: cubicOut }}
							>
								<HugeiconsIcon icon={Tick02Icon} size={16} strokeWidth={2.6} />
							</span>
							<div>
								<p class="text-sm">Ready to share</p>
								<p class="text-xs text-muted">84 MB · plays on any phone, laptop or TV</p>
							</div>
						</div>
						<div
							class="flex items-center gap-2 rounded-lg border border-hairline bg-card px-3 py-2"
							in:fly={{ y: 12, duration: 420, delay: 180, easing: cubicOut }}
						>
							<HugeiconsIcon icon={Link01Icon} size={16} strokeWidth={1.7} class="flex-none text-brand-light" />
							<span class="truncate font-mono text-xs text-muted">alchemist.video/w/8fc21a</span>
							<span class="ml-auto flex-none text-xs text-brand-light">Copy</span>
						</div>
					</div>
				{/if}
			</div>
		{/key}
	</div>
</div>
