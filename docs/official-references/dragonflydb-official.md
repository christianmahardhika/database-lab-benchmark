# DragonflyDB Official Documentation Reference

Source: https://www.dragonflydb.io/docs

## Overview

DragonflyDB is a modern in-memory data store, fully compatible with Redis and Memcached APIs. Written in C++ with a multi-threaded shared-nothing architecture. Claims to be 25x more throughput than Redis on a single instance.

## Storage Engine: Multi-Threaded In-Memory

### Architecture
- **Shared-nothing**: Each thread owns a shard of the keyspace
- **No global locks**: Eliminates contention (vs Redis single-thread)
- **Dashtable**: Custom hash table optimized for memory efficiency
- **VLL (Very Lightweight Locking)**: Per-key locking for transactions

### Key Differences from Redis
| Aspect | Redis | DragonflyDB |
|--------|-------|-------------|
| Threading | Single-threaded (6.0+ I/O threads) | Multi-threaded (shared-nothing) |
| Memory efficiency | ~100 bytes/key overhead | ~40 bytes/key overhead |
| Snapshotting | Fork (2x memory spike) | Forkless (no memory spike) |
| Scaling | Cluster for multi-core | Single instance uses all cores |

### Write Path
```
Command → Route to owning thread → Dashtable insert → ACK
```

### Read Path
```
Command → Route to owning thread → Dashtable lookup → Return
```

## Performance Characteristics (Official Claims)

### Throughput
- **Claim**: "25x more throughput than Redis"
- **Benchmark**: 4M+ ops/sec on single instance (c6gn.16xlarge, 64 vCPU)
- **Why**: Multi-threaded, no global lock, uses all CPU cores

### Latency
- **Claim**: "Sub-millisecond latency comparable to Redis"
- **Typical**: p50 <0.1ms, p99 <1ms (in-memory operations)

### Memory Efficiency
- **Claim**: "Up to 40% less memory than Redis"
- **Why**: Dashtable has lower per-key overhead, forkless snapshots

### Snapshot Performance
- **Claim**: "No memory spikes during persistence"
- **Why**: Forkless snapshotting (vs Redis fork which doubles memory)

## Best Use Cases (from dragonflydb.io)

1. **Drop-in Redis Replacement**
   - Same API, higher throughput
   - When Redis single-thread is bottleneck
   - When memory efficiency matters

2. **High-Throughput Caching**
   - Session stores
   - API response caching
   - Rate limiting at scale

3. **Real-time Applications**
   - Leaderboards
   - Pub/Sub messaging
   - Counters and accumulators

4. **Memory-Constrained Environments**
   - Lower per-key overhead
   - No fork memory spikes

## Anti-Patterns (from docs)

1. **Small-scale deployments**
   - Single-core: Redis performs similarly
   - Benefits show at ≥4 cores

2. **Redis Cluster topology required**
   - DragonflyDB is single-instance multi-thread
   - Not a cluster replacement (yet)

3. **Lua scripting (partial support)**
   - Most scripts work, some edge cases differ

## Configuration for Benchmarks

### Start Command
```bash
dragonfly --maxmemory=1g --proactor_threads=4
```

### Key Settings
```bash
--maxmemory         # Memory limit
--proactor_threads  # Number of threads (default: num CPUs)
--dbfilename        # Snapshot filename
--dir               # Data directory
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| used_memory | RAM usage | <maxmemory |
| instantaneous_ops_per_sec | Throughput | Compare vs Redis |
| connected_clients | Connections | Stable |
| evicted_keys | Memory pressure | Low |
| latency_p99 | Tail latency | <1ms |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H61 | 25x throughput vs Redis | Exp 1: Write throughput comparison (same hardware) |
| H62 | Sub-millisecond p99 latency | Exp 2: Read latency at high concurrency |
| H63 | Better memory efficiency | Exp: Compare memory for same dataset |
| H64 | Multi-core utilization | Exp: Throughput at concurrency 1 vs 10 vs 50 |
| H65 | API-compatible with Redis | Exp: Same benchmark code, just change port |

## Benchmark Commands

```bash
# Use redis-benchmark (compatible)
redis-benchmark -h localhost -p 6380 -c 50 -n 100000 -t set,get -q

# Use memtier_benchmark (recommended by Dragonfly)
memtier_benchmark -s localhost -p 6380 --threads=4 --clients=50 --requests=100000
```

## References

- Documentation: https://www.dragonflydb.io/docs
- Architecture: https://www.dragonflydb.io/docs/getting-started/architecture
- Benchmarks: https://www.dragonflydb.io/benchmarks
- GitHub: https://github.com/dragonflydb/dragonfly
