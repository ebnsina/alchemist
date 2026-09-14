<script lang="ts">
	import type { Asset } from '$lib/api';

	let { assets, loading }: { assets: Asset[]; loading: boolean } = $props();

	// partially_ready is playable: ingest returns the low rungs first and the rest are
	// made on first play. Calling it "in progress" would tell people to wait for
	// something that may never happen on content nobody watches.
	const STATE: Record<string, { label: string; tone: 'ok' | 'work' | 'bad' }> = {
		created: { label: 'Waiting for the file', tone: 'work' },
		uploading: { label: 'Uploading', tone: 'work' },
		uploaded: { label: 'Queued', tone: 'work' },
		probing: { label: 'Looking it over', tone: 'work' },
		mezzanine: { label: 'Preparing', tone: 'work' },
		analyzing: { label: 'Preparing', tone: 'work' },
		encoding: { label: 'Making the sizes', tone: 'work' },
		packaging: { label: 'Almost there', tone: 'work' },
		partially_ready: { label: 'Ready to watch', tone: 'ok' },
		ready: { label: 'Ready to watch', tone: 'ok' },
		failed: { label: 'Did not work', tone: 'bad' }
	};

	const size = (bytes: number | null) =>
		bytes == null
			? '—'
			: bytes >= 1_000_000_000
				? new Intl.NumberFormat('en', {
						style: 'unit',
						unit: 'gigabyte',
						maximumFractionDigits: 1
					}).format(bytes / 1_000_000_000)
				: new Intl.NumberFormat('en', {
						style: 'unit',
						unit: 'megabyte',
						maximumFractionDigits: 0
					}).format(bytes / 1_000_000);

	const length = (secs: number | null) => {
		if (secs == null) return '—';
		const m = Math.floor(secs / 60);
		const s = Math.round(secs % 60);
		return `${m}:${String(s).padStart(2, '0')}`;
	};

	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(
			new Date(iso)
		);
</script>

{#if loading && assets.length === 0}
	<p class="mt-4 text-sm text-muted">Loading…</p>
{:else if assets.length === 0}
	<div class="card mt-4 p-8 text-center">
		<p class="font-semibold">Nothing here yet</p>
		<p class="mx-auto mt-2 max-w-sm text-sm text-muted">
			Send your first video and watch it come back smaller. It takes about a minute for a
			ten-minute clip.
		</p>
		<a href="/app/upload/" class="btn-primary mt-5">Upload a video</a>
	</div>
{:else}
	<div class="card mt-4 overflow-x-auto">
		<table class="w-full text-sm">
			<caption class="vh">Your videos, newest first</caption>
			<thead>
				<tr class="text-xs text-muted">
					<th class="px-5 py-3 text-left font-normal">Video</th>
					<th class="px-5 py-3 text-left font-normal">State</th>
					<th class="px-5 py-3 text-right font-normal">Length</th>
					<th class="px-5 py-3 text-right font-normal">Sent</th>
					<th class="px-5 py-3 text-right font-normal">Added</th>
				</tr>
			</thead>
			<tbody>
				{#each assets as a (a.id)}
					{@const state = STATE[a.state] ?? { label: a.state, tone: 'work' }}
					<tr class="border-t border-hairline">
						<td class="px-5 py-3">
							<code class="font-mono text-xs text-muted">{a.id.slice(0, 8)}</code>
						</td>
						<td class="px-5 py-3">
							<span class="inline-flex items-center gap-2">
								<span
									class="h-1.5 w-1.5 flex-none rounded-full {state.tone === 'ok'
										? 'bg-brand-mid'
										: state.tone === 'bad'
											? 'bg-[#f87171]'
											: 'bg-gold'}"
								></span>
								{state.label}
							</span>
							{#if a.error_code}
								<span class="block text-xs text-muted">{a.error_code}</span>
							{/if}
						</td>
						<td class="px-5 py-3 text-right tabular-nums">{length(a.duration_sec)}</td>
						<td class="px-5 py-3 text-right tabular-nums">{size(a.source_bytes)}</td>
						<td class="px-5 py-3 text-right text-muted">{when(a.created_at)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
