# ScyllaDB Official Documentation Reference

Source: https://docs.scylladb.com/

## Overview

ScyllaDB is a high-performance NoSQL database compatible with Apache Cassandra. Written in C++ using the Seastar framework for async I/O.

## Storage Engine: LSM-Tree

### Architecture
- **MemTable**: In-memory write buffer (default 256MB per shard)
- **SSTables**: Immutable sorted string tables on disk
- **Compaction**: Background merge of SSTables

### Write Path
```
Client Write → Commit Log → MemTable → (flush) → SSTable
```

### Read Path
```
Client Read → MemTable → Bloom Filter → SSTable Index → SSTable Data
```

## Performance Characteristics (Official Claims)

### Write Performance
- **Claim**: "Millions of operations per second on a single node"
- **Why**: Shard-per-core architecture, no JVM overhead
- **Benchmark**: 1M ops/sec with 3-node cluster (official blog)

### Read Performance
- **Claim**: "Sub-millisecond P99 latency"
- **Caveat**: Depends on data in cache, compaction state

### Scalability
- **Claim**: "Linear horizontal scaling"
- **How**: Consistent hashing, automatic sharding

## Best Use Cases (from docs.scylladb.com)

1. **Time-series data**
   - IoT sensor data
   - Metrics and monitoring
   - Financial tick data

2. **High-throughput applications**
   - Gaming leaderboards
   - Real-time bidding
   - Session management

3. **Multi-datacenter deployments**
   - Global applications
   - Disaster recovery

## Anti-Patterns (from docs)

1. **Complex queries**
   - No JOINs
   - Limited aggregations
   - No ad-hoc queries

2. **Small datasets**
   - Overhead not worth it for <100GB

3. **Strong consistency requirements**
   - Default is eventual consistency
   - Can do QUORUM but impacts performance

## Configuration for Benchmarks

### Recommended Settings (from tuning guide)
```yaml
# /etc/scylla/scylla.yaml
commitlog_sync: periodic
commitlog_sync_period_in_ms: 10000
compaction_throughput_mb_per_sec: 64
concurrent_reads: 32
concurrent_writes: 32
```

### Memory Settings
```bash
# Let ScyllaDB manage memory
--memory 8G
--reserve-memory 1G
--overprovisioned  # for shared environments
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| scylla_reactor_utilization | CPU per shard | <90% |
| scylla_storage_proxy_coordinator_write_latency | Write latency | <5ms p99 |
| scylla_storage_proxy_coordinator_read_latency | Read latency | <10ms p99 |
| scylla_compaction_manager_compactions | Active compactions | <4 |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H5 | >10K writes/sec sustained | Exp 1: Write throughput |
| H6 | Time-series optimized | Exp 7, 11: Time-series workload |
| H7 | Linear horizontal scaling | Exp 5: Add nodes, measure throughput |
| H8 | Multi-region capable | N/A (single region test) |

## References

- Architecture: https://docs.scylladb.com/architecture/
- Tuning: https://docs.scylladb.com/operating-scylla/tuning/
- Best Practices: https://docs.scylladb.com/using-scylla/best-practices/
- Benchmarks: https://www.scylladb.com/product/benchmarks/
