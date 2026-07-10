package bench

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDriver struct {
	pool *pgxpool.Pool
	cfg  *Config
}

func (d *PostgresDriver) Name() string { return "postgres" }

func (d *PostgresDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	pool, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	d.pool = pool

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS bench_kv (
			k VARCHAR(64) PRIMARY KEY,
			v BYTEA NOT NULL
		)
	`)
	return err
}

func (d *PostgresDriver) Cleanup() error {
	_, err := d.pool.Exec(context.Background(), "TRUNCATE TABLE bench_kv")
	return err
}

func (d *PostgresDriver) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}

func (d *PostgresDriver) Write(key string, value []byte) error {
	_, err := d.pool.Exec(context.Background(),
		"INSERT INTO bench_kv (k, v) VALUES ($1, $2) ON CONFLICT (k) DO UPDATE SET v = $2",
		key, value)
	return err
}

func (d *PostgresDriver) Read(key string) ([]byte, error) {
	var v []byte
	err := d.pool.QueryRow(context.Background(),
		"SELECT v FROM bench_kv WHERE k = $1", key).Scan(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
