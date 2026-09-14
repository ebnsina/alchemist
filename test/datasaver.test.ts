import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  classifyNetwork, shouldDefaultDataSaver, mbPerHour, abrRestrictions,
  estimateBpsForHeight, DATA_SAVER_MAX_HEIGHT,
} from '../src/network.ts';

test('cellular defaults Data Saver on', () => {
  assert.equal(shouldDefaultDataSaver({ type: 'cellular', effectiveType: '4g' }), true);
});

test('saveData header defaults it on even on wifi', () => {
  assert.equal(shouldDefaultDataSaver({ type: 'wifi', saveData: true }), true);
});

test('wifi and ethernet default it off', () => {
  assert.equal(shouldDefaultDataSaver({ type: 'wifi', effectiveType: '4g' }), false);
  assert.equal(shouldDefaultDataSaver({ type: 'ethernet' }), false);
});

test('slow effectiveType with no type still counts as metered', () => {
  assert.equal(shouldDefaultDataSaver({ effectiveType: '3g' }), true);
  assert.equal(shouldDefaultDataSaver({ effectiveType: 'slow-2g' }), true);
});

test('no Network Information API defaults off rather than guessing metered', () => {
  assert.equal(shouldDefaultDataSaver(undefined), false);
  assert.equal(shouldDefaultDataSaver({}), false);
});

test('network classification', () => {
  assert.equal(classifyNetwork({ type: 'cellular' }), 'cellular');
  assert.equal(classifyNetwork({ type: 'wifi' }), 'wifi');
  assert.equal(classifyNetwork({ effectiveType: '2g' }), 'cellular');
  assert.equal(classifyNetwork({ effectiveType: '4g' }), 'other');
  assert.equal(classifyNetwork(undefined), 'other');
  // BDIX is never guessed from the browser.
  assert.notEqual(classifyNetwork({ type: 'ethernet' }), 'bdix');
});

test('Data Saver caps the ladder at 360p, off restores the tenant ceiling', () => {
  assert.equal(abrRestrictions(true, 720).maxHeight, DATA_SAVER_MAX_HEIGHT);
  assert.equal(abrRestrictions(false, 720).maxHeight, 720);
  assert.equal(abrRestrictions(true, 240).maxHeight, 240); // the cap never raises a lower tenant ceiling
});

test('MB per hour is what the toggle promises', () => {
  // 600 kbps = 0.6e6 * 3600 / 8 / 1e6 = 270 MB/h
  assert.equal(Math.round(mbPerHour(600_000)), 270);
  assert.equal(Math.round(mbPerHour(1_800_000)), 810);
  assert.equal(mbPerHour(0), 0);
  assert.equal(mbPerHour(Number.NaN), 0);
});

test('bitrate estimate prefers the real ladder over the defaults', () => {
  const real = [{ height: 360, bandwidth: 412_000 }, { height: 720, bandwidth: 1_500_000 }];
  assert.equal(estimateBpsForHeight(360, real), 412_000);
  assert.equal(estimateBpsForHeight(720, real), 1_500_000);
  assert.equal(estimateBpsForHeight(144, real), 412_000); // nothing below, fall to the smallest
  assert.equal(estimateBpsForHeight(360), 600_000);
  assert.equal(estimateBpsForHeight(1080), 3_500_000);
});
