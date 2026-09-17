import { error, json, type RequestHandler } from '@sveltejs/kit';
import { generateTranscription } from '@tanstack/ai';
import { transcriberFor } from '$lib/server/ai';

// Speech to subtitles, as a draft.
//
// The audio arrives as a URL rather than bytes: this server holds no session for the
// engine and reads nothing from storage, so the caller passes a signed playback URL it
// already has. That keeps the rule the dashboard is built on -- the web app talks to
// the engine through the public API and nothing else -- and means no audio is stored here.
export const prerender = false;

// Whisper's own limit is 25 MB. An hour of the shared 96k audio track is roughly 43 MB,
// so a long lecture has to be split before it can be sent; that is refused plainly
// rather than truncated into a transcript that stops halfway with no explanation.
const MAX_AUDIO_BYTES = 25 * 1024 * 1024;

export const POST: RequestHandler = async ({ request, fetch }) => {
	let body: { audio_url?: string; language?: string };
	try {
		body = await request.json();
	} catch {
		throw error(400, 'Send a JSON body with audio_url.');
	}
	if (!body.audio_url) throw error(400, 'Include audio_url.');

	let url: URL;
	try {
		url = new URL(body.audio_url);
	} catch {
		throw error(400, 'audio_url is not a URL.');
	}
	if (url.protocol !== 'https:' && url.hostname !== 'localhost' && url.hostname !== '127.0.0.1') {
		// Anything else makes this endpoint a way to have our server fetch arbitrary
		// addresses on somebody else's behalf.
		throw error(400, 'audio_url has to be https.');
	}

	let adapter;
	try {
		adapter = await transcriberFor();
	} catch (e) {
		// A missing key is a deployment mistake, not something the customer did.
		throw error(503, e instanceof Error ? e.message : 'Transcription is not configured here.');
	}

	const res = await fetch(url, { headers: { Accept: 'audio/*' } });
	if (!res.ok) throw error(502, 'That audio could not be fetched.');
	const audio = await res.arrayBuffer();
	if (audio.byteLength === 0) throw error(502, 'That audio was empty.');
	if (audio.byteLength > MAX_AUDIO_BYTES) {
		throw error(413, 'That audio is over 25 MB, which is as much as one request can carry.');
	}

	try {
		const result = await generateTranscription({
			adapter,
			audio,
			language: body.language || undefined,
			// WebVTT straight out, which is exactly what PUT /v1/assets/{id}/captions
			// takes. Asking for json and assembling cues here would be a second, worse
			// implementation of a format the model already writes.
			responseFormat: 'vtt'
		});
		const vtt = typeof result === 'string' ? result : ((result as { text?: string })?.text ?? '');
		if (!vtt.trimStart().startsWith('WEBVTT')) {
			// The captions endpoint refuses anything that does not, so catching it here
			// names the real cause rather than letting it look like a bad upload.
			throw error(502, 'The model did not return WebVTT. Check TRANSCRIBE_MODEL.');
		}
		// Never stored here and never uploaded from here. The customer reviews it and
		// PUTs it to the captions endpoint themselves: a machine transcript going live
		// unread is how a lecture ends up captioned with the wrong terms.
		return json({ vtt, language: body.language ?? null, draft: true });
	} catch (e) {
		if (e && typeof e === 'object' && 'status' in e) throw e;
		throw error(502, e instanceof Error ? e.message : 'Transcription failed.');
	}
};
