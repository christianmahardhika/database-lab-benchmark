# CockroachDB Official Documentation Reference

Source: https://www.cockroachlabs.com/docs/

## Overview

CockroachDB is a distributed SQL database built for cloud-native applications. PostgreSQL-compatible wire protocol. Uses Pebble (LSM-tree) storage engine with Raft consensus for replication.

## Storage Engine: Pebble (LSM-Tree) + Raft

### Architecture
- **Pebble**: LSM-tree storage engine (RocksDB replacement, written in Go)
- **Raft**: Consensus protocol for replication
- **Ranges**: Data partitioned into 512MB ranges (auto-split)
- **Leaseholder**: Node that serves reads for a range

### Write Path
```
SQL → Transaction Coordinator → Raft Consensus → Pebble (LSM) → Disk
```

### Read Path
```
SQL → Gateway Node → Route to Leaseholder → Pebble Read → Return
```

### Transaction Model
- **Serializable isolation** by default (strongest level)
- **MVCC**: Multi-version concurrency control
- **Timestamp ordering**: Clock synchronization via HLC (Hybrid Logical Clock)

## Performance Characteristics (Official Claims)

### Distributed SQL
- **Claim**: "Serializable ACID transactions across distributed nodes"
- **Why**: Raft consensus + MVCC + HLC timestamps
- **Tradeoff**: Higher latency than single-node for cross-range transactions

### Scalability
- **Claim**: "Linear horizontal scaling for reads and writes"
- **How**: Auto-sharding via ranges, leaseholder distribution
- **Benchmark**: Linear throughput scaling up to 256 nodes (official)

### Survivability
- **Claim**: "Survive any single failure (node, zone, region) without interruption"
- **How**: Raft replication (3+ replicas), automatic leader election

### PostgreSQL Compatibility
- **Claim**: "Wire-compatible with PostgreSQL"
- **Reality**: Most PG features supported, some gaps (extensions, stored procedures)

## Best Use Cases (from cockroachlabs.com)

1. **Global Applications**
   - Multi-region deployments
   - Data locality (pin data near users)
   - Low-latency global reads

2. **Financial Services**
   - Serializable transactions
   - No lost writes
   - Audit trails

3. **Cloud-Native Microservices**
   - Kubernetes-native
   - Auto-scaling
   - Zero-downtime upgrades

4. **Multi-Cloud / Hybrid**
   - Run across AWS + GCP + Azure
   - Avoid vendor lock-in

## Anti-Patterns (from docs)

1. **Single-node workloads**
   - Overhead of consensus not justified
   - PostgreSQL faster for single-node

2. **Ultra-low latency (<1ms)**
   - Raft consensus adds latency
   - Use Redis/in-memory for sub-ms

3. **Heavy analytics/OLAP**
   - Not columnar
   - Use ClickHouse for analytical queries

4. **Simple key-value at extreme scale**
   - SQL overhead unnecessary
   - Use ScyllaDB/DynamoDB

## Configuration for Benchmarks

### Node Start
```bash
cockroach start-single-node \
  --insecure \
  --listen-addr=localhost:26257 \
  --http-addr=localhost:8080 \
  --store=cockroach-data \
  --cache=.25 \
  --max-sql-memory=.25
```

### SQL Settings
```sql
SET CLUSTER SETTING kv.range_merge.queue_enabled = true;
SET CLUSTER SETTING kv.snapshot_rebalance.max_rate = '64MiB';
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| sql.txn.commit.count | Transaction throughput | Baseline |
| sql.txn.latency-p99 | Transaction latency | <50ms single-region |
| ranges | Number of ranges | Auto-managed |
| liveness.livenodes | Cluster health | All nodes live |
| capacity.used | Storage utilization | <80% |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H57 | Serializable ACID distributed | Exp: Concurrent writes, verify no anomalies |
| H58 | Linear horizontal scaling | Exp: Single node throughput as baseline |
| H59 | PG wire-compatible | Exp: Use pgx driver, same queries as PG |
| H60 | Higher latency than single-node PG | Exp: Compare p99 vs PostgreSQL standalone |

## Comparison: CockroachDB vs PostgreSQL

| Aspect | CockroachDB | PostgreSQL |
|--------|-------------|------------|
| Distribution | Native | Extensions (Citus) |
| Isolation | Serializable (default) | Read Committed (default) |
| Write latency | Higher (consensus) | Lower (single-node) |
| Horizontal scale | Native | Manual sharding |
| Storage engine | Pebble (LSM) | Heap + B-Tree |
| Failover | Automatic | Manual or tools |
| SQL compatibility | ~95% PG | Full |

## References

- Documentation: https://www.cockroachlabs.com/docs/
- Architecture: https://www.cockroachlabs.com/docs/stable/architecture/overview
- Performance: https://www.cockroachlabs.com/docs/stable/performance
- Benchmarks: https://www.cockroachlabs.com/blog/cockroachdb-2point1-performance/
