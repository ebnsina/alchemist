<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Upload01Icon, Link01Icon, Tick02Icon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import { startUpload, putFile, completeUpload, importFromUrl, ApiError } from '$lib/api';

	type Stage = 'idle' | 'sending' | 'queued' | 'error';

	let stage: Stage = $state('idle');
	let pct = $state(0);
	let assetId = $state('');
	let error = $state('');
	let url = $state('');
	let dragging = $state(false);

	async function send(file: File) {
		stage = 'sending';
		pct = 0;
		error = '';
		try {
			// Three calls, because the bytes never pass through the API: ask for a
			// target, PUT straight to storage, then tell the API it landed.
			const { asset_id, upload_url } = await startUpload();
			await putFile(upload_url, file, (p) => (pct = p));
			await completeUpload(asset_id);
			assetId = asset_id;
			stage = 'queued';
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'That did not work. Try again.';
			stage = 'error';
		}
	}

	async function fromUrl(e: SubmitEvent) {
		e.preventDefault();
		if (!url.trim()) return;
		stage = 'sending';
		error = '';
		try {
			const res = await importFromUrl(url.trim());
			assetId = res.asset_id;
			stage = 'queued';
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'That did not work. Try again.';
			stage = 'error';
		}
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		const file = e.dataTransfer?.files?.[0];
		if (file) send(file);
	}
</script>

<Seo title="Upload — Alchemist" description="Send a video to Alchemist." />

<h1 class="text-2xl font-semibold tracking-tight">Upload</h1>
<p class="mt-1 text-sm text-secondary">
	Drop a file, or point us at one that already lives somewhere else.
</p>

{#if stage === 'queued'}
	<div class="card mt-6 p-8 text-center" in:fly={{ y: 12, duration: 360, easing: cubicOut }}>
		<span
			class="mx-auto flex h-11 w-11 items-center justify-center rounded-full bg-tertiary text-on-primary"
		>
			<HugeiconsIcon icon={Tick02Icon} size={20} strokeWidth={2.6} />
		</span>
		<p class="mt-4 font-semibold">We have it</p>
		<p class="mx-auto mt-2 max-w-sm text-sm text-secondary">
			It is in the queue now. A ten-minute video is usually watchable in about a minute — you
			can leave this page.
		</p>
		<code class="mt-4 block font-mono text-xs text-secondary">{assetId}</code>
		<div class="mt-5 flex justify-center gap-3">
			<a href="/app/" class="btn-primary">See it in the list</a>
			<button
				type="button"
				class="btn-secondary"
				onclick={() => {
					stage = 'idle';
					url = '';
					pct = 0;
				}}
			>
				Send another
			</button>
		</div>
	</div>
{:else}
	<div
		class="card mt-6 p-8 text-center transition-colors {dragging ? 'ring-2 ring-tertiary' : ''}"
		ondragover={(e) => {
			e.preventDefault();
			dragging = true;
		}}
		ondragleave={() => (dragging = false)}
		ondrop={onDrop}
		role="region"
		aria-label="Drop a video file here"
	>
		<HugeiconsIcon
			icon={Upload01Icon}
			size={34}
			strokeWidth={1.6}
			class="mx-auto text-tertiary"
		/>

		{#if stage === 'sending'}
			<p class="mt-4 text-sm">Sending… {pct}%</p>
			<div class="mx-auto mt-3 h-1.5 max-w-sm overflow-hidden rounded-full bg-outline">
				<div
					class="h-full rounded-full bg-tertiary transition-[width] duration-200 ease-out"
					style="width: {pct}%"
				></div>
			</div>
		{:else}
			<p class="mt-4 font-semibold">Drop a video here</p>
			<p class="mt-1 text-sm text-secondary">Any format your camera, phone or editor made.</p>
			<label class="btn-primary mt-5 cursor-pointer">
				Choose a file
				<input
					type="file"
					accept="video/*"
					class="vh"
					onchange={(e) => {
						const f = e.currentTarget.files?.[0];
						if (f) send(f);
					}}
				/>
			</label>
		{/if}
	</div>

	<div class="card mt-4 p-6">
		<h2 class="flex items-center gap-2 text-sm font-semibold">
			<HugeiconsIcon icon={Link01Icon} size={16} strokeWidth={1.7} class="text-tertiary" />
			Already online somewhere?
		</h2>
		<p class="mt-1.5 text-sm text-secondary">
			Give us the address and we will fetch it ourselves. Nothing to upload from here.
		</p>
		<form class="mt-4 flex flex-col gap-3 sm:flex-row" onsubmit={fromUrl}>
			<label class="flex-1">
				<span class="vh">Video address</span>
				<input
					bind:value={url}
					class="field"
					type="url"
					placeholder="https://example.com/lecture.mp4"
				/>
			</label>
			<button type="submit" class="btn-secondary flex-none" disabled={!url.trim() || stage === 'sending'}>
				Fetch it
			</button>
		</form>
	</div>
{/if}

{#if error}
	<p class="mt-4 text-sm text-danger" role="alert">{error}</p>
{/if}
