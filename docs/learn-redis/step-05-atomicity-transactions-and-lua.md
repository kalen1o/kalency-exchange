# Step 05: Atomicity, Transactions, and Lua

## Goal
Handle concurrent updates safely using Redis atomic tools.

## Time Box
2 to 3 days

## Learn
- Single-command atomicity
- `MULTI` and `EXEC`
- Optimistic locking with `WATCH`
- Lua scripts for multi-step atomic logic

## Hands-On
1. Build an atomic counter endpoint:
   - `INCR` and `EXPIRE` for request counting
2. Try transaction flow:
   - `WATCH wallet:1001:balance`
   - `MULTI`
   - `DECRBY wallet:1001:balance 50`
   - `EXEC`
3. Write a Lua script for rate limiting:
   - Increment key
   - Set TTL only on first increment
   - Return current count

## Exit Criteria
- You can identify race conditions in naive Redis logic
- You can choose between transaction and Lua for atomic updates
- You can implement a simple, safe rate limiter

## References
- https://redis.io/docs/latest/develop/interact/transactions/
- https://redis.io/docs/latest/develop/programmability/eval-intro/
