---
inclusion: manual
---

# Database Characteristics & Production Use Cases

## PostgreSQL 16 (B-Tree, MVCC)

### Storage Engine Internals
- **Index**: B+ Tree (default), also supports GiST, GIN, BRIN, Hash
- **Concurrency**: MVCC — readers never block writers, writers never block readers
- **WAL**: Write-ahead log for crash recovery; basis for replication
- **Buffer Pool**: shared_buffers caches 8KB pages; OS page cache as second tier
- **Vacuuming**: Dead tuples from MVCC must be reclaimed (autovacuum)

### Strengths
- Complex OLTP with JOINs, CTEs, window functions, subqueries
- ACID compliance with serializable isolation
- Rich extension ecosystem (PostGIS, pg_trgm, pgvector, TimescaleDB)
- Excellent query planner with cost-based optimization
- JSONB for semi-structured data without sacrificing SQL

### Weaknesses
- Write amplification from B-tree page splits + WAL + MVCC tuple copies
- Table bloat under heavy UPDATE workloads without aggressive vacuum
- Single-node scaling ceiling (no native sharding until Citus)
- Connection overhead (~10MB/conn) — needs PgBouncer for high concurrency

### Best Production Use Cases
- SaaS application backends (multi-tenant)
- Financial systems requiring ACID + complex reporting
- Geospatial applications (PostGIS)
- Full-text search (pg_trgm + GIN indexes)
- AI/ML vector similarity (pgvector)

### Production Sizing (starting point)
```
shared_buffers = 25% of RAM
effective_cache_size = 75% of RAM
work_mem = RAM / (max_connections * 4)
maintenance_work_mem = RAM / 16
max_connections = 200 (behind PgBouncer)
checkpoint_completion_target = 0.9
wal_buffers = 64MB (for write-heavy)
```

---

## MySQL 8.0 (InnoDB, B+ Tree)

### Storage Engine Internals
- **Index**: Clustered B+ Tree (primary key IS the table); secondary indexes store PK
- **Concurrency**: MVCC via undo logs in system tablespace
- **Redo Log**: Circular buffer for crash recovery (ib_logfile)
- **Buffer Pool**: innodb_buffer_pool — caches data + index pages
- **Change Buffer**: Defers secondary index updates for non-unique indexes

### Strengths
- Optimized for simple OLTP (point lookups, PK scans)
- InnoDB clustered index = PK range scans are extremely fast
- Group Replication for multi-primary HA
- Excellent replication ecosystem (semi-sync, GTID, parallel apply)
- Huge community, battle-tested at scale (Meta, Uber, Shopify)

### Weaknesses
- Query planner less sophisticated than PostgreSQL for complex analytics
- Schema changes historically painful (improved with Online DDL)
- No native JSONB-level integration (JSON functions exist but less mature)
- Double-write buffer adds write overhead for crash safety
- Clustered index means secondary index lookups require double traversal

### Best Production Use Cases
- Web applications (read-heavy OLTP)
- E-commerce product catalogs with PK-based access
- Session stores and user authentication
- Content management systems (WordPress, etc.)
- High-throughput simple transactional workloads

### Production Sizing
```
innodb_buffer_pool_size = 70-80% of RAM
innodb_log_file_size = 1-2GB
innodb_flush_log_at_trx_commit = 1 (ACID) or 2 (performance)
innodb_flush_method = O_DIRECT
max_connections = 500 (behind ProxySQL)
innodb_io_capacity = match your disk IOPS
```

---

## MongoDB 7.0 (WiredTiger, B-Tree + LSM hybrid)

### Storage Engine Internals
- **Index**: B-Tree (WiredTiger default); supports compound, multikey, text, geospatial
- **Concurrency**: Document-level locking (WiredTiger)
- **Journaling**: WiredTiger journal (WAL equivalent) — 100ms default flush
- **Compression**: Snappy (data), zlib/zstd optional; prefix compression for indexes
- **Cache**: WiredTiger internal cache (50% RAM - 1GB default)

### Strengths
- Flexible schema: no migrations for adding/removing fields
- Native horizontal scaling (sharding) with automatic balancing
- Rich query language for nested documents and arrays
- Aggregation pipeline for complex data transformations
- Change Streams for real-time event-driven architectures

### Weaknesses
- No multi-document ACID until 4.0 (and still with performance cost)
- No JOINs — $lookup exists but is slow at scale
- WiredTiger cache pressure under heavy writes can cause eviction stalls
- Sharding key choice is irreversible and critical for performance
- Larger storage footprint (document overhead, indexes per shard)

### Best Production Use Cases
- Content management with varying schemas
- Product catalogs with heterogeneous attributes
- Real-time analytics via aggregation pipeline
- IoT device management (schema varies per device type)
- Event sourcing / activity streams
- Rapid prototyping where schema evolves fast

### Production Sizing
```
wiredTiger.engineConfig.cacheSizeGB = (RAM - 1GB) / 2
storage.journal.commitIntervalMs = 100 (default, tune for durability)
replication.replSetName = "rs0"  # Always use replica set
net.maxIncomingConnections = 65536
operationProfiling.slowOpThresholdMs = 100
```

---

## ScyllaDB 5.4 (LSM-Tree, Cassandra-compatible)

### Storage Engine Internals
- **Write path**: Commit log → Memtable → Flush to SSTable (immutable sorted files)
- **Compaction**: Size-tiered (STCS), Leveled (LCS), Time-window (TWCS)
- **Concurrency**: Seastar framework — shard-per-core, no locks, no context switches
- **Bloom Filters**: Per-SSTable probabilistic filter to skip unnecessary reads
- **Read path**: Memtable → Row cache → Bloom filter → SSTable (merge multiple levels)

### Strengths
- 5-10x write throughput vs B-Tree databases (sequential I/O, no page splits)
- Predictable low-latency at high throughput (shard-per-core eliminates contention)
- Linear horizontal scaling (consistent hashing, vnodes)
- Multi-DC replication with tunable consistency
- Wire-compatible with Apache Cassandra (CQL protocol)

### Weaknesses
- Write amplification from compaction (10-50x depending on strategy)
- Read amplification: must check multiple SSTables (mitigated by bloom filters)
- No JOINs, no transactions across partitions
- Schema design is query-driven (denormalization required)
- Range scans across partitions are expensive (ALLOW FILTERING = full scan)

### Best Production Use Cases
- Time-series / event logging (high write throughput, TTL for retention)
- IoT sensor data ingestion (millions of writes/sec)
- User activity feeds and messaging (write-heavy, simple reads)
- DNS/CDN metadata (low-latency point lookups, geographically distributed)
- Fraud detection feature stores (high write, simple read patterns)

### Production Sizing
```
# ScyllaDB auto-tunes most settings
--smp <num_cores>           # Dedicate cores
--memory <RAM>              # Dedicate memory
--overprovisioned 0         # Production: disable overprovisioned mode
Compaction: TWCS for time-series, LCS for read-heavy, STCS for write-heavy
RF=3 for production, CL=QUORUM for strong consistency
```

---

## ClickHouse (MergeTree, Columnar)

### Storage Engine Internals
- **Storage**: Column-oriented — each column stored separately, heavily compressed
- **Engine Family**: MergeTree (base), ReplacingMergeTree, AggregatingMergeTree, etc.
- **Write Path**: Parts written as directories; background merges combine parts
- **Compression**: LZ4 (default), ZSTD, Delta, DoubleDelta, Gorilla per-column
- **Indexing**: Sparse primary index (granules of 8192 rows), skip indexes (minmax, bloom)

### Strengths
- 100M+ rows/sec scan speed for analytical queries
- 10-20x compression ratio on typical columnar data
- Vectorized query execution (SIMD instructions)
- Materialized views for real-time aggregation
- SQL-compatible with extensions for analytics

### Weaknesses
- Not designed for point lookups or OLTP (no row-level updates)
- Eventual consistency in distributed mode (no transactions)
- ALTER TABLE limitations (some operations require table recreation)
- JOIN performance depends heavily on table sizes and distribution
- No MVCC — mutations are async background operations

### Best Production Use Cases
- Real-time analytics dashboards (Grafana, Metabase backends)
- Log analytics (replacement for ELK stack at scale)
- Ad-tech click/impression analytics
- Network monitoring and observability
- Business intelligence on billions of rows
- A/B testing result analysis

### Production Sizing
```
max_memory_usage = 80% of RAM
max_threads = number of CPU cores
merge_tree.max_bytes_to_merge_at_max_space_in_pool = (disk_size * 0.2)
Replicated tables: ReplicatedMergeTree with ZooKeeper/ClickHouse Keeper
Sharding: Distributed table engine across shards
```

---

## TimescaleDB (PostgreSQL extension, Hypertables)

### Storage Engine Internals
- **Base**: PostgreSQL B-Tree engine with automatic time-based partitioning
- **Hypertables**: Transparent chunking by time interval (default 7 days)
- **Compression**: Columnar compression on old chunks (90-95% reduction)
- **Continuous Aggregates**: Materialized views that auto-refresh
- **Data Retention**: Automated chunk dropping based on age policies

### Strengths
- Full PostgreSQL SQL compatibility (JOINs, CTEs, window functions on time-series)
- Automatic partitioning without manual management
- Compression converts old data to columnar format
- Continuous aggregates for real-time dashboards
- Native PostgreSQL tooling (pg_dump, replication, extensions)

### Weaknesses
- Single-node PostgreSQL scaling limits apply
- Write throughput lower than native LSM databases for pure append workloads
- Chunk management adds some overhead vs raw PostgreSQL
- Multi-node (distributed hypertables) requires enterprise license
- Memory usage can spike during compression/decompression

### Best Production Use Cases
- Application metrics and monitoring (where you need SQL JOINs with metadata)
- Financial tick data with complex analytics
- IoT when you need relational JOINs with device metadata
- DevOps observability (metrics + logs correlation)
- Energy/utility metering with regulatory reporting

### Production Sizing
```
# Same as PostgreSQL base + TimescaleDB specifics
timescaledb.max_background_workers = 8
chunk_time_interval = '1 day' to '7 days' (based on ingest rate)
Compression: enable on chunks older than X (e.g., 7 days)
Retention: drop_chunks older than Y (e.g., 90 days)
Continuous aggregates: refresh every 1h for dashboards
```

---

## Redis 7 (In-Memory Hash Table)

### Storage Engine Internals
- **Data Structure**: Hash table (main dict) with incremental rehashing
- **Types**: Strings, Lists (quicklist), Sets (intset/hashtable), Sorted Sets (skiplist+HT), Hashes, Streams
- **Persistence**: RDB (point-in-time snapshots), AOF (append-only log), hybrid
- **Eviction**: LRU, LFU, TTL-based, noeviction (OOM error)
- **Threading**: Single-threaded command execution (6.0+ I/O threads for network)

### Strengths
- Sub-millisecond latency for all operations (memory-resident)
- Rich data structures (sorted sets, streams, HyperLogLog, bitmaps)
- Pub/Sub and Streams for real-time messaging
- Lua scripting for atomic multi-step operations
- Cluster mode for horizontal scaling (hash slots)

### Weaknesses
- Dataset limited by RAM (expensive at scale)
- Single-threaded command execution (CPU bound on complex ops)
- Persistence adds latency (fork for RDB, fsync for AOF)
- Cluster mode: no multi-key operations across slots (unless same hash tag)
- No query language — application must know key patterns

### Best Production Use Cases
- Session store / token cache (TTL-based expiry)
- Rate limiting (INCR + EXPIRE atomic)
- Leaderboards (sorted sets)
- Real-time pub/sub messaging
- Distributed locks (Redlock algorithm)
- Caching layer for hot database queries
- Job queues (Lists or Streams)

### Production Sizing
```
maxmemory = 70-80% of available RAM
maxmemory-policy = allkeys-lfu (general cache) or volatile-lfu (with TTLs)
appendonly yes + appendfsync everysec (balanced durability)
Cluster: 3 masters + 3 replicas minimum
tcp-backlog = 511
timeout = 300 (close idle connections)
```

---

## Decision Matrix: Choosing the Right Database

| Requirement | Best Choice | Runner-up |
|-------------|-------------|-----------|
| Complex OLTP + JOINs | PostgreSQL | MySQL |
| Simple high-throughput OLTP | MySQL | PostgreSQL |
| Write-heavy time-series | ScyllaDB | TimescaleDB |
| Time-series with SQL analytics | TimescaleDB | ClickHouse |
| Sub-ms caching | Redis | — |
| Flexible schema + rapid iteration | MongoDB | PostgreSQL (JSONB) |
| Billions-row analytics | ClickHouse | TimescaleDB |
| Multi-DC geo-distribution | ScyllaDB | MongoDB |
| Full-text search | PostgreSQL (GIN) | MongoDB (Atlas Search) |
| Real-time leaderboards | Redis (ZSET) | — |
