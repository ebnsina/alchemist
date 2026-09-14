import { test } from 'node:test';
import assert from 'node:assert/strict';
import { expiresAt, msUntilExpiry, isExpired, siblingURL, withSignatureOf } from '../src/signed-url.ts';

const SIGNED = 'https://play.example/playback/t1/a1/master.m3u8?exp=1000&kid=k1&sig=abc';

test('exp is read as epoch seconds', () => {
  assert.equal(expiresAt(SIGNED), 1_000_000);
  assert.equal(msUntilExpiry(SIGNED, 400_000), 600_000);
});

test('an unsigned URL never expires rather than expiring immediately', () => {
  assert.equal(expiresAt('https://play.example/a.m3u8'), null);
  assert.equal(msUntilExpiry('https://play.example/a.m3u8'), Infinity);
  assert.equal(isExpired('https://play.example/a.m3u8'), false);
});

test('a garbage exp is treated as unsigned, not as expired', () => {
  assert.equal(expiresAt('https://play.example/a.m3u8?exp=nonsense'), null);
  assert.equal(expiresAt('https://play.example/a.m3u8?exp=-5'), null);
});

test('expiry is exact at the boundary', () => {
  assert.equal(isExpired(SIGNED, 999_999), false);
  assert.equal(isExpired(SIGNED, 1_000_000), true);
  assert.equal(isExpired(SIGNED, 1_000_001), true);
});

test('sibling URLs keep the whole signature', () => {
  assert.equal(siblingURL(SIGNED, 'poster.jpg'),
    'https://play.example/playback/t1/a1/poster.jpg?exp=1000&kid=k1&sig=abc');
});

test('the signature is copied onto same-origin URIs that lack one', () => {
  const out = new URL(withSignatureOf(SIGNED, 'sprite.jpg'));
  assert.equal(out.pathname, '/playback/t1/a1/sprite.jpg');
  assert.equal(out.searchParams.get('sig'), 'abc');
});

test('a cross-origin URI is left alone — we do not leak the signature', () => {
  assert.equal(withSignatureOf(SIGNED, 'https://cdn.other/x.jpg'), 'https://cdn.other/x.jpg');
});
