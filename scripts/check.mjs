// One runnable check for the two things on this site that can silently break:
// the CSS-driven bilingual swap, and the toggles that replace a client framework.
// Usage: npm run build && npm run verify   (serves build/ itself)
import { chromium } from 'playwright';
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

// Reduced motion must actually remove motion, not merely shorten it.
for (const motion of ['reduce', 'no-preference']) {
	const c = await b.newContext({ viewport: { width: 1280, height: 800 }, reducedMotion: motion === 'reduce' ? 'reduce' : 'no-preference' });
	const pg = await c.newPage();
	await pg.goto(B + '/');
	const durs = await pg.evaluate(() =>
		[...document.querySelectorAll('.btn, nav a, a.panel, .seg button, .iconbtn, .tlink')].flatMap((e) => {
			const cs = getComputedStyle(e);
			return [cs.transitionDuration, cs.animationName === 'none' ? '0s' : cs.animationName];
		})
	);
	const moving = durs.filter((d) => d !== '0s' && d !== '0.01s');
	if (motion === 'reduce' && moving.length) problems.push(`reduced motion still animates: ${[...new Set(moving)].join(', ')}`);
	if (motion === 'no-preference' && !moving.length) problems.push('hover transitions were removed entirely, not just under reduced motion');
	await c.close();
}

// The toggles must work and persist with no framework on the page.
const ctx = await b.newContext({ viewport: { width: 1280, height: 700 }, colorScheme: 'dark' });
const p = await ctx.newPage();
await p.goto(B + '/');
if ((await p.evaluate(() => document.documentElement.lang)) !== 'en') problems.push('default language is not English');
if (await p.evaluate(() => !!document.querySelector('script[type="module"][src]'))) problems.push('a module script is being shipped — the client runtime leaked back in');
await p.click('[data-set-lang="bn"]');
const after = await p.evaluate(() => [document.documentElement.lang, localStorage.getItem('alc-lang'), document.querySelector('[data-set-lang="bn"]').getAttribute('aria-pressed')]);
if (after.join() !== 'bn,bn,true') problems.push(`language toggle did not take: ${after}`);
await p.click('[data-toggle-theme]');
await p.reload();
const kept = await p.evaluate(() => [document.documentElement.lang, document.documentElement.dataset.theme]);
if (kept[0] !== 'bn' || kept[1] !== 'light') problems.push(`choices did not survive a reload: ${kept}`);
await ctx.close();
await b.close();
server.close();

// Page weight, because a bandwidth argument has to hold on its own site.
const fonts = ['fonts/google-sans-flex-latin.woff2', 'fonts/geist-mono-latin.woff2'].reduce((a, f) => a + statSync(`build/${f}`).size, 0);
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
console.log('\n(gzipped; fonts are raw woff2 and are cached across every route)\n');

if (problems.length) {
	console.error('FAILED:\n' + problems.map((p) => '  - ' + p).join('\n'));
	process.exit(1);
}
console.log('OK — no overflow at 360px, no language leakage, reduced motion respected, toggles persist, no framework JS shipped.');
