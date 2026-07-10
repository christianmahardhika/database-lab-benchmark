package bench

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CockroachDriver uses pgx — CockroachDB is PostgreSQL wire-compatible.
type CockroachDriver struct {
	pool *pgxpool.Pool
	cfg  *Config
}

func (d *CockroachDriver) Name() string { return "cockroachdb" }

func (d *CockroachDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	pool, err := pgxpool.New(context.Background(), cfg.CockroachDSN)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	d.pool = pool

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS bench_kv (
			k VARCHAR(64) PRIMARY KEY,
			v BYTES NOT NULL
		)
	`)
	return err
}

func (d *CockroachDriver) Cleanup() error {
	_, err := d.pool.Exec(context.Background(), "TRUNCATE TABLE bench_kv")
	return err
}

func (d *CockroachDriver) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}

func (d *CockroachDriver) Write(key string, value []byte) error {
	_, err := d.pool.Exec(context.Background(),
		"UPSERT INTO bench_kv (k, v) VALUES ($1, $2)",
		key, value)
	return err
}

func (d *CockroachDriver) Read(key string) ([]byte, error) {
	var v []byte
	err := d.pool.QueryRow(context.Background(),
		"SELECT v FROM bench_kv WHERE k = $1", key).Scan(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
