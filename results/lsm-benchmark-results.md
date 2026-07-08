# B-Tree vs LSM-Tree Benchmark Results

**Date:** 2026-06-14
**Environment:** Docker containers, 2 CPU cores, 1GB RAM each

## Databases Tested
- **PostgreSQL 15** (B-Tree) - pgbench native
- **ScyllaDB 5.4** (LSM-Tree) - cassandra-stress native

## Results Summary

### Native Benchmark Tools (Fair Comparison)

| Metric | PostgreSQL (B-Tree) | ScyllaDB (LSM-Tree) | Winner | Ratio |
|--------|---------------------|---------------------|--------|-------|
| Write TPS | 1,920 | 9,751 | **ScyllaDB** | 5.1x |
| Write Latency | 2.08ms | 0.4ms | **ScyllaDB** | 5.2x |
| Read TPS | 14,877 | 11,976 | **PostgreSQL** | 1.2x |
| Read Latency | 0.27ms | 0.3ms | **PostgreSQL** | 1.1x |
| Mixed (80/20) | ~4,000* | 8,354 | **ScyllaDB** | 2.1x |

*PostgreSQL mixed estimated from pgbench write-heavy test

### Python Driver Comparison (High Overhead)

| Metric | PostgreSQL | ScyllaDB | Notes |
|--------|------------|----------|-------|
| Write (100k) | 23,562/s | 1,014/s | Driver overhead dominates |
| Point Query | 0.21ms | 0.93ms | Network round-trip |
| Range Scan | 30.5ms | 52.1ms | ALLOW FILTERING penalty |

## Key Insights

### 1. LSM-Tree Wins: Write-Heavy Workloads
- **5x faster writes** due to sequential WAL + memtable flush
- ScyllaDB batches writes to memtable, then flushes to SSTables
- PostgreSQL must update B-tree indexes in-place (random I/O)

### 2. B-Tree Wins: Read-Heavy Workloads  
- **1.2x faster reads** due to sorted page structure
- Single B-tree traversal vs multi-level SSTable + bloom filter checks
- PostgreSQL buffer pool caching very effective

### 3. Driver Overhead Matters
- Python cassandra-driver adds ~10x overhead vs native stress tool
- psycopg2 much more efficient for PostgreSQL
- **Always use native benchmarks for fair comparison**

### 4. Range Scans: B-Tree Dominates
- ScyllaDB requires `ALLOW FILTERING` for non-partition key ranges
- B-tree natural ordering = efficient range access
- LSM scattered SSTables = multiple seeks

## When to Use What

| Use Case | Recommendation | Why |
|----------|----------------|-----|
| Time-series / Logs | LSM (ScyllaDB, Cassandra) | Write-heavy, append-only |
| OLTP / Transactions | B-Tree (PostgreSQL) | Read-heavy, range queries |
| IoT Sensor Data | LSM | High write throughput |
| E-commerce Catalog | B-Tree | Complex queries, joins |
| Chat Messages | LSM | Write-heavy, simple reads |
| Financial Ledger | B-Tree | ACID, complex reporting |

## Architecture Comparison

```
B-Tree (PostgreSQL)           LSM-Tree (ScyllaDB)
├── Buffer Pool               ├── Memtable (in-memory)
├── WAL                       ├── WAL  
├── B-Tree Index              ├── SSTable Level 0
│   ├── Root Page            │   ├── SSTable Level 1
│   ├── Branch Pages         │   ├── SSTable Level 2
│   └── Leaf Pages           │   └── ... (compaction)
└── Heap/Data Pages           └── Bloom Filters
```

## Write Amplification

- **B-Tree:** ~10-30x (WAL + page splits + index updates)
- **LSM-Tree:** ~10-50x (compaction overhead across levels)

### B-Tree Write Path (DDIA Theory)
```
1 INSERT = 
  + WAL write (1x)
  + Page read (if not cached)
  + Page write (1x)
  + Potential page splits (Nx cascade)
```

### LSM-Tree Write Path
```
1 INSERT = 
  + WAL write (1x)
  + Memtable (memory only)
  + Flush to L0 (1x)
  + Compact L0→L1 (1x)
  + Compact L1→L2 (1x)
  + ... Ln-1→Ln (1x each)
```

LSM trades write amplification for write throughput - compaction happens in background.

## Read Amplification

- **B-Tree:** 3-4 page reads (predictable, single tree traversal)
- **LSM-Tree:** 6+ file reads worst case (mitigated by bloom filters)

### Why B-Tree Wins Reads
- Single sorted structure = one traversal path
- Buffer pool caching effective for hot pages
- Depth formula: `log₅₀₀(500M) ≈ 4 levels`

### Why LSM Reads Can Suffer
- Data scattered across multiple SSTables/levels
- Must check each level until found (or bloom filter says "not here")
- Bloom filters reduce this: 1% false positive = skip 99% unnecessary reads

## Next Steps
- [ ] Test compaction impact on read latency
- [ ] Add RocksDB direct benchmarks via TiKV
- [ ] Test under sustained load (24h)
