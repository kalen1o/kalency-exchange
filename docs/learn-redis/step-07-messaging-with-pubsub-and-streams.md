# Step 07: Messaging with Pub/Sub and Streams

## Goal
Understand event-driven patterns in Redis and choose the right primitive.

## Time Box
2 to 3 days

## Learn
- Pub/Sub basics and delivery limitations
- Streams fundamentals
- Consumer groups and acknowledgement
- Retry and dead-letter ideas at application layer

## Hands-On
1. Pub/Sub experiment:
   - Terminal A: `SUBSCRIBE events:user`
   - Terminal B: `PUBLISH events:user "user_created:1001"`
2. Stream experiment:
   - `XADD orders * order_id 9001 amount 120`
   - `XGROUP CREATE orders workers 0 MKSTREAM`
   - `XREADGROUP GROUP workers c1 COUNT 1 STREAMS orders >`
   - `XACK orders workers <id>`
3. Build a worker that:
   - reads stream events
   - retries failures
   - records failed events for later replay

## Exit Criteria
- You can explain Pub/Sub vs Streams clearly
- You can run a consumer group end to end
- You understand at-least-once processing implications

## References
- https://redis.io/docs/latest/develop/interact/pubsub/
- https://redis.io/docs/latest/develop/data-types/streams/
