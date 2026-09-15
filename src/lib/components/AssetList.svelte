<script lang="ts">
	import type { Asset } from '$lib/api';

	let { assets, loading }: { assets: Asset[]; loading: boolean } = $props();

	// partially_ready is playable: ingest returns the low rungs first and the rest are
	// made on first play. Calling it "in progress" would tell people to wait for
	// something that may never happen on content nobody watches.
	const STATE: Record<string, { chip: string; done?: boolean; bad?: boolean }> = {
		created: { chip: 'Waiting' },
		uploading: { chip: 'Sending' },
		uploaded: { chip: 'Queued' },
		probing: { chip: 'Probing' },
		mezzanine: { chip: 'Prep' },
		analyzing: { chip: 'Prep' },
		encoding: { chip: 'Encoding' },
		packaging: { chip: 'Packing' },
		partially_ready: { chip: 'Ready', done: true },
		ready: { chip: 'Ready', done: true },
		failed: { chip: 'Failed', bad: true }
	};

	const size = (bytes: number | null) =>
		bytes == null
			? null
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
		if (secs == null) return null;
		const m = Math.floor(secs / 60);
		const s = Math.round(secs % 60);
		return `${m}:${String(s).padStart(2, '0')}`;
	};

	const when = (iso: string) =>
		new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(
			new Date(iso)
		);

	const caption = (a: Asset) =>
		[length(a.duration_sec), size(a.source_bytes), when(a.created_at)].filter(Boolean).join(' · ');
</script>

{#if loading && assets.length === 0}
	<!-- The shape of the answer rather than a spinner: the first list arrives after an
	     upload and a queue, which is long enough to look broken. -->
	<div class="mt-2">
		{#each [132, 104, 158, 118, 146, 96] as w, i (i)}
			<div class="node">
				<span class="sk circle h-6 w-6 flex-none"></span>
				<span class="flex min-w-0 flex-1 items-baseline gap-3">
					<span class="sk h-3" style="width: {w}px"></span>
					<span class="sk h-2.5 w-20 opacity-60"></span>
				</span>
				<span class="sk ml-auto h-4 w-[54px] rounded-full"></span>
			</div>
		{/each}
	</div>
{:else if assets.length === 0}
	<div class="py-6 text-center">
		<p class="title">Nothing here yet</p>
		<p class="sub mx-auto mt-2 max-w-sm">
			Send your first video and watch it come back smaller. A ten-minute clip takes about a
			minute.
		</p>
		<a href="/app/upload/" class="btn-solid mt-5">Upload a video</a>
	</div>
{:else}
	<div class="mt-2">
		{#each assets as a (a.id)}
			{@const s = STATE[a.state] ?? { chip: a.state }}
			{@const pct = s.done ? 100 : s.bad ? 0 : 45}
			<a href="/app/videos/{a.id}/" class="node" class:opacity-55={s.done}>
				<span class={s.bad ? 'badge-hollow' : s.done ? 'badge-done' : 'badge'}>
					{s.bad ? '!' : s.done ? '✓' : '→'}
				</span>
				<span class="flex min-w-0 flex-1 items-baseline gap-3">
					<span class="title truncate font-mono">{a.id.slice(0, 8)}</span>
					<span class="sub truncate">{caption(a)}</span>
				</span>
				{#if !s.done && !s.bad}
					<span class="mini ml-auto hidden w-[120px] flex-none sm:block">
						<i
							class="block h-full rounded-full bg-ink transition-[width] duration-500"
							style="width: {pct}%"
						></i>
					</span>
				{/if}
				<span class="chip {s.done || s.bad ? '' : 'chip-on'} ml-auto sm:ml-0">{s.chip}</span>
			</a>
		{/each}
	</div>
{/if}
