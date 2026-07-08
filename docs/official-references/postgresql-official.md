# PostgreSQL Official Documentation Reference

Source: https://www.postgresql.org/docs/

## Overview

PostgreSQL is an advanced open-source relational database with 35+ years of active development. Known for reliability, feature robustness, and standards compliance.

## Storage Engine: B-Tree (Heap + Indexes)

### Architecture
- **Heap**: Main table storage (unordered rows)
- **TOAST**: Large value storage (compressed, out-of-line)
- **Indexes**: B-Tree (default), Hash, GiST, GIN, BRIN, SP-GiST
- **WAL**: Write-ahead log for durability

### Page Structure
```
Page (8KB default)
├── PageHeader (24 bytes)
├── ItemId Array (line pointers)
├── Free Space
├── Items (tuples)
└── Special Space (index-specific)
```

### Write Path
```
INSERT → WAL buffer → Shared buffers → Background writer → Disk
```

### Read Path
```
SELECT → Check shared_buffers → Read from disk if miss → Return
```

### MVCC (Multi-Version Concurrency Control)
- Each row has xmin (created), xmax (deleted)
- Readers don't block writers
- VACUUM cleans dead tuples

## Performance Characteristics (Official)

### From PostgreSQL Wiki
- **OLTP**: Excellent, full ACID
- **OLAP**: Good with proper indexing
- **Concurrent users**: Thousands with connection pooling

### Scaling
- **Vertical**: Effective up to ~64 cores
- **Horizontal**: Read replicas, Citus for sharding

## Best Use Cases (from postgresql.org)

1. **Complex OLTP**
   - Multi-table transactions
   - Complex queries with JOINs
   - Referential integrity required

2. **Geospatial (PostGIS)**
   - Location-based services
   - GIS applications

3. **Full-text search**
   - Built-in tsvector/tsquery
   - No external search engine needed

4. **JSON/Document hybrid**
   - JSONB with indexing
   - Relational + document in one DB

5. **Data warehousing**
   - Window functions
   - CTEs, recursive queries
   - Parallel query execution

## Anti-Patterns (from docs/wiki)

1. **Very high write throughput**
   - >50K writes/sec challenging
   - Consider LSM-based alternatives

2. **Simple key-value at scale**
   - Overhead for simple lookups
   - Redis/DynamoDB simpler

3. **Massive horizontal scale**
   - Native sharding limited
   - Need Citus or application-level

## Configuration for Benchmarks

### Memory Settings
```sql
-- /etc/postgresql/postgresql.conf
shared_buffers = 4GB              -- 25% of RAM
effective_cache_size = 12GB       -- 75% of RAM
work_mem = 64MB                   -- Per-operation
maintenance_work_mem = 1GB        -- For VACUUM, CREATE INDEX
```

### WAL Settings
```sql
wal_buffers = 64MB
checkpoint_completion_target = 0.9
max_wal_size = 4GB
min_wal_size = 1GB
```

### Connection Settings
```sql
max_connections = 200
```

### Parallelism
```sql
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
parallel_tuple_cost = 0.01
parallel_setup_cost = 100
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| cache_hit_ratio | Buffer cache efficiency | >99% |
| tup_returned/tup_fetched | Index efficiency | Ratio high |
| xact_commit/xact_rollback | Transaction health | Low rollbacks |
| deadlocks | Concurrency issues | 0 |
| temp_files | work_mem adequacy | Minimal |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H1 | Best for complex OLTP | Exp 3, 8: Complex queries |
| H2 | ACID with serializable | Exp 8: Isolation test |
| H3 | <5ms p99 read latency | Exp 2: Read latency |
| H4 | Vertical scaling to 64 cores | Exp 8: Multi-thread |

## Index Types Comparison

| Type | Use Case | When to Use |
|------|----------|-------------|
| B-Tree | Equality, range | Default, most queries |
| Hash | Equality only | Exact match, no range |
| GiST | Geometric, full-text | PostGIS, text search |
| GIN | Arrays, JSONB, full-text | Contains, text search |
| BRIN | Large sequential data | Time-series, logs |

## References

- Documentation: https://www.postgresql.org/docs/current/
- Wiki: https://wiki.postgresql.org/
- Performance Tips: https://wiki.postgresql.org/wiki/Performance_Optimization
- Tuning Guide: https://wiki.postgresql.org/wiki/Tuning_Your_PostgreSQL_Server
