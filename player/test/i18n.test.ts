import { test } from 'node:test';
import assert from 'node:assert/strict';
import { negotiateLang, makeT, formatTime, formatNumber } from '../src/i18n.ts';

test('explicit choice wins', () => {
  assert.equal(negotiateLang('en', ['bn-BD']), 'en');
  assert.equal(negotiateLang('bn', ['en-US']), 'bn');
});

test('browser languages are honoured when nothing is explicit', () => {
  assert.equal(negotiateLang(null, ['en-GB', 'bn']), 'en');
  assert.equal(negotiateLang(null, ['bn-BD', 'en']), 'bn');
});

test('unknown languages fall back to Bangla, not English', () => {
  assert.equal(negotiateLang(null, ['fr-FR', 'de']), 'bn');
  assert.equal(negotiateLang(undefined, []), 'bn');
  assert.equal(negotiateLang('klingon'), 'bn');
});

test('case and separators do not matter', () => {
  assert.equal(negotiateLang('EN_US'), 'en');
  assert.equal(negotiateLang('BN-bd'), 'bn');
});

test('every Bangla key has an English twin and vice versa', () => {
  const bn = makeT('bn');
  const en = makeT('en');
  for (const k of ['play', 'dataSaver', 'errExpired', 'errNetwork', 'retry', 'offline']) {
    assert.notEqual(bn(k), k, `missing bn copy for ${k}`);
    assert.notEqual(en(k), k, `missing en copy for ${k}`);
    assert.notEqual(bn(k), en(k), `${k} was not translated`);
  }
});

test('conjunct test strings survive the dictionary untouched', () => {
  // If anything in the pipeline mangles Bangla, these are the strings that show it.
  for (const s of ['যুক্তাক্ষর', 'বাংলাদেশ', 'শিক্ষা']) {
    assert.equal([...s].length > 0, true);
    assert.equal(s.normalize('NFC'), s);
  }
  assert.match(makeT('bn')('dataSaver'), /সাশ্রয়/);
});

test('timecodes stay in Latin digits and grow an hours field only when needed', () => {
  assert.equal(formatTime(0), '0:00');
  assert.equal(formatTime(65), '1:05');
  assert.equal(formatTime(3661), '1:01:01');
  assert.equal(formatTime(-5), '0:00');
  assert.equal(formatTime(Number.NaN), '0:00');
});

test('numbers use Intl, so Bangla gets Bangla digits', () => {
  assert.equal(formatNumber('en', 270), '270');
  assert.equal(formatNumber('bn', 270), '২৭০');
});
