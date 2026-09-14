// Shaka is loaded on demand and split by manifest type: the HLS-only build is
// ~65 KB (gzipped) lighter than the combined one, and on a metered BD connection
// that is the viewer's money, not a build statistic.

import type shakaTypes from 'shaka-player/dist/shaka-player.compiled.js';

export type Shaka = typeof shakaTypes;
export type ShakaPlayer = InstanceType<Shaka['Player']>;
export type ShakaTextTrack = Parameters<NonNullable<ShakaPlayer['selectTextTrack']>>[0] extends infer T ? NonNullable<T> : never;

let cached: Promise<Shaka> | null = null;

export function loadShaka(src: string): Promise<Shaka> {
  if (cached) return cached;
  const wantsDash = /\.mpd(\?|$)/i.test(src);
  cached = (wantsDash
    ? import('shaka-player/dist/shaka-player.dash.js')
    : import('shaka-player/dist/shaka-player.hls.js')
  ).then((m) => {
    const lib = ((m as { default?: unknown }).default ?? m) as Shaka;
    lib.polyfill.installAll();
    return lib;
  });
  return cached;
}
