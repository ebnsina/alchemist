// The iframe entry point. Host writes one tag and is done:
//   <iframe src="https://play.example/e/{asset}?t={signed playback URL}"
//           allow="autoplay; fullscreen; encrypted-media" allowfullscreen></iframe>

import { AlchemistPlayerUI } from './ui.ts';
import { siblingURL } from './signed-url.ts';
import { makeT, negotiateLang, type Lang } from './i18n.ts';
import { injectStyles } from './styles.ts';
import { icons } from './icons.ts';
import { VERSION } from './version.ts';
import type { NetworkKind } from './network.ts';

const params = new URLSearchParams(location.search);
const mount = document.getElementById('alc-root') as HTMLElement;

function fatal(lang: Lang, messageKey: string): void {
  injectStyles();
  const t = makeT(lang);
  mount.innerHTML =
    `<div class="alc" lang="${lang}"><div class="alc-stage">` +
    `<div class="alc-panel" role="alert"><div class="alc-panel-icon">${icons.alert}</div>` +
    `<h2></h2><p></p></div></div></div>`;
  (mount.querySelector('h2') as HTMLElement).textContent = t('errTitle');
  (mount.querySelector('p') as HTMLElement).textContent = t(messageKey);
}

function boot(): void {
  const lang: Lang | 'auto' = (() => {
    const raw = params.get('lang');
    return raw === 'bn' || raw === 'en' ? raw : 'auto';
  })();
  const resolvedLang = lang === 'auto' ? negotiateLang(null, navigator.languages) : lang;

  const token = params.get('t');
  if (!token) { fatal(resolvedLang, 'errGeneric'); return; }

  let src: string;
  try {
    src = new URL(token, location.href).toString();
  } catch {
    fatal(resolvedLang, 'errGeneric');
    return;
  }

  // /e/{asset} is the address the host wrote; the token is what actually authorizes.
  // A mismatch means the two were assembled from different assets — refuse rather
  // than quietly play the wrong video.
  const pathAsset = decodeURIComponent(location.pathname.split('/').filter(Boolean).pop() ?? '');
  const tokenAsset = new URL(src).pathname.split('/').filter(Boolean).slice(-2, -1)[0] ?? '';
  if (pathAsset && tokenAsset && pathAsset !== tokenAsset && pathAsset !== 'embed.html') {
    fatal(resolvedLang, 'errNotFound');
    return;
  }

  const saver = params.get('saver');
  const player = new AlchemistPlayerUI(mount, {
    src,
    lang,
    poster: params.get('poster') ?? siblingURL(src, 'poster.jpg'),
    thumbnails: params.get('thumbs') ?? undefined,
    dataSaver: saver === '1' ? true : saver === '0' ? false : 'auto',
    autoplay: params.get('autoplay') === '1',
    muted: params.get('muted') === '1' || params.get('autoplay') === '1',
    loop: params.get('loop') === '1',
    maxHeight: Number(params.get('max')) || 720,
    network: (params.get('network') as NetworkKind | null) ?? undefined,
    country: params.get('country') ?? 'BD',
    beacon: params.get('beacon') !== '0',
  });

  bridge(player);
}

/**
 * An iframe cannot hand a JS object to its host, so events and commands cross by
 * postMessage. `needs-refresh` is the one that matters: without it the host has no
 * way to learn the signature is about to die.
 */
function bridge(player: AlchemistPlayerUI): void {
  const post = (type: string, detail?: unknown) => {
    try { parent.postMessage({ alchemist: VERSION, type, detail }, '*'); } catch { /* no parent */ }
  };

  for (const type of [
    'ready', 'playing', 'pause', 'ended', 'buffering', 'error', 'needs-refresh',
    'datasaverchange', 'languagechange', 'qualitychange', 'statechange', 'timeupdate',
  ]) {
    player.addEventListener(type, (e) => post(type, (e as CustomEvent).detail));
  }

  window.addEventListener('message', (e) => {
    const msg = e.data as { alchemist?: unknown; command?: string; value?: unknown } | null;
    if (!msg || typeof msg !== 'object' || !msg.alchemist || typeof msg.command !== 'string') return;
    switch (msg.command) {
      case 'play': void player.play(); break;
      case 'pause': player.pause(); break;
      case 'seek': player.seek(Number(msg.value) || 0); break;
      case 'mute': player.setMuted(Boolean(msg.value)); break;
      case 'volume': player.setVolume(Number(msg.value)); break;
      case 'dataSaver': player.setDataSaver(Boolean(msg.value)); break;
      case 'lang': player.setLang(msg.value === 'en' ? 'en' : 'bn'); break;
      case 'quality': player.setQuality(msg.value === 'auto' ? 'auto' : Number(msg.value)); break;
      // The whole point of needs-refresh: the host fetches a new signed URL and
      // hands it back, and playback continues from the same second.
      case 'refresh': void player.refresh(String(msg.value)); break;
      case 'destroy': void player.destroy(); break;
    }
  });

  post('embed-ready', { version: VERSION });
}

boot();
