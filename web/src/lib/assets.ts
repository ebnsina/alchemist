// One vocabulary for what a video is doing, shared by the table and the list.
// partially_ready is playable: the low sizes finish first and the rest are made on
// first play, so "in progress" would tell people to wait for something that may
// never happen on content nobody watches.
export const ASSET_STATE: Record<string, { chip: string; done?: boolean; bad?: boolean }> = {
	created: { chip: 'Waiting' },
	uploading: { chip: 'Sending' },
	uploaded: { chip: 'Queued' },
	probing: { chip: 'Probing' },
	mezzanine: { chip: 'Prep' },
	analyzing: { chip: 'Prep' },
	encoding: { chip: 'Encoding' },
	packaging: { chip: 'Packing' },
	// Armed is not on air. Until a frame arrives there is nothing to watch.
	live_armed: { chip: 'Waiting' },
	live: { chip: 'On air' },
	live_ended: { chip: 'Recorded', done: true },
	partially_ready: { chip: 'Ready', done: true },
	ready: { chip: 'Ready', done: true },
	failed: { chip: 'Failed', bad: true }
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
