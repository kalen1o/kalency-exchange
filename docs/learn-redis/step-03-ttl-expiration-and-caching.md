# Step 03: TTL, Expiration, and Caching Patterns

## Goal
Use Redis as a cache correctly, including expiration strategy and invalidation.

## Time Box
2 to 3 days

## Learn
- TTL lifecycle (`EXPIRE`, `TTL`, `PERSIST`)
- Cache-aside pattern
- Write-through vs write-around basics
- Cache invalidation strategy tied to business events
- Why stale data happens and how to mitigate it

## Hands-On
1. Practice expiration:
   - `SET product:500:name "Keyboard" EX 60`
   - `TTL product:500:name`
2. Build a tiny cache-aside flow in your backend language:
   - Read from Redis first
   - On miss, query DB and set cache with TTL
3. Add cache invalidation:
   - On product update, delete `product:{id}:*` keys
4. Simulate stampede protection:
   - Add jitter to TTL (random extra seconds)
   - Optionally use a short lock key during misses

## Exit Criteria
- You can implement cache-aside end to end
- You can choose TTL values based on data volatility
- You can describe stampede and stale cache risks

## References
- https://redis.io/glossary/cache-invalidation/
