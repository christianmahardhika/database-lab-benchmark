# SQLite Official Documentation Reference

Source: https://www.sqlite.org/docs.html

## Overview

SQLite is a self-contained, serverless, zero-configuration, transactional SQL database engine. Most widely deployed database in the world (billions of instances). Single-file database, embedded in the application process.

## Storage Engine: B-Tree (Single File)

### Architecture
- **Single file**: Entire database in one file (`.db`)
- **Page-based**: Fixed-size pages (default 4096 bytes)
- **B-Tree**: Tables stored as B-Trees (rowid) or B+Trees (WITHOUT ROWID)
- **WAL mode**: Write-ahead logging for concurrent reads

### Page Structure
```
Database file → Pages (4KB each)
├── Page 1: File header + schema table
├── Pages 2-N: B-Tree interior/leaf pages
└── Overflow pages for large values
```

### Write Path (WAL mode)
```
INSERT → WAL file (append) → Checkpoint (batch write to main DB)
```

### Read Path
```
SELECT → B-Tree traversal → Page cache (in-process) → Return
```

### Locking Modes
| Mode | Readers | Writers | Use Case |
|------|---------|---------|----------|
| DELETE (default) | Multiple | One at a time | Simple apps |
| WAL | Multiple concurrent | One at a time | High-read concurrency |
| WAL2 | Multiple concurrent | Improved | Experimental |

## Performance Characteristics (Official Claims)

### Read Performance
- **Claim**: "Faster than filesystem for small blobs (<100KB)"
- **Benchmark**: 35% faster than fread() for 10KB blobs (official)
- **Why**: Single syscall vs open+read+close, page cache efficient

### Write Performance
- **Claim**: "50,000-100,000 INSERTs per second (transaction batched)"
- **Single INSERT**: ~60/sec (one transaction per INSERT, fsync)
- **Batched**: 50K+/sec (all in one transaction)
- **WAL mode**: Better write concurrency

### Suitability
- **Claim**: "Appropriate for databases up to ~281TB"
- **Practical**: Excellent up to ~1GB, good up to ~100GB
- **Caveat**: Single writer, so write-heavy > 1 thread limited

## Best Use Cases (from sqlite.org)

1. **Embedded/Mobile Applications**
   - Android/iOS local storage
   - Desktop applications
   - IoT devices

2. **Application File Format**
   - Replace custom binary formats
   - Self-describing, portable, queryable

3. **Testing and Prototyping**
   - No server setup needed
   - In-memory mode for unit tests

4. **Small-to-Medium Websites**
   - <100K hits/day
   - Read-heavy workloads
   - Single-machine deployment

5. **Data Analysis / ETL**
   - Import CSV, query with SQL
   - Intermediate storage for pipelines

## Anti-Patterns (from sqlite.org)

1. **High-concurrency writes**
   - Single writer limitation
   - Use PostgreSQL/MySQL for multi-writer

2. **Client-server architecture**
   - No network protocol
   - Embedded only (same process)

3. **Very large datasets (>1TB)**
   - B-Tree depth increases, slower
   - Use dedicated DBMS

4. **High write throughput requirements**
   - fsync overhead per transaction
   - Batch transactions or use WAL

## Configuration for Benchmarks

### WAL Mode (recommended)
```sql
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;  -- WAL makes this safe
PRAGMA wal_autocheckpoint=1000;
```

### Performance PRAGMAs
```sql
PRAGMA cache_size=-64000;      -- 64MB page cache
PRAGMA mmap_size=268435456;    -- 256MB memory-mapped I/O
PRAGMA temp_store=MEMORY;      -- Temp tables in RAM
PRAGMA page_size=4096;         -- Default, good for SSD
```

### Busy Timeout
```sql
PRAGMA busy_timeout=5000;      -- Wait 5s for lock instead of failing
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| Inserts/sec (batched) | Write throughput | 50K-100K |
| Single-row read latency | Point lookup | <0.01ms (in-process) |
| Database file size | Storage | Monitor growth |
| WAL file size | Write backlog | Auto-checkpointed |
| Page cache hit ratio | Memory efficiency | >99% for hot data |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H74 | Fastest for single-process reads | Exp 2: Read latency vs network DBs |
| H75 | 50K+ inserts/sec (batched) | Exp 1: Write throughput in single tx |
| H76 | Zero network latency advantage | Exp: Compare vs localhost PostgreSQL |
| H77 | Single-writer limitation visible | Exp: Concurrent write throughput |

## Comparison: SQLite vs Alternatives

| Aspect | SQLite | PostgreSQL | Redis |
|--------|--------|------------|-------|
| Deployment | Embedded (no server) | Client-server | Client-server |
| Concurrency | Single writer, multi reader | Full MVCC | Single-thread |
| Latency | ~0ms (in-process) | ~0.5ms (localhost) | ~0.1ms (localhost) |
| Write throughput | 50K-100K/s (batched) | 20K-50K/s | 100K+/s |
| Durability | Full (fsync) | Full (WAL) | Optional (AOF) |
| Max size | ~281TB (practical ~100GB) | Unlimited | RAM-limited |
| SQL | Full (most of SQL92) | Full (SQL:2016) | None |

## References

- Documentation: https://www.sqlite.org/docs.html
- When to Use: https://www.sqlite.org/whentouse.html
- Performance: https://www.sqlite.org/fasterthanfs.html
- WAL Mode: https://www.sqlite.org/wal.html
- Limits: https://www.sqlite.org/limits.html
