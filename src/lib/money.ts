/**
 * Currency formatting through Intl, so digit grouping, decimal marks and symbol
 * placement come from the locale rather than from hand-typed strings.
 *
 * `narrowSymbol` is required: the default for BDT renders the letters "BDT", not
 * the taka sign. These run at build time; the site prerenders and ships no JS.
 */
const cache = new Map<string, Intl.NumberFormat>();

function formatter(currency: string, min: number, max: number) {
	const key = `${currency}|${min}|${max}`;
	let f = cache.get(key);
	if (!f) {
		f = new Intl.NumberFormat(currency === 'USD' ? 'en-US' : 'en-BD', {
			style: 'currency',
			currency,
			currencyDisplay: 'narrowSymbol',
			minimumFractionDigits: min,
			maximumFractionDigits: max
		});
		cache.set(key, f);
	}
	return f;
}

export function money(amount: number, decimals = 2, currency = 'BDT') {
	return formatter(currency, decimals, decimals).format(amount);
}

/** The same formatting, split so a part can be styled on its own if needed. */
export function parts(amount: number, decimals = 2, currency = 'BDT') {
	return formatter(currency, decimals, decimals).formatToParts(amount);
}
