// The player core: shaka plus everything Bangladesh needs on top of it.
// No DOM chrome here — src/ui.ts draws that against this API.

import { loadShaka, type Shaka, type ShakaPlayer, type ShakaTextTrack } from './shaka.ts';
import { classifyShakaError, expiredSignature, unsupportedBrowser, type ClassifiedError } from './errors.ts';
import {
  abrRestrictions, classifyNetwork, estimateBpsForHeight, getConnection,
  mbPerHour, shouldDefaultDataSaver, type NetworkKind,
} from './network.ts';
import { isExpired, msUntilExpiry, siblingURL, viewerLabel } from './signed-url.ts';
import { newSessionId, sendBeacon } from './beacon.ts';
import { negotiateLang, type Lang } from './i18n.ts';
import { loadThumbnails, type Tile } from './thumbnails.ts';

export type PlayerState = 'idle' | 'loading' | 'ready' | 'playing' | 'paused' | 'buffering' | 'ended' | 'error';

export interface AlchemistOptions {
  /** Signed HLS or DASH URL from `GET /v1/assets/{id}`. */
  src: string;
  poster?: string;
  /** Signed sprite.vtt URL. Omit to derive it from `src`. */
  thumbnails?: string | false;
  /** 'auto' negotiates from the browser; default lands on Bangla. */
  lang?: Lang | 'auto';
  /** 'auto' turns it on for cellular and Save-Data, off otherwise. */
  dataSaver?: boolean | 'auto';
  /** Top rung of the tenant's ladder. BD default is 720p. */
  maxHeight?: number;
  autoplay?: boolean;
  muted?: boolean;
  loop?: boolean;
  /** Override when the host knows it is serving from BDIX. Never guessed. */
  network?: NetworkKind;
  country?: string;
  /**
   * True while the broadcast is going out. The live playlist is EXT-X-PLAYLIST-TYPE:EVENT
   * so a viewer can seek back to its start, and shaka reads EVENT as a growing VOD --
   * its own isLive() is false for the whole broadcast. Omit and shaka decides.
   */
  live?: boolean;
  /** Set false to send no QoE telemetry at all. */
  beacon?: boolean;
  /** How long before expiry to ask the host for a fresh URL. */
  refreshLeadMs?: number;
}

export interface QualityOption {
  height: number;
  bandwidth: number;
  mbPerHour: number;
  id: number;
}

export interface PlayerStats {
  sessionId: string;
  startupMs: number | null;
  rebufferCount: number;
  rebufferMs: number;
  avgBitrateBps: number | null;
  network: NetworkKind;
}

const DEFAULTS = { maxHeight: 720, refreshLeadMs: 120_000 };

export class AlchemistPlayer extends EventTarget {
  readonly video: HTMLVideoElement;

  private opts: Required<Pick<AlchemistOptions, 'maxHeight' | 'refreshLeadMs'>> & AlchemistOptions;
  private player: ShakaPlayer | null = null;
  private _state: PlayerState = 'idle';
  private _lang: Lang;
  private _dataSaver: boolean;
  private _error: ClassifiedError | null = null;
  private _tiles: Tile[] = [];
  private _quality: number | 'auto' = 'auto';

  private expiryTimer: ReturnType<typeof setTimeout> | null = null;
  private sampler: ReturnType<typeof setInterval> | null = null;
  private loadStartedAt = 0;
  private bufferingSince = 0;
  private bitrateSamples: number[] = [];
  private endSent = false;
  private destroyed = false;
  private stats: PlayerStats;
  private abort = new AbortController();

  constructor(el: HTMLElement | HTMLVideoElement, options: AlchemistOptions) {
    super();
    this.opts = { ...DEFAULTS, ...options };

    this.video = el instanceof HTMLVideoElement ? el : document.createElement('video');
    if (this.video !== el) el.appendChild(this.video);
    this.video.playsInline = true;
    this.video.preload = 'metadata';
    this.video.crossOrigin = 'anonymous';
    if (options.muted) this.video.muted = true;
    if (options.loop) this.video.loop = true;
    if (options.poster) this.video.poster = options.poster;

    this._lang = options.lang && options.lang !== 'auto'
      ? options.lang
      : negotiateLang(null, typeof navigator !== 'undefined' ? navigator.languages : []);

    const conn = getConnection();
    this._dataSaver = options.dataSaver === undefined || options.dataSaver === 'auto'
      ? shouldDefaultDataSaver(conn)
      : options.dataSaver;

    this.stats = {
      sessionId: newSessionId(),
      startupMs: null,
      rebufferCount: 0,
      rebufferMs: 0,
      avgBitrateBps: null,
      network: options.network ?? classifyNetwork(conn),
    };

    this.wireVideoEvents();
    this.wirePageEvents();
    // Deferred one microtask so a caller that subscribes right after `new` still
    // catches an error raised on the very first line of boot — an already-dead
    // signature, for one.
    queueMicrotask(() => void this.boot());
  }

  // ---- public API -------------------------------------------------------

  get src(): string { return this.opts.src; }
  get state(): PlayerState { return this._state; }
  get lang(): Lang { return this._lang; }
  get dataSaver(): boolean { return this._dataSaver; }
  get error(): ClassifiedError | null { return this._error; }
  get thumbnailTiles(): readonly Tile[] { return this._tiles; }
  get currentTime(): number { return this.video.currentTime; }
  get duration(): number { return Number.isFinite(this.video.duration) ? this.video.duration : 0; }

  /** True while a broadcast is still going out. The host overrides; shaka answers otherwise. */
  get isLive(): boolean { return this.opts.live ?? this.player?.isLive() ?? false; }
  get paused(): boolean { return this.video.paused; }
  get muted(): boolean { return this.video.muted; }
  get volume(): number { return this.video.volume; }
  get quality(): number | 'auto' { return this._quality; }
  /** The signed viewer label, when the URL carries one. Drawn as a drifting watermark. */
  get viewerLabel(): string | null { return viewerLabel(this.opts.src); }

  async play(): Promise<void> {
    try {
      await this.video.play();
    } catch {
      // Autoplay refused. Muted retry is the only thing a browser will allow.
      if (!this.video.muted) {
        this.video.muted = true;
        try { await this.video.play(); } catch { /* viewer will press play */ }
      }
    }
  }

  pause(): void { this.video.pause(); }
  seek(seconds: number): void { this.video.currentTime = Math.max(0, seconds); }
  setMuted(m: boolean): void { this.video.muted = m; }
  setVolume(v: number): void { this.video.volume = Math.min(1, Math.max(0, v)); this.video.muted = v === 0; }

  setLang(lang: Lang): void {
    if (lang === this._lang) return;
    this._lang = lang;
    this.emit('languagechange', { lang });
  }

  setDataSaver(on: boolean): void {
    if (on === this._dataSaver) return;
    this._dataSaver = on;
    this.applyAbr();
    this.emit('datasaverchange', { dataSaver: on, mbPerHour: this.estimatedMbPerHour() });
  }

  /** `'auto'` hands control back to ABR. A number pins the nearest rung at or below it. */
  setQuality(height: number | 'auto'): void {
    this._quality = height;
    if (!this.player) return;
    if (height === 'auto') {
      this.player.configure('abr.enabled', true);
      this.applyAbr();
    } else {
      const track = this.qualities()
        .filter((q) => q.height <= height)
        .sort((a, b) => b.height - a.height)[0] ?? this.qualities()[0];
      if (!track) return;
      this.player.configure('abr.enabled', false);
      const variant = this.player.getVariantTracks().find((t) => t.id === track.id);
      if (variant) this.player.selectVariantTrack(variant, true);
    }
    this.emit('qualitychange', { quality: this._quality });
  }

  qualities(): QualityOption[] {
    if (!this.player) return [];
    const seen = new Map<number, QualityOption>();
    for (const t of this.player.getVariantTracks()) {
      const h = t.height ?? 0;
      if (!h) continue;
      const existing = seen.get(h);
      if (!existing || t.bandwidth < existing.bandwidth) {
        seen.set(h, { height: h, bandwidth: t.bandwidth, mbPerHour: mbPerHour(t.bandwidth), id: t.id });
      }
    }
    return [...seen.values()].sort((a, b) => b.height - a.height);
  }

  captionTracks(): ShakaTextTrack[] {
    return this.player?.getTextTracks() ?? [];
  }

  /** `null` turns captions off. */
  setCaptions(id: number | null): void {
    if (!this.player) return;
    if (id === null) { this.player.selectTextTrack(null); this.emit('captionschange', { id: null }); return; }
    const track = this.captionTracks().find((t) => t.id === id);
    if (!track) return;
    this.player.selectTextTrack(track);
    this.emit('captionschange', { id });
  }

  /** What Data Saver actually promises, in MB per hour at the current cap. */
  estimatedMbPerHour(): number {
    const rungs = this.qualities().map((q) => ({ height: q.height, bandwidth: q.bandwidth }));
    const ceiling = this._dataSaver ? 360 : this.opts.maxHeight;
    return mbPerHour(estimateBpsForHeight(ceiling, rungs.length ? rungs : undefined));
  }

  /**
   * Swap in a freshly signed URL, in place, keeping the playhead. This is what a
   * host calls after a `needs-refresh` event.
   */
  async refresh(src: string, extra?: { poster?: string; thumbnails?: string }): Promise<void> {
    const at = this.video.currentTime;
    const wasPlaying = !this.video.paused;
    this.opts.src = src;
    if (extra?.poster) this.opts.poster = extra.poster;
    if (extra?.thumbnails) this.opts.thumbnails = extra.thumbnails;
    this._error = null;
    this.endSent = false;
    // A signature that was already dead at construction means boot never got as far
    // as creating the shaka player, so a refresh has to start it now.
    if (!this.player) { await this.boot(at); return; }
    await this.loadSrc(at);
    if (wasPlaying) await this.play();
  }

  getStats(): PlayerStats {
    return { ...this.stats, avgBitrateBps: this.averageBitrate() };
  }

  async destroy(): Promise<void> {
    if (this.destroyed) return;
    this.destroyed = true;
    this.flushBeacon();
    this.clearTimers();
    this.abort.abort();
    try { await this.player?.destroy(); } catch { /* already gone */ }
    this.player = null;
  }

  // ---- internals --------------------------------------------------------

  private emit(type: string, detail?: unknown): void {
    this.dispatchEvent(new CustomEvent(type, { detail }));
  }

  private setState(next: PlayerState): void {
    if (this._state === next) return;
    this._state = next;
    this.emit('statechange', { state: next });
  }

  private fail(err: ClassifiedError): void {
    this._error = err;
    this.setState('error');
    if (this.opts.beacon !== false) {
      sendBeacon(this.opts.src, { ...this.getStats(), errorCode: err.code, country: this.opts.country });
    }
    if (err.kind === 'expired') this.emit('needs-refresh', { reason: 'expired' });
    this.emit('error', err);
  }

  private async boot(startAt?: number): Promise<void> {
    this.setState('loading');
    if (isExpired(this.opts.src)) { this.fail(expiredSignature()); return; }

    let lib: Shaka;
    try {
      lib = await loadShaka(this.opts.src);
    } catch {
      this.fail(unsupportedBrowser());
      return;
    }
    if (this.destroyed) return;
    if (!lib.Player.isBrowserSupported()) { this.fail(unsupportedBrowser()); return; }

    const player = new lib.Player();
    this.player = player;
    await player.attach(this.video);

    player.configure({
      abr: {
        enabled: true,
        useNetworkInformation: true,
        // A pessimistic first guess beats starting at 720p and immediately
        // dropping: the wasted segment is already paid for by then.
        defaultBandwidthEstimate: 700_000,
        restrictToElementSize: true,
        restrictions: abrRestrictions(this._dataSaver, this.opts.maxHeight),
      },
      // Top-level restrictions are the hard filter; abr.restrictions alone is only a
      // preference the ABR manager may abandon. Data Saver has to be a real ceiling.
      restrictions: abrRestrictions(this._dataSaver, this.opts.maxHeight),
      streaming: {
        bufferingGoal: this._dataSaver ? 10 : 18,
        rebufferingGoal: 3,
        bufferBehind: 20,
        // BD mobile drops packets; giving up after two tries strands a viewer
        // who would have recovered on the third.
        retryParameters: { maxAttempts: 5, baseDelay: 500, backoffFactor: 1.8, timeout: 25_000, stallTimeout: 8_000 },
      },
      manifest: { retryParameters: { maxAttempts: 4, baseDelay: 500, backoffFactor: 2, timeout: 20_000 } },
      drm: { retryParameters: { maxAttempts: 3, baseDelay: 500, backoffFactor: 2, timeout: 20_000 } },
    });

    player.addEventListener('error', (e) => {
      const detail = (e as unknown as { detail?: { category?: number; code?: number; data?: unknown[] } }).detail;
      this.fail(classifyShakaError(detail, navigator.onLine !== false));
    });
    player.addEventListener('buffering', (e) => {
      this.onBuffering(Boolean((e as unknown as { buffering?: boolean }).buffering));
    });
    player.addEventListener('trackschanged', () => this.emit('tracksavailable', { qualities: this.qualities() }));
    player.addEventListener('adaptation', () => this.emit('adaptation', { qualities: this.qualities() }));

    await this.loadSrc(startAt);
    void this.fetchThumbnails();
  }

  private async loadSrc(startAt?: number): Promise<void> {
    if (!this.player) return;
    this.setState('loading');
    this.loadStartedAt = Date.now();
    this.armExpiryWatch();
    // The EME Clear Key licence is a sibling of the manifest, so the same signature
    // authorizes it. Shaka POSTs to this URI verbatim, query included — no filter needed.
    this.player.configure('drm.servers', { 'org.w3.clearkey': siblingURL(this.opts.src, 'key') });
    try {
      await this.player.load(this.opts.src, startAt);
      if (this.destroyed) return;
      this.applyAbr();
      this.startSampler();
      this.setState('ready');
      this.emit('ready', { qualities: this.qualities(), mbPerHour: this.estimatedMbPerHour() });
      if (this.opts.autoplay) void this.play();
    } catch (e) {
      if (this.destroyed) return;
      const err = e as { category?: number; code?: number; data?: unknown[] };
      // A load that fails on an already-dead signature is expiry, whatever shaka says.
      if (isExpired(this.opts.src)) this.fail(expiredSignature());
      else this.fail(classifyShakaError(err, navigator.onLine !== false));
    }
  }

  private applyAbr(): void {
    if (!this.player) return;
    const limits = abrRestrictions(this._dataSaver, this.opts.maxHeight);
    this.player.configure('restrictions', limits);
    this.player.configure('abr.restrictions', limits);
    // Toggling Data Saver on must stop spending the viewer's money now, not at the
    // end of a buffer they already paid for.
    this.player.configure('abr.clearBufferSwitch', this._dataSaver);
    this.player.configure('streaming.bufferingGoal', this._dataSaver ? 10 : 18);
    if (this._quality === 'auto') this.player.configure('abr.enabled', true);
  }

  private armExpiryWatch(): void {
    if (this.expiryTimer) clearTimeout(this.expiryTimer);
    const left = msUntilExpiry(this.opts.src);
    if (!Number.isFinite(left)) return;
    const lead = Math.max(0, left - this.opts.refreshLeadMs);
    this.expiryTimer = setTimeout(() => {
      // Ask early and let the host swap the URL mid-playback, instead of dying
      // on the next segment request.
      this.emit('needs-refresh', { reason: 'expiring', expiresInMs: msUntilExpiry(this.opts.src) });
    }, lead);
  }

  private wireVideoEvents(): void {
    const v = this.video;
    const on = (name: string, fn: () => void) => v.addEventListener(name, fn);
    on('playing', () => {
      if (this.stats.startupMs === null && this.loadStartedAt) {
        this.stats.startupMs = Date.now() - this.loadStartedAt;
      }
      this.onBuffering(false);
      this.setState('playing');
      this.emit('playing');
    });
    on('pause', () => { if (this._state !== 'ended' && this._state !== 'error') this.setState('paused'); this.emit('pause'); });
    on('ended', () => { this.setState('ended'); this.flushBeacon(); this.emit('ended'); });
    on('timeupdate', () => this.emit('timeupdate', { currentTime: v.currentTime, duration: this.duration }));
    on('progress', () => this.emit('progress'));
    on('volumechange', () => this.emit('volumechange', { volume: v.volume, muted: v.muted }));
    on('waiting', () => this.onBuffering(true));
  }

  private wirePageEvents(): void {
    if (typeof window === 'undefined') return;
    const { signal } = this.abort;
    // pagehide fires on mobile Safari where unload does not.
    window.addEventListener('pagehide', () => this.flushBeacon(), { signal });
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'hidden') this.flushBeacon();
    }, { signal });
    window.addEventListener('offline', () => {
      if (this._state !== 'error') this.emit('offline');
    }, { signal });
    window.addEventListener('online', () => this.emit('online'), { signal });
  }

  private onBuffering(active: boolean): void {
    if (active) {
      if (this.bufferingSince) return;
      this.bufferingSince = Date.now();
      this.stats.rebufferCount++;
      if (this._state !== 'error') this.setState('buffering');
      this.emit('buffering', { buffering: true });
    } else {
      if (this.bufferingSince) {
        this.stats.rebufferMs += Date.now() - this.bufferingSince;
        this.bufferingSince = 0;
      }
      this.emit('buffering', { buffering: false });
    }
  }

  private startSampler(): void {
    if (this.sampler) clearInterval(this.sampler);
    this.sampler = setInterval(() => {
      const bw = this.player?.getStats()?.streamBandwidth;
      if (typeof bw === 'number' && bw > 0 && !this.video.paused) this.bitrateSamples.push(bw);
    }, 5000);
  }

  private averageBitrate(): number | null {
    if (!this.bitrateSamples.length) return null;
    return Math.round(this.bitrateSamples.reduce((a, b) => a + b, 0) / this.bitrateSamples.length);
  }

  private async fetchThumbnails(): Promise<void> {
    if (this.opts.thumbnails === false) return;
    const url = this.opts.thumbnails ?? siblingURL(this.opts.src, 'sprite.vtt');
    const tiles = await loadThumbnails(url, this.abort.signal);
    if (this.destroyed || !tiles.length) return;
    this._tiles = tiles;
    this.emit('thumbnails', { count: tiles.length });
  }

  private flushBeacon(): void {
    if (this.endSent || this.opts.beacon === false) return;
    if (this.stats.startupMs === null && !this.stats.rebufferCount) return; // nothing ever played
    this.endSent = true;
    sendBeacon(this.opts.src, { ...this.getStats(), errorCode: null, country: this.opts.country });
  }

  private clearTimers(): void {
    if (this.expiryTimer) clearTimeout(this.expiryTimer);
    if (this.sampler) clearInterval(this.sampler);
    this.expiryTimer = this.sampler = null;
  }
}
