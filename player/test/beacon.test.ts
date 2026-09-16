import { test } from 'node:test';
import assert from 'node:assert/strict';
import { buildPayload, beaconURL, newSessionId } from '../src/beacon.ts';

const SIGNED = 'https://play.example/playback/t1/a1/master.m3u8?exp=1780000000&kid=k1&sig=abc';

test('beacon URL reuses the playback signature verbatim', () => {
  const u = new URL(beaconURL(SIGNED));
  assert.equal(u.pathname, '/playback/t1/a1/beacon');
  assert.equal(u.searchParams.get('exp'), '1780000000');
  assert.equal(u.searchParams.get('kid'), 'k1');
  assert.equal(u.searchParams.get('sig'), 'abc');
});

test('beacon URL works from a relative playback URL too', () => {
  const u = new URL(beaconURL('/playback/t1/a1/manifest.mpd?exp=1&kid=k&sig=s', 'http://localhost:8099/'));
  assert.equal(u.href, 'http://localhost:8099/playback/t1/a1/beacon?exp=1&kid=k&sig=s');
});

test('payload matches the origin contract and clamps like the origin does', () => {
  const p = buildPayload({
    sessionId: 'x'.repeat(200),
    startupMs: 999_999_999,
    rebufferCount: -4,
    rebufferMs: 1234.7,
    avgBitrateBps: 600_000,
    errorCode: null,
    network: 'cellular',
    country: 'BD',
  });
  assert.deepEqual(Object.keys(p).sort(), [
    'avg_bitrate_bps', 'country', 'error_code', 'network',
    'rebuffer_count', 'rebuffer_ms', 'session_id', 'startup_ms',
  ]);
  assert.equal(p.session_id.length, 64);
  assert.equal(p.startup_ms, 600_000);
  assert.equal(p.rebuffer_count, 0);
  assert.equal(p.rebuffer_ms, 1235);
  assert.equal(p.error_code, null);
  assert.equal(p.network, 'cellular');
});

test('missing values become null or zero, never undefined', () => {
  const p = buildPayload({ sessionId: 's1' });
  assert.equal(p.startup_ms, null);
  assert.equal(p.avg_bitrate_bps, null);
  assert.equal(p.error_code, null);
  assert.equal(p.rebuffer_count, 0);
  assert.equal(p.rebuffer_ms, 0);
  assert.equal(p.network, 'other');
  assert.equal(p.country, 'BD');
  assert.equal(JSON.stringify(p).includes('undefined'), false);
});

test('an error code is carried through, truncated', () => {
  assert.equal(buildPayload({ sessionId: 's', errorCode: 'shaka-1001' }).error_code, 'shaka-1001');
  assert.equal(buildPayload({ sessionId: 's', errorCode: 'e'.repeat(100) }).error_code?.length, 64);
});

test('session ids are unique and fit the 64 char limit', () => {
  const a = newSessionId(), b = newSessionId();
  assert.notEqual(a, b);
  assert.ok(a.length <= 64);
});
