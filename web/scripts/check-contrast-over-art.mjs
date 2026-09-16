// Contrast of text that sits over artwork rather than over a background-color.
//
// The usual walk-up-the-DOM check is useless here: the aurora and the page's
// washes and slabs are sibling layers, not ancestors, so compositing parent
// backgrounds reports the body colour and passes everything. This hides the
// glyphs, screenshots each viewport down the page, and samples the real pixel
// under each text node instead.
//
// Run against a built site: npm run build, then npm run check:contrast

import { chromium } from 'playwright';
import { spawn } from 'node:child_process';

// Starts the site's own adapter-node server. `vite preview` does not serve an
// adapter-node build, and hand-serving build/ would measure something the deployment
// never runs.
const PORT = 4321;
const app = spawn('node', ['build/index.js'], {
	env: { ...process.env, PORT: String(PORT), ORIGIN: `http://localhost:${PORT}` },
	stdio: ['ignore', 'ignore', 'inherit']
});
const BASE = `http://localhost:${PORT}`;
process.on('exit', () => app.kill());
for (let i = 0; i < 60; i++) {
	try {
		await fetch(BASE + '/');
		break;
	} catch {
		await new Promise((r) => setTimeout(r, 250));
	}
}

const W = 1280, H = 900;
const PAGES = ['/', '/contact/', '/login/', '/signup/'];

const lin = c => { c /= 255; return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4; };
const lum = ([r, g, b]) => 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);

const browser = await chromium.launch();
// Reduced motion, deliberately: the same markup and colours, with nothing animating
// mid-sample.
const context = await browser.newContext({
	viewport: { width: W, height: H },
	deviceScaleFactor: 1,
	reducedMotion: 'reduce'
});
const p = await context.newPage();
let fails = 0, checked = 0;

// Both themes. A palette that only passes in one of them is half a system, and the
// dark values are the ones nobody looks at with a meter.
const THEMES = ['light', 'dark'];

for (const theme of THEMES) {
for (const route of PAGES) {
	await p.goto(BASE + route, { waitUntil: 'networkidle' });
	await p.evaluate(t => document.documentElement.setAttribute('data-theme', t), theme);
	// Scroll-reveal starts elements at opacity 0. Jumping straight to an offset can
	// outrun the observer, and an unrevealed card samples as bare page instead of
	// its own surface. The resting state is the final state, so pin it.
	await p.waitForTimeout(800);
	const total = await p.evaluate(() => document.body.scrollHeight);

	for (let top = 0; top < total; top += H) {
		// html has scroll-behavior: smooth, so a plain scrollTo is still animating
		// when the screenshot lands and every sample reads the wrong row.
		await p.evaluate(y => window.scrollTo({ top: y, behavior: 'instant' }), top);
		await p.waitForTimeout(350);

		const spots = await p.evaluate(({ W, H }) => [...document.querySelectorAll('[class*="text-"], h1, h2, h3, p, li, a, span')]
			.filter(e => e.textContent.trim() && e.children.length === 0)
			// Screen-reader-only text is clipped to a pixel. It is never seen, and
			// sampling it reads whatever is painted behind the clip.
			.filter(e => {
				const r = e.getBoundingClientRect();
				return r.width > 4 && r.height > 4;
			})
			// Drop anything another element paints over — the fixed header, a
			// decorative plate — since the pixel there is not this element's ground.
			.filter(e => {
				const r = e.getBoundingClientRect();
				const hit = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
				return hit && (hit === e || e.contains(hit) || hit.contains(e));
			})
			.map(e => {
				const r = e.getBoundingClientRect();
				// Canvas resolves color-mix() and oklab() the way a regex cannot.
				const cv = document.createElement('canvas').getContext('2d');
				cv.fillStyle = '#000';
				cv.fillStyle = getComputedStyle(e).color;
				cv.fillRect(0, 0, 1, 1);
				const [cr, cg, cb, ca] = cv.getImageData(0, 0, 1, 1).data;
				return { t: e.textContent.trim().slice(0, 20), x: Math.round(r.left + r.width / 2),
					y: Math.round(r.top + r.height / 2), c: [cr, cg, cb], a: ca / 255 };
			})
			.filter(s => s.y > 4 && s.y < H - 4 && s.x > 4 && s.x < W - 4), { W, H });
		if (!spots.length) continue;

		await p.addStyleTag({ content: '.contrast-probe{color:transparent!important}' });
		await p.evaluate(() => document.documentElement.classList.add('contrast-probe'));
		await p.addStyleTag({ content: '.contrast-probe *{color:transparent!important}' });
		await p.waitForTimeout(200);
		const png = (await p.screenshot()).toString('base64');
		await p.evaluate(() => document.documentElement.classList.remove('contrast-probe'));

		const px = await p.evaluate(async ({ png, spots }) => {
			const img = new Image();
			await new Promise(r => { img.onload = r; img.src = 'data:image/png;base64,' + png; });
			const c = document.createElement('canvas'); c.width = img.width; c.height = img.height;
			const ctx = c.getContext('2d'); ctx.drawImage(img, 0, 0);
			return spots.map(s => ({ ...s, bg: [...ctx.getImageData(s.x, s.y, 1, 1).data].slice(0, 3) }));
		}, { png, spots });

		for (const s of px) {
			checked++;
			const fg = [0, 1, 2].map(i => s.c[i] * s.a + s.bg[i] * (1 - s.a));
			const A = lum(fg), B = lum(s.bg);
			const r = (Math.max(A, B) + 0.05) / (Math.min(A, B) + 0.05);
			if (r < 4.5) {
				fails++;
				console.log(`  FAIL ${theme} ${route} ${s.t.padEnd(20)} bg=rgb(${s.bg.join(',')}) ${r.toFixed(2)}:1`);
			}
		}
	}
}
}

console.log(fails ? `  ${fails} failing of ${checked}` : `  all ${checked} text samples pass AA in both themes`);
await browser.close();
process.exit(fails ? 1 : 0);
