<script lang="ts">
	import { page } from '$app/state';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Tick02Icon,
		Delete02Icon,
		RectangularIcon,
		MobileNavigator01Icon,
		StopIcon,
		ComputerIcon,
		CropIcon,
		TextIcon,
		ImageAdd02Icon,
		Cancel01Icon,
		VolumeOffIcon,
		Scissor01Icon
	} from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import Timeline from '$lib/components/Timeline.svelte';
	import Tooltip from '$lib/components/Tooltip.svelte';
	import Player from '$lib/components/Player.svelte';
	import { setCrumbs } from '$lib/crumbs.svelte';
	import { getBranding, type EditOverlay, type Branding } from '$lib/api';
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';
	import {
		getAsset,
		listEdits,
		createEdit,
		deleteEdit,
		ApiError,
		type AssetDetail,
		type Edit
	} from '$lib/api';

	let asset = $state<AssetDetail | null>(null);
	let edits = $state<Edit[]>([]);
	let loading = $state(true);
	let error = $state('');
	let busy = $state(false);

	const id = $derived(page.params.id ?? '');

	// The edit being built. Times in seconds, crop in fractions of the source — the
	// API takes fractions precisely so the browser's render size never matters.
	let start = $state(0);
	let end = $state(0);
	let aspect = $state('');
	let width = $state(0);
	let mute = $state(false);
	let cropping = $state(false);
	let crop = $state({ x: 0, y: 0, w: 1, h: 1 });
	let current = $state(0);
	let playable = $state(true);
	let media = $state<HTMLVideoElement | null>(null);
	let playing = $state(false);
	let overlays = $state<EditOverlay[]>([]);
	// What the inspector is talking about. Adding something selects it, so its
	// controls are the first thing on screen rather than something to go and find.
	let selected = $state<number | null>(null);
	let brand = $state<Branding | null>(null);

	// The inspector shows what is selected. Nothing selected means the clip itself.
	const picked = $derived(selected !== null ? (overlays[selected] ?? null) : null);

	// Laid out as the frame itself: nine cells, five of them real positions. Picking a
	// corner by pointing at the corner beats picking it from a dropdown of words.
	const CORNERS = [
		{ id: 'top-left', label: 'Top left' },
		null,
		{ id: 'top-right', label: 'Top right' },
		null,
		{ id: 'centre', label: 'Middle' },
		null,
		{ id: 'bottom-left', label: 'Bottom left' },
		null,
		{ id: 'bottom-right', label: 'Bottom right' }
	] as const;

	function addText() {
		overlays = [
			...overlays,
			{ kind: 'text', text: 'Your text', colour: '#ffffff', shadow: true, at: 'bottom-left', scale: 0.06, opacity: 1 }
		];
		selected = overlays.length - 1;
	}

	function addLogo() {
		overlays = [
			...overlays,
			{ kind: 'image', source: 'logo', at: 'bottom-right', scale: 0.16, opacity: 0.7 }
		];
		selected = overlays.length - 1;
	}

	function dropOverlay(i: number) {
		overlays = overlays.filter((_, n) => n !== i);
		selected = null;
	}

	function patch(i: number, change: Partial<EditOverlay>) {
		overlays = overlays.map((o, n) => (n === i ? { ...o, ...change } : o));
	}

	// Where an overlay sits on the preview, so what you place is what you get.
	function box(o: EditOverlay) {
		if (o.at === 'free') {
			return `left:${(o.x ?? 0) * 100}%;top:${(o.y ?? 0) * 100}%;opacity:${o.opacity ?? 1}`;
		}
		const m = (o.margin ?? 0.04) * 100;
		const at = o.at ?? 'bottom-right';
		const v = at.startsWith('top') ? `top:${m}%` : at === 'centre' ? 'top:50%' : `bottom:${m}%`;
		const h = at.endsWith('left') ? `left:${m}%` : at === 'centre' ? 'left:50%' : `right:${m}%`;
		const centre = at === 'centre' ? 'transform:translate(-50%,-50%);' : '';
		return `${v};${h};${centre}opacity:${o.opacity ?? 1}`;
	}

	const SHAPES = [
		{ id: '', label: 'As it is', hint: 'Keep the shape it was filmed in', icon: RectangularIcon },
		{ id: '9:16', label: 'Reel', hint: 'Upright, for phones', icon: MobileNavigator01Icon },
		{ id: '1:1', label: 'Square', hint: 'For feeds', icon: StopIcon },
		{ id: '16:9', label: 'Wide', hint: 'For players and TVs', icon: ComputerIcon }
	];
	const WIDTHS = [0, 1920, 1280, 1080, 720];

	async function load() {
		error = '';
		try {
			const [a, e] = await Promise.all([getAsset(id), listEdits(id)]);
			asset = a;
			edits = e.edits;
			getBranding()
				.then((b) => (brand = b))
				.catch(() => {});
			if (end === 0 && a.duration_seconds) end = a.duration_seconds;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (id) load();
	});

	$effect(() => {
		setCrumbs([
			{ label: 'Studio', href: '/app/studio/' },
			{ label: id.slice(0, 8), href: `/app/videos/${id}/` },
			{ label: 'Editing' }
		]);
	});

	// Keep asking while an edit is rendering, and stop the moment none is.
	$effect(() => {
		if (!edits.some((e) => e.state === 'queued' || e.state === 'rendering')) return;
		const t = setInterval(load, 4000);
		return () => clearInterval(t);
	});

	// Back to the video as it is, so an experiment can be abandoned without reloading.
	function reset() {
		start = 0;
		end = asset?.duration_seconds ?? 0;
		aspect = '';
		width = 0;
		mute = false;
		cropping = false;
		crop = { x: 0, y: 0, w: 1, h: 1 };
		overlays = [];
		selected = null;
		error = '';
	}

	// Starting at the whole frame left nothing to drag: every clamp correctly refused
	// to move a box that already filled its bounds. It opens inset instead, so the
	// first thing you see is something you can obviously grab.
	function startCrop() {
		cropping = !cropping;
		if (cropping && crop.w >= 1 && crop.h >= 1) {
			crop = { x: 0.12, y: 0.12, w: 0.76, h: 0.76 };
		}
	}

	// Crop by dragging the box on the frame. Fractions of the frame throughout, so
	// the numbers mean the same thing whatever size the preview happens to be.
	let grabbing = $state<string | null>(null);
	let grabFrom = { x: 0, y: 0, crop: { x: 0, y: 0, w: 1, h: 1 } };

	function frameBox(): DOMRect | null {
		return document.querySelector('video')?.getBoundingClientRect() ?? null;
	}

	// Dragging an overlay puts it wherever it is dropped: `at` becomes "free" and the
	// API reads x/y instead of a corner. Picking a corner again goes back to snapping.
	let dragging = $state<number | null>(null);
	let dragFrom = { x: 0, y: 0, ox: 0, oy: 0, moved: false };

	function grabOverlay(e: PointerEvent, i: number) {
		const r = frameBox();
		const el = (e.currentTarget as HTMLElement).getBoundingClientRect();
		if (!r || r.width === 0) return;
		e.preventDefault();
		selected = i;
		dragging = i;
		dragFrom = {
			x: e.clientX,
			y: e.clientY,
			ox: (el.left - r.left) / r.width,
			oy: (el.top - r.top) / r.height,
			moved: false
		};
	}

	function dragOverlay(e: PointerEvent) {
		if (dragging === null) return;
		const r = frameBox();
		const el = document.getElementById(`ov-${dragging}`)?.getBoundingClientRect();
		if (!r || r.width === 0 || !el) return;
		const dx = (e.clientX - dragFrom.x) / r.width;
		const dy = (e.clientY - dragFrom.y) / r.height;
		if (!dragFrom.moved && Math.abs(dx) < 0.004 && Math.abs(dy) < 0.004) return;
		dragFrom.moved = true;
		// Clamped so what was dragged stays reachable: the API clamps too, but a thing
		// you cannot see is a thing you cannot drag back.
		const clamp = (v: number, hi: number) => Math.min(hi, Math.max(0, v));
		patch(dragging, {
			at: 'free',
			x: clamp(dragFrom.ox + dx, 1 - el.width / r.width),
			y: clamp(dragFrom.oy + dy, 1 - el.height / r.height)
		});
	}

	function grab(e: PointerEvent, what: string) {
		e.preventDefault();
		grabbing = what;
		grabFrom = { x: e.clientX, y: e.clientY, crop: { ...crop } };
	}

	function dragCrop(e: PointerEvent) {
		if (!grabbing) return;
		const r = frameBox();
		if (!r || r.width === 0) return;
		const dx = (e.clientX - grabFrom.x) / r.width;
		const dy = (e.clientY - grabFrom.y) / r.height;
		const f = grabFrom.crop;
		const clamp = (v: number, lo: number, hi: number) => Math.min(hi, Math.max(lo, v));

		if (grabbing === 'move') {
			crop = {
				...crop,
				x: clamp(f.x + dx, 0, 1 - f.w),
				y: clamp(f.y + dy, 0, 1 - f.h)
			};
			return;
		}
		// A corner moves two edges, and the box never inverts or leaves the frame.
		let { x, y, w, h } = f;
		const min = 0.05;
		if (grabbing.includes('w')) {
			const nx = clamp(f.x + dx, 0, f.x + f.w - min);
			w = f.x + f.w - nx;
			x = nx;
		} else {
			w = clamp(f.w + dx, min, 1 - f.x);
		}
		if (grabbing.includes('n')) {
			const ny = clamp(f.y + dy, 0, f.y + f.h - min);
			h = f.y + f.h - ny;
			y = ny;
		} else {
			h = clamp(f.h + dy, min, 1 - f.y);
		}
		crop = { x, y, w, h };
	}

	// Arrow keys move a corner by one percent, five with shift.
	function nudge(e: KeyboardEvent, corner: string) {
		const step = (e.shiftKey ? 5 : 1) / 100;
		const d = { ArrowLeft: [-step, 0], ArrowRight: [step, 0], ArrowUp: [0, -step], ArrowDown: [0, step] }[
			e.key
		];
		if (!d) return;
		e.preventDefault();
		grabFrom = { x: 0, y: 0, crop: { ...crop } };
		grabbing = corner;
		dragCrop({ clientX: d[0] * (frameBox()?.width ?? 1), clientY: d[1] * (frameBox()?.height ?? 1) } as PointerEvent);
		grabbing = null;
	}

	function toggle() {
		if (!media) return;
		if (media.paused) media.play().catch(() => {});
		else media.pause();
	}

	function seek(t: number) {
		current = t;
		if (media) media.currentTime = t;
	}

	async function apply() {
		error = '';
		busy = true;
		try {
			await createEdit(id, {
				start_sec: start || undefined,
				end_sec: end && asset?.duration_seconds && end < asset.duration_seconds ? end : undefined,
				...(cropping ? { crop_x: crop.x, crop_y: crop.y, crop_w: crop.w, crop_h: crop.h } : {}),
				...(overlays.length ? { overlays } : {}),
				aspect: aspect || undefined,
				width: width || undefined,
				mute: mute || undefined
			});
			reset();
			await load();
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}

	async function forget(e: Edit) {
		error = '';
		try {
			await deleteEdit(e.id);
			await load();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong.';
		}
	}

	const poster = $derived(
		asset?.playback?.poster ? PUBLIC_ALCHEMIST_API + asset.playback.poster : null
	);
	const duration = $derived(asset?.duration_seconds ?? 0);
	const changed = $derived(
		start > 0 ||
			(duration > 0 && end < duration) ||
			cropping ||
			aspect !== '' ||
			width > 0 ||
			mute ||
			overlays.length > 0
	);

	// What the result will be, said in words. A list of numbers is not a description.
	const summary = $derived.by(() => {
		if (!changed) return 'Nothing changed yet. Pick a shape, or drag the ends of the bar.';
		const bits: string[] = [];
		if (start > 0 || (duration > 0 && end < duration)) bits.push(`${(end - start).toFixed(1)}s of it`);
		if (cropping) bits.push('a cropped area');
		const shape = SHAPES.find((s) => s.id === aspect);
		if (aspect) bits.push(shape ? shape.label.toLowerCase() : aspect);
		if (width) bits.push(`${width}px wide`);
		if (mute) bits.push('no sound');
		if (overlays.length) bits.push(`${overlays.length} thing${overlays.length > 1 ? 's' : ''} on top`);
		return 'A new video: ' + bits.join(', ') + '. The one you started from stays as it is.';
	});

	const STATE: Record<string, string> = {
		queued: 'Waiting to start',
		rendering: 'Making it',
		done: 'Ready',
		failed: 'Did not work'
	};

	const describe = (o: Edit['ops']) => {
		const bits: string[] = [];
		if (o.start_sec || o.end_sec) bits.push(`${(o.start_sec ?? 0).toFixed(1)}s–${(o.end_sec ?? 0).toFixed(1)}s`);
		if (o.crop_w) bits.push('cropped');
		if (o.aspect) bits.push(o.aspect);
		if (o.width) bits.push(`${o.width}px`);
		if (o.mute) bits.push('muted');
		if (o.overlays?.length) bits.push(`${o.overlays.length} overlay${o.overlays.length > 1 ? 's' : ''}`);
		return bits.length ? bits.join(' · ') : 'a copy';
	};
</script>

<!-- Dragging continues outside the box, so the listeners live on the window. -->
<svelte:window
	onpointermove={(e) => {
		dragCrop(e);
		dragOverlay(e);
	}}
	onpointerup={() => {
		grabbing = null;
		dragging = null;
	}}
	onpointercancel={() => {
		grabbing = null;
		dragging = null;
	}}
/>

<Seo title="Studio — Alchemist" description="Cut, crop and reshape a video." />

<header class="flex flex-wrap items-center justify-between gap-4">
	<div class="min-w-0">
		<div class="flex items-center gap-2.5">
			<h1 class="text-2xl font-semibold tracking-tight">Studio</h1>
			{#if changed}
				<!-- Unsaved work, said plainly: nothing here exists until it is made. -->
				<span class="chip chip-on">Unsaved changes</span>
			{/if}
		</div>
		<p class="sub mt-0.5">Makes a new video — the one you started from never changes.</p>
	</div>

	<div class="flex flex-none items-center gap-2">
		{#if changed}
			<button type="button" class="btn" onclick={reset} disabled={busy}>Start over</button>
		{/if}
		<button
			type="button"
			class="btn-solid"
			onclick={apply}
			disabled={busy || !changed}
			aria-disabled={busy || !changed}
		>
			{busy ? 'Making it…' : 'Make it'}
		</button>
	</div>
</header>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="mt-6 grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
		<div class="sk aspect-video w-full"></div>
		<div class="sk h-72"></div>
	</div>
{:else if !asset}
	<p class="sub mt-6">We could not find that video.</p>
{:else if !asset.playback}
	<div class="card mt-6">
		<p class="title">Not ready to edit yet</p>
		<p class="sub mt-2">
			Editing opens once a video has finished processing — until then there is nothing to cut.
		</p>
	</div>
{:else}
	<!-- Three zones, the way an editor is laid out: tools down the side, the stage in
	     the middle, and the timeline across the bottom of both. -->
	<div class="editor mt-4">
		<aside class="editor__tools">
			{#if picked}
				<!-- Contextual: the thing you just added or clicked is what this panel is
				     about, and it is the first thing on screen. -->
				<section class="picked">
					<div class="flex items-center justify-between gap-2">
						<p class="label text-accent">
							{picked.kind === 'text' ? 'Text' : 'Your logo'}
						</p>
						<div class="flex items-center gap-1">
							<button
								type="button"
								class="icon-btn"
								onclick={() => dropOverlay(selected!)}
								aria-label="Remove this"
							>
								<HugeiconsIcon icon={Delete02Icon} size={15} strokeWidth={1.8} />
							</button>
							<button
								type="button"
								class="icon-btn"
								onclick={() => (selected = null)}
								aria-label="Done with this"
							>
								<HugeiconsIcon icon={Cancel01Icon} size={15} strokeWidth={1.8} />
							</button>
						</div>
					</div>

					{#if picked.kind === 'text'}
						<label class="mt-3 block">
							<span class="mono mb-1 block">What it says</span>
							<input
								class="field py-1.5 text-sm"
								type="text"
								value={picked.text ?? ''}
								maxlength="120"
								oninput={(e) => patch(selected!, { text: e.currentTarget.value })}
							/>
						</label>
						<div class="mt-3 flex items-center gap-3">
							<label class="flex items-center gap-2">
								<span class="mono">Colour</span>
								<input
									type="color"
									value={picked.colour ?? '#ffffff'}
									oninput={(e) => patch(selected!, { colour: e.currentTarget.value })}
								/>
							</label>
							<label class="mono flex items-center gap-2">
								<input
									type="checkbox"
									checked={picked.shadow ?? false}
									onchange={(e) => patch(selected!, { shadow: e.currentTarget.checked })}
								/>
								Shadow
							</label>
						</div>
					{/if}

					<div class="mt-3">
						<span class="mono mb-1.5 block">
							Where it sits{picked.at === 'free' ? ' — dropped where you put it' : ', or drag it on the video'}
						</span>
						<div class="corners">
							{#each CORNERS as c, i (i)}
								{#if c}
									<button
										type="button"
										class="corner"
										class:corner--on={(picked.at ?? 'bottom-right') === c.id}
										onclick={() => patch(selected!, { at: c.id })}
										aria-label={c.label}
										aria-pressed={(picked.at ?? 'bottom-right') === c.id}
									></button>
								{:else}
									<span class="corner corner--off"></span>
								{/if}
							{/each}
						</div>
					</div>

					<div class="mt-3 grid grid-cols-2 gap-3">
						<label class="block">
							<span class="mono mb-1 block">Size</span>
							<input
								type="range"
								min="0.02"
								max="0.6"
								step="0.01"
								value={picked.scale ?? 0.15}
								oninput={(e) => patch(selected!, { scale: Number(e.currentTarget.value) })}
							/>
						</label>
						<label class="block">
							<span class="mono mb-1 block">Fade</span>
							<input
								type="range"
								min="0.1"
								max="1"
								step="0.05"
								value={picked.opacity ?? 1}
								oninput={(e) => patch(selected!, { opacity: Number(e.currentTarget.value) })}
							/>
						</label>
					</div>
				</section>
			{/if}

			{#if overlays.length > 0}
				<section>
					<p class="label">On top</p>
					<ul class="mt-2 grid gap-1">
						{#each overlays as o, i (i)}
							<li>
								<button
									type="button"
									class="layer"
									class:layer--on={selected === i}
									onclick={() => (selected = i)}
								>
									<HugeiconsIcon
										icon={o.kind === 'text' ? TextIcon : ImageAdd02Icon}
										size={14}
										strokeWidth={1.8}
										class="flex-none"
									/>
									<span class="min-w-0 flex-1 truncate text-left text-sm">
										{o.kind === 'text' ? o.text || 'Text' : 'Your logo'}
									</span>
								</button>
							</li>
						{/each}
					</ul>
				</section>
			{/if}

			{#if cropping}
				<section>
					<p class="label">Crop</p>
					<p class="sub mt-2">
						Drag the box on the video to move it, or a corner to resize it.
					</p>
				</section>
			{/if}

			<section class="rounded-md border border-sunk p-3.5" class:tier--lead={changed}>
				<p class="label">What you will get</p>
				<p class="sub mt-2">{summary}</p>
			</section>
		</aside>

		<div class="editor__left">
		<div class="editor__toolbar">
			<span class="label">Add</span>
			<Tooltip icon={TextIcon} label="Add text" onclick={addText} />
			<Tooltip
				icon={ImageAdd02Icon}
				label={brand?.logo_url ? 'Add your logo' : 'Upload a logo first'}
				disabled={!brand?.logo_url}
				onclick={addLogo}
			/>
			<span class="editor__sep"></span>
			<Tooltip icon={CropIcon} label="Crop the frame" on={cropping} onclick={startCrop} />
			<Tooltip icon={VolumeOffIcon} label="Drop the sound" on={mute} onclick={() => (mute = !mute)} />
			<span class="editor__sep"></span>
			<span class="label">Shape</span>
			{#each SHAPES as sh (sh.id)}
				<Tooltip
					icon={sh.icon}
					label="{sh.label} — {sh.hint}"
					on={aspect === sh.id}
					onclick={() => (aspect = sh.id)}
				/>
			{/each}
			<span class="editor__sep"></span>
			<label class="label" for="studio-size">Size</label>
			<select id="studio-size" bind:value={width} class="editor__select">
				{#each WIDTHS as px (px)}
					<option value={px}>{px === 0 ? 'As filmed' : `${px}px`}</option>
				{/each}
			</select>
			{#if overlays.length}
				<span class="mono ml-auto" style="color:rgba(255,255,255,.55)">
					{overlays.length} on top
				</span>
			{/if}
		</div>

		<div class="editor__stage">
			<Player
				hls={asset.playback.hls}
				poster={asset.playback.poster}
				bind:currentTime={current}
				bind:playable
				bind:element={media}
				bind:playing
				footer={false}
				fit
			>
				{#snippet overlay()}
					{#each overlays as o, i (i)}
						<!-- Clickable on the frame too: selecting a thing by pointing at it is
						     how every editor works, and it keeps the panel in step. -->
						<button
							type="button"
							id="ov-{i}"
							class="ov"
							class:ov--on={selected === i}
							class:ov--drag={dragging === i}
							style={box(o)}
							onpointerdown={(e) => grabOverlay(e, i)}
							onclick={() => (selected = i)}
							aria-label="Drag to move this {o.kind === 'text' ? 'text' : 'logo'}"
						>
							{#if o.kind === 'text'}
								<span
									class="ov__text"
									style="color:{o.colour ?? '#ffffff'}; font-size:{(o.scale ?? 0.06) * 100}cqw;
										{o.shadow ? 'text-shadow:0 2px 4px rgba(0,0,0,.6);' : ''}"
								>{o.text || 'Your text'}</span>
							{:else if brand?.logo_url}
								<img
									src={PUBLIC_ALCHEMIST_API + brand.logo_url}
									alt=""
									style="width:{(o.scale ?? 0.16) * 100}cqw"
								/>
							{/if}
						</button>
					{/each}
					{#if cropping}
						<div
							class="cropbox"
							style="left:{crop.x * 100}%; top:{crop.y * 100}%; width:{crop.w * 100}%; height:{crop.h * 100}%"
							onpointerdown={(e) => grab(e, 'move')}
							role="application"
							aria-label="Crop area — drag to move, drag a corner to resize"
						>
							{#each [['nw', 'Top left'], ['ne', 'Top right'], ['sw', 'Bottom left'], ['se', 'Bottom right']] as [h, name] (h)}
								<!-- A real button: draggable with a pointer, and nudgeable with arrow
								     keys for anyone without one. -->
								<button
									type="button"
									class="crophandle crophandle--{h}"
									aria-label="{name} corner"
									onpointerdown={(e) => {
										e.stopPropagation();
										grab(e, h);
									}}
									onkeydown={(e) => nudge(e, h)}
								></button>
							{/each}
						</div>
					{/if}
				{/snippet}
			</Player>
			{#if !playable}
				<p class="sub px-4 pb-3 text-center">
					This browser cannot play the stream, so there is nothing to cut against by eye.
					Type the times below instead — the edit itself works either way.
				</p>
			{/if}
		</div>

			<div class="editor__timeline">
				<Timeline
					{duration}
					bind:start
					bind:end
					bind:current
					{playing}
					muted={mute}
					onplay={toggle}
					onseek={seek}
				/>
			</div>
		</div>
	</div>

	<h2 class="mt-10 text-lg font-semibold tracking-tight">What you have made</h2>
	{#if edits.length === 0}
		<p class="sub mt-3">Nothing yet. Anything you make from this video shows up here.</p>
	{:else}
		<ul class="card mt-4 divide-y divide-sunk p-0">
			{#each edits as e (e.id)}
				<li class="flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3.5">
					<div class="min-w-0 flex-1">
						<p class="text-sm font-medium">{describe(e.ops)}</p>
						<p class="mono mt-0.5">{STATE[e.state] ?? e.state}</p>
					</div>
					{#if e.state === 'done' && e.output_asset_id}
						<a href="/app/videos/{e.output_asset_id}/" class="btn btn-sm flex-none">
							<HugeiconsIcon icon={Tick02Icon} size={14} strokeWidth={2.4} />
							Open it
						</a>
					{:else if e.state === 'failed'}
						<span class="chip flex-none text-red">Failed</span>
					{:else}
						<span class="chip chip-on flex-none">{STATE[e.state]}</span>
					{/if}
					<button
						type="button"
						class="icon-btn flex-none"
						onclick={() => forget(e)}
						aria-label="Forget this edit"
					>
						<HugeiconsIcon icon={Delete02Icon} size={15} strokeWidth={1.8} />
					</button>
				</li>
			{/each}
		</ul>
		<p class="sub mt-3">
			Forgetting an edit removes the record, not the video it made. Delete that through the
			video itself.
		</p>
	{/if}
{/if}

<style>
	/* Tools, stage, timeline. On a phone the three stack and the page scrolls, because
	   a three-pane editor at 360px is not an editor. */
	.editor {
		display: grid;
		gap: 12px;
		grid-template-areas: 'left' 'inspector';
		min-height: 0;
	}
	@media (min-width: 1024px) {
		.editor {
			/* Viewer left, inspector right, timeline the full width underneath — the
			   arrangement every editor uses, so nobody has to learn this one. It fills
			   the panel rather than scrolling the page: a timeline you have to scroll
			   to reach is not a timeline. */
			grid-template-columns: minmax(0, 1fr) 320px;
			/* Rows size to content. Trying to fit the video to the viewport broke the
			   player three different ways; a stage that is its natural size and a page
			   that scrolls is worse in theory and better in practice. */
			align-items: start;
			grid-template-areas: 'left inspector';
		}
	}
	/* Its own row, so the flex stage cannot squeeze it: as a flex child it collapsed
	   to a sliver and its buttons spilled over the video, where clicks landed on the
	   video instead of the button. */
	.editor__left {
		grid-area: left;
		display: grid;
		gap: 12px;
		min-width: 0;
	}
	.editor__toolbar {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
		padding: 8px 10px;
		border: 1px solid var(--color-sunk);
		border-radius: var(--radius-md);
		background: #0b0b0d;
	}
	.editor__toolbar :global(.label) {
		color: rgba(255, 255, 255, 0.5);
		margin-right: 2px;
	}
	.editor__select {
		height: 32px;
		padding: 0 26px 0 10px;
		border-radius: var(--radius-sm);
		background: rgba(255, 255, 255, 0.1)
			url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16' fill='none' stroke='%23fff' stroke-width='1.8'%3E%3Cpath d='M4 6.5 8 10.5 12 6.5'/%3E%3C/svg%3E")
			no-repeat right 6px center / 14px;
		color: #fff;
		font-family: var(--font-mono);
		font-size: 11px;
		appearance: none;
	}
	.editor__select option {
		color: var(--color-ink);
	}
	.editor__sep {
		width: 1px;
		height: 20px;
		margin: 0 4px;
		background: rgba(255, 255, 255, 0.14);
	}

	.editor__tools {
		grid-area: inspector;
		display: grid;
		align-content: start;
		gap: 18px;
		padding: 16px;
		border: 1px solid var(--color-sunk);
		border-radius: var(--radius-md);
		background: var(--color-card);
		/* Its own scroll, so a long inspector never pushes the timeline off screen. */
	}
	/* The viewer sits on a dark ground whatever the theme, the way a viewer does:
	   the frame is what is being judged, and a light surround skews it. */
	.editor__stage {
		display: grid;
		place-items: center;
		/* The video's native controls draw on its bottom edge; flush against a
		   clipped, rounded container they were sliced in half. */
		padding: 10px;
		min-height: 0;
		border: 1px solid var(--color-sunk);
		border-radius: var(--radius-md);
		background: #0b0b0d;
		overflow: hidden;
	}
	.editor__timeline {
		min-width: 0;
		border: 1px solid var(--color-sunk);
		border-radius: var(--radius-md);
		background: var(--color-card);
		overflow: hidden;
	}

	/* The preview is the container overlays are sized against, so a fraction of the
	   frame means the same on screen as it will in the render. */
	/* inline-size, not size: full containment makes the box ignore its contents, and
	   the video is what gives it a height. */
	:global(.editor__stage .card) {
		container-type: inline-size;
	}
	.ov {
		position: absolute;
		white-space: nowrap;
		line-height: 1;
		/* The overlay layer is pointer-events:none so the player keeps its controls;
		   each overlay turns them back on for itself. */
		pointer-events: auto;
		cursor: grab;
		touch-action: none;
		outline: 1px dashed transparent;
		outline-offset: 3px;
	}
	.ov--drag {
		cursor: grabbing;
	}
	.ov:hover {
		outline-color: rgba(255, 255, 255, 0.5);
	}
	.ov--on {
		outline: 2px solid var(--color-brand);
		outline-offset: 3px;
	}

	/* The nine-box: where a thing sits is a place on the frame, not a word in a list. */
	.corners {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 3px;
		width: 84px;
	}
	.corner {
		height: 24px;
		border-radius: var(--radius-sk);
		background: var(--color-sunk);
		border: 1px solid var(--color-muted);
		transition: background 120ms ease;
	}
	.corner:hover {
		background: var(--color-muted);
	}
	.corner--on {
		background: var(--color-brand);
		border-color: var(--color-brand);
	}
	.corner--off {
		background: transparent;
		border-color: transparent;
	}

	.picked {
		padding: 14px;
		border-radius: var(--radius-md);
		border: 1.5px solid var(--color-brand);
		background: var(--color-card);
	}
	.layer {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 6px 8px;
		border-radius: var(--radius-sk);
		color: var(--color-dim);
		transition: background 120ms ease;
	}
	.layer:hover {
		background: var(--color-sunk);
	}
	.layer--on {
		background: var(--color-sunk);
		color: var(--color-ink);
	}
	.ov__text {
		font-family: var(--font-sans);
		font-weight: 600;
	}

	.cropbox {
		position: absolute;
		border: 2px solid var(--color-brand);
		background: color-mix(in oklab, var(--color-brand) 12%, transparent);
		box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.45);
		pointer-events: auto;
		cursor: move;
		touch-action: none;
	}
	.crophandle {
		position: absolute;
		height: 14px;
		width: 14px;
		background: var(--color-brand);
		border: 2px solid var(--color-on-brand);
		border-radius: var(--radius-sk);
		touch-action: none;
	}
	.crophandle--nw {
		top: -8px;
		left: -8px;
		cursor: nwse-resize;
	}
	.crophandle--ne {
		top: -8px;
		right: -8px;
		cursor: nesw-resize;
	}
	.crophandle--sw {
		bottom: -8px;
		left: -8px;
		cursor: nesw-resize;
	}
	.crophandle--se {
		bottom: -8px;
		right: -8px;
		cursor: nwse-resize;
	}
</style>
