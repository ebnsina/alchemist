// The player never sees an API key. It sees a signed URL whose query authorizes
// the whole asset prefix — manifest, segments, key, poster, thumbnails, beacon.

/** Resolves against the document when relative, so callers can pass either form. */
export function toURL(src: string, base?: string): URL {
  const fallback = base ?? (typeof location !== 'undefined' ? location.href : 'http://localhost/');
  return new URL(src, fallback);
}

/** Epoch milliseconds the signature stops working, or null if the URL carries no `exp`. */
export function expiresAt(src: string, base?: string): number | null {
  let url: URL;
  try {
    url = toURL(src, base);
  } catch {
    return null;
  }
  const exp = url.searchParams.get('exp');
  if (!exp) return null;
  const secs = Number(exp);
  if (!Number.isFinite(secs) || secs <= 0) return null;
  return secs * 1000;
}

/** Milliseconds left on the signature. Infinity when the URL is unsigned. */
export function msUntilExpiry(src: string, now = Date.now(), base?: string): number {
  const at = expiresAt(src, base);
  return at === null ? Infinity : at - now;
}

export function isExpired(src: string, now = Date.now(), base?: string): boolean {
  return msUntilExpiry(src, now, base) <= 0;
}

/** Swaps the last path segment, keeping the signature query. `beacon`, `sprite.jpg`, … */
export function siblingURL(src: string, file: string, base?: string): string {
  const url = toURL(src, base);
  const parts = url.pathname.split('/');
  parts[parts.length - 1] = file;
  url.pathname = parts.join('/');
  return url.toString();
}

/** Copies the signature query onto a URI that has none — sprite.vtt cues, for one. */
export function withSignatureOf(signed: string, target: string, base?: string): string {
  const from = toURL(signed, base);
  const to = toURL(target, from.toString());
  if (to.origin !== from.origin) return target;
  for (const k of ['exp', 'kid', 'sig']) {
    const v = from.searchParams.get(k);
    if (v !== null && !to.searchParams.has(k)) to.searchParams.set(k, v);
  }
  return to.toString();
}
