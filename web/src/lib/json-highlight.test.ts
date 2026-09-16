import { tokenizeJson } from './json-highlight';

// Run with: npx vitest run  (or node --test after a build). Kept tiny on purpose:
// the only thing worth proving is that every character survives and keys are keys.
export function demo() {
	const src = '{"id":"a-1","count":42,"ok":true,"none":null,"nested":{"k":"v"}}';
	const tokens = tokenizeJson(src);

	const rebuilt = tokens.map((t) => t.text).join('');
	if (rebuilt !== src) throw new Error(`lost characters: ${rebuilt}`);

	const keys = tokens.filter((t) => t.kind === 'key').map((t) => t.text);
	const expected = ['"id"', '"count"', '"ok"', '"none"', '"nested"', '"k"'];
	if (keys.join(',') !== expected.join(',')) throw new Error(`keys: ${keys.join(',')}`);

	const strings = tokens.filter((t) => t.kind === 'string').map((t) => t.text);
	if (strings.join(',') !== '"a-1","v"') throw new Error(`strings: ${strings.join(',')}`);

	// A colon inside a string must not turn the next token into a key.
	const tricky = tokenizeJson('{"url":"https://x.test/a:b","n":1}');
	const trickyKeys = tricky.filter((t) => t.kind === 'key').map((t) => t.text);
	if (trickyKeys.join(',') !== '"url","n"') throw new Error(`tricky keys: ${trickyKeys.join(',')}`);

	return 'ok';
}
