# Redis Official Documentation Reference

Source: https://redis.io/docs/

## Overview

Redis is an in-memory data structure store. Used as database, cache, message broker, and streaming engine. Single-threaded event loop (6.0+ has I/O threads).

## Storage Engine: In-Memory + Optional Persistence

### Data Structures
| Type | Use Case | Commands |
|------|----------|----------|
| String | Cache, counters | GET, SET, INCR |
| Hash | Objects, field-value | HGET, HSET, HMGET |
| List | Queues, timelines | LPUSH, RPOP, LRANGE |
| Set | Tags, unique items | SADD, SMEMBERS, SINTER |
| Sorted Set | Leaderboards, ranges | ZADD, ZRANGE, ZRANK |
| Stream | Event log, messaging | XADD, XREAD, XRANGE |

### Persistence Options
| Mode | Description | Performance |
|------|-------------|-------------|
| RDB | Point-in-time snapshots | Fast, data loss risk |
| AOF | Append-only log | Slower, minimal loss |
| RDB+AOF | Combined | Balanced |
| None | Pure cache | Fastest |

### Eviction Policies
```
noeviction       - Error when memory full
allkeys-lru      - LRU eviction any key
volatile-lru     - LRU eviction keys with TTL
allkeys-random   - Random eviction
volatile-random  - Random eviction keys with TTL
volatile-ttl     - Evict shortest TTL first
```

## Performance Characteristics (Official Claims)

### From redis.io
- **Latency**: "Sub-millisecond response times"
- **Throughput**: "Millions of requests per second"
- **Benchmark**: 100K+ ops/sec single instance

### Typical Performance
```
# redis-benchmark -q
PING_INLINE: 142857.14 requests per second
SET: 128205.13 requests per second
GET: 131578.95 requests per second
INCR: 135135.14 requests per second
```

## Best Use Cases (from redis.io)

1. **Caching**
   - Session storage
   - API response cache
   - Database query cache

2. **Real-time Leaderboards**
   - Sorted sets for ranking
   - O(log N) updates

3. **Rate Limiting**
   - Atomic INCR with TTL
   - Sliding window counters

4. **Pub/Sub Messaging**
   - Real-time notifications
   - Chat applications

5. **Queues**
   - Job queues with lists
   - Reliable queues with streams

## Anti-Patterns (from docs)

1. **Data larger than RAM**
   - In-memory only
   - Use Redis on Flash or disk DB

2. **Complex queries**
   - No query language
   - Application-side filtering

3. **Transactions across keys**
   - Limited transaction support
   - MULTI/EXEC not true ACID

## Configuration for Benchmarks

### Memory
```conf
# /etc/redis/redis.conf
maxmemory 4gb
maxmemory-policy allkeys-lru
```

### Persistence (for benchmark: disable)
```conf
save ""
appendonly no
```

### Performance
```conf
tcp-backlog 511
tcp-keepalive 300
timeout 0
```

### Threads (Redis 6.0+)
```conf
io-threads 4
io-threads-do-reads yes
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| used_memory | RAM usage | <maxmemory |
| connected_clients | Client connections | <10K |
| instantaneous_ops_per_sec | Throughput | Baseline |
| keyspace_hits/misses | Cache hit ratio | >95% |
| evicted_keys | Memory pressure | Minimal |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H13 | Sub-millisecond latency | Exp 2: Read latency |
| H14 | Real-time leaderboards | Exp: ZADD/ZRANGE ops |
| H15 | Pub/Sub messaging | N/A (not in scope) |
| H16 | Rate limiting | N/A (not in scope) |

## Benchmark Commands

```bash
# Built-in benchmark
redis-benchmark -h localhost -p 6379 -c 50 -n 100000

# Specific commands
redis-benchmark -t set,get -n 100000 -q

# Pipeline (batch)
redis-benchmark -t set -n 100000 -P 16 -q
```

## Pipelining
```python
# Without pipeline: 100 round trips
for i in range(100):
    r.set(f'key:{i}', f'value:{i}')

# With pipeline: 1 round trip
pipe = r.pipeline()
for i in range(100):
    pipe.set(f'key:{i}', f'value:{i}')
pipe.execute()
```

## References

- Documentation: https://redis.io/docs/
- Data Types: https://redis.io/docs/data-types/
- Persistence: https://redis.io/docs/management/persistence/
- Benchmarks: https://redis.io/docs/management/optimization/benchmarks/
- Cluster: https://redis.io/docs/management/scaling/
