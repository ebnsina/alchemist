// The chrome. Everything the viewer touches, in Bangla or English, keyboard-first.

import { AlchemistPlayer, type AlchemistOptions } from './player.ts';
import { injectStyles } from './styles.ts';
import { icons, type IconName } from './icons.ts';
import { formatNumber, formatTime, makeT, type T } from './i18n.ts';
import { tileAt } from './thumbnails.ts';

export interface UIOptions extends AlchemistOptions {
  /** Hide all chrome and drive the player from your own UI. */
  controls?: boolean;
  /** Seconds the controls stay up after the pointer stops. */
  idleMs?: number;
}

const el = <K extends keyof HTMLElementTagNameMap>(
  tag: K, cls?: string, attrs?: Record<string, string>,
): HTMLElementTagNameMap[K] => {
  const node = document.createElement(tag);
  if (cls) node.className = cls;
  for (const [k, v] of Object.entries(attrs ?? {})) node.setAttribute(k, v);
  return node;
};

const iconBtn = (name: IconName, label: string, cls = 'alc-btn'): HTMLButtonElement => {
  const b = el('button', cls, { type: 'button', 'aria-label': label, title: label });
  b.innerHTML = icons[name];
  return b;
};

export class AlchemistPlayerUI extends AlchemistPlayer {
  readonly root: HTMLElement;
  private t: T;
  private idleMs: number;
  private idleTimer: ReturnType<typeof setTimeout> | null = null;
  private toastTimer: ReturnType<typeof setTimeout> | null = null;
  private driftTimer: ReturnType<typeof setInterval> | null = null;
  private scrubbing = false;
  private menuOpen = false;
  private chromeless: boolean;

  private ui!: {
    stage: HTMLElement; centre: HTMLElement; bar: HTMLElement;
    playBtn: HTMLButtonElement; bigBtn: HTMLButtonElement; spinner: HTMLElement;
    muteBtn: HTMLButtonElement; volume: HTMLInputElement;
    time: HTMLElement; seek: HTMLButtonElement;
    played: HTMLElement; buffered: HTMLElement; knob: HTMLElement;
    preview: HTMLElement; previewImg: HTMLElement; previewTime: HTMLElement;
    saver: HTMLButtonElement; saverRate: HTMLElement; saverLabel: HTMLElement;
    capsBtn: HTMLButtonElement; menuBtn: HTMLButtonElement; fsBtn: HTMLButtonElement;
    menu: HTMLElement; panel: HTMLElement; panelTitle: HTMLElement;
    panelBody: HTMLElement; panelAction: HTMLButtonElement;
    toast: HTMLElement; toastText: HTMLElement; live: HTMLElement;
    watermark: HTMLElement | null;
  };

  constructor(host: HTMLElement, options: UIOptions) {
    injectStyles();
    const root = el('div', 'alc');
    const stage = el('div', 'alc-stage');
    const video = el('video');
    stage.appendChild(video);
    root.appendChild(stage);
    host.appendChild(root);

    super(video, options);
    this.root = root;
    this.chromeless = options.controls === false;
    this.idleMs = options.idleMs ?? 2800;
    this.t = makeT(this.lang);

    this.build(stage);
    this.root.setAttribute('lang', this.lang);
    if (this.chromeless) {
      this.ui.bar.hidden = true;
      this.ui.centre.hidden = true;
      this.video.controls = true;
    }
    this.wire();
    this.render();
  }

  // ---- construction -----------------------------------------------------

  private build(stage: HTMLElement): void {
    const t = this.t;

    const centre = el('div', 'alc-centre');
    const bigBtn = iconBtn('play', t('play'), 'alc-big');
    const spinner = el('div', 'alc-spinner');
    spinner.hidden = true;
    centre.append(bigBtn, spinner);

    const scrim = el('div', 'alc-scrim');
    const bar = el('div', 'alc-bar');

    // Seek bar is a real slider for assistive tech, not a div with a mousedown.
    const seek = el('button', 'alc-seek', {
      type: 'button', role: 'slider', 'aria-label': t('seek'),
      'aria-valuemin': '0', 'aria-valuemax': '0', 'aria-valuenow': '0',
    });
    const track = el('div', 'alc-track');
    const buffered = el('div', 'alc-buffered');
    const played = el('div', 'alc-played');
    const knob = el('div', 'alc-knob');
    track.append(buffered, played, knob);
    seek.appendChild(track);

    const preview = el('div', 'alc-preview');
    const previewImg = el('div', 'alc-preview-img');
    const previewTime = el('span', 'alc-preview-time');
    preview.append(previewImg, previewTime);

    const row = el('div', 'alc-row');
    const playBtn = iconBtn('play', t('play'));

    const vol = el('div', 'alc-vol');
    const muteBtn = iconBtn('volume', t('mute'));
    const volume = el('input', '', {
      type: 'range', min: '0', max: '1', step: '0.05', value: '1', 'aria-label': t('volume'),
    }) as HTMLInputElement;
    vol.append(muteBtn, volume);

    const time = el('div', 'alc-time');
    const spacer = el('div', 'alc-spacer');

    const saver = el('button', 'alc-pill', {
      type: 'button', 'aria-pressed': String(this.dataSaver), 'aria-describedby': 'alc-saver-hint',
    });
    const saverIcon = el('span');
    saverIcon.innerHTML = icons.dataSaver;
    const saverLabel = el('span', 'alc-pill-label');
    const saverRate = el('span', 'alc-pill-rate');
    saver.append(saverIcon.firstElementChild as SVGElement, saverLabel, saverRate);

    const capsBtn = iconBtn('captions', t('captions'));
    capsBtn.hidden = true;
    const menuBtn = iconBtn('settings', t('settings'));
    menuBtn.setAttribute('aria-haspopup', 'true');
    menuBtn.setAttribute('aria-expanded', 'false');
    const fsBtn = iconBtn('fullscreen', t('fullscreen'));

    row.append(playBtn, vol, time, spacer, saver, capsBtn, menuBtn, fsBtn);
    bar.append(preview, seek, row);

    const menu = el('div', 'alc-menu', { role: 'menu', 'aria-label': t('settings') });
    menu.hidden = true;

    const panel = el('div', 'alc-panel', { role: 'alert', 'aria-live': 'assertive' });
    panel.hidden = true;
    const panelIcon = el('div', 'alc-panel-icon');
    panelIcon.innerHTML = icons.alert;
    const panelTitle = el('h2');
    const panelBody = el('p');
    const panelAction = el('button', 'alc-action', { type: 'button' });
    panel.append(panelIcon, panelTitle, panelBody, panelAction);

    const toast = el('div', 'alc-toast', { role: 'status' });
    const toastIcon = el('span');
    toastIcon.innerHTML = icons.dataSaver;
    const toastText = el('span');
    toast.append(toastIcon.firstElementChild as SVGElement, toastText);

    const live = el('span', 'alc-sr', { 'aria-live': 'polite' });

    // The label rides inside the signature, so there is nothing to configure and
    // nothing to draw when the URL carries none.
    let watermark: HTMLElement | null = null;
    const label = this.viewerLabel;
    if (label) {
      watermark = el('div', 'alc-wm');
      watermark.textContent = label;
      stage.appendChild(watermark);
    }

    stage.append(centre, scrim, bar, menu, panel, toast, live);

    this.ui = {
      stage, centre, bar, playBtn, bigBtn, spinner, muteBtn, volume, time, seek,
      played, buffered, knob, preview, previewImg, previewTime,
      saver, saverRate, saverLabel, capsBtn, menuBtn, fsBtn,
      menu, panel, panelTitle, panelBody, panelAction, toast, toastText, live, watermark,
    };
  }

  // ---- wiring -----------------------------------------------------------

  private wire(): void {
    const u = this.ui;

    const toggle = () => (this.paused ? void this.play() : this.pause());
    u.playBtn.addEventListener('click', toggle);
    u.bigBtn.addEventListener('click', toggle);
    this.video.addEventListener('click', () => { if (!this.chromeless) toggle(); });

    u.muteBtn.addEventListener('click', () => this.setMuted(!this.muted));
    u.volume.addEventListener('input', () => this.setVolume(Number(u.volume.value)));
    u.fsBtn.addEventListener('click', () => this.toggleFullscreen());
    u.saver.addEventListener('click', () => this.setDataSaver(!this.dataSaver));
    u.capsBtn.addEventListener('click', () => this.openMenu('captions'));
    u.menuBtn.addEventListener('click', () => (this.menuOpen ? this.closeMenu() : this.openMenu()));
    u.panelAction.addEventListener('click', () => this.onPanelAction());

    this.wireSeek();
    this.wireKeyboard();
    this.wireChromeVisibility();

    for (const evt of ['statechange', 'timeupdate', 'progress', 'volumechange', 'ready', 'tracksavailable', 'adaptation', 'qualitychange']) {
      this.addEventListener(evt, () => this.render());
    }
    this.addEventListener('error', () => this.render());
    this.addEventListener('datasaverchange', () => {
      this.render();
      const mb = formatNumber(this.lang, Math.round(this.estimatedMbPerHour()));
      this.toast(
        this.dataSaver
          ? `${this.t('dataSaverOn')} · ${this.t('perHour')} ${mb} ${this.t('megabytes')}`
          : this.t('dataSaverOff'),
      );
    });
    this.addEventListener('languagechange', () => {
      this.t = makeT(this.lang);
      this.root.setAttribute('lang', this.lang);
      this.relabel();
      this.render();
      if (this.menuOpen) this.openMenu();
    });
    this.addEventListener('needs-refresh', (e) => {
      const reason = (e as CustomEvent).detail?.reason;
      if (reason === 'expiring') this.toast(this.t('errExpired'), 'warn');
    });
    this.addEventListener('offline', () => this.toast(this.t('offline'), 'warn'));

    if (this.ui.watermark) {
      this.drift();
      this.driftTimer = setInterval(() => this.drift(), 7000);
    }

    document.addEventListener('fullscreenchange', () => { this.render(); this.drift(); });
    document.addEventListener('click', (e) => {
      if (this.menuOpen && !this.root.contains(e.target as Node)) this.closeMenu();
    });
  }

  private wireSeek(): void {
    const u = this.ui;
    const ratio = (clientX: number) => {
      const r = u.seek.getBoundingClientRect();
      return r.width ? Math.min(1, Math.max(0, (clientX - r.left) / r.width)) : 0;
    };
    const scrubTo = (clientX: number) => {
      if (!this.duration) return;
      this.seek(ratio(clientX) * this.duration);
    };

    u.seek.addEventListener('pointerdown', (e) => {
      this.scrubbing = true;
      this.root.setAttribute('data-scrubbing', '');
      u.seek.setPointerCapture(e.pointerId);
      scrubTo(e.clientX);
    });
    u.seek.addEventListener('pointermove', (e) => {
      if (this.duration) this.showPreview(ratio(e.clientX));
      if (this.scrubbing) scrubTo(e.clientX);
    });
    const end = () => {
      this.scrubbing = false;
      this.root.removeAttribute('data-scrubbing');
      this.hidePreview();
    };
    u.seek.addEventListener('pointerup', end);
    u.seek.addEventListener('pointercancel', end);
    u.seek.addEventListener('pointerleave', () => { if (!this.scrubbing) this.hidePreview(); });

    u.seek.addEventListener('keydown', (e) => {
      const step = e.shiftKey ? 30 : 5;
      if (e.key === 'ArrowLeft') { this.seek(this.currentTime - step); e.preventDefault(); }
      else if (e.key === 'ArrowRight') { this.seek(this.currentTime + step); e.preventDefault(); }
      else if (e.key === 'Home') { this.seek(0); e.preventDefault(); }
      else if (e.key === 'End') { this.seek(this.duration); e.preventDefault(); }
    });
  }

  private wireKeyboard(): void {
    this.root.tabIndex = -1;
    this.root.addEventListener('keydown', (e) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === 'INPUT' || (e.target === this.ui.seek && e.key.startsWith('Arrow'))) return;
      switch (e.key) {
        case ' ': case 'k': this.paused ? void this.play() : this.pause(); break;
        case 'ArrowLeft': this.seek(this.currentTime - 5); break;
        case 'ArrowRight': this.seek(this.currentTime + 5); break;
        case 'ArrowUp': this.setVolume(this.volume + 0.1); break;
        case 'ArrowDown': this.setVolume(this.volume - 0.1); break;
        case 'm': this.setMuted(!this.muted); break;
        case 'f': this.toggleFullscreen(); break;
        case 's': this.setDataSaver(!this.dataSaver); break;
        case 'c': this.openMenu('captions'); break;
        case 'Escape': if (this.menuOpen) this.closeMenu(); else return; break;
        default: return;
      }
      e.preventDefault();
      this.showChrome();
    });
  }

  private wireChromeVisibility(): void {
    const show = () => this.showChrome();
    for (const evt of ['pointermove', 'pointerdown', 'focusin']) {
      this.root.addEventListener(evt, show);
    }
    this.root.addEventListener('pointerleave', () => {
      if (!this.paused && !this.menuOpen) this.setChrome(false);
    });
    this.showChrome();
  }

  private showChrome(): void {
    this.setChrome(true);
    if (this.idleTimer) clearTimeout(this.idleTimer);
    this.idleTimer = setTimeout(() => {
      if (!this.paused && !this.menuOpen && !this.scrubbing) this.setChrome(false);
    }, this.idleMs);
  }

  private setChrome(on: boolean): void {
    if (this.chromeless) return;
    this.root.setAttribute('data-chrome', on ? 'on' : 'off');
  }

  // ---- rendering --------------------------------------------------------

  private render(): void {
    const u = this.ui;
    const t = this.t;
    const state = this.state;

    const playing = !this.paused && state !== 'ended';
    const playIcon = state === 'ended' ? 'replay' : playing ? 'pause' : 'play';
    u.playBtn.innerHTML = icons[playIcon];
    u.playBtn.setAttribute('aria-label', t(playIcon));
    u.playBtn.title = t(playIcon);
    u.bigBtn.innerHTML = icons[state === 'ended' ? 'replay' : 'play'];
    u.bigBtn.hidden = this.chromeless || playing || state === 'loading' || state === 'buffering' || state === 'error';
    u.spinner.hidden = !(state === 'loading' || state === 'buffering');
    u.centre.hidden = this.chromeless || (u.bigBtn.hidden && u.spinner.hidden);

    const volIcon = this.muted || this.volume === 0 ? 'mute' : this.volume < 0.5 ? 'volumeLow' : 'volume';
    u.muteBtn.innerHTML = icons[volIcon];
    u.muteBtn.setAttribute('aria-label', this.muted ? t('unmute') : t('mute'));
    u.volume.value = String(this.muted ? 0 : this.volume);

    const cur = this.currentTime;
    const dur = this.duration;
    u.time.innerHTML = `<b>${formatTime(cur)}</b> / ${formatTime(dur)}`;
    const pct = dur ? (cur / dur) * 100 : 0;
    u.played.style.width = `${pct}%`;
    u.knob.style.left = `${pct}%`;
    u.buffered.style.width = `${this.bufferedPercent()}%`;
    u.seek.setAttribute('aria-valuemax', String(Math.floor(dur)));
    u.seek.setAttribute('aria-valuenow', String(Math.floor(cur)));
    u.seek.setAttribute('aria-valuetext', `${formatTime(cur)} / ${formatTime(dur)}`);

    u.saver.setAttribute('aria-pressed', String(this.dataSaver));
    u.saver.setAttribute('aria-label', this.dataSaver ? t('dataSaverOn') : t('dataSaverOff'));
    u.saver.title = t('dataSaverHint');
    u.saverLabel.textContent = t('dataSaver');
    const mb = Math.round(this.estimatedMbPerHour());
    u.saverRate.textContent = mb ? `${formatNumber(this.lang, mb)} ${t('megabytes')}` : '';

    u.capsBtn.hidden = this.captionTracks().length === 0;
    u.fsBtn.innerHTML = icons[this.isFullscreen() ? 'fullscreenExit' : 'fullscreen'];
    u.fsBtn.setAttribute('aria-label', this.isFullscreen() ? t('exitFullscreen') : t('fullscreen'));

    this.renderPanel();
    if (state === 'paused' || state === 'ended' || state === 'error') this.setChrome(true);
  }

  private relabel(): void {
    const u = this.ui;
    u.seek.setAttribute('aria-label', this.t('seek'));
    u.volume.setAttribute('aria-label', this.t('volume'));
    u.menuBtn.setAttribute('aria-label', this.t('settings'));
    u.menuBtn.title = this.t('settings');
    u.capsBtn.setAttribute('aria-label', this.t('captions'));
    u.menu.setAttribute('aria-label', this.t('settings'));
  }

  private renderPanel(): void {
    const u = this.ui;
    const err = this.error;
    if (!err) { u.panel.hidden = true; return; }
    u.panel.hidden = false;
    u.panelTitle.textContent = this.t('errTitle');
    u.panelBody.textContent = this.t(err.messageKey);
    u.panelAction.textContent = this.t('retry');
    // An expired link cannot be retried from inside the player; the host reloads it.
    u.panelAction.hidden = err.kind === 'expired' && !this.canReload();
  }

  private canReload(): boolean {
    return typeof location !== 'undefined';
  }

  private onPanelAction(): void {
    const err = this.error;
    if (err?.kind === 'expired') { location.reload(); return; }
    void this.refresh(this.src);
  }

  private bufferedPercent(): number {
    const b = this.video.buffered;
    const dur = this.duration;
    if (!b.length || !dur) return 0;
    return (b.end(b.length - 1) / dur) * 100;
  }

  // ---- preview thumbnails ----------------------------------------------

  private showPreview(ratio: number): void {
    const tiles = this.thumbnailTiles;
    const u = this.ui;
    const at = ratio * this.duration;
    u.previewTime.textContent = formatTime(at);
    const tile = tiles.length ? tileAt(tiles, at) : undefined;
    if (tile) {
      u.previewImg.style.width = `${tile.w}px`;
      u.previewImg.style.height = `${tile.h}px`;
      u.previewImg.style.backgroundImage = `url("${tile.url}")`;
      u.previewImg.style.backgroundPosition = `-${tile.x}px -${tile.y}px`;
      u.previewImg.hidden = false;
    } else {
      u.previewImg.hidden = true;
    }
    const width = u.seek.getBoundingClientRect().width;
    const half = (tile?.w ?? 60) / 2;
    const x = Math.min(width - half, Math.max(half, ratio * width));
    u.preview.style.left = `${x}px`;
    u.preview.setAttribute('data-show', '');
  }

  private hidePreview(): void {
    this.ui.preview.removeAttribute('data-show');
  }

  // ---- menu -------------------------------------------------------------

  private openMenu(focus?: 'quality' | 'captions' | 'language'): void {
    const u = this.ui;
    const t = this.t;
    u.menu.innerHTML = '';

    const section = (title: string) => {
      const h = el('div', 'alc-menu-title');
      h.textContent = title;
      u.menu.appendChild(h);
    };
    const item = (label: string, checked: boolean, note: string, onPick: () => void) => {
      const b = el('button', 'alc-item', { type: 'button', role: 'menuitemradio', 'aria-checked': String(checked) });
      const tick = el('span');
      tick.innerHTML = icons.check;
      b.appendChild(tick.firstElementChild as SVGElement);
      const text = el('span');
      text.textContent = label;
      b.appendChild(text);
      if (note) {
        const n = el('span', 'alc-item-note');
        n.textContent = note;
        b.appendChild(n);
      }
      b.addEventListener('click', () => { onPick(); this.closeMenu(); this.showChrome(); });
      u.menu.appendChild(b);
      return b;
    };

    // Quality — the ladder tops out at 720p, so the list shows what exists, no more.
    section(t('quality'));
    let first: HTMLElement | null = item(t('auto'), this.quality === 'auto', '', () => this.setQuality('auto'));
    for (const q of this.qualities()) {
      const mb = Math.round(q.mbPerHour);
      item(
        `${formatNumber(this.lang, q.height)}p`,
        this.quality === q.height,
        `${formatNumber(this.lang, mb)} ${t('megabytes')}`,
        () => this.setQuality(q.height),
      );
    }

    const caps = this.captionTracks();
    if (caps.length) {
      section(t('captions'));
      const capItem = item(t('captionsOff'), false, '', () => this.setCaptions(null));
      if (focus === 'captions') first = capItem;
      for (const c of caps) {
        item(c.label || c.language || '—', Boolean(c.active), '', () => this.setCaptions(c.id));
      }
    }

    section(t('language'));
    const langItem = item('বাংলা', this.lang === 'bn', '', () => this.setLang('bn'));
    item('English', this.lang === 'en', '', () => this.setLang('en'));
    if (focus === 'language') first = langItem;

    const hint = el('div', 'alc-hint', { id: 'alc-saver-hint' });
    hint.textContent = t('dataSaverHint');
    u.menu.appendChild(hint);

    u.menu.hidden = false;
    this.menuOpen = true;
    u.menuBtn.setAttribute('aria-expanded', 'true');
    first?.focus();
    this.showChrome();
  }

  private closeMenu(): void {
    this.ui.menu.hidden = true;
    this.menuOpen = false;
    this.ui.menuBtn.setAttribute('aria-expanded', 'false');
  }

  // ---- misc -------------------------------------------------------------

  private toast(message: string, tone: 'info' | 'warn' = 'info'): void {
    const u = this.ui;
    u.toastText.textContent = message;
    u.toast.setAttribute('data-tone', tone);
    u.toast.setAttribute('data-show', '');
    u.live.textContent = message;
    if (this.toastTimer) clearTimeout(this.toastTimer);
    this.toastTimer = setTimeout(() => u.toast.removeAttribute('data-show'), 3200);
  }

  /** One transform every seven seconds, so no static overlay or crop can cover it. */
  private drift(): void {
    const wm = this.ui.watermark;
    if (!wm) return;
    const x = Math.random() * Math.max(0, this.ui.stage.clientWidth - wm.offsetWidth);
    const y = Math.random() * Math.max(0, this.ui.stage.clientHeight - wm.offsetHeight);
    wm.style.transform = `translate(${Math.round(x)}px,${Math.round(y)}px)`;
  }

  private isFullscreen(): boolean {
    return document.fullscreenElement === this.root;
  }

  toggleFullscreen(): void {
    if (this.isFullscreen()) {
      void document.exitFullscreen?.();
    } else if (this.root.requestFullscreen) {
      void this.root.requestFullscreen().catch(() => this.toast(this.t('errGeneric'), 'warn'));
    } else {
      // Old iOS WebView has no Fullscreen API on elements, only on the video.
      const legacy = this.video as HTMLVideoElement & { webkitEnterFullscreen?: () => void };
      legacy.webkitEnterFullscreen?.();
    }
  }

  override async destroy(): Promise<void> {
    if (this.idleTimer) clearTimeout(this.idleTimer);
    if (this.toastTimer) clearTimeout(this.toastTimer);
    if (this.driftTimer) clearInterval(this.driftTimer);
    await super.destroy();
    this.root.remove();
  }
}
