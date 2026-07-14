# Database Lab Benchmark

Comprehensive benchmark suite for comparing **19 databases** across different storage engines, topologies, and use cases. Prove or disprove vendor claims with real data.

## 🎯 Goal

Validate vendor claims (hypotheses H1–H91) with reproducible benchmark data across six storage engine families:

| Family | Databases | Storage Model |
|--------|-----------|---------------|
| **B-Tree** | PostgreSQL 16, MySQL 8.0, MongoDB 7 (WiredTiger), SQLite | Sorted pages, in-place updates |
| **LSM-Tree** | ScyllaDB 5.4, Cassandra 4.1, CockroachDB (Pebble) | Memtable → SSTable flush, compaction |
| **Columnar** | ClickHouse | Column-oriented, MergeTree family |
| **Time-Series** | TimescaleDB, InfluxDB 2.7, Prometheus 2.53 | Hypertables / TSM / TSDB |
| **In-Memory** | Redis 7, Valkey 7.2, DragonflyDB | Hash table, optional persistence |
| **Specialized** | OpenSearch (Inverted Index), Neo4j (Graph), Milvus (Vector), Qdrant (Vector), pgvector (PG Vector) | Domain-optimized |

## 📊 Benchmark Workloads (Go)

All 19 databases run the same standardized benchmark:

| Workload | Description | Key Metric |
|----------|-------------|------------|
| `write` | Concurrent upsert throughput | ops/s, p50/p95/p99 latency |
| `read` | Random point lookup latency | ops/s, p50/p95/p99 latency |
| `mixed` | 80% read / 20% write | ops/s, p50/p95/p99 latency |

```bash
cd benchmarks/go
go run . write    # Write throughput for all 19 DBs
go run . read     # Read latency for all 19 DBs
go run . mixed    # Mixed workload (80/20)
go run . all      # Run all benchmarks

# Selective
BENCH_DBS=postgres,mysql,redis go run . write
BENCH_ROWS=1000000 BENCH_CONCURRENCY=50 go run . all
```

## 🐳 Lab Topologies (Near-Production)

| Database | Standalone | Replication | Cluster/Sharding | Pooling | Notes |
|----------|:---:|:---:|:---:|:---:|-------|
| PostgreSQL | ✅ | ✅ Streaming | — | ✅ PgBouncer | Primary + Replica |
| MySQL | ✅ | ✅ GTID | ✅ Group Repl | — | Source + Replica |
| MongoDB | ✅ | ✅ Replica Set (3) | ✅ Sharded (2 shards) | — | Full topology |
| Redis | ✅ | ✅ Sentinel (3+3) | ✅ Cluster (6 nodes) | — | Full topology |
| Valkey | ✅ | ✅ Sentinel (3+3) | — | — | Redis-compatible HA |
| DragonflyDB | ✅ | ✅ Master-Replica | — | — | Multi-threaded HA |
| ScyllaDB | ✅ | — | ✅ 3-node cluster | — | Shard-per-core |
| Cassandra | ✅ | — | ✅ 3-node cluster | — | Gossip-based ring |
| ClickHouse | ✅ | — | ✅ 2×2 + Keeper | — | Sharded + Replicated |
| CockroachDB | ✅ | — | ✅ 3-node Raft | — | Distributed SQL |
| TimescaleDB | ✅ | ✅ Streaming | — | — | PG-based replication |
| OpenSearch | ✅ | — | ✅ 3-node cluster | — | Shard distribution |
| Neo4j | ✅ | — | ✅ 3 Core + Reader | — | Raft consensus |
| InfluxDB | ✅ | ❌ OSS only | ❌ Enterprise only | — | Single-node limitation |
| Prometheus | ✅ | — | ❌ Use Thanos/Mimir | — | Monitoring TSDB |
| Milvus | ✅ | — | ✅ Distributed | — | Separated compute/storage |
| Qdrant | ✅ | — | ✅ 3-node (Raft) | — | Rust vector DB |
| pgvector | ✅ | (inherits PG) | — | — | PG extension for vectors |
| SQLite | ✅ | — | — | — | Embedded, no server |

## 📋 Experiments & Hypotheses

| # | Experiment | Databases | Validates |
|---|------------|-----------|-----------|
| 1 | Write Throughput | All 19 | H1, H5, H17, H53, H61, H66, H72, H74, H75, H83 |
| 2 | Read Latency | All 19 | H3, H13, H52, H54, H62, H67, H70, H74, H76, H78, H88 |
| 3 | Complex Query | PG, MySQL, Mongo, CockroachDB | H1, H9, H58, H59 |
| 4 | Schema Evolution | PG, MySQL, Mongo, Scylla | H9, H24, H26 |
| 5 | Horizontal Scaling | Scylla, Mongo, CockroachDB, CH, Qdrant | H7, H28, H48, H57, H82 |
| 6 | Write Amplification | PG, Scylla, Cassandra | H19 |
| 7 | Time-Series Ingest | Timescale, InfluxDB, Prometheus, Scylla, CH | H6, H40, H53, H54, H83-H86 |
| 8 | MySQL vs PostgreSQL | MySQL, PG | H21-H25 |
| 9 | MongoDB vs SQL | Mongo, PG, MySQL | H26-H30 |
| 10 | OLAP Analytics | ClickHouse, PG, TimescaleDB | H36-H39 |
| 11 | Time-Series Battle | Timescale, InfluxDB, Prometheus, Mongo TS, CH | H40-H43, H53-H56, H83-H86 |
| 12 | In-Memory Battle | Redis vs Valkey vs DragonflyDB | H61-H69 |
| 13 | Graph Traversal | Neo4j vs PG (recursive CTE) | H49-H52 |
| 14 | Vector Search | Milvus vs Qdrant vs pgvector | H70-H73, H78-H82, H87-H91 |
| 15 | Full-Text Search | OpenSearch vs PG tsvector | H44-H48 |
| 16 | Distributed SQL | CockroachDB vs PG (latency cost) | H57-H60 |

### DDIA & Database Internals — Deep Experiments

| # | Experiment | Databases | Category | Validates |
|---|------------|-----------|----------|-----------|
| WA-01 | Write Amplification Ratio | PG, MySQL, ScyllaDB, Cassandra | Storage Engine | disk_bytes / logical_bytes |
| WA-02 | WAL/Binlog Growth | PG, MySQL, MongoDB | Storage Engine | MB/s WAL growth |
| WA-03 | Compaction I/O | ScyllaDB, Cassandra | Storage Engine | IOPS during compaction |
| RA-01 | Read Amplification (key exists) | All | Storage Engine | IOPS per read |
| RA-02 | Negative Lookup (key not exists) | ScyllaDB, Cassandra | Storage Engine | Bloom filter efficiency |
| RUM-01 | RUM Triangle Plot | All 19 | Storage Engine | R/U/M radar chart |
| BT-01 | Page Split Frequency | PG, MySQL | B-Tree | Splits/1000 inserts |
| BT-02 | Page Split Latency Spike | PG, MySQL | B-Tree | p99 during splits |
| BT-03 | Fillfactor Impact | PG | B-Tree | Split reduction % |
| BT-04 | Tree Depth at Scale | PG, MySQL | B-Tree | Depth at 1M/10M/100M rows |
| LSM-01 | Compaction Latency Spike | ScyllaDB, Cassandra | LSM-Tree | p99 read during compaction |
| LSM-02 | Compaction I/O Saturation | ScyllaDB, Cassandra | LSM-Tree | IOPS % used |
| LSM-03 | Level Fanout Strategies | ScyllaDB | LSM-Tree | Throughput stability |
| LSM-05 | Space Amplification | ScyllaDB, Cassandra | LSM-Tree | Actual vs logical ratio |
| DM-01 | Document Locality | MongoDB, PG (JSONB) | Data Model | Partial vs full read latency |
| DM-02 | JOIN at Scale | PG, MySQL, CockroachDB | Data Model | 2/3/5-table JOIN latency |
| GR-01 | Depth-1 Traversal | Neo4j, PG | Graph | Direct neighbor latency |
| GR-02 | Depth-3 Traversal | Neo4j, PG | Graph | 3-hop latency |
| REP-01 | Replication Lag (sync) | PG, MySQL (semisync) | Replication | ms write→replica visible |
| REP-02 | Replication Lag (async) | PG, MongoDB | Replication | Eventual consistency time |
| CON-01 | Leader Election Time | CockroachDB, Neo4j | Consensus | Failover duration (ms) |
| CON-02 | Consensus Round-Trip | CockroachDB (Raft) | Consensus | Write latency overhead |
| BAT-01 | Bulk Insert (COPY) | PG, MySQL, ClickHouse | Batch | rows/sec via bulk loader |
| BAT-02 | Batch vs Single Insert | All | Batch | Throughput ratio |
| TS-01 | Tumbling Window Aggregate | TimescaleDB, InfluxDB, CH | Time-Series | 5-min agg latency |
| TS-02 | Sliding Window | TimescaleDB, InfluxDB | Time-Series | Moving average latency |
| DIST-01 | Hot Key / Skewed Writes | Redis Cluster, Cassandra | Distributed | Hotspot node CPU |
| DIST-02 | Zipfian Distribution | All distributed | Distributed | Throughput degradation |
| PART-01 | Hash vs Range Partition | CockroachDB, Cassandra | Distributed | Query latency |
| PART-02 | Cross-Partition Query | CockroachDB, ScyllaDB | Distributed | Scatter-gather overhead |
| MEM-01 | Cold vs Hot Cache | PG, MySQL | In-Memory | First vs subsequent read |
| MEM-02 | Memory Pressure | Redis, DragonflyDB | In-Memory | Throughput at 80%/90%/100% mem |
| MEM-03 | Persistence Overhead | Redis, Valkey | In-Memory | AOF vs none write latency |
| VEC-01 | Recall vs Latency | Milvus, Qdrant, pgvector | Vector | recall@10 at target latency |
| VEC-02 | Index Build Time | Milvus, Qdrant, pgvector | Vector | HNSW build on 1M vectors |
| VEC-03 | Dimension Scaling | All vector DBs | Vector | 128 vs 768 vs 1536 dim |
| VEC-04 | Hybrid Search | Milvus, Qdrant | Vector | Vector + metadata filter |

## 🚀 Quick Start

### Option A: Go Benchmark (Recommended)

```bash
# Start database(s)
cd docker/postgres && docker compose up -d
cd docker/redis && docker compose up -d

# Run benchmark
cd benchmarks/go
BENCH_DBS=postgres,redis go run . all

# Cleanup
cd docker/postgres && docker compose down -v
cd docker/redis && docker compose down -v
```

### Option B: Near-Production Topology

```bash
# Start a 3-node cluster
cd docker/scylladb && docker compose -f cluster.yml up -d

# Wait for all nodes to be UN (Up/Normal)
docker exec scylla-node1 nodetool status

# Run benchmark against cluster
cd benchmarks/go
BENCH_DBS=scylladb BENCH_SCYLLA_HOSTS=localhost go run . all

# Cleanup
cd docker/scylladb && docker compose -f cluster.yml down -v
```

### Option C: EC2 (Full suite, recommended for accurate results)

```bash
cd terraform
terraform init && terraform apply -var="key_name=<your-key>"
ssh ubuntu@<ip>
./scripts/run_all.sh
```

## 💰 Cost Estimate

| Instance | Spec | Hourly | 4 Hours |
|----------|------|--------|---------|
| m6i.2xlarge (on-demand) | 8 vCPU, 32GB | $0.384 | ~$1.50 |
| m6i.2xlarge (spot) | 8 vCPU, 32GB | $0.12 | ~$0.50 |
| m6i.4xlarge (on-demand) | 16 vCPU, 64GB | $0.768 | ~$3.00 |

## 📁 Structure

```
database-lab-benchmark/
├── benchmarks/
│   ├── go/                   # Go benchmark runner (all 19 DBs)
│   │   ├── bench/            # Drivers: driver_postgres.go, driver_qdrant.go, etc.
│   │   └── main.go           # CLI: write, read, mixed, all, list
│   └── write-throughput/     # Python LSM-tree benchmarks
├── docker/                   # Docker Compose per database
│   ├── postgres/             # docker-compose.yml, replication.yml, pooling.yml
│   ├── mysql/                # docker-compose.yml, replication.yml, group-replication.yml
│   ├── mongodb/              # docker-compose.yml, replication.yml, sharding.yml
│   ├── redis/                # docker-compose.yml, replication.yml, cluster.yml
│   ├── valkey/               # docker-compose.yml, replication.yml
│   ├── dragonflydb/          # docker-compose.yml, replication.yml
│   ├── scylladb/             # docker-compose.yml, cluster.yml
│   ├── cassandra/            # docker-compose.yml, cluster.yml
│   ├── clickhouse/           # docker-compose.yml, cluster.yml + config/
│   ├── cockroachdb/          # docker-compose.yml, cluster.yml
│   ├── timescaledb/          # docker-compose.yml, replication.yml
│   ├── opensearch/            # docker-compose.yml, cluster.yml
│   ├── neo4j/                # docker-compose.yml, cluster.yml
│   ├── influxdb/             # docker-compose.yml (OSS: no clustering)
│   ├── prometheus/           # docker-compose.yml (single-node TSDB)
│   ├── milvus/               # docker-compose.yml, cluster.yml
│   ├── qdrant/               # docker-compose.yml, cluster.yml
│   ├── pgvector/             # docker-compose.yml (PG + vector extension)
│   └── sqlite/               # README.md (embedded, no Docker needed)
├── docs/
│   ├── official-references/  # Vendor claims & config per DB (19 files)
│   ├── REPLICATION.md
│   ├── BACKUP.md
│   └── POOLING.md
├── results/                  # Benchmark JSON outputs
└── scripts/                  # Orchestration
```

## 📋 Key Hypotheses (H1–H91)

| ID | Claim | Source | How We Test |
|----|-------|--------|-------------|
| H1 | PostgreSQL best for complex OLTP | postgresql.org | Exp 3, 8: Complex queries |
| H5 | ScyllaDB >10K writes/sec sustained | scylladb.com | Exp 1: Write throughput |
| H13 | Redis sub-millisecond latency | redis.io | Exp 2: Read latency |
| H21 | MySQL best for web apps | mysql.com | Exp 8: Simple OLTP |
| H26 | MongoDB flexible schema, sub-ms | mongodb.com | Exp 9: Document workload |
| H36 | ClickHouse 100M+ rows/sec scan | clickhouse.com | Exp 10: Full scan |
| H40 | TimescaleDB 10-100x faster than PG | timescale.com | Exp 11: Time-range queries |
| H44 | OpenSearch near real-time search | opensearch.org | Exp 15: Index + search |
| H49 | Neo4j constant-time traversal | neo4j.com | Exp 13: Graph vs recursive CTE |
| H53 | InfluxDB 750K+ values/sec | influxdata.com | Exp 7: Time-series ingest |
| H57 | CockroachDB serializable distributed | cockroachlabs.com | Exp 16: ACID + latency cost |
| H61 | DragonflyDB 25x throughput vs Redis | dragonflydb.io | Exp 12: Side-by-side |
| H66 | Valkey same perf as Redis | valkey.io | Exp 12: Drop-in comparison |
| H70 | Milvus ms-level vector search | milvus.io | Exp 14: ANN search |
| H74 | SQLite fastest single-process reads | sqlite.org | Exp 2: In-process vs network |
| H78 | Qdrant highest RPS among vector DBs | qdrant.tech | Exp 14: Vector DB battle |
| H83 | Prometheus 2M+ samples/sec ingest | prometheus.io | Exp 7: Remote write throughput |
| H87 | pgvector 2,300 QPS with HNSW | github.com/pgvector | Exp 14: PG vector search |

Full hypothesis list in `docs/official-references/`.

## 🔬 Benchmark Results

### Experiment 12: In-Memory Battle (July 14, 2026)

**Setup**: AWS EC2 m6i.2xlarge (spot), ap-southeast-3 (Jakarta), 8 vCPU / 32GB RAM
**Config**: 10K rows, 10 goroutines, 100B values, 1 run (light mode)

| Workload | Redis 7 | Valkey 7.2 | DragonflyDB | Winner |
|----------|---------|------------|-------------|--------|
| **Write throughput** | 61,482 ops/s | **63,175 ops/s** | 58,997 ops/s | Valkey |
| **Read latency** | **63,459 ops/s** | 63,051 ops/s | 60,358 ops/s | Redis |
| **Mixed 80/20** | 63,623 ops/s | **64,475 ops/s** | 61,533 ops/s | Valkey |

**Latency (all workloads)**:
| Metric | Redis | Valkey | DragonflyDB |
|--------|-------|--------|-------------|
| p50 | 0.15ms | 0.15ms | 0.16ms |
| p95 | 0.22-0.23ms | 0.22-0.23ms | 0.24-0.25ms |
| p99 | 0.26-0.28ms | 0.26-0.29ms | 0.29-0.31ms |

**Hypothesis Validation**:
| Hypothesis | Claim | Result |
|------------|-------|--------|
| H13 | Redis sub-millisecond latency | ✅ Confirmed (p99 = 0.28ms) |
| H66 | Valkey same performance as Redis | ✅ Confirmed (within 3%, sometimes faster) |
| H67 | Valkey sub-millisecond latency | ✅ Confirmed (p99 = 0.29ms) |
| H61 | DragonflyDB 25x throughput vs Redis | ❌ Not observed at 10 concurrency / 2 CPU |
| H62 | DragonflyDB sub-millisecond p99 | ✅ Confirmed (p99 = 0.31ms) |

**Key Insight**: At low concurrency (10 goroutines) and limited CPU (2 cores), DragonflyDB's multi-threaded advantage doesn't manifest. Its shard-per-core architecture needs higher concurrency (50+ threads) to outperform Redis's single-threaded model. Re-test with `BENCH_CONCURRENCY=50` on a 4+ CPU instance to validate H61.

### Experiment 12b: High Concurrency H61 Validation (July 14, 2026)

**Setup**: AWS EC2 m6i.2xlarge (on-demand), ap-southeast-3, 8 vCPU / 32GB RAM
**Config**: 100K rows, **50 goroutines**, 100B values, 3 runs (median)
**Images**: Redis 8-alpine, Valkey 8.1-alpine, DragonflyDB latest

| Workload | Redis 8 | Valkey 8.1 | DragonflyDB | Dragonfly vs Redis |
|----------|---------|------------|-------------|-----|
| **Write** | 96,857 ops/s | 96,742 ops/s | **116,073 ops/s** | **1.20x** |
| **Read** | 101,118 ops/s | 101,024 ops/s | **120,480 ops/s** | **1.19x** |
| **Mixed 80/20** | 98,236 ops/s | 99,352 ops/s | **119,233 ops/s** | **1.21x** |

| Metric | Redis 8 | Valkey 8.1 | DragonflyDB |
|--------|---------|------------|-------------|
| p50 | 0.48-0.50ms | 0.48-0.50ms | **0.39ms** |
| p95 | 0.72-0.75ms | 0.73-0.76ms | 0.73-0.75ms |
| p99 | 0.89-0.94ms | 0.89-0.92ms | 0.96-1.02ms |

**H61 Verdict**: DragonflyDB consistently **~1.2x faster** throughput and **~20% lower p50 latency** on 2 CPU threads. The "25x" claim requires 64-core machines. At 2 threads, the multi-threaded architecture still provides a clear median latency advantage.

### Experiment 8: MySQL 8.4 vs PostgreSQL 17 (July 14, 2026)

**Setup**: Same instance (m6i.2xlarge on-demand, ap-southeast-3)
**Config**: 100K rows, 50 goroutines, 100B values, 3 runs (median)
**Images**: postgres:17-alpine, mysql:8.4

| Workload | PostgreSQL 17 | MySQL 8.4 | Winner |
|----------|--------------|-----------|--------|
| **Write throughput** | 4,189 ops/s | **7,584 ops/s** | MySQL (1.81x) |
| **Read throughput** | **29,210 ops/s** | 20,806 ops/s | PostgreSQL (1.40x) |
| **Mixed 80/20** | 17,344 ops/s | **19,593 ops/s** | MySQL (1.13x) |

| Metric | PostgreSQL 17 | MySQL 8.4 |
|--------|--------------|-----------|
| Write p50 | 11.55ms | **6.26ms** |
| Write p99 | **13.51ms** | 11.38ms |
| Read p50 | 1.33ms | **1.27ms** |
| Read p99 | **22.68ms** | 46.38ms |
| Mixed p50 | 2.72ms | **1.28ms** |
| Mixed p99 | **5.13ms** | 48.33ms |

**Hypothesis Validation**:
| Hypothesis | Claim | Result |
|------------|-------|--------|
| H1 | PostgreSQL best for read-heavy OLTP | ✅ Read throughput 1.40x faster |
| H21 | MySQL best for web apps (writes) | ✅ Write throughput 1.81x faster |
| H22 | MySQL simple OLTP fast | ✅ Lower p50 across all workloads |

**Key Insight**: MySQL 8.4 shows **even stronger write advantage** (1.81x vs previous 1.53x on 8.0) — InnoDB improvements in 8.4 are real. However, MySQL **tail latency is terrible** (p99 = 46-48ms) while PostgreSQL stays tight (p99 = 5-22ms). For latency-sensitive apps, PG's predictability wins; for raw write throughput, MySQL dominates.

## 📖 References

- [DDIA Chapter 3: Storage & Retrieval](https://dataintensive.net/)
- [PostgreSQL Docs](https://postgresql.org/docs/)
- [ScyllaDB Docs](https://docs.scylladb.com/)
- [MongoDB Docs](https://docs.mongodb.com/)
- [Redis Docs](https://redis.io/docs/)
- [ClickHouse Docs](https://clickhouse.com/docs/)
- [OpenSearch Docs](https://opensearch.org/docs/)
- [Neo4j Docs](https://neo4j.com/docs/)
- [CockroachDB Docs](https://cockroachlabs.com/docs/)
- [TimescaleDB Docs](https://docs.timescale.com/)
- [InfluxDB Docs](https://docs.influxdata.com/)
- [Prometheus Docs](https://prometheus.io/docs/)
- [DragonflyDB Docs](https://dragonflydb.io/docs)
- [Valkey Docs](https://valkey.io/docs/)
- [Milvus Docs](https://milvus.io/docs)
- [Qdrant Docs](https://qdrant.tech/documentation/)
- [pgvector](https://github.com/pgvector/pgvector)
- [SQLite Docs](https://sqlite.org/docs.html)

## License

MIT
