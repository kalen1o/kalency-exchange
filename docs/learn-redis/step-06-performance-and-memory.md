# Step 06: Performance and Memory Optimization

## Goal
Write Redis usage patterns that stay fast under real traffic.

## Time Box
2 days

## Learn
- Pipelining to reduce network overhead
- Avoiding expensive commands in production (`KEYS *`)
- Big keys and memory pressure risks
- Eviction policies overview (`allkeys-lru`, `volatile-ttl`, others)

## Hands-On
1. Compare sequential commands vs pipelined commands in your app
2. Inspect memory:
   - `INFO memory`
   - `MEMORY USAGE key:name`
3. Find risky keys:
   - `SCAN` pattern exploration
4. Configure and test eviction in local Redis:
   - set maxmemory
   - set eviction policy
   - generate load and observe behavior

## Exit Criteria
- You can explain and use pipelining
- You avoid anti-pattern commands in normal request paths
- You understand memory limits and eviction impact

## References
- https://redis.io/docs/latest/develop/using-commands/pipelining/
