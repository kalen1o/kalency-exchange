# Step 02: Core Data Structures

## Goal
Master Redis data structures and choose the right one for each backend problem.

## Time Box
2 to 3 days

## Learn
- Strings
- Hashes
- Lists
- Sets
- Sorted sets
- Naming conventions for keys (`service:entity:id:field`)

## Hands-On
1. Strings and counters:
   - `SET page:home:views 0`
   - `INCR page:home:views`
2. Hashes for user profile:
   - `HSET user:1001 name "Ana" plan "pro"`
   - `HGETALL user:1001`
3. Lists as queue:
   - `LPUSH jobs:email "job1" "job2"`
   - `RPOP jobs:email`
4. Sets for unique tags:
   - `SADD post:44:tags redis backend cache`
   - `SMEMBERS post:44:tags`
5. Sorted sets for leaderboard:
   - `ZADD game:leaderboard 1200 alice 950 bob`
   - `ZRANGE game:leaderboard 0 -1 WITHSCORES`

## Exit Criteria
- You can pick an appropriate data type for a use case
- You can model simple entities and lists of work in Redis
- You can explain tradeoffs between list, set, and sorted set

## References
- https://redis.io/docs/latest/develop/data-types/
