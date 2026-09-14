/**
 * Currency formatting through Intl, so digit grouping, decimal marks and symbol
 * placement come from the locale rather than from strings typed by hand.
 *
 * `narrowSymbol` matters: the default for BDT is the string "BDT", not the taka
 * sign. And Bangla is not a digit swap — bn-BD renders ২,৪০০.০০৳ with the symbol
 * trailing, which is why the locale does this rather than a replace().
 *
 * These run at build time; the site prerenders and ships no JavaScript.
 */
const cache = new Map<string, Intl.NumberFormat>();

function formatter(locale: string, currency: string, min: number, max: number) {
	const key = `${locale}|${currency}|${min}|${max}`;
	let f = cache.get(key);
	if (!f) {
		f = new Intl.NumberFormat(locale, {
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

/** Taka, in the digits and symbol placement of the given language. */
export function taka(amount: number, lang: 'en' | 'bn' = 'en', decimals = 2) {
	return formatter(lang === 'bn' ? 'bn-BD' : 'en-BD', 'BDT', decimals, decimals).format(amount);
}

/** US dollars, for the comparison column. */
export function usd(amount: number, lang: 'en' | 'bn' = 'en', decimals = 0) {
	return formatter(lang === 'bn' ? 'bn-BD' : 'en-US', 'USD', decimals, decimals).format(amount);
}

/** The same formatting, split so the symbol can be styled apart from the digits. */
export function parts(
	amount: number,
	lang: 'en' | 'bn' = 'en',
	decimals = 2,
	currency = 'BDT'
) {
	const locale = lang === 'bn' ? 'bn-BD' : currency === 'USD' ? 'en-US' : 'en-BD';
	return formatter(locale, currency, decimals, decimals).formatToParts(amount);
}
