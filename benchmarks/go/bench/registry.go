package bench

import "fmt"

// driverFactory maps database names to constructor functions.
var driverFactory = map[string]func() DBDriver{
	"postgres":      func() DBDriver { return &PostgresDriver{} },
	"mysql":         func() DBDriver { return &MySQLDriver{} },
	"mongodb":       func() DBDriver { return &MongoDriver{} },
	"redis":         func() DBDriver { return &RedisDriver{variant: "redis"} },
	"valkey":        func() DBDriver { return &RedisDriver{variant: "valkey"} },
	"dragonfly":     func() DBDriver { return &RedisDriver{variant: "dragonfly"} },
	"scylladb":      func() DBDriver { return &CQLDriver{variant: "scylladb"} },
	"cassandra":     func() DBDriver { return &CQLDriver{variant: "cassandra"} },
	"clickhouse":    func() DBDriver { return &ClickHouseDriver{} },
	"timescaledb":   func() DBDriver { return &TimescaleDriver{} },
	"cockroachdb":   func() DBDriver { return &CockroachDriver{} },
	"elasticsearch": func() DBDriver { return &ElasticsearchDriver{} },
	"neo4j":         func() DBDriver { return &Neo4jDriver{} },
	"influxdb":      func() DBDriver { return &InfluxDBDriver{} },
	"milvus":        func() DBDriver { return &MilvusDriver{} },
	"sqlite":        func() DBDriver { return &SQLiteDriver{} },
	"qdrant":        func() DBDriver { return &QdrantDriver{} },
	"pgvector":      func() DBDriver { return &PgvectorDriver{} },
	"prometheus":    func() DBDriver { return &PrometheusDriver{} },
}

// GetDriver creates a new driver instance for the given database name.
func GetDriver(name string) (DBDriver, error) {
	factory, ok := driverFactory[name]
	if !ok {
		return nil, fmt.Errorf("unknown database: %s", name)
	}
	return factory(), nil
}
