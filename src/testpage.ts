// Test bench. Not shipped — it is how you drive the player against a local Alchemist.

import { AlchemistPlayerUI } from './ui.ts';
import { siblingURL, expiresAt, viewerLabel } from './signed-url.ts';
import { buildPayload, beaconURL } from './beacon.ts';

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;

const srcInput = $<HTMLInputElement>('src');
const host = $('host');
const logEl = $('log');
const statsEl = $('stats');
const frame = $<HTMLIFrameElement>('frame');
const embedCard = $('embedCard');

const DEFAULT_ORIGIN = 'http://localhost:8099';
srcInput.value = localStorage.getItem('alc.src') ?? '';

const lines: string[] = [];
function log(msg: string): void {
  lines.unshift(`${new Date().toISOString().slice(11, 23)}  ${msg}`);
  logEl.textContent = lines.slice(0, 120).join('\n');
}

let player: AlchemistPlayerUI | null = null;

async function load(): Promise<void> {
  const src = srcInput.value.trim();
  if (!src) { log('no URL — paste one first'); return; }
  localStorage.setItem('alc.src', src);
  await player?.destroy();
  host.innerHTML = '';
  lines.length = 0;

  const exp = expiresAt(src);
  log(`loading ${new URL(src, DEFAULT_ORIGIN).pathname}`);
  log(exp ? `signature expires ${new Date(exp).toISOString()}` : 'URL carries no exp — unsigned');
  log(`beacon would POST to ${new URL(beaconURL(src, DEFAULT_ORIGIN)).pathname}`);
  log(`clear key licence would POST to ${new URL(siblingURL(src, 'key'), DEFAULT_ORIGIN).pathname}${new URL(src, DEFAULT_ORIGIN).search}`);
  log(viewerLabel(src) ? `viewer label ${viewerLabel(src)} — watermark on` : 'no vl in the URL — no watermark');

  player = new AlchemistPlayerUI(host, {
    src,
    poster: siblingURL(src, 'poster.jpg'),
    lang: 'auto',
    dataSaver: 'auto',
    maxHeight: 720,
    country: 'BD',
  });

  for (const type of ['ready', 'playing', 'pause', 'ended', 'error', 'needs-refresh',
    'datasaverchange', 'languagechange', 'qualitychange', 'buffering', 'thumbnails', 'statechange']) {
    player.addEventListener(type, (e) => {
      const d = (e as CustomEvent).detail;
      log(`${type}${d ? ` ${JSON.stringify(d).slice(0, 180)}` : ''}`);
    });
  }
  (window as unknown as Record<string, unknown>).__alc = player;
  setInterval(renderStats, 1000);
}

function renderStats(): void {
  if (!player) return;
  const s = player.getStats();
  statsEl.textContent = JSON.stringify({
    state: player.state,
    lang: player.lang,
    dataSaver: player.dataSaver,
    quality: player.quality,
    mbPerHour: Math.round(player.estimatedMbPerHour()),
    qualities: player.qualities().map((q) => `${q.height}p @ ${Math.round(q.bandwidth / 1000)}kbps`),
    captions: player.captionTracks().length,
    viewerLabel: player.viewerLabel,
    thumbnails: player.thumbnailTiles.length,
    stats: s,
    beaconBody: buildPayload({ ...s, country: 'BD' }),
  }, null, 2);
}

$('load').addEventListener('click', () => void load());
$('bn').addEventListener('click', () => player?.setLang('bn'));
$('en').addEventListener('click', () => player?.setLang('en'));
$('saver').addEventListener('click', () => player?.setDataSaver(!player.dataSaver));

$('expire').addEventListener('click', () => {
  // Rewrites exp to the past, which is exactly what the origin will 403 on.
  const u = new URL(srcInput.value.trim() || 'http://x/a.m3u8', DEFAULT_ORIGIN);
  u.searchParams.set('exp', String(Math.floor(Date.now() / 1000) - 60));
  srcInput.value = u.toString();
  log('exp rewritten to the past — press Load to watch the expired-signature path');
});

$('label').addEventListener('click', () => {
  const u = new URL(srcInput.value.trim() || 'http://x/a/master.m3u8', DEFAULT_ORIGIN);
  u.searchParams.set('wm', 'STU-2291 · 017•••4456');
  srcInput.value = u.toString();
  log('vl added — press Load to watch the watermark drift (the real one is signed)');
});

$('loadEmbed').addEventListener('click', () => {
  const src = srcInput.value.trim();
  if (!src) return;
  const asset = new URL(src, DEFAULT_ORIGIN).pathname.split('/').filter(Boolean).slice(-2, -1)[0] ?? 'asset';
  embedCard.hidden = false;
  frame.src = `/e/${asset}?t=${encodeURIComponent(src)}`;
  log(`embed → /e/${asset}?t=…`);
});

window.addEventListener('message', (e) => {
  const d = e.data as { alchemist?: string; type?: string; detail?: unknown } | null;
  if (d && typeof d === 'object' && d.alchemist) log(`iframe ▸ ${d.type} ${JSON.stringify(d.detail ?? {}).slice(0, 140)}`);
});

// Conjunct check: the strings that break with a naive font stack.
const SAMPLES = ['যুক্তাক্ষর', 'বাংলাদেশ', 'শিক্ষা', 'ডেটা সাশ্রয়', 'ভিডিও চালানো যাচ্ছে না'];
$('conjuncts').innerHTML = SAMPLES.map((s) => {
  const cps = [...s].map((c) => 'U+' + c.codePointAt(0)!.toString(16).toUpperCase().padStart(4, '0')).join(' ');
  return `<div>${s}<small>${cps}</small></div>`;
}).join('');
