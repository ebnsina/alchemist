import { chromium } from 'playwright';
import { PNG } from 'pngjs';
const b = await chromium.launch({ args: ['--use-gl=swiftshader', '--enable-unsafe-swiftshader'] });
const c = await b.newContext({ viewport: { width: 1280, height: 900 } });
const p = await c.newPage();
await p.goto('http://localhost:8765/');
await p.waitForTimeout(2600);
// Hide the hero text so the screenshot is pure background under each box.
const boxes = await p.evaluate(() => {
	const sel = ['h1', '.pill', 'p.text-base, p.text-lg', '.btn-ghost', 'ul.mt-6 li'];
	const out = [];
	for (const s of sel)
		for (const el of document.querySelectorAll('section:first-of-type ' + s)) {
			const r = el.getBoundingClientRect();
			if (r.width < 4 || r.top > 900) continue;
			const cs = getComputedStyle(el);
			out.push({ s, x: r.x, y: r.y, w: r.width, h: r.height, color: cs.webkitTextFillColor || cs.color });
			el.style.visibility = 'hidden';
		}
	return out;
});
const buf = await p.screenshot({ clip: { x: 0, y: 0, width: 1280, height: 900 } });
const png = PNG.sync.read(buf);
const lum = ([r, g, bb]) => { const v = [r, g, bb].map(x => x / 255).map(x => x <= .03928 ? x / 12.92 : ((x + .055) / 1.055) ** 2.4); return .2126 * v[0] + .7152 * v[1] + .0722 * v[2]; };
const cr = (a, l2) => { const l1 = lum(a); return (Math.max(l1, l2) + .05) / (Math.min(l1, l2) + .05); };
const parse = s => s.match(/[\d.]+/g).map(Number).slice(0, 3);
for (const bx of boxes) {
  let worst = 99, worstPx = null;
  for (let y = Math.max(0, bx.y | 0); y < Math.min(900, bx.y + bx.h); y += 2)
    for (let x = Math.max(0, bx.x | 0); x < Math.min(1280, bx.x + bx.w); x += 2) {
      const i = (1280 * y + x) << 2;
      const L = lum([png.data[i], png.data[i + 1], png.data[i + 2]]);
      const r = cr(parse(bx.color), L);
      if (r < worst) { worst = r; worstPx = [png.data[i], png.data[i + 1], png.data[i + 2]]; }
    }
  console.log(bx.s.padEnd(24), bx.color.padEnd(22), 'worst', worst.toFixed(2) + ':1', 'over rgb(' + worstPx + ')');
}
await b.close();
