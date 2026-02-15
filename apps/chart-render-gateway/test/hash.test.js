import test from 'node:test';
import assert from 'node:assert/strict';

import { buildCacheKey } from '../src/hash.js';

test('buildCacheKey is stable for same payload shape', () => {
  const key1 = buildCacheKey({ symbol: 'BTC-USD', timeframe: '1m', from: 'a', to: 'b', width: 800, height: 450, theme: 'light' });
  const key2 = buildCacheKey({ height: 450, width: 800, to: 'b', from: 'a', theme: 'light', timeframe: '1m', symbol: 'BTC-USD' });
  assert.equal(key1, key2);
});
