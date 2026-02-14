# Step 04: Persistence and Durability

## Goal
Understand what data Redis keeps after restart and how to configure persistence safely.

## Time Box
1 to 2 days

## Learn
- RDB snapshots
- AOF logging
- Tradeoffs: performance vs durability
- Basic backup and restore mindset

## Hands-On
1. Inspect persistence settings:
   - `CONFIG GET save`
   - `CONFIG GET appendonly`
2. Force snapshot:
   - `BGSAVE`
3. Enable AOF in a local config and restart Redis
4. Compare behavior:
   - Write keys
   - Restart server
   - Confirm what survives

## Exit Criteria
- You understand when Redis can lose data
- You can explain RDB vs AOF in practical terms
- You can set realistic expectations for Redis durability in production

## References
- https://redis.io/docs/latest/operate/oss_and_stack/management/persistence/
