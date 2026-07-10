package bench

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

// CQLDriver handles ScyllaDB and Cassandra (both use CQL protocol).
type CQLDriver struct {
	variant string // "scylladb" or "cassandra"
	session *gocql.Session
	cfg     *Config
}

func (d *CQLDriver) Name() string { return d.variant }

func (d *CQLDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	port := cfg.ScyllaPort
	if d.variant == "cassandra" {
		port = cfg.CassandraPort
	}

	cluster := gocql.NewCluster(cfg.ScyllaHosts)
	cluster.Port = port
	cluster.Consistency = gocql.One
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second
	cluster.NumConns = cfg.Concurrency

	// Connect without keyspace first to create it
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	err = session.Query(`
		CREATE KEYSPACE IF NOT EXISTS benchmark
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
	`).Exec()
	if err != nil {
		session.Close()
		return fmt.Errorf("create keyspace: %w", err)
	}
	session.Close()

	// Reconnect with keyspace
	cluster.Keyspace = "benchmark"
	session, err = cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("connect to keyspace: %w", err)
	}

	err = session.Query(`
		CREATE TABLE IF NOT EXISTS bench_kv (
			k TEXT PRIMARY KEY,
			v BLOB
		)
	`).Exec()
	if err != nil {
		session.Close()
		return fmt.Errorf("create table: %w", err)
	}

	d.session = session
	return nil
}

func (d *CQLDriver) Cleanup() error {
	return d.session.Query("TRUNCATE bench_kv").Exec()
}

func (d *CQLDriver) Close() error {
	if d.session != nil {
		d.session.Close()
	}
	return nil
}

func (d *CQLDriver) Write(key string, value []byte) error {
	return d.session.Query("INSERT INTO bench_kv (k, v) VALUES (?, ?)", key, value).Exec()
}

func (d *CQLDriver) Read(key string) ([]byte, error) {
	var v []byte
	err := d.session.Query("SELECT v FROM bench_kv WHERE k = ?", key).Scan(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
