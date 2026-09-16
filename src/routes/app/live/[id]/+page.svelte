<script lang="ts">
	import { page } from '$app/state';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Tick02Icon, Copy01Icon, Video01Icon, StopIcon } from '@hugeicons/core-free-icons';
	import Seo from '$lib/Seo.svelte';
	import { setCrumbs } from '$lib/crumbs.svelte';
	import {
		getLiveStream,
		startLiveStream,
		stopLiveStream,
		replaceLiveKey,
		ApiError,
		type LiveStream
	} from '$lib/api';

	const id = $derived(page.params.id ?? '');

	let stream = $state<LiveStream | null>(null);
	let loading = $state(true);
	let error = $state('');
	let ingest = $state('');
	// Shown once, then gone from memory as well as from the screen.
	let freshKey = $state('');
	let replacing = $state(false);
	let confirming = $state(false);
	let confirmingStop = $state(false);
	let stopping = $state(false);
	let copied = $state('');

	// Browser publishing. Nothing here runs until the customer presses a button: a page
	// that takes the camera on load is the one dark pattern this feature invites.
	let media = $state<MediaStream | null>(null);
	let cams = $state<MediaDeviceInfo[]>([]);
	let mics = $state<MediaDeviceInfo[]>([]);
	let camId = $state('');
	let micId = $state('');
	let camError = $state('');
	let phase = $state<'off' | 'preview' | 'connecting' | 'on'>('off');
	let preview = $state<HTMLVideoElement | null>(null);
	let peer: RTCPeerConnection | null = null;

	const SOURCE: Record<LiveStream['protocol'], string> = {
		camera: 'From this browser',
		srt: 'SRT encoder',
		rtmp: 'RTMP encoder'
	};

	const STATE: Record<LiveStream['state'], string> = {
		idle: 'Made, never started.',
		armed: 'Waiting for your encoder.',
		live: 'On air now.',
		ended: 'Finished. If anything went out, it is under Recordings.'
	};

	const said = (e: unknown) => (e instanceof ApiError ? e.message : 'Something went wrong.');

	$effect(() => {
		if (!id) return;
		load();
	});

	// The name, not the id: a trail reading "6080efa2" tells nobody which stream this is.
	$effect(() => {
		setCrumbs([{ label: 'Streams', href: '/app/live/' }, { label: stream?.name ?? 'Stream' }]);
	});

	async function load() {
		loading = true;
		try {
			stream = await getLiveStream(id);
			error = '';
		} catch (e) {
			error = said(e);
		} finally {
			loading = false;
		}
	}

	async function start() {
		try {
			ingest = (await startLiveStream(id)).ingest_url;
			await load();
		} catch (e) {
			error = said(e);
		}
	}

	async function replace() {
		replacing = true;
		try {
			freshKey = (await replaceLiveKey(id)).stream_key;
			confirming = false;
			await load();
		} catch (e) {
			error = said(e);
		} finally {
			replacing = false;
		}
	}

	async function copy(text: string, what: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied = what;
			setTimeout(() => (copied = ''), 2000);
		} catch {
			copied = '';
		}
	}

	// Every one of these is a different thing for the customer to do next, so none of
	// them may collapse into "something went wrong" -- and no browser text reaches them.
	function whyNoCamera(e: unknown) {
		const name = e instanceof DOMException ? e.name : '';
		if (name === 'NotAllowedError' || name === 'SecurityError')
			return 'This page is not allowed to use your camera. Allow it from the icon in your address bar, then try again.';
		if (name === 'NotFoundError' || name === 'OverconstrainedError')
			return 'We could not find a camera on this device. Plug one in, or use a phone or laptop that has one.';
		if (name === 'NotReadableError' || name === 'AbortError')
			return 'Another app or tab already has the camera. Close it, then try again.';
		return 'We could not open your camera. Try again, or open this page in a different browser.';
	}

	async function openCamera() {
		camError = '';
		// getUserMedia does not exist at all on an insecure address, so this is not a
		// refusal to explain away -- it is the address being wrong.
		if (!navigator.mediaDevices?.getUserMedia) {
			camError =
				'Browsers only hand over a camera on a secure address. Open this dashboard over https, or on localhost.';
			return;
		}
		try {
			releaseCamera();
			media = await navigator.mediaDevices.getUserMedia({
				video: camId ? { deviceId: { exact: camId } } : true,
				audio: micId ? { deviceId: { exact: micId } } : true
			});
		} catch (e) {
			camError = whyNoCamera(e);
			return;
		}
		// Device labels are blank until permission has been given once, so the lists
		// are only worth reading after a camera has actually opened.
		const devices = await navigator.mediaDevices.enumerateDevices();
		cams = devices.filter((d) => d.kind === 'videoinput');
		mics = devices.filter((d) => d.kind === 'audioinput');
		camId ||= media.getVideoTracks()[0]?.getSettings().deviceId ?? '';
		micId ||= media.getAudioTracks()[0]?.getSettings().deviceId ?? '';
		phase = 'preview';
	}

	// One POST: the offer goes up as SDP and the answer comes back in the body. The key
	// travels in the header because the ingest server ignores credentials in the query
	// string and refuses the publish with the same 401 a wrong key would get.
	async function whip(url: string, token: string, stream: MediaStream) {
		const pc = new RTCPeerConnection();
		peer = pc;
		for (const track of stream.getTracks()) pc.addTrack(track, stream);
		await pc.setLocalDescription(await pc.createOffer());
		await gathered(pc);
		const res = await fetch(url, {
			method: 'POST',
			headers: { 'Content-Type': 'application/sdp', Authorization: `Bearer publisher:${token}` },
			body: pc.localDescription?.sdp ?? ''
		});
		if (!res.ok) throw new Error('refused');
		await pc.setRemoteDescription({ type: 'answer', sdp: await res.text() });
	}

	// No trickle: one POST carries every candidate. Capped, because a gathering state
	// that never completes would otherwise leave the page saying "connecting" forever.
	const gathered = (pc: RTCPeerConnection) =>
		pc.iceGatheringState === 'complete'
			? Promise.resolve()
			: new Promise<void>((done) => {
					const t = setTimeout(done, 3000);
					pc.addEventListener('icegatheringstatechange', () => {
						if (pc.iceGatheringState !== 'complete') return;
						clearTimeout(t);
						done();
					});
				});

	async function goLive() {
		if (!media) return;
		phase = 'connecting';
		error = '';
		try {
			const armed = await startLiveStream(id);
			if (!armed.publish_token) throw new Error('refused');
			await whip(armed.ingest_url, armed.publish_token, media);
			phase = 'on';
			await load();
		} catch (e) {
			peer?.close();
			peer = null;
			phase = 'preview';
			error =
				e instanceof ApiError
					? e.message
					: 'We could not put you on air. Check your connection and try again.';
		}
	}

	// Stopping releases the camera as well as the connection. A recording light still
	// on after Stop is the thing a customer would never forgive, and never forget.
	function stopBroadcast() {
		peer?.close();
		peer = null;
		releaseCamera();
		phase = 'off';
		load();
	}

	// Both halves, always. Closing the connection alone leaves an encoder somewhere
	// else still sending, and telling the server alone leaves this browser publishing
	// with the camera light on.
	async function endBroadcast() {
		stopping = true;
		error = '';
		try {
			await stopLiveStream(id);
		} catch (e) {
			// Already over is not a fault: the broadcast is in the state they asked for.
			if (!(e instanceof ApiError && e.code === 'not_broadcasting')) error = said(e);
		} finally {
			confirmingStop = false;
			stopping = false;
			stopBroadcast();
		}
	}

	function releaseCamera() {
		media?.getTracks().forEach((t) => t.stop());
		media = null;
		if (preview) preview.srcObject = null;
	}

	// The video element does not exist until the preview renders, so the camera is
	// attached when both are there. Setting it inside openCamera silently did nothing
	// and left the customer looking at a black box while actually broadcasting.
	$effect(() => {
		if (preview && media) preview.srcObject = media;
	});

	// Leaving the page is a stop too, camera included.
	$effect(() => () => {
		peer?.close();
		releaseCamera();
	});

	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(
			new Date(iso)
		);
</script>

<Seo
	title={(stream?.name ?? 'Stream') + ' — Alchemist'}
	description="One live stream and its settings."
/>

{#if loading && !stream}
	<div class="sk h-8 w-64"></div>
{:else if error && !stream}
	<div class="card mt-6 py-8 text-center">
		<p class="title">We could not open that stream</p>
		<p class="sub mx-auto mt-2 max-w-sm">{error}</p>
		<a href="/app/live/" class="btn-solid mt-5">Back to your streams</a>
	</div>
{:else if stream}
	<header>
		<h1 class="text-2xl font-semibold tracking-tight">{stream.name}</h1>
		<p class="sub mt-1">
			{SOURCE[stream.protocol]} · made {when(stream.created_at)} · {STATE[stream.state]}
		</p>
	</header>

	{#if error}
		<!-- Beside the controls that produce it. At the foot of the page it sat below
		     three panels, where nobody pressing a button at the top would ever see it. -->
		<p class="mt-4 text-sm text-red" role="alert">{error}</p>
	{/if}

	{#if freshKey}
		<div class="card mt-6 p-6" in:fly={{ y: 12, duration: 340, easing: cubicOut }}>
			<p class="text-sm font-semibold">Your new stream key</p>
			<p class="mt-1 text-xs text-dim">
				Copy it now — this is the only time it is on screen. The previous key stopped working
				the moment this one was made.
			</p>
			<div class="mt-3 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
				<code class="truncate font-mono text-xs">{freshKey}</code>
				<button
					type="button"
					class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
					onclick={() => copy(freshKey, 'key')}
				>
					<HugeiconsIcon
						icon={copied === 'key' ? Tick02Icon : Copy01Icon}
						size={13}
						strokeWidth={2}
					/>
					{copied === 'key' ? 'Copied' : 'Copy'}
				</button>
			</div>
			<button type="button" class="btn btn-sm mt-4" onclick={() => (freshKey = '')}>
				I have saved it
			</button>
		</div>
	{/if}

	{#if stream.protocol === 'camera'}
		<section class="card mt-6 p-6">
			<h2 class="text-lg font-semibold tracking-tight">Go live from this browser</h2>

			{#if phase === 'off'}
				<p class="sub mt-2 max-w-lg">
					We will ask for your camera and microphone when you press this. Nothing is switched
					on before then, and nothing goes out until you press Go live on the next screen.
				</p>
				<button type="button" class="btn-solid mt-4" onclick={openCamera}>
					<HugeiconsIcon icon={Video01Icon} size={14} strokeWidth={2} />
					Turn on my camera
				</button>
			{:else}
				<!-- Muted because it is your own face on your own speakers: unmuted it howls. -->
				<video
					bind:this={preview}
					class="mt-4 w-full rounded-xl border border-sunk bg-sunk"
					style="aspect-ratio: 16 / 9"
					autoplay
					muted
					playsinline
				><track kind="captions" /></video>

				{#if cams.length > 1 || mics.length > 1}
					<div class="mt-4 grid gap-3 sm:grid-cols-2">
						{#if cams.length > 1}
							<label class="block">
								<span class="label">Camera</span>
								<select
									class="select mt-2"
									bind:value={camId}
									disabled={phase !== 'preview'}
									onchange={openCamera}
								>
									{#each cams as d (d.deviceId)}
										<option value={d.deviceId}>{d.label || 'Camera'}</option>
									{/each}
								</select>
							</label>
						{/if}
						{#if mics.length > 1}
							<label class="block">
								<span class="label">Microphone</span>
								<select
									class="select mt-2"
									bind:value={micId}
									disabled={phase !== 'preview'}
									onchange={openCamera}
								>
									{#each mics as d (d.deviceId)}
										<option value={d.deviceId}>{d.label || 'Microphone'}</option>
									{/each}
								</select>
							</label>
						{/if}
					</div>
					{#if phase !== 'preview'}
						<p class="sub mt-2">Swapping camera or microphone means stopping first.</p>
					{/if}
				{/if}

				<div class="mt-5 flex flex-wrap items-center gap-3">
					{#if phase === 'preview'}
						<button type="button" class="btn-solid" onclick={goLive}>Go live</button>
						<button type="button" class="btn" onclick={stopBroadcast}>Turn the camera off</button>
						<p class="sub">Only you can see this so far.</p>
					{:else if phase === 'connecting'}
						<button type="button" class="btn-solid" disabled>Going live…</button>
						<p class="sub">Handing your picture over. This takes a second or two.</p>
					{:else if confirmingStop}
						<button type="button" class="btn-solid" onclick={endBroadcast} disabled={stopping}>
							{stopping ? 'Stopping…' : 'Yes, stop it'}
						</button>
						<button type="button" class="btn" onclick={() => (confirmingStop = false)}>
							Keep going
						</button>
						<p class="sub">
							Viewers watching now will see it end. What has gone out so far is kept as a
							recording.
						</p>
					{:else}
						<button type="button" class="btn" onclick={() => (confirmingStop = true)}>
							<HugeiconsIcon icon={StopIcon} size={14} strokeWidth={2} />
							Stop the broadcast
						</button>
						<span class="chip chip-on">On air</span>
						<p class="sub">Viewers can watch now. Stopping also turns your camera off.</p>
					{/if}
				</div>
			{/if}

			{#if camError}
				<p class="mt-4 text-sm text-red" role="alert">{camError}</p>
			{/if}
		</section>
	{:else}
	<section class="card mt-6 p-6">
		<h2 class="text-lg font-semibold tracking-tight">Where your encoder connects</h2>
		{#if ingest}
			<!-- OBS joins Server and Stream Key with a slash, so one pasted URL gets the
			     key appended twice and lands on a path nothing authorised. -->
			<p class="label mt-4">Server</p>
			<div class="mt-2 flex items-center gap-2 rounded-xl border border-sunk bg-bg px-3 py-2">
				<code class="truncate font-mono text-xs">{ingest}</code>
				<button
					type="button"
					class="ml-auto flex flex-none items-center gap-1.5 text-xs text-ink"
					title="Copy the server address"
					onclick={() => copy(ingest, 'url')}
				>
					<HugeiconsIcon
						icon={copied === 'url' ? Tick02Icon : Copy01Icon}
						size={14}
						strokeWidth={2}
					/>
					<span class="vh">{copied === 'url' ? 'Copied' : 'Copy'}</span>
				</button>
			</div>
			<p class="sub mt-2">
				Swap <code class="font-mono">YOUR_STREAM_KEY</code> for your key. We only keep a
				fingerprint of it, so we cannot fill it in for you.
			</p>

			<p class="label mt-5">Stream Key</p>
			<div class="mt-2 rounded-xl border border-sunk bg-bg px-3 py-2">
				<code class="font-mono text-xs text-dim">Leave this empty</code>
			</div>
			<p class="sub mt-2">
				Your key is already in the server address above. Putting it here as well sends it
				twice and nothing will connect.
			</p>
		{:else if stream.state === 'armed' || stream.state === 'live'}
			<!-- The address is minted once per broadcast and never returned again, so
			     pressing Start here only ever answered "already waiting for an encoder". -->
			<p class="sub mt-2 max-w-lg">
				The address was shown when this stream was started, and we cannot show it again.
				If you no longer have it, stop the broadcast below and start it again — that
				gives you a fresh one.
			</p>
		{:else}
			<p class="sub mt-2">
				The address appears when you start the stream. Starting holds a slot for your encoder;
				nothing goes out until one connects.
			</p>
			<button type="button" class="btn-solid mt-4" onclick={start}>Start this stream</button>
		{/if}
	</section>

	{#if stream.state === 'armed' || stream.state === 'live'}
		<section class="card mt-4 p-6">
			<h2 class="text-lg font-semibold tracking-tight">End this broadcast</h2>
			<p class="sub mt-2 max-w-lg">
				Use this to finish a class early, or when your encoder was closed badly and the stream
				is still holding a slot. Whatever has gone out so far is kept as a recording.
			</p>
			{#if confirmingStop}
				<p class="mt-4 text-sm">
					Anyone watching will see it end. Your encoder will stop being accepted, and the
					recording starts being turned into an ordinary video.
				</p>
				<div class="mt-4 flex flex-wrap gap-2">
					<button type="button" class="btn-solid" onclick={endBroadcast} disabled={stopping}>
						{stopping ? 'Stopping…' : 'Yes, end it'}
					</button>
					<button type="button" class="btn" onclick={() => (confirmingStop = false)}>
						Keep it going
					</button>
				</div>
			{:else}
				<button type="button" class="btn mt-4" onclick={() => (confirmingStop = true)}>
					<HugeiconsIcon icon={StopIcon} size={14} strokeWidth={2} />
					Stop the broadcast
				</button>
			{/if}
		</section>
	{/if}

	<section class="card mt-4 p-6">
		<h2 class="text-lg font-semibold tracking-tight">Stream key</h2>
		<p class="sub mt-2">
			We keep only a fingerprint of your key, so there is no way to show it to you again — not
			to us either. If you did not save it, or it has been somewhere it should not, make a new
			one.
		</p>
		{#if confirming}
			<p class="mt-4 text-sm">
				The old key stops working immediately. An encoder already sending keeps going until it
				disconnects; anything that reconnects needs the new key.
			</p>
			<div class="mt-4 flex flex-wrap gap-2">
				<button type="button" class="btn-solid" onclick={replace} disabled={replacing}>
					{replacing ? 'Making one…' : 'Yes, replace it'}
				</button>
				<button type="button" class="btn" onclick={() => (confirming = false)}>Keep it</button>
			</div>
		{:else}
			<button type="button" class="btn mt-4" onclick={() => (confirming = true)}>
				Replace the stream key
			</button>
		{/if}
	</section>
	{/if}

	{#if stream.asset_id && stream.state !== 'armed'}
		<section class="card mt-4 p-6">
			<h2 class="text-lg font-semibold tracking-tight">The recording</h2>
			<!-- Offered only once something has actually gone out: on an armed stream this
			     linked to a recording of nothing. -->
			<p class="sub mt-2">Everything sent on this stream is kept as an ordinary video.</p>
			<a href="/app/videos/{stream.asset_id}/" class="btn mt-4">Open the recording</a>
		</section>
	{/if}

{/if}
