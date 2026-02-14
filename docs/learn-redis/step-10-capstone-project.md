# Step 10: Capstone Project

## Goal
Combine everything into a realistic backend service using Redis in multiple roles.

## Time Box
4 to 7 days

## Project
Build a mini backend service (any language) with:
- Cache-aside for product or profile reads
- Rate limiter middleware using Redis atomic logic
- Background job queue or stream consumer
- Session or token metadata storage

## Requirements
1. Add observability:
   - command latency metrics
   - cache hit/miss tracking
2. Add operational readiness:
   - key naming convention doc
   - TTL policy doc
   - incident runbook stub
3. Add tests:
   - unit tests for Redis wrappers
   - integration tests with local Redis container

## Exit Criteria
- You can explain all Redis design decisions in your project
- You can show concurrency-safe logic for at least one critical flow
- You can run the service locally with repeatable setup

## Optional Stretch
- Add distributed lock with clear timeout and fallback behavior
- Add stream retry + dead-letter flow
- Add load test and document bottlenecks
