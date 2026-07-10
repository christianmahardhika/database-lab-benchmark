# Valkey Official Documentation Reference

Source: https://valkey.io/docs/

## Overview

Valkey is a high-performance open-source key-value store, forked from Redis 7.2.4 (post-license change). Maintained by Linux Foundation. API-compatible with Redis, community-driven development.

## Storage Engine: In-Memory Hash Table (Redis-compatible)

### Architecture
- **Same as Redis**: Single-threaded event loop + I/O threads
- **Data structures**: String, Hash, List, Set, Sorted Set, Stream
- **Persistence**: RDB snapshots + AOF (append-only file)
- **Eviction**: Same policies as Redis (LRU, LFU, TTL-based)

### Key Differences from Redis
| Aspect | Redis (post 7.4) | Valkey |
|--------|-------------------|--------|
| License | SSPL/RSALv2 (not OSI) | BSD-3-Clause (true open source) |
| Governance | Redis Ltd | Linux Foundation |
| API | Redis API | Identical (fork) |
| Performance | Baseline | Same or better (community patches) |
| Modules | Redis modules | Valkey modules (compatible) |

### Write/Read Path
```
Same as Redis — single-threaded event loop, all operations O(1) or O(log N)
```

## Performance Characteristics (Official Claims)

### Throughput
- **Claim**: "Same or better performance than Redis 7.2"
- **Baseline**: 100K+ ops/sec single instance
- **Why**: Same codebase, community performance patches

### Latency
- **Claim**: "Sub-millisecond latency (same as Redis)"
- **Typical**: p50 <0.1ms, p99 <1ms for simple commands

### Compatibility
- **Claim**: "100% API compatible with Redis"
- **How**: Direct fork, same protocol (RESP)
- **Goal**: Drop-in replacement

## Best Use Cases (from valkey.io)

1. **Redis replacement (license concerns)**
   - Same API, truly open source
   - No licensing restrictions
   - Community-driven roadmap

2. **Caching**
   - Same performance as Redis
   - Session management
   - API response cache

3. **Real-time Data Structures**
   - Sorted sets for leaderboards
   - Streams for event sourcing
   - Pub/Sub for messaging

4. **Cloud Provider Managed Services**
   - AWS ElastiCache (Valkey)
   - No license fees passed to users

## Anti-Patterns (same as Redis)

1. **Data larger than RAM** — In-memory only
2. **Complex queries** — No query language
3. **ACID transactions** — Limited MULTI/EXEC
4. **Durable primary store** — Use with persistence or as cache only

## Configuration for Benchmarks

### Memory
```conf
maxmemory 1gb
maxmemory-policy allkeys-lru
```

### Persistence (disable for benchmark)
```conf
save ""
appendonly no
```

### Threading
```conf
io-threads 4
io-threads-do-reads yes
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| used_memory | RAM usage | <maxmemory |
| instantaneous_ops_per_sec | Throughput | Compare vs Redis |
| connected_clients | Client count | Stable |
| keyspace_hits/misses | Cache efficiency | >95% hits |
| latency | Command latency | Sub-ms |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H66 | Same performance as Redis 7.2 | Exp 1: Side-by-side write throughput |
| H67 | Sub-millisecond latency | Exp 2: Read latency comparison |
| H68 | API compatible (same driver works) | Exp: Same go-redis client, different port |
| H69 | Drop-in replacement | Exp: All operations produce same results |

## Benchmark Commands

```bash
# Use redis-benchmark (same protocol)
redis-benchmark -h localhost -p 6381 -c 50 -n 100000 -t set,get -q

# Valkey also ships its own CLI
valkey-benchmark -h localhost -p 6381 -c 50 -n 100000 -t set,get -q
```

## References

- Documentation: https://valkey.io/docs/
- GitHub: https://github.com/valkey-io/valkey
- Compatibility: https://valkey.io/docs/compatibility/
- Blog: https://valkey.io/blog/
