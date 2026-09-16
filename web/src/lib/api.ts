import { PUBLIC_ALCHEMIST_API } from '$env/static/public';

// The API is the source of truth for what went wrong; this file only maps its stable
// codes to sentences a person can act on. Raw server text never reaches the page.
const MESSAGES: Record<string, string> = {
	invalid_request: 'Something in that form did not come through. Try again.',
	invalid_org: 'Tell us what to call your organisation.',
	invalid_email: 'That email address does not look right.',
	weak_password: 'Use at least 10 characters.',
	email_taken: 'There is already an account with that email.',
	invalid_credentials: 'That email and password do not match.',
	no_session: 'You are not signed in.',
	too_many_attempts: 'Too many tries. Wait a minute, then have another go.',
	internal_error: 'Something went wrong on our side. Please try again.',
	offline: 'We could not reach Alchemist. Check your connection and try again.',
	upload_failed: 'The upload did not finish. Try it again.',
	not_found: 'We could not find that.',
	invalid_message: 'Tell us a little more — a sentence or two is plenty.',
	asset_not_found: 'We could not find that video.',
	quota_exceeded: 'You have reached this month\u2019s limit. Upgrade or wait for the reset.',
	session_required: 'Sign in to change this. An API key cannot.',
	not_permitted: 'Only an owner or an admin can change who is on the team.',
	invalid_role: 'Pick admin or member. Only an owner can add another owner.',
	already_a_member: 'That person is already on the team.',
	last_owner: 'Make someone else an owner first. An account cannot be left without one.',
	cannot_remove_self: 'You cannot remove yourself. Ask another owner to do it.',
	invite_not_found: 'That invite has expired or has already been used. Ask for a new one.',
	invalid_image: 'Send a PNG, JPEG or WebP image.',
	image_too_large: 'That image is over 1 MB. Try a smaller one.',
	invalid_url: 'That does not look like a valid https address.',
	unknown_event: 'We do not send an event by that name.',
	invalid_date: 'Use a real date.',
	invalid_range: 'The end of the period has to come after the start.',
	unknown_profile: 'We do not have an encoding preset by that name.',
	invalid_edit: 'That edit does not describe a video we can make.',
	not_ready: 'That video is still being processed. Editing opens once it is ready.',
	unknown_provider:
		'We cannot move a library from that service. Some hosts never hand back the original file.',
	invalid_state: 'That has already moved on. Reload the page to see where it is now.',
	live_not_enabled: 'Live is not on this plan. Talk to us and we will switch it on.',
	invalid_protocol: 'Pick your camera, SRT or RTMP.',
	stream_not_found: 'We could not find that stream.',
	stream_busy: 'That stream is already waiting for an encoder.',
	not_broadcasting: 'That stream is not on air, so there is nothing to stop.'
};

export class ApiError extends Error {
	code: string;
	constructor(code: string) {
		super(MESSAGES[code] ?? MESSAGES.internal_error);
		this.code = code;
	}
}

async function call<T>(path: string, init: RequestInit = {}): Promise<T> {
	if (!PUBLIC_ALCHEMIST_API) {
		// A missing base URL is a build mistake, not a user error, and silently
		// posting to the site's own origin would look like a server fault instead.
		throw new Error('PUBLIC_ALCHEMIST_API is not set');
	}
	let res: Response;
	try {
		res = await fetch(PUBLIC_ALCHEMIST_API + path, {
			...init,
			credentials: 'include',
			headers: { 'Content-Type': 'application/json', ...(init.headers ?? {}) }
		});
	} catch {
		throw new ApiError('offline');
	}
	if (res.status === 204) return undefined as T;
	const body = await res.json().catch(() => null);
	if (!res.ok) throw new ApiError(body?.error?.code ?? 'internal_error');
	return body as T;
}

export type SignupResult = { tenant_id: string; org: string; email: string; api_key: string };
export type Session = { user_id: string; tenant_id: string; email: string; org: string };

export const signup = (org: string, email: string, password: string) =>
	call<SignupResult>('/v1/auth/signup', {
		method: 'POST',
		body: JSON.stringify({ org, email, password })
	});

export const login = (email: string, password: string) =>
	call<{ tenant_id: string; email: string }>('/v1/auth/login', {
		method: 'POST',
		body: JSON.stringify({ email, password })
	});

export const logout = () => call<void>('/v1/auth/logout', { method: 'POST' });

export type ContactRequest = {
	name: string;
	email: string;
	org: string;
	message: string;
	website: string;
};
export const contact = (body: ContactRequest) =>
	call<{ status: string }>('/v1/contact', { method: 'POST', body: JSON.stringify(body) });
export const session = () => call<Session>('/v1/auth/session');

export type Asset = {
	id: string;
	state: string;
	error_code: string | null;
	title: string | null;
	duration_sec: number | null;
	source_bytes: number | null;
	height: number | null;
	created_at: string;
};
export type ApiKey = {
	id: string;
	name: string;
	created_at: string;
	revoked_at: string | null;
};
export type UsageLine = { kind: string; quantity: number; unit: string };

// Every list endpoint takes the same five and answers with a total, so one helper
// builds the query for all of them and no page invents its own parameter names.
export type ListQuery = {
	limit?: number;
	offset?: number;
	q?: string;
	sort?: string;
	order?: 'asc' | 'desc';
	filters?: Record<string, string | string[] | undefined>;
};

export const listParams = (l: ListQuery = {}) => {
	const p = new URLSearchParams();
	if (l.limit != null) p.set('limit', String(l.limit));
	if (l.offset) p.set('offset', String(l.offset));
	if (l.q?.trim()) p.set('q', l.q.trim());
	if (l.sort) p.set('sort', l.sort);
	if (l.order) p.set('order', l.order);
	for (const [k, v] of Object.entries(l.filters ?? {})) {
		if (v == null || v === '') continue;
		for (const one of Array.isArray(v) ? v : [v]) p.append(k, one);
	}
	const s = p.toString();
	return s ? `?${s}` : '';
};

export const listAssets = (l: ListQuery = {}) =>
	call<{ assets: Asset[]; total: number }>(`/v1/assets${listParams(l)}`);
export const deleteAsset = (id: string) => call<void>(`/v1/assets/${id}`, { method: 'DELETE' });
export const patchAsset = (id: string, title: string) =>
	call<Asset>(`/v1/assets/${id}`, { method: 'PATCH', body: JSON.stringify({ title }) });
export const listKeys = (l: ListQuery = {}) =>
	call<{ keys: ApiKey[]; total: number }>(`/v1/keys${listParams(l)}`);
export const renameKey = (id: string, name: string) =>
	call<ApiKey>(`/v1/keys/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) });
export const createKey = (name: string) =>
	call<{ id: string; name: string; api_key: string }>('/v1/keys', {
		method: 'POST',
		body: JSON.stringify({ name })
	});
export const revokeKey = (id: string) => call<void>(`/v1/keys/${id}`, { method: 'DELETE' });
export const usage = (from?: string, to?: string) => {
	const q = new URLSearchParams();
	if (from) q.set('from', from);
	if (to) q.set('to', to);
	const qs = q.toString();
	return call<{ from: string; to: string; lines: UsageLine[] }>(
		'/v1/usage' + (qs ? '?' + qs : '')
	);
};

export const startUpload = () =>
	call<{ asset_id: string; upload_url: string; expires_in_seconds: number }>('/v1/uploads', {
		method: 'POST'
	});
export const completeUpload = (id: string) =>
	call<{ asset_id: string; state: string }>(`/v1/assets/${id}/complete`, { method: 'POST' });
export const importFromUrl = (url: string) =>
	call<{ asset_id: string; state: string }>('/v1/assets', {
		method: 'POST',
		body: JSON.stringify({ url })
	});

// The browser PUTs the bytes straight to storage, so they never pass through the
// API — and so progress is the only thing we can report while it happens.
export function putFile(url: string, file: File, onProgress: (pct: number) => void) {
	return new Promise<void>((resolve, reject) => {
		const xhr = new XMLHttpRequest();
		xhr.open('PUT', url);
		xhr.upload.onprogress = (e) => {
			if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100));
		};
		xhr.onload = () =>
			xhr.status >= 200 && xhr.status < 300
				? resolve()
				: reject(new ApiError('upload_failed'));
		xhr.onerror = () => reject(new ApiError('upload_failed'));
		xhr.send(file);
	});
}

export type Rendition = {
	height: number;
	codec: string;
	bitrate_bps: number;
	state: string;
	chunks_done: number;
	chunks_total: number;
	bytes: number | null;
	lazy: boolean;
};
export type AssetDetail = {
	id: string;
	state: string;
	error_code?: string;
	duration_seconds?: number;
	width?: number;
	height?: number;
	source_bytes?: number;
	created_at?: string;
	renditions: Rendition[];
	playback?: Playback;
};

export type Playback = {
	hls: string;
	dash: string;
	poster: string;
	thumbnails: string;
	// Encrypted media is packaged cenc and HLS cannot carry cenc, so the API says
	// which of the two to hand a player rather than the page guessing.
	preferred: 'hls' | 'dash';
	encrypted: boolean;
};

export const getAsset = (id: string) => call<AssetDetail>(`/v1/assets/${id}`);

export type Chunk = {
	rendition: string;
	index: number;
	start_sec: number;
	end_sec: number;
	done: boolean;
};
export type Activity = {
	step: string;
	state: string;
	attempt: number;
	failures: number;
	queued_at: string;
	started_at: string | null;
	finished_at: string | null;
};

export const getChunks = (id: string) => call<{ chunks: Chunk[] }>(`/v1/assets/${id}/chunks`);
export const getActivity = (id: string) => call<{ activity: Activity[] }>(`/v1/assets/${id}/activity`);

export type Member = {
	id: string;
	email: string;
	role: 'owner' | 'admin' | 'member';
	created_at: string;
	last_login_at: string | null;
	you: boolean;
};
export type Invite = {
	id: string;
	email: string;
	role: string;
	created_at: string;
	expires_at: string;
};

export const listMembers = (l: ListQuery = {}) =>
	call<{ members: Member[]; invites: Invite[]; total: number }>(`/v1/members${listParams(l)}`);
export const inviteMember = (email: string, role: string) =>
	call<{ id: string; email: string; role: string; token: string; expires_at: string }>(
		'/v1/members/invites',
		{ method: 'POST', body: JSON.stringify({ email, role }) }
	);
export const withdrawInvite = (id: string) =>
	call<void>(`/v1/members/invites/${id}`, { method: 'DELETE' });
export const setMemberRole = (id: string, role: string) =>
	call<{ id: string; role: string }>(`/v1/members/${id}`, {
		method: 'PATCH',
		body: JSON.stringify({ role })
	});
export const removeMember = (id: string) => call<void>(`/v1/members/${id}`, { method: 'DELETE' });

// Redeeming an invite happens before there is an account, so neither of these is
// behind a session.
export const previewInvite = (token: string) =>
	call<{ email: string; role: string; org: string }>(
		`/v1/auth/invite?token=${encodeURIComponent(token)}`
	);
export const acceptInvite = (token: string, password: string) =>
	call<{ tenant_id: string; email: string }>('/v1/auth/invite', {
		method: 'POST',
		body: JSON.stringify({ token, password })
	});

export type Branding = { name: string; logo_url: string | null; logo_updated_at: string | null };

export const getBranding = () => call<Branding>('/v1/branding');
export const deleteLogo = () => call<void>('/v1/branding/logo', { method: 'DELETE' });

// The logo goes through the API rather than a presigned PUT: it is under a megabyte,
// and the round trip to get a target costs more than the upload saves.
export const putLogo = (file: File) =>
	call<{ logo_url: string; logo_updated_at: string }>('/v1/branding/logo', {
		method: 'PUT',
		body: file,
		headers: { 'Content-Type': file.type }
	});

export type Webhook = { id: string; url: string; events: string[]; active: boolean };
export const WEBHOOK_EVENTS = ['asset.ready', 'asset.failed', 'rendition.ready'] as const;

export const listWebhooks = (l: ListQuery = {}) =>
	call<{ webhooks: Webhook[]; total: number }>(`/v1/webhooks${listParams(l)}`);
export const getWebhook = (id: string) => call<Webhook>(`/v1/webhooks/${id}`);
export const patchWebhook = (id: string, body: { url?: string; active?: boolean }) =>
	call<Webhook>(`/v1/webhooks/${id}`, { method: 'PATCH', body: JSON.stringify(body) });
export const deleteWebhook = (id: string) =>
	call<void>(`/v1/webhooks/${id}`, { method: 'DELETE' });
export const createWebhook = (url: string, events: string[]) =>
	call<{ id: string; url: string; secret: string }>('/v1/webhooks', {
		method: 'POST',
		body: JSON.stringify({ url, events })
	});

export type BucketSource = {
	id: string;
	bucket: string;
	prefix: string;
	active: boolean;
	last_synced_at: string | null;
	last_error: string | null;
	imported_objects: number;
};
export type NewBucketSource = {
	endpoint: string;
	region: string;
	bucket: string;
	prefix: string;
	access_key_id: string;
	secret_access_key: string;
};

export const listBucketSources = (l: ListQuery = {}) =>
	call<{ bucket_sources: BucketSource[]; total: number }>(`/v1/bucket-sources${listParams(l)}`);
export const patchBucketSource = (id: string, body: { active?: boolean }) =>
	call<BucketSource>(`/v1/bucket-sources/${id}`, { method: 'PATCH', body: JSON.stringify(body) });
export const deleteBucketSource = (id: string) =>
	call<void>(`/v1/bucket-sources/${id}`, { method: 'DELETE' });
export const connectBucket = (body: NewBucketSource) =>
	call<{ id: string }>('/v1/bucket-sources', { method: 'POST', body: JSON.stringify(body) });

export type Whoami = {
	tenant_id: string;
	name: string;
	ladder_profile: string;
	live_enabled: boolean;
};
export const whoami = () => call<Whoami>('/v1/whoami');

export type LadderRung = { height: number; codec: string; maxrate_bps: number; lazy: boolean };
export type LadderProfile = {
	name: string;
	description: string;
	rungs: LadderRung[];
	current: boolean;
};

export const listProfiles = () => call<{ profiles: LadderProfile[] }>('/v1/ladder-profiles');
export const setProfile = (profile: string) =>
	call<{ profile: string }>('/v1/ladder-profile', {
		method: 'PUT',
		body: JSON.stringify({ profile })
	});

export type MigrationProvider = {
	name: string;
	label: string;
	secret: { label: string; hint: string; kind: 'key' | 'text' };
	config: { key: string; label: string; hint: string }[];
};
export type Migration = {
	id: string;
	provider: string;
	state: 'previewing' | 'scanning' | 'done' | 'failed' | 'paused';
	preview_done: boolean;
	total: number;
	handled: number;
	imported: number;
	skipped: number;
	last_error: string | null;
	created_at: string;
};
export type MigrationItem = {
	remote_id: string;
	title: string;
	asset_id: string | null;
	state: 'pending' | 'imported' | 'skipped' | 'failed';
	reason: string | null;
};

export const listMigrationProviders = () =>
	call<{ providers: MigrationProvider[] }>('/v1/migration-providers');
export const listMigrations = (l: ListQuery = {}) =>
	call<{ migrations: Migration[]; total: number }>(`/v1/migrations${listParams(l)}`);
export const listMigrationItems = (id: string) =>
	call<{ items: MigrationItem[] }>(`/v1/migrations/${id}/items`);
export const startMigration = (provider: string, secret: string, config: Record<string, string>) =>
	call<{ id: string; provider: string; state: string }>('/v1/migrations', {
		method: 'POST',
		body: JSON.stringify({ provider, secret, config })
	});
export const confirmMigration = (id: string) =>
	call<{ id: string; state: string }>(`/v1/migrations/${id}/confirm`, { method: 'POST' });
export const pauseMigration = (id: string) =>
	call<{ id: string; state: string }>(`/v1/migrations/${id}/pause`, { method: 'POST' });
export const resumeMigration = (id: string) =>
	call<{ id: string; state: string }>(`/v1/migrations/${id}/resume`, { method: 'POST' });
export const deleteMigration = (id: string) =>
	call<void>(`/v1/migrations/${id}`, { method: 'DELETE' });

export type EditOverlay = {
	kind: 'text' | 'image';
	text?: string;
	colour?: string;
	shadow?: boolean;
	/** Names what to put on, never where it is. 'logo' is this account's own logo. */
	source?: 'logo';
	at?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'centre' | 'free';
	/** Top-left as fractions of the frame. Read only when `at` is 'free'. */
	x?: number;
	y?: number;
	scale?: number;
	margin?: number;
	opacity?: number;
	start_sec?: number;
	end_sec?: number;
};

export type EditOps = {
	start_sec?: number;
	end_sec?: number;
	crop_x?: number;
	crop_y?: number;
	crop_w?: number;
	crop_h?: number;
	aspect?: string;
	width?: number;
	mute?: boolean;
	overlays?: EditOverlay[];
};
export type Edit = {
	id: string;
	source_asset_id: string;
	output_asset_id: string | null;
	state: 'queued' | 'rendering' | 'done' | 'failed';
	error_code: string | null;
	ops: EditOps;
	created_at: string;
};

export const listEdits = (assetId?: string) =>
	call<{ edits: Edit[] }>('/v1/edits' + (assetId ? `?asset_id=${assetId}` : ''));
export const createEdit = (assetId: string, ops: EditOps) =>
	call<{ id: string; state: string }>('/v1/edits', {
		method: 'POST',
		body: JSON.stringify({ asset_id: assetId, ops })
	});
export const deleteEdit = (id: string) => call<void>(`/v1/edits/${id}`, { method: 'DELETE' });

// 'camera' is what the customer publishes from, not what goes over the wire: a
// browser webcam rather than an encoder they would have to install.
export type LiveSource = 'camera' | 'srt' | 'rtmp';

export type LiveStream = {
	id: string;
	name: string;
	protocol: LiveSource;
	state: 'idle' | 'armed' | 'live' | 'ended';
	asset_id?: string;
	created_at: string;
};

export const listLiveStreams = (l: ListQuery = {}) =>
	call<{ live_streams: LiveStream[]; total: number }>(`/v1/live-streams${listParams(l)}`);
export const renameLiveStream = (id: string, name: string) =>
	call<LiveStream>(`/v1/live-streams/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) });
export const getLiveStream = (id: string) => call<LiveStream>(`/v1/live-streams/${id}`);

// The key comes back on this one call and is never returned again, exactly like an
// API key.
export const createLiveStream = (name: string, protocol: LiveSource) =>
	call<LiveStream & { stream_key: string }>('/v1/live-streams', {
		method: 'POST',
		body: JSON.stringify({ name, protocol })
	});

// publish_token comes back only for a camera stream, because the browser is the
// encoder and has no key to paste. It is minted for this broadcast and no other.
export const startLiveStream = (id: string) =>
	call<{
		stream_id: string;
		session_id: string;
		asset_id: string;
		ingest_url: string;
		publish_token?: string;
	}>(`/v1/live-streams/${id}/start`, { method: 'POST' });

// Ending a broadcast on demand. The worker holds the encoder in another process, so
// this records the intent and returns; the broadcast is over within a second or two.
export const stopLiveStream = (id: string) =>
	call<{ stream_id: string; session_id: string }>(`/v1/live-streams/${id}/stop`, {
		method: 'POST'
	});

export const deleteLiveStream = (id: string) =>
	call<void>(`/v1/live-streams/${id}`, { method: 'DELETE' });

// Also shown once. There is no reading an existing key back, so this is what a
// customer who did not write theirs down actually needs.
export const replaceLiveKey = (id: string) =>
	call<{ stream_id: string; stream_key: string; state: string }>(
		`/v1/live-streams/${id}/key`,
		{ method: 'POST' }
	);
