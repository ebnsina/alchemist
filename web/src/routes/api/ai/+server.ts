import { error, type RequestHandler } from '@sveltejs/kit';
import { chat, toServerSentEventsResponse } from '@tanstack/ai';
import { adapterFor, MODELS, type ModelKind } from '$lib/server/ai';

// The one route on this site that runs per request, and the only reason there is a
// server at all. The provider key never leaves this process.
export const prerender = false;

export const POST: RequestHandler = async ({ request }) => {
	let body: { kind?: ModelKind; messages?: unknown };
	try {
		body = await request.json();
	} catch {
		throw error(400, 'Send a JSON body.');
	}

	const kind = (body.kind ?? 'chat') as ModelKind;
	const task = MODELS[kind];
	if (!task) throw error(400, 'Unknown task.');
	if (!Array.isArray(body.messages) || body.messages.length === 0) {
		throw error(400, 'Include at least one message.');
	}

	let adapter;
	try {
		adapter = await adapterFor(kind);
	} catch (e) {
		// A missing key is a deployment mistake, not something the visitor did, so it
		// says so rather than reading as a model failure.
		throw error(503, e instanceof Error ? e.message : 'AI is not configured here.');
	}

	// The controller is passed to both: a reader that goes away should stop the
	// generation rather than leaving it running and still being billed for it.
	const abortController = new AbortController();
	const stream = chat({
		adapter,
		messages: body.messages,
		systemPrompts: [task.system],
		abortController
	});

	return toServerSentEventsResponse(stream, { abortController });
};
