---
inclusion: auto
---

# Database Lab Benchmark — Steering Guide

## Project Purpose

This is a hands-on database laboratory for validating vendor claims (hypotheses H1–H43) about database engines through reproducible benchmarks. The lab covers **16 databases** across seven storage engine families:

| Family | Databases | Storage Model |
|--------|-----------|---------------|
| B-Tree | PostgreSQL 16, MySQL 8.0, MongoDB 7 (WiredTiger), SQLite | Sorted pages, in-place updates |
| LSM-Tree | ScyllaDB 5.4, Cassandra 4.1, CockroachDB (Pebble) | Memtable → SSTable flush, compaction |
| Columnar | ClickHouse | Column-oriented, MergeTree family |
| Time-Series | TimescaleDB, InfluxDB 2.7 | Hypertables / TSM engine |
| In-Memory | Redis 7, Valkey 7.2, DragonflyDB | Hash table, optional persistence |
| Search | Elasticsearch 8 | Inverted index (Lucene) |
| Graph | Neo4j 5 | Native graph storage |
| Vector | Milvus 2.4 | HNSW / IVF indexes |

## Lab Architecture

```
database-lab-benchmark/
├── terraform/        # AWS EC2 (m6i.2xlarge, 8 vCPU / 32GB) infrastructure
├── docker/           # One docker-compose per database (standalone + topologies)
├── benchmarks/       # Python experiment scripts per workload pattern
├── scripts/          # Orchestration (run_all.sh)
├── results/          # Benchmark outputs (JSON + markdown summaries)
└── docs/             # Operational guides (replication, backup, pooling)
```

## Running Experiments

### Benchmark Tool (Go)
The benchmark suite is written in Go for minimal overhead and true concurrent I/O via goroutines.

```bash
cd benchmarks/go

# Run all databases, write benchmark
go run . write

# Specific databases only
BENCH_DBS=postgres,mysql,redis go run . write

# Full suite with higher concurrency
BENCH_ROWS=1000000 BENCH_CONCURRENCY=50 go run . all

# List all available databases
go run . list
```

### Starting databases
```bash
cd docker/<database>
docker compose up -d
# wait for healthcheck
docker compose ps
```

### EC2 (full suite, recommended)
```bash
cd terraform && terraform apply -var="key_name=<your-key>"
ssh ubuntu@<ip>
cd database-lab-benchmark/benchmarks/go && go run . all
```

## Conventions

- All benchmarks written in Go (`benchmarks/go/`) — single binary, all 16 databases
- Results saved as JSON to `results/` directory
- Docker containers use resource limits (2 CPU, 1-2GB RAM) for fair comparison
- Port mappings avoid conflicts:
  - PG=5432, TimescaleDB=5433, MySQL=3306, Mongo=27017
  - Redis=6379, Valkey=6381, DragonflyDB=6380
  - ScyllaDB=9043, Cassandra=9042
  - ClickHouse=8123/9000, Elasticsearch=9200
  - CockroachDB=26257, Neo4j=7687, InfluxDB=8086, Milvus=19530
- SQLite runs embedded (no Docker needed)

## Key Principles for Lab Work

1. **Fair comparison**: Same hardware limits, same data volume, same measurement methodology
2. **Reproduce vendor claims**: Each experiment maps to specific hypotheses (H1-H43)
3. **Explain the WHY**: Results must tie back to storage engine internals (DDIA Ch. 3)
4. **Production relevance**: Include operational aspects (replication, backup, pooling, monitoring)

## When Generating Benchmark Scripts

- Benchmarks are in Go (`benchmarks/go/`) — add new drivers by implementing the `DBDriver` interface
- Interface: `Setup`, `Cleanup`, `Close`, `Write(key, value)`, `Read(key)`
- Register new drivers in `bench/registry.go`
- Runner handles concurrency via goroutines, warmup, multiple runs, and percentile calculation
- Use `pgx` for anything PostgreSQL-wire compatible (PG, TimescaleDB, CockroachDB)
- Use `gocql` for anything CQL-compatible (ScyllaDB, Cassandra)
- Use `go-redis` for anything Redis-protocol compatible (Redis, Valkey, DragonflyDB)

## When Working With Docker Configs

- Keep resource limits consistent across databases for fair benchmarks
- Use healthchecks so dependent services wait properly
- Volume mount data dirs for persistence across restarts
- Use alpine/slim images where possible to reduce pull time
