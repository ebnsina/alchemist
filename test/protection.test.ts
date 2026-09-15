import { test } from 'node:test';
import assert from 'node:assert/strict';
import { siblingURL, viewerLabel, withSignatureOf } from '../src/signed-url.ts';
import { classifyShakaError } from '../src/errors.ts';
import { makeT } from '../src/i18n.ts';

const SIGNED = 'https://o.test/playback/t1/a1/master.m3u8?exp=1800000000&kid=k7&sig=abc123&wm=STU-2291';

test('the Clear Key licence URL is the manifest sibling, signature intact', () => {
  const key = new URL(siblingURL(SIGNED, 'key'));
  assert.equal(key.pathname, '/playback/t1/a1/key');
  // Shaka POSTs to this URI verbatim, so the signature has to already be on it.
  assert.equal(key.searchParams.get('exp'), '1800000000');
  assert.equal(key.searchParams.get('kid'), 'k7');
  assert.equal(key.searchParams.get('sig'), 'abc123');
});

test('the viewer label is read from the signed URL, never configured separately', () => {
  assert.equal(viewerLabel(SIGNED), 'STU-2291');
  assert.equal(viewerLabel('https://o.test/a/master.m3u8?exp=1&sig=b'), null);
  assert.equal(viewerLabel('https://o.test/a/master.m3u8?wm=%20%20'), null);
  assert.equal(viewerLabel('not a url at all'), null);
});

test('the label travels with the signature onto sprite URIs, or thumbnails 403', () => {
  const tile = withSignatureOf(SIGNED, 'sprite.jpg');
  assert.equal(new URL(tile).searchParams.get('wm'), 'STU-2291');
});

test('a browser with no Clear Key key system gets its own copy, not a raw DRM failure', () => {
  for (const code of [6000, 6001, 6020]) {
    const c = classifyShakaError({ category: 6, code });
    assert.equal(c.code, 'key_system_unavailable');
    assert.equal(c.retryable, false); // Safari will not grow Clear Key on a retry
    for (const lang of ['bn', 'en'] as const) {
      const msg = makeT(lang)(c.messageKey);
      assert.notEqual(msg, c.messageKey);
      assert.doesNotMatch(msg, /\d{4}|shaka|clear ?key|EME|DRM/i);
    }
  }
});

test('a licence that was fetched but refused is still the ordinary DRM message', () => {
  assert.equal(classifyShakaError({ category: 6, code: 6007 }).messageKey, 'errDrm');
});
