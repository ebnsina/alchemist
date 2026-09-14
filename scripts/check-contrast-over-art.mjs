// Contrast of text that sits over artwork rather than over a background-color.
//
// The usual walk-up-the-DOM check is useless here: the aurora is a sibling layer,
// not an ancestor, so compositing parent backgrounds reports the body colour and
// passes everything. This hides the glyphs, screenshots, and samples the real
// pixel under each text node instead.
//
// Run against a served build: npm run preview, then npm run check:contrast

import { chromium } from 'playwright';
const b = await chromium.launch();
const p = await b.newPage({ viewport: { width: 1280, height: 900 }, deviceScaleFactor: 1 });
await p.goto('http://localhost:4321/', { waitUntil: 'networkidle' });
await p.waitForTimeout(1000);
const spots = await p.evaluate(() => [...document.querySelectorAll('[class*="text-muted"], h1, h2, p, li, a')]
  .filter(e => e.getBoundingClientRect().top < 900 && e.textContent.trim() && e.children.length === 0)
  .map(e => { const r = e.getBoundingClientRect();
    const m = getComputedStyle(e).color.match(/[\d.]+/g);
    return { t: e.textContent.trim().slice(0,20), x: Math.round(r.left+r.width/2),
             y: Math.round(r.top+r.height/2), c: [+m[0],+m[1],+m[2]],
             a: m.length<4?1:parseFloat(m[3]) }; })
  .filter(s => s.y > 0 && s.y < 900 && s.x > 0 && s.x < 1280));
await p.addStyleTag({ content: '*{color:transparent!important}' });
await p.waitForTimeout(250);
const png = (await p.screenshot()).toString('base64');
const px = await p.evaluate(async ({ png, spots }) => {
  const img = new Image();
  await new Promise(r => { img.onload = r; img.src = 'data:image/png;base64,' + png; });
  const c = document.createElement('canvas'); c.width = img.width; c.height = img.height;
  const ctx = c.getContext('2d'); ctx.drawImage(img, 0, 0);
  return spots.map(s => ({ ...s, bg: [...ctx.getImageData(s.x, s.y, 1, 1).data].slice(0,3) }));
}, { png, spots });
const lin = c => { c/=255; return c<=0.03928 ? c/12.92 : ((c+0.055)/1.055)**2.4; };
const lum = ([r,g,bl]) => 0.2126*lin(r)+0.7152*lin(g)+0.0722*lin(bl);
let fails = 0;
for (const s of px) {
  const fg = [0,1,2].map(i => s.c[i]*s.a + s.bg[i]*(1-s.a));
  const A = lum(fg), B = lum(s.bg);
  const r = (Math.max(A,B)+0.05)/(Math.min(A,B)+0.05);
  if (r < 4.5) { fails++; console.log(`  FAIL ${s.t.padEnd(22)} bg=rgb(${s.bg.join(',')}) ${r.toFixed(2)}:1`); }
}
console.log(fails ? `  ${fails} failing` : '  all hero text passes AA');
await b.close();
process.exit(fails ? 1 : 0);
