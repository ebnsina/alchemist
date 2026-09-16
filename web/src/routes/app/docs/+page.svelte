<script lang="ts">
	import { PUBLIC_ALCHEMIST_API } from '$env/static/public';
	import Seo from '$lib/Seo.svelte';
	import Json from '$lib/components/Json.svelte';

	// The reference lives next to the thing it describes. Anything that would go
	// stale — the spec itself — is linked, not copied.
	const endpoints = [
		{
			method: 'POST',
			path: '/v1/uploads',
			summary: 'Ask for somewhere to put a file',
			body: null,
			response: `{
  "asset_id": "0f4c1e52-5f1a-4a2e-9a55-0a1f9d2b77c1",
  "upload_url": "https://storage.example/alchemist/src/0f4c…?X-Amz-Signature=…",
  "expires_in_seconds": 900
}`,
			note: 'PUT the bytes to upload_url yourself. They never pass through us, which is why a large file does not tie up a request.'
		},
		{
			method: 'POST',
			path: '/v1/assets/{id}/complete',
			summary: 'Tell us the bytes landed',
			body: null,
			response: `{ "asset_id": "0f4c1e52-5f1a-4a2e-9a55-0a1f9d2b77c1", "state": "uploaded" }`,
			note: 'This is what starts the work.'
		},
		{
			method: 'POST',
			path: '/v1/assets',
			summary: 'Or point us at a file already online',
			body: `{ "url": "https://example.com/lecture.mp4" }`,
			response: `{ "asset_id": "0f4c1e52-5f1a-4a2e-9a55-0a1f9d2b77c1", "state": "uploaded" }`,
			note: 'We fetch it. Nothing to upload from your side.'
		},
		{
			method: 'GET',
			path: '/v1/assets/{id}',
			summary: 'State, sizes and the links to play it',
			body: null,
			response: `{
  "id": "0f4c1e52-5f1a-4a2e-9a55-0a1f9d2b77c1",
  "state": "partially_ready",
  "duration_seconds": 612.4,
  "width": 1920,
  "height": 1080,
  "renditions": [
    { "height": 1080, "codec": "h264", "bitrate_bps": 4200000,
      "state": "ready", "chunks_done": 41, "chunks_total": 41, "lazy": false },
    { "height": 480, "codec": "h264", "bitrate_bps": 900000,
      "state": "pending", "chunks_done": 0, "chunks_total": 0, "lazy": true }
  ],
  "playback": {
    "hls": "/playback/{tenant}/{asset}/master.m3u8?exp=…&kid=k1&sig=…",
    "dash": "/playback/{tenant}/{asset}/manifest.mpd?exp=…&kid=k1&sig=…",
    "poster": "/playback/{tenant}/{asset}/poster.jpg?exp=…&kid=k1&sig=…",
    "thumbnails": "/playback/{tenant}/{asset}/sprite.vtt?exp=…&kid=k1&sig=…"
  }
}`,
			note: 'partially_ready is playable. The low sizes are made on ingest and the rest on first play, so waiting for "ready" hangs forever on a video nobody has watched.'
		},
		{
			method: 'POST',
			path: '/v1/webhooks',
			summary: 'Be told instead of asking',
			body: `{ "url": "https://yoursite.example/hooks/alchemist", "events": ["asset.ready", "asset.failed"] }`,
			response: `{ "id": "8b1a…", "secret": "whsec_…" }`,
			note: 'Deliveries carry X-Alchemist-Signature: sha256=<hmac of the raw body>. Verify it before trusting the payload.'
		},
		{
			method: 'GET',
			path: '/v1/usage',
			summary: 'What you are being billed for',
			body: null,
			response: `{
  "from": "2026-09-01T00:00:00Z",
  "to": "2026-10-01T00:00:00Z",
  "lines": [
    { "kind": "ingest", "quantity": 24750, "unit": "seconds" }
  ]
}`,
			note: 'Only processing is metered so far, in seconds. Storage and delivery are not counted yet and no line appears for them — do not read their absence as zero.'
		}
	];

	const errorShape = `{
  "error": {
    "code": "quota_exceeded",
    "message": "You have reached this month's limit."
  }
}`;
</script>

<Seo title="API reference — Alchemist" description="How to talk to Alchemist from your own code." />

<h1 class="text-2xl font-semibold tracking-tight">API reference</h1>
<p class="mt-1 max-w-2xl text-sm text-dim">
	Everything here uses the key from <a href="/app/keys/" class="text-ink">API keys</a>,
	sent as <code class="font-mono text-xs">Authorization: Bearer …</code>. The base address is
	<code class="font-mono text-xs">{PUBLIC_ALCHEMIST_API}</code>.
</p>

<section class="card mt-6 p-6">
	<h2 class="text-sm font-semibold">The shape of a normal day</h2>
	<ol class="mt-3 grid gap-2 text-sm text-dim">
		<li>1. Ask for an upload target, or hand us a URL.</li>
		<li>2. We make the sizes your viewers need.</li>
		<li>3. Ask for the asset, and play the links it gives you.</li>
	</ol>
</section>

<div class="mt-8 grid gap-6">
	{#each endpoints as e (e.method + e.path)}
		<article class="card p-6">
			<header class="flex flex-wrap items-center gap-3">
				<!-- bg-secondary is not a token in this palette, so every POST badge drew
				     no background at all. -->
				<span
					class="rounded-lg px-2 py-0.5 font-mono text-xs font-semibold {e.method === 'GET'
						? 'bg-solid/15 text-ink'
						: 'bg-sunk text-dim'}"
				>
					{e.method}
				</span>
				<code class="font-mono text-sm">{e.path}</code>
			</header>
			<p class="mt-2 text-sm text-dim">{e.summary}</p>

			{#if e.body}
				<div class="mt-4">
					<Json source={e.body} label="Request" />
				</div>
			{/if}
			<div class="mt-3">
				<Json source={e.response} label="Response" />
			</div>
			{#if e.note}
				<p class="mt-3 text-xs text-dim">{e.note}</p>
			{/if}
		</article>
	{/each}
</div>

<section class="card mt-6 p-6">
	<h2 class="text-sm font-semibold">When something goes wrong</h2>
	<p class="mt-2 max-w-2xl text-sm text-dim">
		Errors are an object with a code and a message. Branch on the code: it is stable. The
		message is written for a person, may be reworded, and comes back in Bangla when the request
		sends <code class="font-mono text-xs">Accept-Language: bn</code>.
	</p>
	<div class="mt-4">
		<Json source={errorShape} label="Error" />
	</div>
</section>

<p class="mt-6 text-sm text-dim">
	Something here not matching what you get back? <a href="/contact/" class="link">Tell us</a> —
	the endpoints above are the ones this dashboard itself uses.
</p>
