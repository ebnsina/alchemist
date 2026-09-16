// QoE telemetry. The endpoint always answers 204 and the viewer must never learn
// this exists, so nothing here retries, throws, or reaches the UI.

import { siblingURL } from './signed-url.ts';
import type { NetworkKind } from './network.ts';

export interface QoEStats {
  sessionId: string;
  startupMs?: number | null;
  rebufferCount?: number;
  rebufferMs?: number;
  avgBitrateBps?: number | null;
  errorCode?: string | null;
  network?: NetworkKind;
  country?: string;
}

export interface BeaconPayload {
  session_id: string;
  startup_ms: number | null;
  rebuffer_count: number;
  rebuffer_ms: number;
  avg_bitrate_bps: number | null;
  error_code: string | null;
  network: NetworkKind;
  country: string;
}

const clampInt = (v: number | null | undefined, lo: number, hi: number): number | null => {
  if (v === null || v === undefined || !Number.isFinite(v)) return null;
  return Math.min(hi, Math.max(lo, Math.round(v)));
};

export function buildPayload(stats: QoEStats): BeaconPayload {
  return {
    session_id: String(stats.sessionId).slice(0, 64),
    startup_ms: clampInt(stats.startupMs, 0, 600_000),
    rebuffer_count: clampInt(stats.rebufferCount, 0, 10_000) ?? 0,
    rebuffer_ms: clampInt(stats.rebufferMs, 0, 86_400_000) ?? 0,
    avg_bitrate_bps: clampInt(stats.avgBitrateBps, 0, 100_000_000),
    error_code: stats.errorCode ? String(stats.errorCode).slice(0, 64) : null,
    network: stats.network ?? 'other',
    country: (stats.country ?? 'BD').slice(0, 8),
  };
}

export function beaconURL(playbackURL: string, base?: string): string {
  return siblingURL(playbackURL, 'beacon', base);
}

export function newSessionId(): string {
  const c = typeof crypto !== 'undefined' ? crypto : undefined;
  if (c?.randomUUID) return c.randomUUID();
  return `s-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

/**
 * Sent as text/plain so it stays a CORS-simple request: no preflight, which the
 * origin's beacon route does not answer. We never read the response, so the
 * missing Access-Control-Allow-Origin on it costs us nothing.
 */
export function sendBeacon(playbackURL: string, stats: QoEStats, base?: string): void {
  let url: string;
  let body: string;
  try {
    url = beaconURL(playbackURL, base);
    body = JSON.stringify(buildPayload(stats));
  } catch {
    return;
  }
  try {
    if (typeof navigator !== 'undefined' && navigator.sendBeacon) {
      if (navigator.sendBeacon(url, new Blob([body], { type: 'text/plain;charset=UTF-8' }))) return;
    }
    if (typeof fetch === 'function') {
      void fetch(url, {
        method: 'POST',
        body,
        keepalive: true,
        mode: 'no-cors',
        headers: { 'Content-Type': 'text/plain;charset=UTF-8' },
      }).catch(() => {});
    }
  } catch {
    // Telemetry is never worth a viewer-visible failure.
  }
}
