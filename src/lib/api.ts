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
	quota_exceeded: 'You have reached this month\u2019s limit. Upgrade or wait for the reset.'
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

export const listAssets = () => call<{ assets: Asset[] }>('/v1/assets');
export const listKeys = () => call<{ keys: ApiKey[] }>('/v1/keys');
export const createKey = (name: string) =>
	call<{ id: string; name: string; api_key: string }>('/v1/keys', {
		method: 'POST',
		body: JSON.stringify({ name })
	});
export const revokeKey = (id: string) => call<void>(`/v1/keys/${id}`, { method: 'DELETE' });
export const usage = () => call<{ from: string; to: string; lines: UsageLine[] }>('/v1/usage');

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
	playback?: { hls: string; dash: string; poster: string; thumbnails: string };
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
