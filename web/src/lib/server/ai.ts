import { env } from '$env/dynamic/private';

// Provider-agnostic on purpose: which model answers is a deployment decision, not
// something baked into a page. Nothing here reaches the browser — SvelteKit refuses
// to bundle $lib/server into client code, which is the whole reason the key is safe.

export type ModelKind = 'chat' | 'title' | 'chapters' | 'captions';

// Tasks name what they are for rather than a model, so a cheaper model can be put
// behind the small jobs later without touching a caller.
export const MODELS: Record<ModelKind, { purpose: string; system: string }> = {
	chat: {
		purpose: 'Answering questions about a video library',
		system: 'You help someone understand their own video library. Be brief and concrete. If you do not know, say so.'
	},
	title: {
		purpose: 'Suggesting a title and description',
		system: 'Suggest one plain title and a two-sentence description. No marketing language, no emoji.'
	},
	chapters: {
		purpose: 'Proposing chapters from a transcript',
		system: 'Propose chapter markers as timestamp and short label. Only use moments present in the transcript.'
	},
	captions: {
		purpose: 'Cleaning up a machine transcript',
		system: 'Fix punctuation and obvious transcription errors. Never invent words that were not spoken.'
	}
};

type Provider = 'anthropic' | 'openai' | 'gemini' | 'ollama';

// Each adapter reads its own key from the environment and deliberately refuses to
// take one as an argument, so a key cannot be captured in application code by
// accident. We only check it is there, and name the variable when it is not.
const KEY_VAR: Record<Provider, string> = {
	anthropic: 'ANTHROPIC_API_KEY',
	openai: 'OPENAI_API_KEY',
	gemini: 'GEMINI_API_KEY',
	// Ollama runs on your own machine and authenticates with nothing. The empty name
	// is what says "this one needs no key" without a second flag to forget.
	ollama: ''
};

function configured(): { provider: Provider; model: string } {
	const provider = (env.AI_PROVIDER ?? '') as Provider;
	if (!provider) throw new Error('AI_PROVIDER is not set. Nothing here guesses a provider.');
	if (!KEY_VAR[provider]) {
		throw new Error(`AI_PROVIDER=${provider} is not one this deployment has an adapter for.`);
	}
	const model = env.AI_MODEL ?? '';
	if (!model) throw new Error('AI_MODEL is not set.');
	const keyVar = KEY_VAR[provider];
	if (keyVar && !env[keyVar]) {
		throw new Error(`${keyVar} is not set, so ${provider} cannot be reached.`);
	}
	return { provider, model };
}

/**
 * adapterFor returns the configured provider's adapter. The kind is taken now so a
 * per-task model can be introduced later without changing a single caller.
 *
 * Adding a provider is one case plus one package: TanStack AI ships adapters for
 * Gemini and Ollama too, and they are deliberately not installed until wanted rather
 * than carried as dead weight.
 */
export async function adapterFor(_kind: ModelKind) {
	const { provider, model } = configured();

	switch (provider) {
		// The adapters type their model argument as a literal union of the models they
		// know. AI_MODEL is a string from the environment, so it is narrowed to that
		// union here: an unknown name then fails at the provider with its own message
		// rather than failing to compile every time a model is released.
		case 'anthropic': {
			const { anthropicText } = await import('@tanstack/ai-anthropic');
			return anthropicText(model as Parameters<typeof anthropicText>[0]);
		}
		case 'openai': {
			const { openaiText } = await import('@tanstack/ai-openai');
			return openaiText(model as Parameters<typeof openaiText>[0]);
		}
		case 'gemini': {
			const { geminiText } = await import('@tanstack/ai-gemini');
			return geminiText(model as Parameters<typeof geminiText>[0]);
		}
		case 'ollama': {
			// Self-hosted, which is this project's preference wherever a solid option
			// exists: nothing leaves the machine and there is no per-token bill.
			const { ollamaText } = await import('@tanstack/ai-ollama');
			return ollamaText(model as Parameters<typeof ollamaText>[0]);
		}
		default:
			throw new Error(
				`AI_PROVIDER=${provider} is not installed. Add @tanstack/ai-${provider} and a case here.`
			);
	}
}

/** Whether AI is usable at all, so the UI can hide what would only 503. */
export function aiEnabled(): boolean {
	try {
		configured();
		return true;
	} catch {
		return false;
	}
}

/**
 * Transcription is its own provider choice, because it is its own market.
 *
 * TanStack AI ships a transcription adapter for OpenAI and not for the others: Gemini
 * and Ollama's audio support is speech *generation*, which is the opposite direction.
 * So rather than pretend there are four options, this reads its own variables and
 * leans on the fact that whisper's API shape is the one everybody copied.
 *
 * TRANSCRIBE_BASE_URL is the whole self-hosted story. faster-whisper-server,
 * whisper.cpp's server and several others expose an OpenAI-compatible
 * /v1/audio/transcriptions, so pointing this at one keeps the audio on your own
 * machine and costs nothing per minute. Unset, it goes to OpenAI.
 */
export function transcriptionConfigured(): { model: string; baseURL?: string } {
	const model = env.TRANSCRIBE_MODEL ?? '';
	if (!model) throw new Error('TRANSCRIBE_MODEL is not set. Nothing here guesses a model.');
	const baseURL = env.TRANSCRIBE_BASE_URL || undefined;
	// A self-hosted server usually wants no key at all, so one is only insisted on
	// when the request is going to OpenAI.
	if (!baseURL && !env.OPENAI_API_KEY) {
		throw new Error('OPENAI_API_KEY is not set, and no TRANSCRIBE_BASE_URL was given.');
	}
	return { model, baseURL };
}

export async function transcriberFor() {
	const { model, baseURL } = transcriptionConfigured();
	const { createOpenaiTranscription } = await import('@tanstack/ai-openai');
	// A local server is given a placeholder rather than nothing: the client insists on
	// a key it will then send to a server that ignores it.
	const apiKey = env.OPENAI_API_KEY || 'self-hosted';
	return createOpenaiTranscription(
		model as Parameters<typeof createOpenaiTranscription>[0],
		apiKey,
		baseURL ? { baseURL } : undefined
	);
}

/** Whether transcription is usable, so the dashboard can hide a button that would 503. */
export function transcriptionEnabled(): boolean {
	try {
		transcriptionConfigured();
		return true;
	} catch {
		return false;
	}
}
