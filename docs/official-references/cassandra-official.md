# Cassandra Official Documentation Reference

Source: https://cassandra.apache.org/doc/latest/

## Overview

Apache Cassandra is a distributed NoSQL database for managing large amounts of data across many commodity servers. Written in Java, uses LSM-tree storage. Designed for high availability with no single point of failure.

## Storage Engine: LSM-Tree (Java)

### Architecture
- **MemTable**: In-memory write buffer
- **Commit Log**: Write-ahead log for durability
- **SSTables**: Immutable sorted string tables on disk
- **Bloom Filters**: Probabilistic structure to avoid unnecessary disk reads
- **Compaction**: Background merge (STCS, LCS, TWCS strategies)

### Write Path
```
Client Write → Commit Log (durability) → MemTable → (flush) → SSTable
```

### Read Path
```
Client Read → MemTable → Bloom Filter → Partition Index → SSTable Data
```

### Compaction Strategies
| Strategy | Best For | Behavior |
|----------|----------|----------|
| STCS (SizeTiered) | Write-heavy | Merge similar-sized SSTables |
| LCS (Leveled) | Read-heavy | Guaranteed max 10% space overhead |
| TWCS (TimeWindow) | Time-series | Compact within time windows |

## Performance Characteristics (Official Claims)

### Write Performance
- **Claim**: "Optimized for fast writes at any scale"
- **Why**: Append-only commit log + memtable flush (no random I/O)
- **Benchmark**: ~10K-50K writes/sec per node (depends on consistency level)

### Read Performance
- **Claim**: "Single-digit millisecond reads for partition key lookups"
- **Caveat**: Performance degrades with multi-partition queries
- **Bloom filter**: Reduces unnecessary SSTable reads

### Scalability
- **Claim**: "Linear horizontal scaling — double nodes = double throughput"
- **How**: Consistent hash ring, no master node
- **Proven**: Used at Apple (400K+ nodes), Netflix, Discord

### Availability
- **Claim**: "No single point of failure"
- **How**: Peer-to-peer, all nodes equal, tunable consistency

## Best Use Cases (from cassandra.apache.org)

1. **High-Write Throughput**
   - Event logging
   - Sensor/IoT data
   - Activity tracking

2. **Time-Series Data**
   - TWCS compaction strategy
   - TTL-based auto-expiration
   - Wide partitions for time ranges

3. **Multi-Datacenter Deployments**
   - NetworkTopologyStrategy
   - Asynchronous replication
   - Tunable consistency per query

4. **Always-On Applications**
   - No single point of failure
   - Graceful degradation
   - Rolling upgrades

## Anti-Patterns (from docs)

1. **Ad-hoc queries**
   - Must design tables around query patterns
   - No flexible WHERE clauses
   - No JOINs

2. **Small datasets**
   - Overhead not justified for <10GB
   - Use PostgreSQL/MySQL instead

3. **Strong consistency (all reads)**
   - QUORUM reads reduce throughput
   - Consider CockroachDB for serializable

4. **Heavy read-modify-write patterns**
   - No built-in read-before-write
   - LWT (Lightweight Transactions) are slow

## Configuration for Benchmarks

### Memory Settings
```yaml
# cassandra.yaml
memtable_heap_space: 256mb
key_cache_size_in_mb: 100
row_cache_size_in_mb: 0  # disable for benchmark
```

### Compaction
```yaml
compaction_throughput_mb_per_sec: 64
concurrent_compactors: 2
```

### JVM Settings
```bash
# cassandra-env.sh
MAX_HEAP_SIZE="1G"
HEAP_NEWSIZE="256M"
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| write_latency | Coordinator write latency | <5ms p99 |
| read_latency | Coordinator read latency | <10ms p99 |
| pending_compactions | Compaction backlog | <10 |
| sstable_count | SSTables per table | Depends on strategy |
| heap_usage | JVM heap | <75% |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H5 | Fast writes (append-only LSM) | Exp 1: Write throughput |
| H6 | Compare vs ScyllaDB (Java vs C++) | Exp 1: Same workload, side-by-side |
| H7 | Linear scaling (add nodes) | Future: Multi-node lab |
| H8 | Time-series optimized (TWCS) | Exp: Time-series workload |

## Comparison: Cassandra vs ScyllaDB

| Aspect | Cassandra | ScyllaDB |
|--------|-----------|----------|
| Language | Java (JVM) | C++ (Seastar) |
| Threading | Thread-per-core pool | Shard-per-core |
| GC pauses | Yes (JVM) | No (manual memory) |
| Tail latency | Higher (GC spikes) | Lower (predictable) |
| Compatibility | Original | CQL-compatible |
| Tooling | Mature (nodetool) | Growing |
| Community | Larger | Smaller |

## References

- Documentation: https://cassandra.apache.org/doc/latest/
- Architecture: https://cassandra.apache.org/doc/latest/cassandra/architecture/
- Data Modeling: https://cassandra.apache.org/doc/latest/cassandra/data_modeling/
- Performance: https://cassandra.apache.org/doc/latest/cassandra/operating/
