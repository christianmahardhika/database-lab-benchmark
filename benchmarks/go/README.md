# Database Lab Benchmark (Go)

High-performance benchmark suite for 16 databases, written in Go for minimal client-side overhead.

## Why Go over Python?

| Aspect | Python | Go |
|--------|--------|-----|
| Concurrency | GIL limits true parallelism | Goroutines — true concurrent I/O |
| Driver overhead | High (psycopg2 ~10x slower than pgbench) | Minimal (pgx is near-native) |
| Memory | ~50MB per benchmark | ~10MB |
| Startup | 1-2s (import time) | Instant |
| Measurement accuracy | time.time() ~ms resolution | time.Now() ~ns resolution |

## Databases (16)

### Server-based (Docker)
| Database | Driver | Port | Engine |
|----------|--------|------|--------|
| PostgreSQL 16 | pgx | 5432 | B-Tree |
| MySQL 8.0 | go-sql-driver | 3306 | B+ Tree (InnoDB) |
| MongoDB 7.0 | mongo-driver | 27017 | B-Tree (WiredTiger) |
| Redis 7 | go-redis | 6379 | In-Memory Hash |
| Valkey 7.2 | go-redis | 6381 | In-Memory Hash (Redis fork) |
| DragonflyDB | go-redis | 6380 | Multi-thread Memory |
| ScyllaDB 5.4 | gocql | 9043 | LSM-Tree (C++) |
| Cassandra 4.1 | gocql | 9042 | LSM-Tree (Java) |
| ClickHouse | clickhouse-go | 9000 | MergeTree (Columnar) |
| TimescaleDB | pgx | 5433 | B-Tree + Hypertable |
| CockroachDB | pgx | 26257 | LSM (Pebble) + Raft |
| OpenSearch 2.15 | opensearch-go | 9200 | Inverted Index |
| Neo4j 5 | neo4j-go-driver | 7687 | Native Graph |
| InfluxDB 2.7 | influxdb-client-go | 8086 | TSM (Time-Series) |
| Milvus 2.4 | milvus-sdk-go | 19530 | HNSW (Vector) |

### Embedded (in-process)
| Database | Driver | Engine |
|----------|--------|--------|
| SQLite | modernc.org/sqlite | B-Tree (single-file) |

## Quick Start

```bash
# Start databases you want to test
cd ../../docker/postgres && docker compose up -d
cd ../../docker/redis && docker compose up -d

# Run benchmark
cd benchmarks/go
BENCH_DBS=postgres,redis go run . write

# Run all workloads
go run . all
```

## Architecture

```
bench/
├── config.go           # Configuration from env vars
├── registry.go         # Driver factory registry
├── runner.go           # Concurrent benchmark engine (goroutine pool)
├── result.go           # Percentile calculation, JSON output
├── driver_postgres.go  # PostgreSQL via pgx
├── driver_mysql.go     # MySQL via go-sql-driver
├── driver_mongodb.go   # MongoDB via official driver
├── driver_redis.go     # Redis/Valkey/Dragonfly via go-redis
├── driver_cql.go       # ScyllaDB/Cassandra via gocql
├── driver_clickhouse.go
├── driver_timescale.go
├── driver_cockroach.go
├── driver_opensearch.go
├── driver_neo4j.go
├── driver_influxdb.go
├── driver_milvus.go
└── driver_sqlite.go
```

## Adding a New Database

1. Create `bench/driver_<name>.go` implementing `DBDriver` interface:
   ```go
   type DBDriver interface {
       Name() string
       Setup(cfg *Config) error
       Cleanup() error
       Close() error
       Write(key string, value []byte) error
       Read(key string) ([]byte, error)
   }
   ```

2. Register in `bench/registry.go`:
   ```go
   "newdb": func() DBDriver { return &NewDBDriver{} },
   ```

3. Add config fields in `bench/config.go`

4. Add Docker compose in `docker/newdb/docker-compose.yml`
