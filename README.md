# Database Lab Benchmark

Comprehensive benchmark suite for comparing **16 databases** across different storage engines, topologies, and use cases. Prove or disprove vendor claims with real data.

## 🎯 Goal

Validate vendor claims (hypotheses H1–H77) with reproducible benchmark data across five storage engine families:

| Family | Databases | Storage Model |
|--------|-----------|---------------|
| **B-Tree** | PostgreSQL 16, MySQL 8.0, MongoDB 7 (WiredTiger), SQLite | Sorted pages, in-place updates |
| **LSM-Tree** | ScyllaDB 5.4, Cassandra 4.1, CockroachDB (Pebble) | Memtable → SSTable flush, compaction |
| **Columnar** | ClickHouse | Column-oriented, MergeTree family |
| **Time-Series** | TimescaleDB, InfluxDB 2.7 | Hypertables / TSM engine |
| **In-Memory** | Redis 7, Valkey 7.2, DragonflyDB | Hash table, optional persistence |
| **Specialized** | Elasticsearch (Inverted Index), Neo4j (Graph), Milvus (Vector) | Domain-optimized |

## 📊 Benchmark Workloads (Go)

All 16 databases run the same standardized benchmark:

| Workload | Description | Key Metric |
|----------|-------------|------------|
| `write` | Concurrent upsert throughput | ops/s, p50/p95/p99 latency |
| `read` | Random point lookup latency | ops/s, p50/p95/p99 latency |
| `mixed` | 80% read / 20% write | ops/s, p50/p95/p99 latency |

```bash
cd benchmarks/go
go run . write    # Write throughput for all 16 DBs
go run . read     # Read latency for all 16 DBs
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
| Elasticsearch | ✅ | — | ✅ 3-node cluster | — | Shard distribution |
| Neo4j | ✅ | — | ✅ 3 Core + Reader | — | Raft consensus |
| InfluxDB | ✅ | ❌ OSS only | ❌ Enterprise only | — | Single-node limitation |
| Milvus | ✅ | — | ✅ Distributed | — | Separated compute/storage |
| SQLite | ✅ | — | — | — | Embedded, no server |

## 📋 Experiments & Hypotheses

| # | Experiment | Databases | Validates |
|---|------------|-----------|-----------|
| 1 | Write Throughput | All 16 | H1, H5, H17, H53, H61, H66, H72, H74, H75 |
| 2 | Read Latency | All 16 | H3, H13, H52, H54, H62, H67, H70, H74, H76 |
| 3 | Complex Query | PG, MySQL, Mongo, CockroachDB | H1, H9, H58, H59 |
| 4 | Schema Evolution | PG, MySQL, Mongo, Scylla | H9, H24, H26 |
| 5 | Horizontal Scaling | Scylla, Mongo, CockroachDB, CH | H7, H28, H48, H57 |
| 6 | Write Amplification | PG, Scylla, Cassandra | H19 |
| 7 | Time-Series Ingest | Timescale, InfluxDB, Scylla, CH | H6, H40, H53, H54 |
| 8 | MySQL vs PostgreSQL | MySQL, PG | H21-H25 |
| 9 | MongoDB vs SQL | Mongo, PG, MySQL | H26-H30 |
| 10 | OLAP Analytics | ClickHouse, PG, TimescaleDB | H36-H39 |
| 11 | Time-Series Battle | Timescale, InfluxDB, Mongo TS, CH | H40-H43, H53-H56 |
| 12 | In-Memory Battle | Redis vs Valkey vs DragonflyDB | H61-H69 |
| 13 | Graph Traversal | Neo4j vs PG (recursive CTE) | H49-H52 |
| 14 | Vector Search | Milvus (HNSW vs FLAT) | H70-H73 |
| 15 | Full-Text Search | Elasticsearch vs PG tsvector | H44-H48 |
| 16 | Distributed SQL | CockroachDB vs PG (latency cost) | H57-H60 |

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
│   ├── go/                   # Go benchmark runner (all 16 DBs)
│   │   ├── bench/            # Drivers: driver_postgres.go, driver_redis.go, etc.
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
│   ├── elasticsearch/        # docker-compose.yml, cluster.yml
│   ├── neo4j/                # docker-compose.yml, cluster.yml
│   ├── influxdb/             # docker-compose.yml (OSS: no clustering)
│   ├── milvus/               # docker-compose.yml, cluster.yml
│   └── sqlite/               # README.md (embedded, no Docker needed)
├── docs/
│   ├── official-references/  # Vendor claims & config per DB
│   ├── REPLICATION.md
│   ├── BACKUP.md
│   └── POOLING.md
├── results/                  # Benchmark JSON outputs
└── scripts/                  # Orchestration
```

## 📋 Key Hypotheses (H1–H77)

| ID | Claim | Source | How We Test |
|----|-------|--------|-------------|
| H1 | PostgreSQL best for complex OLTP | postgresql.org | Exp 3, 8: Complex queries |
| H5 | ScyllaDB >10K writes/sec sustained | scylladb.com | Exp 1: Write throughput |
| H13 | Redis sub-millisecond latency | redis.io | Exp 2: Read latency |
| H21 | MySQL best for web apps | mysql.com | Exp 8: Simple OLTP |
| H26 | MongoDB flexible schema, sub-ms | mongodb.com | Exp 9: Document workload |
| H36 | ClickHouse 100M+ rows/sec scan | clickhouse.com | Exp 10: Full scan |
| H40 | TimescaleDB 10-100x faster than PG | timescale.com | Exp 11: Time-range queries |
| H44 | Elasticsearch near real-time search | elastic.co | Exp 15: Index + search |
| H49 | Neo4j constant-time traversal | neo4j.com | Exp 13: Graph vs recursive CTE |
| H53 | InfluxDB 750K+ values/sec | influxdata.com | Exp 7: Time-series ingest |
| H57 | CockroachDB serializable distributed | cockroachlabs.com | Exp 16: ACID + latency cost |
| H61 | DragonflyDB 25x throughput vs Redis | dragonflydb.io | Exp 12: Side-by-side |
| H66 | Valkey same perf as Redis | valkey.io | Exp 12: Drop-in comparison |
| H70 | Milvus ms-level vector search | milvus.io | Exp 14: ANN search |
| H74 | SQLite fastest single-process reads | sqlite.org | Exp 2: In-process vs network |

Full hypothesis list in `docs/official-references/`.

## 🔬 Previous Results

From local benchmark (June 2026):

| Metric | PostgreSQL | ScyllaDB | Ratio |
|--------|------------|----------|-------|
| Write TPS | 1,920 | 9,751 | 5.1x |
| Write Latency | 2.08ms | 0.4ms | 5.2x |
| Read TPS | 14,877 | 11,976 | 0.8x |
| Write Amplification | 1.6x | 9.2x | 5.8x |

## 📖 References

- [DDIA Chapter 3: Storage & Retrieval](https://dataintensive.net/)
- [PostgreSQL Docs](https://postgresql.org/docs/)
- [ScyllaDB Docs](https://docs.scylladb.com/)
- [MongoDB Docs](https://docs.mongodb.com/)
- [Redis Docs](https://redis.io/docs/)
- [ClickHouse Docs](https://clickhouse.com/docs/)
- [Elasticsearch Docs](https://elastic.co/guide/)
- [Neo4j Docs](https://neo4j.com/docs/)
- [CockroachDB Docs](https://cockroachlabs.com/docs/)
- [TimescaleDB Docs](https://docs.timescale.com/)
- [InfluxDB Docs](https://docs.influxdata.com/)
- [DragonflyDB Docs](https://dragonflydb.io/docs)
- [Valkey Docs](https://valkey.io/docs/)
- [Milvus Docs](https://milvus.io/docs)
- [SQLite Docs](https://sqlite.org/docs.html)

## License

MIT
