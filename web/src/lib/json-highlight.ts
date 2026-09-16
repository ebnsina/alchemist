// A JSON highlighter, because pulling in a syntax-highlighting library to color
// four token types would be more bytes than the dashboard itself.
//
// Returns tokens rather than HTML: the component renders them, so nothing here has
// to escape anything and no markup can leak from the data.
export type Token = { text: string; kind: 'key' | 'string' | 'number' | 'literal' | 'plain' };

export function tokenizeJson(source: string): Token[] {
	const out: Token[] = [];
	let i = 0;

	while (i < source.length) {
		const ch = source[i];

		if (ch === '"') {
			// Walk to the closing quote, honouring escapes, then decide whether this
			// string is a key by what follows it.
			let j = i + 1;
			while (j < source.length) {
				if (source[j] === '\\') j += 2;
				else if (source[j] === '"') break;
				else j += 1;
			}
			const text = source.slice(i, j + 1);
			const after = source.slice(j + 1).match(/^\s*:/);
			out.push({ text, kind: after ? 'key' : 'string' });
			i = j + 1;
			continue;
		}

		const literal = source.slice(i).match(/^(true|false|null)\b/);
		if (literal) {
			out.push({ text: literal[0], kind: 'literal' });
			i += literal[0].length;
			continue;
		}

		const number = source.slice(i).match(/^-?\d+(\.\d+)?([eE][+-]?\d+)?/);
		if (number) {
			out.push({ text: number[0], kind: 'number' });
			i += number[0].length;
			continue;
		}

		// Everything else — braces, commas, whitespace — travels as one run.
		let j = i;
		while (j < source.length && !'"'.includes(source[j]) && !/[-\d]/.test(source[j]) &&
			!/^(true|false|null)\b/.test(source.slice(j))) {
			j += 1;
		}
		if (j === i) j += 1;
		out.push({ text: source.slice(i, j), kind: 'plain' });
		i = j;
	}
	return out;
}
