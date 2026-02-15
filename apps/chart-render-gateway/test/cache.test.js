import test from 'node:test';
import assert from 'node:assert/strict';

import { MemoryCache } from '../src/cache.js';

test('MemoryCache returns cached value before expiry', async () => {
  const cache = new MemoryCache();
  cache.set('k1', { ok: true }, 1_000);
  const value = cache.get('k1');
  assert.deepEqual(value, { ok: true });
});

test('MemoryCache expires value after ttl', async () => {
  const cache = new MemoryCache();
  cache.set('k2', { ok: true }, 5);
  await new Promise((resolve) => setTimeout(resolve, 20));
  assert.equal(cache.get('k2'), null);
});
