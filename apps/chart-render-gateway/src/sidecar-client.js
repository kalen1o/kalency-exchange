export async function renderWithSidecar(baseUrl, payload) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 5000);
  try {
    const response = await fetch(baseUrl, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(payload),
      signal: controller.signal,
    });
    if (!response.ok) {
      const body = await response.text();
      throw new Error(`sidecar request failed: ${response.status} ${body}`);
    }
    const data = await response.json();
    return {
      renderId: data.renderId,
      artifactType: data.artifactType,
      artifact: data.artifact,
      meta: data.meta || payload,
    };
  } finally {
    clearTimeout(timeout);
  }
}
