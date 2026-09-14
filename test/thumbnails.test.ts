import { test } from 'node:test';
import assert from 'node:assert/strict';
import { parseSpriteVTT, parseVttTime, tileAt } from '../src/thumbnails.ts';

const VTT_URL = 'https://play.example/playback/t1/a1/sprite.vtt?exp=1000&kid=k1&sig=abc';
const VTT = `WEBVTT

00:00:00.000 --> 00:00:05.000
sprite.jpg#xywh=0,0,160,90

00:00:05.000 --> 00:00:10.000
sprite.jpg#xywh=160,0,160,90

01:00:00.000 --> 01:00:05.000
sprite.jpg#xywh=320,0,160,90
`;

test('vtt timestamps parse with and without an hours field', () => {
  assert.equal(parseVttTime('00:00:05.500'), 5.5);
  assert.equal(parseVttTime('01:02:03.000'), 3723);
  assert.equal(parseVttTime('02:03'), 123);
  assert.ok(Number.isNaN(parseVttTime('nope')));
});

test('tiles carry the signature the origin did not add to the VTT', () => {
  const tiles = parseSpriteVTT(VTT, VTT_URL);
  assert.equal(tiles.length, 3);
  const first = tiles[0]!;
  assert.deepEqual([first.x, first.y, first.w, first.h], [0, 0, 160, 90]);
  const u = new URL(first.url);
  assert.equal(u.pathname, '/playback/t1/a1/sprite.jpg');
  assert.equal(u.searchParams.get('sig'), 'abc');
});

test('malformed cues are skipped, not thrown on', () => {
  const tiles = parseSpriteVTT('WEBVTT\n\nbroken\n\n00:00:00.000 --> junk\nsprite.jpg#xywh=0,0,1,1\n', VTT_URL);
  assert.equal(tiles.length, 0);
});

test('CRLF input parses the same', () => {
  assert.equal(parseSpriteVTT(VTT.replace(/\n/g, '\r\n'), VTT_URL).length, 3);
});

test('tile lookup lands in the right cue and clamps past the end', () => {
  const tiles = parseSpriteVTT(VTT, VTT_URL);
  assert.equal(tileAt(tiles, 0)!.x, 0);
  assert.equal(tileAt(tiles, 7)!.x, 160);
  assert.equal(tileAt(tiles, 99_999)!.x, 320);
});
