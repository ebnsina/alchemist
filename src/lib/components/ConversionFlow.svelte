<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Video01Icon, Tick02Icon, Link01Icon } from '@hugeicons/core-free-icons';
	import { fly, fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	const steps = [
		{ label: 'They upload', hold: 2600 },
		{ label: 'We encode', hold: 4200 },
		{ label: 'You get a link', hold: 3400 }
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

<div class="card-flush w-full p-6 text-left">
	<ol class="label-caps mb-6 flex items-center justify-between gap-2">
		{#each steps as s, i (s.label)}
			<li
				class="flex items-center gap-2 whitespace-nowrap transition-colors duration-500 {i === stage
					? 'text-primary'
					: 'text-secondary'}"
			>
				<span
					class="status-dot transition-colors duration-500 {i <= stage
						? 'bg-tertiary'
						: 'bg-outline'}"
				></span>
				<span class="hidden sm:inline">{s.label}</span>
			</li>
		{/each}
	</ol>

	<div class="relative h-[168px] sm:h-[152px]">
		{#key stage}
			<div
				class="absolute inset-0"
				in:fly={{ y: 10, duration: 360, delay: 120, easing: cubicOut }}
				out:fade={{ duration: 110 }}
			>
				{#if stage === 0}
					<div
						class="flex h-full flex-col items-center justify-center gap-2 rounded-md border border-dashed border-outline"
					>
						<div
							class="flex items-center gap-2 rounded-md border border-outline bg-neutral px-3 py-2"
							in:fly={{ y: -20, duration: 480, delay: 240, easing: cubicOut }}
						>
							<HugeiconsIcon
								icon={Video01Icon}
								size={18}
								strokeWidth={1.6}
								class="flex-none text-tertiary"
							/>
							<span class="body-sm">lecture-week-4.mov</span>
							<span class="numeric text-secondary">1.2 GB</span>
						</div>
						<p class="body-sm text-secondary" in:fade={{ duration: 360, delay: 560 }}>
							Straight to storage, never through your servers
						</p>
					</div>
				{:else if stage === 1}
					<div class="flex h-full flex-col justify-center gap-4">
						<div class="flex items-baseline justify-between">
							<p class="body-sm">Making every size their viewers need</p>
							<p class="numeric text-secondary">{progress}%</p>
						</div>
						<div class="h-1.5 overflow-hidden rounded-full bg-outline">
							<div
								class="h-full rounded-full bg-tertiary transition-[width] duration-100 ease-linear"
								style="width: {progress}%"
							></div>
						</div>
						<ul class="flex flex-wrap gap-2">
							{#each outputs as o (o.name)}
								{#if elapsed >= o.at}
									<li
										class="numeric flex items-center gap-2 rounded-sm border border-outline px-2.5 py-1"
										in:scale={{ start: 0.9, duration: 260, easing: cubicOut }}
									>
										<HugeiconsIcon
											icon={Tick02Icon}
											size={12}
											strokeWidth={2.6}
											class="flex-none text-success"
										/>
										{o.name}
									</li>
								{/if}
							{/each}
						</ul>
					</div>
				{:else}
					<div class="flex h-full flex-col justify-center gap-4">
						<div class="flex items-center gap-2">
							<span
								class="flex h-9 w-9 flex-none items-center justify-center rounded-full bg-tertiary text-on-primary"
								in:scale={{ start: 0.6, duration: 360, easing: cubicOut }}
							>
								<HugeiconsIcon icon={Tick02Icon} size={16} strokeWidth={2.6} />
							</span>
							<div>
								<p class="body-sm font-medium">Ready to play</p>
								<p class="body-sm text-secondary">84 MB · signed, expiring, plays anywhere</p>
							</div>
						</div>
						<div
							class="flex items-center gap-2 rounded-md border border-outline bg-neutral px-3 py-2"
							in:fly={{ y: 10, duration: 360, delay: 160, easing: cubicOut }}
						>
							<HugeiconsIcon
								icon={Link01Icon}
								size={15}
								strokeWidth={1.7}
								class="flex-none text-tertiary"
							/>
							<span class="numeric truncate text-secondary">alchemist.video/w/8fc21a</span>
						</div>
					</div>
				{/if}
			</div>
		{/key}
	</div>
</div>
