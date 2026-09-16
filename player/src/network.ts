// Everything about the viewer's connection and what it costs them.
// 89% of BD internet users are metered, so bitrate is the viewer's money.

export type NetworkKind = 'bdix' | 'cellular' | 'wifi' | 'other';

/** The subset of the Network Information API we rely on. Absent on many WebViews. */
export interface ConnectionLike {
  type?: string;
  effectiveType?: string;
  saveData?: boolean;
  downlink?: number;
}

/** Data Saver pins the ladder here. 360p is the BD default rung. */
export const DATA_SAVER_MAX_HEIGHT = 360;

/** Ladder rungs, in bits per second, matching the backend's BD-mobile profile. */
export const LADDER_BPS: Record<number, number> = {
  144: 150_000,
  240: 300_000,
  360: 600_000,
  480: 1_000_000,
  720: 1_800_000,
  1080: 3_500_000,
};

export function getConnection(nav?: Navigator): ConnectionLike | undefined {
  const n = (nav ?? (typeof navigator !== 'undefined' ? navigator : undefined)) as
    | (Navigator & { connection?: ConnectionLike; mozConnection?: ConnectionLike; webkitConnection?: ConnectionLike })
    | undefined;
  return n?.connection ?? n?.mozConnection ?? n?.webkitConnection;
}

/**
 * What kind of network this is, as far as the browser will admit.
 * BDIX is an origin-side fact the player cannot observe, so it is never guessed —
 * a host that knows better passes `network` explicitly.
 */
export function classifyNetwork(conn?: ConnectionLike): NetworkKind {
  if (!conn) return 'other';
  if (conn.type === 'cellular') return 'cellular';
  if (conn.type === 'wifi' || conn.type === 'ethernet') return 'wifi';
  // No `type` (Chrome desktop, most WebViews) but an effectiveType we can read.
  if (conn.effectiveType && ['slow-2g', '2g', '3g'].includes(conn.effectiveType)) return 'cellular';
  return 'other';
}

/**
 * Data Saver defaults ON for anything that looks metered or slow, OFF otherwise.
 * Unknown connections default OFF: guessing "metered" on a broadband viewer caps
 * them at 360p for no reason, and they have no idea why.
 */
export function shouldDefaultDataSaver(conn?: ConnectionLike): boolean {
  if (!conn) return false;
  if (conn.saveData === true) return true;
  if (classifyNetwork(conn) === 'cellular') return true;
  if (conn.effectiveType && ['slow-2g', '2g', '3g'].includes(conn.effectiveType)) return true;
  return false;
}

/** Megabytes an hour at a given bitrate. What the toggle actually promises. */
export function mbPerHour(bitsPerSecond: number): number {
  if (!Number.isFinite(bitsPerSecond) || bitsPerSecond <= 0) return 0;
  return (bitsPerSecond * 3600) / 8 / 1e6;
}

/**
 * Shaka ABR restrictions for a data-saver state. Capping by height rather than
 * bandwidth keeps it true whatever the tenant's ladder bitrates are.
 */
export function abrRestrictions(dataSaver: boolean, maxHeight: number) {
  return {
    minHeight: 0,
    maxHeight: dataSaver ? Math.min(DATA_SAVER_MAX_HEIGHT, maxHeight) : maxHeight,
    minWidth: 0,
    maxWidth: Infinity,
    minPixels: 0,
    maxPixels: Infinity,
    minFrameRate: 0,
    maxFrameRate: Infinity,
    minBandwidth: 0,
    maxBandwidth: Infinity,
  };
}

/** Nearest ladder rung at or below a height, for the MB/hour estimate. */
export function estimateBpsForHeight(height: number, known?: Array<{ height: number; bandwidth: number }>): number {
  if (known && known.length) {
    const fit = known.filter((v) => v.height <= height).sort((a, b) => b.height - a.height)[0];
    if (fit) return fit.bandwidth;
    const smallest = known.slice().sort((a, b) => a.height - b.height)[0];
    if (smallest) return smallest.bandwidth;
  }
  const rungs = Object.keys(LADDER_BPS).map(Number).sort((a, b) => a - b);
  let pick = rungs[0] as number;
  for (const r of rungs) if (r <= height) pick = r;
  return LADDER_BPS[pick] as number;
}
