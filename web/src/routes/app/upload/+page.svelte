<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Upload01Icon,
		Link01Icon,
		Tick02Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon,
		FileValidationIcon,
		VideoReplayIcon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Steps from '$lib/components/Steps.svelte';
	import { startUpload, putFile, completeUpload, importFromUrl, ApiError } from '$lib/api';

	const steps = [
		{
			key: 'how',
			icon: VideoReplayIcon,
			title: 'Where is the video?',
			hint: 'On this computer, or already online somewhere.'
		},
		{
			key: 'what',
			icon: FileValidationIcon,
			title: 'Which one?',
			hint: 'Anything a camera, phone or editor produced. Up to 8K.'
		},
		{
			key: 'send',
			icon: Upload01Icon,
			title: 'Send it',
			hint: 'The file goes straight to storage — never through our servers.'
		}
	];

	let step = $state(0);
	let method = $state<'file' | 'url'>('file');
	let file = $state<File | null>(null);
	let url = $state('');
	let pct = $state(0);
	let busy = $state(false);
	let error = $state('');
	let assetId = $state('');
	let dragging = $state(false);

	const ready = $derived(method === 'file' ? file !== null : url.trim().length > 0);

	function pick(e: Event) {
		file = (e.target as HTMLInputElement).files?.[0] ?? null;
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		const f = e.dataTransfer?.files?.[0];
		if (f) {
			file = f;
			method = 'file';
			if (step === 0) step = 1;
		}
	}

	function back() {
		error = '';
		if (step > 0) step -= 1;
	}

	async function next() {
		error = '';
		if (step === 0) {
			step = 1;
			return;
		}
		if (step === 1) {
			if (!ready) return;
			step = 2;
			return;
		}
		busy = true;
		pct = 0;
		try {
			if (method === 'file' && file) {
				// Three calls, because the bytes never pass through the API: ask for a
				// target, PUT straight to storage, then tell the API it landed.
				const { asset_id, upload_url } = await startUpload();
				await putFile(upload_url, file, (p) => (pct = p));
				await completeUpload(asset_id);
				assetId = asset_id;
			} else {
				assetId = (await importFromUrl(url.trim())).asset_id;
			}
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'That did not work. Try again.';
		} finally {
			busy = false;
		}
	}

	function again() {
		step = 0;
		file = null;
		url = '';
		assetId = '';
		pct = 0;
		error = '';
	}

	const size = (n: number) =>
		new Intl.NumberFormat('en', { style: 'unit', unit: 'megabyte', maximumFractionDigits: 1 }).format(
			n / 1_000_000
		);
</script>

<Seo title="Upload — Alchemist" description="Send a video to Alchemist." />

<h1 class="text-2xl font-semibold tracking-tight">Upload</h1>
<p class="sub mt-1 max-w-xl">
	One video at a time. We work out which sizes your viewers need, so there is nothing here to
	configure.
</p>

<div class="card mt-6 max-w-2xl">
	{#if assetId}
		<div class="py-4 text-center" in:fly={{ y: 12, duration: 360, easing: cubicOut }}>
			<span
				class="mx-auto flex h-11 w-11 items-center justify-center rounded-full bg-solid text-on-solid"
			>
				<HugeiconsIcon icon={Tick02Icon} size={20} strokeWidth={2.6} />
			</span>
			<p class="title mt-4">We have it. Work has started.</p>
			<p class="sub mx-auto mt-2 max-w-sm">
				The small sizes finish first, so it is usually watchable within a minute. You do not
				have to wait here — we will keep going without you.
			</p>
			<div class="mt-6 flex flex-wrap justify-center gap-2">
				<a href="/app/videos/{assetId}/" class="btn-solid">Watch it come together</a>
				<button type="button" class="btn" onclick={again}>Send another</button>
			</div>
		</div>
	{:else}
		<Steps {steps} {step}>
			<div class="mt-5">
				{#if step === 0}
					<div class="grid gap-2.5 sm:grid-cols-2">
						{#each [{ id: 'file', icon: Upload01Icon, title: 'From this computer', why: 'Drag it in or browse for it.' }, { id: 'url', icon: Link01Icon, title: 'From a link', why: 'We fetch it. Nothing to upload.' }] as opt (opt.id)}
							<button
								type="button"
								class="choice"
								class:choice--on={method === opt.id}
								onclick={() => (method = opt.id as 'file' | 'url')}
								aria-pressed={method === opt.id}
							>
								<HugeiconsIcon icon={opt.icon} size={20} strokeWidth={1.7} />
								<span class="mt-2.5 block text-sm font-medium">{opt.title}</span>
								<span class="sub mt-1 block">{opt.why}</span>
							</button>
						{/each}
					</div>
					<!-- The third way in is a standing connection rather than an action, so
					     it lives on its own page. Signposted here because this is where
					     somebody looks for it. -->
					<p class="sub mt-3.5">
						Already have a library somewhere? <a href="/app/sources/" class="link">
							Connect a bucket
						</a> and we take everything in it, and everything added to it later.
					</p>
				{:else if step === 1 && method === 'file'}
					<label
						class="drop"
						class:drop--over={dragging}
						ondragover={(e) => {
							e.preventDefault();
							dragging = true;
						}}
						ondragleave={() => (dragging = false)}
						ondrop={onDrop}
					>
						<input type="file" accept="video/*" class="vh" onchange={pick} />
						<HugeiconsIcon icon={Upload01Icon} size={22} strokeWidth={1.7} class="text-faint" />
						{#if file}
							<span class="mt-3 block text-sm font-medium">{file.name}</span>
							<span class="mono mt-1 block">{size(file.size)} · Click to choose another</span>
						{:else}
							<span class="mt-3 block text-sm font-medium">Drop a video here</span>
							<span class="sub mt-1 block">or click to browse</span>
						{/if}
					</label>
				{:else if step === 1}
					<label class="block">
						<span class="vh">Link to the video</span>
						<input
							bind:value={url}
							class="field"
							type="url"
							placeholder="https://example.com/lecture.mp4"
							required
						/>
					</label>
					<p class="sub mt-2.5">
						It has to be reachable from the public internet. We will not follow a link into a
						private network.
					</p>
				{:else}
					<dl class="grid gap-3 rounded-md border border-sunk p-4">
						<div class="flex items-baseline justify-between gap-4">
							<dt class="label">What</dt>
							<dd class="truncate text-sm font-medium">{file ? file.name : url}</dd>
						</div>
						{#if file}
							<div class="flex items-baseline justify-between gap-4">
								<dt class="label">Size</dt>
								<dd class="num text-sm">{size(file.size)}</dd>
							</div>
						{/if}
						<div class="flex items-baseline justify-between gap-4">
							<dt class="label">Goes</dt>
							<dd class="text-sm">{method === 'file' ? 'Straight to storage' : 'We fetch it'}</dd>
						</div>
					</dl>

					{#if busy && method === 'file'}
						<div class="mt-4">
							<div class="flex items-baseline justify-between gap-3">
								<span class="label">Sending</span>
								<span class="num text-sm">{pct}%</span>
							</div>
							<div class="mini mt-2">
								<div
									class="h-full rounded-full bg-brand transition-[width] duration-200"
									style="width: {pct}%"
								></div>
							</div>
						</div>
					{/if}
				{/if}

				{#if error}
					<p class="mt-3 text-sm text-red" role="alert">{error}</p>
				{/if}

				<div class="mt-6 flex items-center gap-2">
					{#if step > 0 && !busy}
						<button type="button" class="btn flex-none" onclick={back} aria-label="Back">
							<HugeiconsIcon icon={ArrowLeft01Icon} size={16} strokeWidth={2.2} />
						</button>
					{/if}
					<button
						type="button"
						class="btn-solid flex-1"
						onclick={next}
						disabled={busy || (step === 1 && !ready)}
						aria-disabled={busy || (step === 1 && !ready)}
					>
						{#if busy}
							Sending…
						{:else if step < 2}
							Next
							<HugeiconsIcon icon={ArrowRight01Icon} size={16} strokeWidth={2.2} />
						{:else}
							Send it
						{/if}
					</button>
				</div>
			</div>
		</Steps>
	{/if}
</div>

<style>
	.choice {
		padding: 16px;
		border-radius: var(--radius-md);
		border: 1px solid var(--color-sunk);
		text-align: left;
		color: var(--color-faint);
		transition:
			border-color 140ms ease,
			color 140ms ease;
	}
	.choice--on {
		border-color: var(--color-brand);
		color: var(--color-accent);
	}
	.drop {
		display: block;
		padding: 32px 20px;
		border-radius: var(--radius-md);
		border: 1.5px dashed var(--color-muted);
		text-align: center;
		cursor: pointer;
		transition:
			border-color 140ms ease,
			background 140ms ease;
	}
	.drop:hover,
	.drop--over {
		border-color: var(--color-brand);
		background: var(--color-sunk);
	}
</style>
