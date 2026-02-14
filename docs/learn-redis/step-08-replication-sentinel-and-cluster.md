# Step 08: Replication, Sentinel, and Cluster

## Goal
Learn Redis high availability and horizontal scaling concepts used in production.

## Time Box
2 to 3 days

## Learn
- Primary-replica replication
- Failover basics
- Sentinel role
- Cluster hash slots and sharding model
- Application implications during failover

## Hands-On
1. Read and diagram:
   - single node
   - primary + replica
   - Sentinel-managed setup
   - Cluster setup
2. Run a local replication test with Docker compose
3. Simulate primary restart and observe reconnection behavior
4. Document how your backend client handles failover

## Exit Criteria
- You can explain when to use Sentinel vs Cluster
- You understand consistency implications during failover
- You can list required client configuration for HA

## References
- https://redis.io/docs/latest/operate/oss_and_stack/management/replication/
- https://redis.io/docs/latest/operate/oss_and_stack/management/sentinel/
- https://redis.io/docs/latest/operate/oss_and_stack/management/scaling/
