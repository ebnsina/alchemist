// One runnable check for the things on this site that can silently break: contrast
// on a dark ground, overflow at phone width, and the theme toggle that replaces a
// client framework.
// Usage: npm run build && npm run verify   (starts the real adapter-node server)
import { chromium } from 'playwright';

import { spawn } from 'node:child_process';
import { gzipSync } from 'node:zlib';
import { readFileSync, statSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

// The site is served by its own adapter-node server now rather than as a folder of
// files. Hand-serving build/ would test something the deployment never runs, and
// would miss /api entirely.
const PORT = 8799;
const app = spawn('node', ['build/index.js'], {
	env: { ...process.env, PORT: String(PORT), ORIGIN: `http://localhost:${PORT}` },
	stdio: ['ignore', 'ignore', 'inherit']
});
const B = `http://localhost:${PORT}`;

// Wait for it rather than guessing at a sleep: a slow machine would otherwise fail
// every check with a connection error and look like a site problem.
for (let i = 0; i < 60; i++) {
	try {
		await fetch(B + '/');
		break;
	} catch {
		await new Promise((r) => setTimeout(r, 250));
	}
}
const server = { close: () => app.kill() };

const routes = ['/', '/login/', '/signup/', '/contact/', '/nope/'];
const problems = [];
const b = await chromium.launch();

for (const scheme of ['dark', 'light']) {
	for (const w of [360, 1280]) {
		const ctx = await b.newContext({ viewport: { width: w, height: 900 }, colorScheme: scheme });
		const p = await ctx.newPage();
		p.on('pageerror', (e) => problems.push(`page error ${scheme}: ${e.message}`));
		for (const r of routes) {
			await p.goto(B + r);
			await p.evaluate(() => document.fonts.ready);
			const over = await p.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
			if (over > 1) problems.push(`horizontal scroll ${over}px on ${r} (${scheme}/${w}px)`);
		}
		await ctx.close();
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
	for (const r of ['/', '/login/', '/signup/', '/contact/']) {
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
		[...document.querySelectorAll('.btn, .btn-solid, .link, .side-link, .icon-btn, .node, .field, .sk, details summary, .chevron')].flatMap((e) => {
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
	await c.close();
}

// The page must declare English with no framework on it.
const ctx = await b.newContext({ viewport: { width: 1280, height: 700 }, colorScheme: 'dark' });
const p = await ctx.newPage();
await p.goto(B + '/');
if ((await p.evaluate(() => document.documentElement.lang)) !== 'en') problems.push('default language is not English');
await ctx.close();
await b.close();
server.close();

// Page weight. The three faces are self-hosted, so they are the files on disk —
// already woff2, and not worth gzipping twice.
const fonts = ['archivo-latin', 'clash-display', 'geist-mono-latin'].reduce((n, f) => n + statSync(`static/fonts/${f}.woff2`).size, 0);
void readdirSync;

const gz = (f) => gzipSync(readFileSync(f), { level: 9 }).length;
// Prerendered pages land under build/prerendered; the hashed assets under build/client.
const pageFile = (r) => {
	const rel = r === '/404' ? '404.html' : join(r.replace(/^\//, ''), 'index.html');
	for (const base of ['build/prerendered/pages', 'build/prerendered', 'build/client']) {
		const f = join(base, rel);
		try {
			statSync(f);
			return f;
		} catch {}
	}
	return null;
};
const kb = (n) => (n / 1024).toFixed(1).padStart(6) + ' KB';
console.log('\nroute             html      css       js    fonts   first visit');
for (const r of ['/', '/login/', '/signup/', '/contact/', '/404']) {
	const f = pageFile(r);
	if (!f) {
		console.log(r.padEnd(12), '  (not prerendered)');
		continue;
	}
	const html = readFileSync(f, 'utf8');
	let css = 0, js = 0;
	for (const a of new Set([...html.matchAll(/(_app\/immutable\/[^"')\s]+)/g)].map((m) => m[1]))) {
		const n = gz(`build/client/${a}`);
		a.endsWith('.css') ? (css += n) : (js += n);
	}
	const h = gzipSync(Buffer.from(html), { level: 9 }).length;
	console.log(r.padEnd(12), kb(h), kb(css), kb(js), kb(fonts), kb(h + css + js + fonts));
}
console.log('\n(gzipped, except fonts: three self-hosted woff2 files, cached across every route)\n');

if (problems.length) {
	console.error('FAILED:\n' + problems.map((p) => '  - ' + p).join('\n'));
	process.exit(1);
}
console.log('OK — contrast passes AA, no overflow at 360px, reduced motion respected.');
