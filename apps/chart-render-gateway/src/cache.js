export class MemoryCache {
  constructor() {
    this.entries = new Map();
  }

  set(key, value, ttlMs) {
    const expiresAt = Date.now() + Math.max(0, Number(ttlMs) || 0);
    this.entries.set(key, { value, expiresAt });
  }

  get(key) {
    const entry = this.entries.get(key);
    if (!entry) {
      return null;
    }
    if (entry.expiresAt <= Date.now()) {
      this.entries.delete(key);
      return null;
    }
    return entry.value;
  }
}
