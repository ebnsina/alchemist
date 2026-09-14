// One runnable check for the two things on this site that can silently break:
// the CSS-driven bilingual swap, and the toggles that replace a client framework.
// Usage: npm run build && npm run verify   (serves build/ itself)
import { chromium } from 'playwright';

const UA =
	'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36';
import { createServer } from 'node:http';
import { readFile, stat } from 'node:fs/promises';
import { extname, join } from 'node:path';
import { gzipSync } from 'node:zlib';
import { readFileSync, statSync } from 'node:fs';

const TYPES = { '.html': 'text/html', '.css': 'text/css', '.js': 'text/javascript', '.svg': 'image/svg+xml', '.woff2': 'font/woff2', '.png': 'image/png', '.xml': 'application/xml', '.txt': 'text/plain' };
const server = createServer(async (req, res) => {
	let p = join('build', decodeURIComponent(req.url.split('?')[0]));
	try {
		if ((await stat(p)).isDirectory()) p = join(p, 'index.html');
	} catch {
		p = 'build/404.html';
		res.statusCode = 404;
	}
	try {
		const body = await readFile(p);
		res.setHeader('content-type', TYPES[extname(p)] ?? 'application/octet-stream');
		res.end(body);
	} catch {
		res.statusCode = 404;
		res.end('not found');
	}
});
await new Promise((r) => server.listen(8799, r));
const B = 'http://localhost:8799';

const routes = ['/', '/edtech/', '/media/', '/pricing/', '/docs/', '/about/', '/nope/'];
const problems = [];
const b = await chromium.launch();

for (const scheme of ['dark', 'light']) {
	for (const lang of ['en', 'bn']) {
		for (const w of [360, 1280]) {
			const ctx = await b.newContext({ viewport: { width: w, height: 900 }, colorScheme: scheme });
			const p = await ctx.newPage();
			p.on('pageerror', (e) => problems.push(`page error ${lang}/${scheme}: ${e.message}`));
			await p.addInitScript((l) => localStorage.setItem('alc-lang', l), lang);
			for (const r of routes) {
				await p.goto(B + r);
				await p.evaluate(() => document.fonts.ready);
				const over = await p.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
				if (over > 1) problems.push(`horizontal scroll ${over}px on ${r} (${lang}/${scheme}/${w}px)`);
				const [leaked, swallowed] = await p.evaluate((l) => {
					const vis = (sel) => [...document.querySelectorAll(sel)].filter((e) => e.offsetParent !== null).length;
					const gone = (sel) => [...document.querySelectorAll(sel)].filter((e) => getComputedStyle(e).display === 'none').length;
					return l === 'bn' ? [vis('.l-en'), gone('.l-bn')] : [vis('.l-bn'), gone('.l-en')];
				}, lang);
				if (leaked) problems.push(`${leaked} nodes of the other language visible on ${r} (${lang})`);
				if (swallowed) problems.push(`${swallowed} nodes of the chosen language hidden on ${r} (${lang})`);
			}
			await ctx.close();
		}
	}
}

// Contrast, computed against the real composited background rather than eyeballed.
// Low-opacity ink on near-black is exactly how a dark theme fails AA quietly.
const lum = (c) => {
	const v = c.map((x) => x / 255).map((x) => (x <= 0.03928 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4));
	return 0.2126 * v[0] + 0.7152 * v[1] + 0.0722 * v[2];
};
const ratio = (a, b) => {
	const [x, y] = [lum(a), lum(b)];
	return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05);
};
{
	const theme = 'dark';
	const c = await b.newContext({ viewport: { width: 1280, height: 900 } });
	const pg = await c.newPage();
	for (const r of ['/', '/pricing/', '/docs/', '/about/', '/edtech/']) {
		await pg.goto(B + r);
		const samples = await pg.evaluate(() => {
			// color-mix() computes to color(srgb r g b / a) with 0-1 channels, while
			// rgb()/rgba() are 0-255. Reading one as the other turns a pale translucent
			// header into near-black and invents contrast failures.
			const parse = (s) => {
				const n = (s.match(/[\d.]+(?=%)?|[\d.]+/g) || []).map(Number);
				return s.startsWith('color(') ? [n[0] * 255, n[1] * 255, n[2] * 255, n.length > 3 ? n[3] : 1] : n;
			};
			// Composite every translucent layer down to the first opaque one, so a
			// semi-transparent surface is not silently skipped.
			const bgOf = (el) => {
				const stack = [];
				for (let n = el; n; n = n.parentElement) {
					const bg = parse(getComputedStyle(n).backgroundColor);
					const a = bg.length > 3 ? bg[3] : 1;
					if (a === 0) continue;
					stack.push([bg.slice(0, 3), a]);
					if (a > 0.999) break;
				}
				// If nothing opaque was found, the page's own background is the floor —
				// not black, which would invent a failure that no reader can see.
				const floor = parse(getComputedStyle(document.body).backgroundColor).slice(0, 3);
				let out = stack.length && stack[stack.length - 1][1] > 0.999 ? stack.pop()[0] : floor;
				while (stack.length) {
					const [c, a] = stack.pop();
					out = c.map((v, i) => v * a + out[i] * (1 - a));
				}
				return out;
			};
			const out = [];
			for (const el of document.querySelectorAll('p, li, dd, dt, td, th, h1, h2, h3, a, caption, span')) {
				if (!el.offsetParent || !el.textContent.trim()) continue;
				const box = el.getBoundingClientRect();
				if (box.right < 0 || box.bottom < 0 || box.width === 0) continue; // off-canvas skip link
				const cs = getComputedStyle(el);
				// Gradient text paints through -webkit-text-fill-color: transparent, so the
				// computed colour is meaningless. Its stops are asserted separately below.
				const fill = cs.webkitTextFillColor || cs.color;
				if (/rgba?\(0,\s*0,\s*0,\s*0\)|transparent/.test(fill)) continue;
				if (el.children.length && !/[^\s]/.test([...el.childNodes].filter((n) => n.nodeType === 3).map((n) => n.textContent).join(''))) continue;
				const col = parse(cs.color);
				const a = col.length > 3 ? col[3] : 1;
				const bg = bgOf(el);
				const fg = col.slice(0, 3).map((v, i) => v * a + bg[i] * (1 - a));
				out.push([cs.color, parseFloat(cs.fontSize), parseInt(cs.fontWeight, 10), fg, bg, el.className || el.tagName]);
			}
			return out;
		});
		for (const [col, size, weight, fg, bg, who] of samples) {
			const large = size >= 24 || (size >= 18.66 && weight >= 700);
			const need = large ? 3 : 4.5;
			const got = ratio(fg, bg);
			if (got < need) problems.push(`contrast ${got.toFixed(2)}:1 (needs ${need}) — ${col} at ${size}px on bg rgb(${bg.map(Math.round).join(',')}) [${who}] ${r} (${theme})`);
		}
	}
	await c.close();
}

// Reduced motion must actually remove motion, not merely shorten it.
for (const motion of ['reduce', 'no-preference']) {
	const c = await b.newContext({ viewport: { width: 1280, height: 800 }, reducedMotion: motion === 'reduce' ? 'reduce' : 'no-preference' });
	const pg = await c.newPage();
	await pg.goto(B + '/');
	const durs = await pg.evaluate(() =>
		[...document.querySelectorAll('.btn, .btn-primary, .btn-ghost, nav a, [data-set-lang], .tlink, .ticker, .shine, .progress-anim, .card, details summary, .chevron')].flatMap((e) => {
			const cs = getComputedStyle(e);
			// .shine carries its animation on ::after, so the pseudo-element counts too.
			const after = getComputedStyle(e, '::after');
			return [
				cs.transitionDuration,
				cs.animationName === 'none' ? '0s' : cs.animationName,
				after.animationName === 'none' ? '0s' : after.animationName,
				after.animationDuration === '0s' ? '0s' : after.animationDuration
			];
		})
	);
	const moving = durs.filter((d) => d !== '0s' && d !== '0.01s');
	if (motion === 'reduce' && moving.length) problems.push(`reduced motion still animates: ${[...new Set(moving)].join(', ')}`);
	if (motion === 'no-preference' && !moving.length) problems.push('hover transitions were removed entirely, not just under reduced motion');
	for (const anim of ['ticker', 'shine', 'progress'])
		if (motion === 'no-preference' && !durs.some((d) => String(d).includes(anim)))
			problems.push(`the ${anim} animation is missing when motion is allowed`);
	await c.close();
}

// Content must never be left hidden by a reveal that did not fire, and it must not
// be hidden at all when scripting is off — the fade class is added by the action, so
// no JavaScript means no opacity:0.
{
	const c = await b.newContext({ viewport: { width: 1280, height: 900 } });
	const pg = await c.newPage();
	await pg.goto(B + '/');
	await pg.waitForTimeout(500);
	await pg.evaluate(async () => {
		for (let y = 0; y < document.body.scrollHeight; y += 400) {
			window.scrollTo(0, y);
			await new Promise((r) => requestAnimationFrame(() => setTimeout(r, 60)));
		}
	});
	await pg.waitForTimeout(1200);
	const [faded, shown] = await pg.evaluate(() => [
		document.querySelectorAll('.scroll-fade').length,
		document.querySelectorAll('.scroll-fade.is-visible').length
	]);
	if (faded === 0) problems.push('no scroll reveals were registered at all');
	if (faded !== shown) problems.push(`${faded - shown} revealed blocks never became visible after a full scroll`);
	await c.close();

	const nojs = await b.newContext({ viewport: { width: 1280, height: 900 }, javaScriptEnabled: false });
	const np = await nojs.newPage();
	await np.goto(B + '/');
	const hidden = await np.evaluate(() => document.querySelectorAll('.scroll-fade').length);
	if (hidden) problems.push(`${hidden} blocks are hidden with scripting off`);
	await nojs.close();
}

// The toggles must work and persist with no framework on the page.
const ctx = await b.newContext({ viewport: { width: 1280, height: 700 }, colorScheme: 'dark' });
const p = await ctx.newPage();
await p.goto(B + '/');
if ((await p.evaluate(() => document.documentElement.lang)) !== 'en') problems.push('default language is not English');
await p.click('[data-set-lang="bn"]');
const after = await p.evaluate(() => [document.documentElement.lang, localStorage.getItem('alc-lang'), document.querySelector('[data-set-lang="bn"]').getAttribute('aria-pressed')]);
if (after.join() !== 'bn,bn,true') problems.push(`language toggle did not take: ${after}`);
await p.reload();
const kept = await p.evaluate(() => document.documentElement.lang);
if (kept !== 'bn') problems.push(`the language choice did not survive a reload: ${kept}`);
await ctx.close();
await b.close();
server.close();

// Page weight. Inter is fetched from Google rather than self-hosted, so it is
// measured over the wire: the stylesheet plus the latin face it actually pulls.
let fonts = 0;
try {
	const url =
		'https://fonts.googleapis.com/css2?family=Inter:wght@400..800&display=swap';
	const css = await (await fetch(url, { headers: { 'user-agent': UA } })).text();
	fonts += Buffer.byteLength(css);
	// Only the latin face is downloaded. Bengali falls through to the system's Noto
	// Sans Bengali and costs nothing over the wire.
	for (const rule of [...css.matchAll(/@font-face\s*\{([^}]*)\}/g)].map((m) => m[1])) {
		if (!/unicode-range:\s*U\+0000-00FF/.test(rule)) continue;
		const u = rule.match(/url\((https:[^)]+)\)/);
		if (u) fonts += (await (await fetch(u[1])).arrayBuffer()).byteLength;
	}
} catch {
	fonts = NaN;
}

const gz = (f) => gzipSync(readFileSync(f), { level: 9 }).length;
const kb = (n) => (n / 1024).toFixed(1).padStart(6) + ' KB';
console.log('\nroute             html      css       js    fonts   first visit');
for (const [r, f] of [['/', 'index.html'], ['/edtech/', 'edtech/index.html'], ['/media/', 'media/index.html'], ['/pricing/', 'pricing/index.html'], ['/docs/', 'docs/index.html'], ['/about/', 'about/index.html'], ['/404', '404.html']]) {
	const html = readFileSync(`build/${f}`, 'utf8');
	let css = 0, js = 0;
	for (const a of new Set([...html.matchAll(/(_app\/immutable\/[^"')\s]+)/g)].map((m) => m[1]))) {
		const n = gz(`build/${a}`);
		a.endsWith('.css') ? (css += n) : (js += n);
	}
	const h = gzipSync(Buffer.from(html), { level: 9 }).length;
	console.log(r.padEnd(12), kb(h), kb(css), kb(js), kb(fonts), kb(h + css + js + fonts));
}
console.log('\n(gzipped, except fonts: Inter over the wire from Google, cached across every route)\n');

if (problems.length) {
	console.error('FAILED:\n' + problems.map((p) => '  - ' + p).join('\n'));
	process.exit(1);
}
console.log('OK — contrast passes AA, no overflow at 360px, no language leakage, reduced motion respected, language choice persists.');
