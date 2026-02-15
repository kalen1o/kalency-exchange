import crypto from 'node:crypto';

export function buildCacheKey(payload) {
  const normalized = {
    symbol: String(payload.symbol || '').trim(),
    timeframe: String(payload.timeframe || '').trim(),
    from: String(payload.from || '').trim(),
    to: String(payload.to || '').trim(),
    width: Number(payload.width || 800),
    height: Number(payload.height || 450),
    theme: String(payload.theme || 'light').trim().toLowerCase(),
  };

  const canonical = JSON.stringify(normalized, Object.keys(normalized).sort());
  return crypto.createHash('sha256').update(canonical).digest('hex');
}
