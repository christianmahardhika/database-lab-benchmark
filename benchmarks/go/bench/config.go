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

	// Connection strings
	PostgresDSN      string
	MySQLDSN         string
	MongoURI         string
	RedisAddr        string
	ValkeyAddr       string
	DragonflyAddr    string
	ScyllaHosts      string
	ScyllaPort       int
	CassandraPort    int
	ClickHouseAddr   string
	TimescaleDSN     string
	CockroachDSN     string
	ElasticsearchURL string
	Neo4jURI         string
	Neo4jUser        string
	Neo4jPassword    string
	InfluxDBURL      string
	InfluxDBToken    string
	InfluxDBOrg      string
	InfluxDBBucket   string
	MilvusAddr       string
	SQLitePath       string
	QdrantURL        string
	PgvectorDSN      string
	PrometheusURL    string
}

// LoadConfig creates config with defaults and overrides from env.
func LoadConfig() *Config {
	cfg := &Config{
		NumRows:          100_000,
		Concurrency:      10,
		ValueSize:        100,
		WarmupRows:       10_000,
		Runs:             3,
		OutputDir:        "../../results",
		DBs:              "all",
		PostgresDSN:      "postgres://bench:bench@localhost:5432/benchmark?sslmode=disable",
		MySQLDSN:         "bench:bench@tcp(localhost:3306)/benchmark",
		MongoURI:         "mongodb://localhost:27017",
		RedisAddr:        "localhost:6379",
		ValkeyAddr:       "localhost:6381",
		DragonflyAddr:    "localhost:6380",
		ScyllaHosts:      "localhost",
		ScyllaPort:       9043,
		CassandraPort:    9042,
		ClickHouseAddr:   "localhost:9000",
		TimescaleDSN:     "postgres://bench:bench@localhost:5433/benchmark?sslmode=disable",
		CockroachDSN:     "postgres://root@localhost:26257/defaultdb?sslmode=disable",
		ElasticsearchURL: "http://localhost:9200",
		Neo4jURI:         "bolt://localhost:7687",
		Neo4jUser:        "neo4j",
		Neo4jPassword:    "benchpass123",
		InfluxDBURL:      "http://localhost:8086",
		InfluxDBToken:    "bench-token-secret",
		InfluxDBOrg:      "benchmark",
		InfluxDBBucket:   "benchmark",
		MilvusAddr:       "localhost:19530",
		SQLitePath:       "/tmp/bench_sqlite.db",
		QdrantURL:        "http://localhost:6333",
		PgvectorDSN:      "postgres://bench:bench@localhost:5432/benchmark?sslmode=disable",
		PrometheusURL:    "http://localhost:9090",
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
	envStr("BENCH_ELASTICSEARCH_URL", &c.ElasticsearchURL)
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
		"cockroachdb", "elasticsearch", "neo4j", "influxdb", "milvus", "sqlite",
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
