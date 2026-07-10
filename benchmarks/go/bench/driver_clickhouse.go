package bench

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseDriver struct {
	conn clickhouse.Conn
	cfg  *Config
}

func (d *ClickHouseDriver) Name() string { return "clickhouse" }

func (d *ClickHouseDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.ClickHouseAddr},
		Auth: clickhouse.Auth{
			Database: "benchmark",
			Username: "bench",
			Password: "bench",
		},
		MaxOpenConns: cfg.Concurrency + 10,
	})
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	d.conn = conn

	err = conn.Exec(context.Background(), "CREATE DATABASE IF NOT EXISTS benchmark")
	if err != nil {
		return fmt.Errorf("create db: %w", err)
	}

	err = conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS benchmark.bench_kv (
			k String,
			v String
		) ENGINE = MergeTree()
		ORDER BY k
	`)
	return err
}

func (d *ClickHouseDriver) Cleanup() error {
	return d.conn.Exec(context.Background(), "TRUNCATE TABLE benchmark.bench_kv")
}

func (d *ClickHouseDriver) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func (d *ClickHouseDriver) Write(key string, value []byte) error {
	return d.conn.Exec(context.Background(),
		"INSERT INTO benchmark.bench_kv (k, v) VALUES (?, ?)",
		key, string(value))
}

func (d *ClickHouseDriver) Read(key string) ([]byte, error) {
	var v string
	row := d.conn.QueryRow(context.Background(),
		"SELECT v FROM benchmark.bench_kv WHERE k = ?", key)
	if err := row.Scan(&v); err != nil {
		return nil, err
	}
	return []byte(v), nil
}
