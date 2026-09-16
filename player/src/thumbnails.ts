// Scrub previews come from one sprite sheet indexed by a WebVTT file.
// The origin only rewrites .m3u8 and .mpd, so the sprite.jpg URI inside the VTT
// carries no signature and we have to copy ours onto it.

import { withSignatureOf } from './signed-url.ts';

export interface Tile {
  start: number;
  end: number;
  url: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

export function parseVttTime(s: string): number {
  const m = /^(?:(\d+):)?(\d{1,2}):(\d{2})(?:[.,](\d{1,3}))?$/.exec(s.trim());
  if (!m) return NaN;
  const h = Number(m[1] ?? 0);
  const min = Number(m[2]);
  const sec = Number(m[3]);
  const ms = Number((m[4] ?? '0').padEnd(3, '0'));
  return h * 3600 + min * 60 + sec + ms / 1000;
}

/** `vttURL` is the signed sprite.vtt URL; tile image URLs are resolved against it. */
export function parseSpriteVTT(text: string, vttURL: string, base?: string): Tile[] {
  const lines = text.replace(/\r\n?/g, '\n').split('\n');
  const tiles: Tile[] = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i] as string;
    const arrow = line.indexOf('-->');
    if (arrow < 0) continue;
    const start = parseVttTime(line.slice(0, arrow));
    const end = parseVttTime(line.slice(arrow + 3).split(/\s+/).filter(Boolean)[0] ?? '');
    if (!Number.isFinite(start) || !Number.isFinite(end)) continue;

    const payload = (lines[i + 1] ?? '').trim();
    if (!payload) continue;
    i++;

    const hash = payload.indexOf('#xywh=');
    const ref = hash < 0 ? payload : payload.slice(0, hash);
    const nums = hash < 0 ? [] : payload.slice(hash + 6).split(',').map(Number);
    if (nums.length !== 4 || nums.some((n) => !Number.isFinite(n))) continue;

    tiles.push({
      start,
      end,
      url: withSignatureOf(vttURL, ref, base),
      x: nums[0] as number,
      y: nums[1] as number,
      w: nums[2] as number,
      h: nums[3] as number,
    });
  }
  return tiles;
}

export function tileAt(tiles: readonly Tile[], seconds: number): Tile | undefined {
  for (const t of tiles) if (seconds >= t.start && seconds < t.end) return t;
  return tiles[tiles.length - 1];
}

export async function loadThumbnails(vttURL: string, signal?: AbortSignal, base?: string): Promise<Tile[]> {
  try {
    const res = await fetch(vttURL, { signal });
    if (!res.ok) return [];
    return parseSpriteVTT(await res.text(), vttURL, base);
  } catch {
    return []; // Scrub previews are a nicety; never fail playback over them.
  }
}
