export function validateRenderRequest(payload) {
  if (typeof payload !== 'object' || payload === null) {
    throw new Error('payload must be an object');
  }

  const symbol = String(payload.symbol || '').trim();
  const timeframe = String(payload.timeframe || '').trim();
  if (!symbol) {
    throw new Error('symbol is required');
  }
  if (!timeframe) {
    throw new Error('timeframe is required');
  }

  const width = Number(payload.width || 800);
  const height = Number(payload.height || 450);
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) {
    throw new Error('width and height must be positive numbers');
  }

  return {
    symbol,
    timeframe,
    from: payload.from ? String(payload.from).trim() : '',
    to: payload.to ? String(payload.to).trim() : '',
    width,
    height,
    theme: payload.theme ? String(payload.theme).trim() : 'light',
  };
}
