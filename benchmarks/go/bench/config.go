package bench

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all benchmark parameters.
type Config struct {
	NumRows     int
	Concurrency int
	ValueSize   int
	WarmupRows  int
	Runs        int
	OutputDir   string
	DBs         string

	// Connection strings - ports match docker-compose.yml for simultaneous running
	PostgresDSN   string
	MySQLDSN      string
	MongoURI      string
	RedisAddr     string
	ValkeyAddr    string
	DragonflyAddr string
	ScyllaHosts   string
	ScyllaPort    int
	CassandraPort int
	ClickHouseAddr string
	TimescaleDSN   string
	CockroachDSN   string
	OpenSearchURL  string
	Neo4jURI       string
	Neo4jUser      string
	Neo4jPassword  string
	InfluxDBURL    string
	InfluxDBToken  string
	InfluxDBOrg    string
	InfluxDBBucket string
	MilvusAddr     string
	SQLitePath     string
	QdrantURL      string
	PgvectorDSN    string
	PrometheusURL  string
}

// LoadConfig creates config with defaults and overrides from env.
// Ports match docker-compose.yml configurations for simultaneous running.
func LoadConfig() *Config {
	cfg := &Config{
		NumRows:        100_000,
		Concurrency:   10,
		ValueSize:     100,
		WarmupRows:    10_000,
		Runs:          3,
		OutputDir:     "../../results",
		DBs:           "all",
		// PostgreSQL: 5432 (standard)
		PostgresDSN:   "postgres://bench:bench@localhost:5432/benchmark?sslmode=disable",
		// MySQL: 3306 (standard)
		MySQLDSN:      "bench:bench@tcp(localhost:3306)/benchmark",
		// MongoDB: 27017 (standard), with auth
		MongoURI:      "mongodb://bench:bench@localhost:27017",
		// Redis: 6379 (standard)
		RedisAddr:     "localhost:6379",
		// Valkey: 6381 (offset to avoid Redis conflict)
		ValkeyAddr:    "localhost:6381",
		// DragonflyDB: 6380 (offset to avoid Redis conflict)
		DragonflyAddr: "localhost:6380",
		// ScyllaDB: 9043 (offset to avoid Cassandra conflict)
		ScyllaHosts:   "localhost",
		ScyllaPort:    9043,
		// Cassandra: 9042 (standard)
		CassandraPort: 9042,
		// ClickHouse: 9000 (native protocol)
		ClickHouseAddr: "localhost:9000",
		// TimescaleDB: 5433 (offset to avoid Postgres conflict)
		TimescaleDSN:  "postgres://bench:bench@localhost:5433/benchmark?sslmode=disable",
		// CockroachDB: 26257 (standard)
		CockroachDSN:  "postgres://root@localhost:26257/defaultdb?sslmode=disable",
		// OpenSearch: 9200 (standard)
		OpenSearchURL: "http://localhost:9200",
		// Neo4j: 7687 (Bolt protocol)
		Neo4jURI:      "bolt://localhost:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: "benchpass123",
		// InfluxDB: 8086 (standard)
		InfluxDBURL:    "http://localhost:8086",
		InfluxDBToken:  "bench-token-secret",
		InfluxDBOrg:    "benchmark",
		InfluxDBBucket: "benchmark",
		// Milvus: 19530 (gRPC)
		MilvusAddr:    "localhost:19530",
		// SQLite: embedded (no port)
		SQLitePath:    "/tmp/bench_sqlite.db",
		// Qdrant: 6333 (REST API)
		QdrantURL:     "http://localhost:6333",
		// pgvector: 5434 (offset to avoid Postgres/Timescale conflict)
		PgvectorDSN:   "postgres://bench:bench@localhost:5434/benchmark?sslmode=disable",
		// Prometheus: 9090 (standard)
		PrometheusURL: "http://localhost:9090",
	}
	cfg.loadFromEnv()
	return cfg
}

func (c *Config) loadFromEnv() {
	envInt := func(key string, target *int) {
		if v := os.Getenv(key); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				*target = n
			}
		}
	}
	envStr := func(key string, target *string) {
		if v := os.Getenv(key); v != "" {
			*target = v
		}
	}

	envInt("BENCH_ROWS", &c.NumRows)
	envInt("BENCH_CONCURRENCY", &c.Concurrency)
	envInt("BENCH_VALUE_SIZE", &c.ValueSize)
	envInt("BENCH_WARMUP", &c.WarmupRows)
	envInt("BENCH_RUNS", &c.Runs)
	envStr("BENCH_OUTPUT", &c.OutputDir)
	envStr("BENCH_DBS", &c.DBs)
	envStr("BENCH_POSTGRES_DSN", &c.PostgresDSN)
	envStr("BENCH_MYSQL_DSN", &c.MySQLDSN)
	envStr("BENCH_MONGO_URI", &c.MongoURI)
	envStr("BENCH_REDIS_ADDR", &c.RedisAddr)
	envStr("BENCH_VALKEY_ADDR", &c.ValkeyAddr)
	envStr("BENCH_DRAGONFLY_ADDR", &c.DragonflyAddr)
	envStr("BENCH_SCYLLA_HOSTS", &c.ScyllaHosts)
	envInt("BENCH_SCYLLA_PORT", &c.ScyllaPort)
	envInt("BENCH_CASSANDRA_PORT", &c.CassandraPort)
	envStr("BENCH_CLICKHOUSE_ADDR", &c.ClickHouseAddr)
	envStr("BENCH_TIMESCALE_DSN", &c.TimescaleDSN)
	envStr("BENCH_COCKROACH_DSN", &c.CockroachDSN)
	envStr("BENCH_OPENSEARCH_URL", &c.OpenSearchURL)
	envStr("BENCH_NEO4J_URI", &c.Neo4jURI)
	envStr("BENCH_NEO4J_USER", &c.Neo4jUser)
	envStr("BENCH_NEO4J_PASSWORD", &c.Neo4jPassword)
	envStr("BENCH_INFLUXDB_URL", &c.InfluxDBURL)
	envStr("BENCH_INFLUXDB_TOKEN", &c.InfluxDBToken)
	envStr("BENCH_INFLUXDB_ORG", &c.InfluxDBOrg)
	envStr("BENCH_INFLUXDB_BUCKET", &c.InfluxDBBucket)
	envStr("BENCH_MILVUS_ADDR", &c.MilvusAddr)
	envStr("BENCH_SQLITE_PATH", &c.SQLitePath)
	envStr("BENCH_QDRANT_URL", &c.QdrantURL)
	envStr("BENCH_PGVECTOR_DSN", &c.PgvectorDSN)
	envStr("BENCH_PROMETHEUS_URL", &c.PrometheusURL)
}

// TargetDBs returns the list of databases to benchmark.
func (c *Config) TargetDBs() []string {
	all := []string{
		"postgres", "mysql", "mongodb", "redis", "valkey", "dragonfly",
		"scylladb", "cassandra", "clickhouse", "timescaledb",
		"cockroachdb", "opensearch", "neo4j", "influxdb", "milvus", "sqlite",
		"qdrant", "pgvector", "prometheus",
	}
	if c.DBs == "all" || c.DBs == "" {
		return all
	}
	parts := strings.Split(c.DBs, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
