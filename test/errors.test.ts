import { test } from 'node:test';
import assert from 'node:assert/strict';
import { classifyShakaError } from '../src/errors.ts';
import { makeT } from '../src/i18n.ts';

test('a 403 on any playback request is an expired signature, not a network blip', () => {
  const c = classifyShakaError({ category: 1, code: 1001, data: ['https://x/seg', 403] });
  assert.equal(c.kind, 'expired');
  assert.equal(c.code, 'playback_not_authorized');
  assert.equal(c.retryable, false); // retrying the same dead URL is a dead end
});

test('404 is its own message', () => {
  assert.equal(classifyShakaError({ category: 1, code: 1001, data: ['u', 404] }).kind, 'notFound');
});

test('timeouts and transport failures are retryable network errors', () => {
  assert.equal(classifyShakaError({ category: 1, code: 1003 }).retryable, true);
  assert.equal(classifyShakaError({ category: 1, code: 1002 }).kind, 'network');
});

test('being offline beats whatever shaka said', () => {
  const c = classifyShakaError({ category: 6, code: 6007 }, false);
  assert.equal(c.kind, 'offline');
  assert.equal(c.retryable, true);
});

test('DRM and codec failures do not promise a retry that cannot work', () => {
  assert.equal(classifyShakaError({ category: 6, code: 6001 }).retryable, false);
  assert.equal(classifyShakaError({ category: 4, code: 4032 }).kind, 'unsupported');
  assert.equal(classifyShakaError({ category: 3, code: 3016 }).kind, 'unsupported');
});

test('an unknown error still produces usable copy in both languages', () => {
  const c = classifyShakaError(undefined);
  for (const lang of ['bn', 'en'] as const) {
    const msg = makeT(lang)(c.messageKey);
    assert.notEqual(msg, c.messageKey);
    assert.doesNotMatch(msg, /\d{4}|shaka|manifest|codec|HTTP/i); // no jargon reaches the viewer
  }
});

test('every classified message key exists in both dictionaries', () => {
  const cases = [
    classifyShakaError({ category: 1, code: 1001, data: ['u', 403] }),
    classifyShakaError({ category: 1, code: 1001, data: ['u', 404] }),
    classifyShakaError({ category: 1, code: 1003 }),
    classifyShakaError({ category: 6, code: 6001 }),
    classifyShakaError({ category: 4, code: 4000 }),
    classifyShakaError({ category: 5, code: 5001 }),
    classifyShakaError({ category: 99, code: 1 }, false),
  ];
  for (const c of cases) {
    for (const lang of ['bn', 'en'] as const) {
      assert.notEqual(makeT(lang)(c.messageKey), c.messageKey, `${lang}: ${c.messageKey}`);
    }
  }
});
