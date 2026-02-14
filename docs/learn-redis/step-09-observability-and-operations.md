# Step 09: Observability and Operations

## Goal
Operate Redis safely with monitoring, troubleshooting, and recovery basics.

## Time Box
1 to 2 days

## Learn
- Key runtime metrics: memory, connected clients, ops/sec, evictions, keyspace hits/misses
- Slow query detection with `SLOWLOG`
- Latency debugging
- Backup and restore routine

## Hands-On
1. Collect baseline metrics:
   - `INFO stats`
   - `INFO commandstats`
2. Inspect slow commands:
   - `SLOWLOG GET 20`
3. Review latency tools:
   - `LATENCY DOCTOR`
4. Write a simple runbook:
   - incident checks
   - common causes
   - recovery steps

## Exit Criteria
- You can define a minimal Redis monitoring dashboard
- You can investigate performance issues with built-in commands
- You have a basic incident runbook draft

## References
- https://redis.io/docs/latest/operate/rs/references/cli-utilities/redis-cli/
