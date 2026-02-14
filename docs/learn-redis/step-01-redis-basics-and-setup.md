# Step 01: Redis Basics and Setup

## Goal
Understand what Redis is, where it fits in backend systems, and get a local environment ready.

## Time Box
1 to 2 days

## Learn
- Redis core idea: in-memory key-value store
- Why teams use Redis: low latency, simple operations, flexible data structures
- Common backend use cases: caching, sessions, counters, queues, rate limiting
- Difference between Redis and primary databases (PostgreSQL, MySQL)

## Hands-On
1. Start Redis locally with Docker:
   - `docker run --name redis-dev -p 6379:6379 -d redis:7`
2. Connect with CLI:
   - `docker exec -it redis-dev redis-cli`
3. Run basic commands:
   - `PING`
   - `SET user:1:name "kalen"`
   - `GET user:1:name`
   - `DEL user:1:name`
4. Inspect server info:
   - `INFO`
   - `DBSIZE`
   - `KEYS *` (learning only, not production)

## Exit Criteria
- You can explain when Redis is useful in a backend service
- You can run Redis and use `redis-cli` confidently
- You understand basic key operations (`SET`, `GET`, `DEL`)

## References
- https://redis.io/docs/latest/
- https://redis.io/docs/latest/develop/tools/cli/
