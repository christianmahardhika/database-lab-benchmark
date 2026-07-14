package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/christianmahardhika/database-lab-benchmark/bench"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "write":
		runWriteThroughput()
	case "read":
		runReadLatency()
	case "mixed":
		runMixedWorkload()
	case "all":
		runWriteThroughput()
		runReadLatency()
		runMixedWorkload()
	case "list":
		listDatabases()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func runWriteThroughput() {
	cfg := bench.LoadConfig()
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("  Database Lab Benchmark — Write Throughput (Go)")
	fmt.Printf("  Rows: %d | Concurrency: %d | Value: %d bytes | Runs: %d\n",
		cfg.NumRows, cfg.Concurrency, cfg.ValueSize, cfg.Runs)
	fmt.Println("══════════════════════════════════════════════════════════════")

	dbs := cfg.TargetDBs()
	var results []bench.Result

	for _, db := range dbs {
		fmt.Printf("\n  ▶ %s ...\n", db)
		r, err := bench.RunWrite(db, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "    ✗ ERROR: %v\n", err)
			continue
		}
		results = append(results, r)
		fmt.Printf("    ✓ %s\n", r.Summary())
	}

	fmt.Println("\n── Write Throughput Results ─────────────────────────────────")
	bench.PrintTable(results)
	bench.SaveJSON(results, cfg.OutputDir, "write_throughput")
}

func runReadLatency() {
	cfg := bench.LoadConfig()
	fmt.Println("\n══════════════════════════════════════════════════════════════")
	fmt.Println("  Database Lab Benchmark — Read Latency (Go)")
	fmt.Printf("  Rows: %d | Concurrency: %d | Runs: %d\n",
		cfg.NumRows, cfg.Concurrency, cfg.Runs)
	fmt.Println("══════════════════════════════════════════════════════════════")

	dbs := cfg.TargetDBs()
	var results []bench.Result

	for _, db := range dbs {
		fmt.Printf("\n  ▶ %s ...\n", db)
		r, err := bench.RunRead(db, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "    ✗ ERROR: %v\n", err)
			continue
		}
		results = append(results, r)
		fmt.Printf("    ✓ %s\n", r.Summary())
	}

	fmt.Println("\n── Read Latency Results ────────────────────────────────────")
	bench.PrintTable(results)
	bench.SaveJSON(results, cfg.OutputDir, "read_latency")
}

func runMixedWorkload() {
	cfg := bench.LoadConfig()
	fmt.Println("\n══════════════════════════════════════════════════════════════")
	fmt.Println("  Database Lab Benchmark — Mixed Workload 80/20 (Go)")
	fmt.Printf("  Ops: %d | Concurrency: %d | Runs: %d\n",
		cfg.NumRows, cfg.Concurrency, cfg.Runs)
	fmt.Println("══════════════════════════════════════════════════════════════")

	dbs := cfg.TargetDBs()
	var results []bench.Result

	for _, db := range dbs {
		fmt.Printf("\n  ▶ %s ...\n", db)
		r, err := bench.RunMixed(db, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "    ✗ ERROR: %v\n", err)
			continue
		}
		results = append(results, r)
		fmt.Printf("    ✓ %s\n", r.Summary())
	}

	fmt.Println("\n── Mixed Workload Results ──────────────────────────────────")
	bench.PrintTable(results)
	bench.SaveJSON(results, cfg.OutputDir, "mixed_workload")
}

func listDatabases() {
	fmt.Println("Available databases:")
	fmt.Println()
	fmt.Println("  Server-based (Docker):")
	fmt.Println("    postgres      PostgreSQL 16         B-Tree (MVCC)        :5432")
	fmt.Println("    mysql         MySQL 8.0             B+ Tree (InnoDB)     :3306")
	fmt.Println("    mongodb       MongoDB 7.0           B-Tree (WiredTiger)  :27017")
	fmt.Println("    redis         Redis 7               In-Memory Hash       :6379")
	fmt.Println("    valkey        Valkey 7.2            In-Memory Hash       :6381")
	fmt.Println("    dragonfly     DragonflyDB           Multi-thread Memory  :6380")
	fmt.Println("    scylladb      ScyllaDB 5.4          LSM-Tree (C++)       :9043")
	fmt.Println("    cassandra     Cassandra 4.1         LSM-Tree (Java)      :9042")
	fmt.Println("    clickhouse    ClickHouse            MergeTree (Columnar) :9000")
	fmt.Println("    timescaledb   TimescaleDB           B-Tree + Hypertable  :5433")
	fmt.Println("    cockroachdb   CockroachDB           LSM (Pebble) + Raft  :26257")
	fmt.Println("    opensearch    OpenSearch 2.15       Inverted Index       :9200")
	fmt.Println("    neo4j         Neo4j 5               Native Graph         :7687")
	fmt.Println("    influxdb      InfluxDB 2.7          TSM (Time-Series)    :8086")
	fmt.Println("    milvus        Milvus 2.4            HNSW (Vector)        :19530")
	fmt.Println("    qdrant        Qdrant 1.9            HNSW (Rust Vector)   :6333")
	fmt.Println("    pgvector      pgvector (PG ext)     HNSW/IVFFlat         :5432")
	fmt.Println("    prometheus    Prometheus 2.53       TSDB (Monitoring)    :9090")
	fmt.Println()
	fmt.Println("  Embedded (in-process):")
	fmt.Println("    sqlite        SQLite                B-Tree (single-file)")
	fmt.Println()
	fmt.Println("Set BENCH_DBS to comma-separated list, e.g.:")
	fmt.Println("  BENCH_DBS=postgres,mysql,redis go run . write")
}

func printUsage() {
	usage := `
Database Lab Benchmark (Go) — 16 Databases

Usage:
  go run . <command>

Commands:
  write     Write throughput benchmark
  read      Read latency benchmark (point lookups)
  mixed     Mixed workload (80%% read / 20%% write)
  all       Run all benchmarks
  list      List available databases
  help      Show this help

Environment Variables:
  BENCH_ROWS          Number of rows/ops (default: 100000)
  BENCH_CONCURRENCY   Goroutines (default: 10)
  BENCH_VALUE_SIZE    Value size in bytes (default: 100)
  BENCH_DBS           Comma-separated DBs (default: all)
  BENCH_OUTPUT        Output directory (default: ../../results)
  BENCH_WARMUP        Warmup rows (default: 10000)
  BENCH_RUNS          Runs for median (default: 3)

Connection Overrides:
  BENCH_POSTGRES_DSN      (default: postgres://bench:bench@localhost:5432/benchmark)
  BENCH_MYSQL_DSN         (default: bench:bench@tcp(localhost:3306)/benchmark)
  BENCH_MONGO_URI         (default: mongodb://localhost:27017)
  BENCH_REDIS_ADDR        (default: localhost:6379)
  BENCH_VALKEY_ADDR       (default: localhost:6381)
  BENCH_DRAGONFLY_ADDR    (default: localhost:6380)
  BENCH_SCYLLA_HOSTS      (default: localhost)
  BENCH_SCYLLA_PORT       (default: 9043)
  BENCH_CASSANDRA_PORT    (default: 9042)
  BENCH_CLICKHOUSE_ADDR   (default: localhost:9000)
  BENCH_TIMESCALE_DSN     (default: postgres://bench:bench@localhost:5433/benchmark)
  BENCH_COCKROACH_DSN     (default: postgres://root@localhost:26257/defaultdb?sslmode=disable)
  BENCH_ELASTICSEARCH_URL (default: http://localhost:9200)
  BENCH_NEO4J_URI         (default: bolt://localhost:7687)
  BENCH_INFLUXDB_URL      (default: http://localhost:8086)
  BENCH_MILVUS_ADDR       (default: localhost:19530)

Examples:
  go run . write
  BENCH_ROWS=1000000 BENCH_DBS=postgres,scylladb go run . all
  BENCH_CONCURRENCY=50 BENCH_DBS=redis,dragonfly,valkey go run . write
`
	fmt.Println(strings.TrimSpace(usage))
}
