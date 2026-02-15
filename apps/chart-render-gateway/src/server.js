import http from 'node:http';

import { MemoryCache } from './cache.js';
import { buildCacheKey } from './hash.js';
import { renderWithSidecar } from './sidecar-client.js';
import { validateRenderRequest } from './validate.js';

const port = Number(process.env.PORT || '8085');
const sidecarUrl = process.env.CHARTGPU_SIDECAR_URL || 'http://chartgpu-sidecar:8086/render';
const ttlMs = Number(process.env.CHART_CACHE_TTL_MS || '60000');

const cache = new MemoryCache();

function sendJson(res, status, payload) {
  const body = JSON.stringify(payload);
  res.writeHead(status, {
    'content-type': 'application/json',
    'content-length': Buffer.byteLength(body),
    'access-control-allow-origin': '*',
  });
  res.end(body);
}

function handleOptions(res) {
  res.writeHead(204, {
    'access-control-allow-origin': '*',
    'access-control-allow-methods': 'GET,POST,OPTIONS',
    'access-control-allow-headers': 'content-type',
  });
  res.end();
}

const server = http.createServer(async (req, res) => {
  if (req.method === 'OPTIONS') {
    handleOptions(res);
    return;
  }

  if (req.method === 'GET' && req.url === '/healthz') {
    sendJson(res, 200, { status: 'ok' });
    return;
  }

  if (req.method !== 'POST' || req.url !== '/v1/charts/render') {
    sendJson(res, 404, { error: 'not found' });
    return;
  }

  let body = '';
  req.setEncoding('utf8');
  req.on('data', (chunk) => {
    body += chunk;
  });

  req.on('end', async () => {
    try {
      const payload = body ? JSON.parse(body) : {};
      const validated = validateRenderRequest(payload);
      const cacheKey = buildCacheKey(validated);

      const cached = cache.get(cacheKey);
      if (cached) {
        sendJson(res, 200, {
          cached: true,
          cacheKey,
          ...cached,
        });
        return;
      }

      const rendered = await renderWithSidecar(sidecarUrl, validated);
      const value = {
        renderId: rendered.renderId,
        artifactType: rendered.artifactType,
        artifact: rendered.artifact,
        meta: rendered.meta,
      };
      cache.set(cacheKey, value, ttlMs);

      sendJson(res, 200, {
        cached: false,
        cacheKey,
        ...value,
      });
    } catch (error) {
      sendJson(res, 400, { error: String(error.message || error) });
    }
  });
});

server.listen(port, () => {
  console.log(`chart-render-gateway listening on :${port}`);
});
