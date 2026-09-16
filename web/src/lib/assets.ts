// One vocabulary for what a video is doing, shared by the table and the list.
// partially_ready is playable: the low sizes finish first and the rest are made on
// first play, so "in progress" would tell people to wait for something that may
// never happen on content nobody watches.
import type { Tone } from '$lib/components/Badge.svelte';

export const ASSET_STATE: Record<string, { chip: string; tone: Tone; done?: boolean }> = {
	created: { chip: 'Waiting', tone: 'idle' },
	uploading: { chip: 'Sending', tone: 'busy' },
	uploaded: { chip: 'Queued', tone: 'idle' },
	probing: { chip: 'Probing', tone: 'busy' },
	mezzanine: { chip: 'Prep', tone: 'busy' },
	analyzing: { chip: 'Prep', tone: 'busy' },
	encoding: { chip: 'Encoding', tone: 'busy' },
	packaging: { chip: 'Packing', tone: 'busy' },
	// Armed is not on air. Until a frame arrives there is nothing to watch.
	live_armed: { chip: 'Waiting', tone: 'idle' },
	live: { chip: 'On air', tone: 'good' },
	live_ended: { chip: 'Recorded', tone: 'good', done: true },
	partially_ready: { chip: 'Ready', tone: 'good', done: true },
	ready: { chip: 'Ready', tone: 'good', done: true },
	failed: { chip: 'Failed', tone: 'bad' }
};

export const bytes = (n: number | null | undefined) =>
	n == null
		? '—'
		: n >= 1_000_000_000
			? new Intl.NumberFormat('en', {
					style: 'unit',
					unit: 'gigabyte',
					maximumFractionDigits: 1
				}).format(n / 1_000_000_000)
			: new Intl.NumberFormat('en', {
					style: 'unit',
					unit: 'megabyte',
					maximumFractionDigits: n >= 10_000_000 ? 0 : 1
				}).format(n / 1_000_000);

export const clock = (secs: number | null | undefined) => {
	if (secs == null) return '—';
	const m = Math.floor(secs / 60);
	return `${m}:${String(Math.round(secs % 60)).padStart(2, '0')}`;
};

export const when = (iso: string) =>
	new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso));

// A video with no title is shown by the front of its id, which is what the customer
// sees in logs and URLs anyway.
export const assetName = (a: { title: string | null; id: string }) => a.title || a.id.slice(0, 8);
