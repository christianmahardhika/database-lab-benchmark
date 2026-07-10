package bench

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PgvectorDriver benchmarks pgvector (PostgreSQL extension for vector search).
// Uses the same pgx driver as PostgreSQL.
type PgvectorDriver struct {
	pool *pgxpool.Pool
	cfg  *Config
	seq  int64
}

func (d *PgvectorDriver) Name() string { return "pgvector" }

func (d *PgvectorDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	pool, err := pgxpool.New(context.Background(), cfg.PgvectorDSN)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	d.pool = pool

	// Enable pgvector extension
	_, err = pool.Exec(context.Background(), "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("create extension vector: %w", err)
	}

	// Create table with vector column
	_, err = pool.Exec(context.Background(), fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS bench_vectors (
			id BIGSERIAL PRIMARY KEY,
			k VARCHAR(64) UNIQUE NOT NULL,
			embedding vector(%d) NOT NULL
		)
	`, benchVectorDim))
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	// Create HNSW index
	_, err = pool.Exec(context.Background(), `
		CREATE INDEX IF NOT EXISTS bench_vectors_hnsw_idx
		ON bench_vectors USING hnsw (embedding vector_cosine_ops)
		WITH (m = 16, ef_construction = 128)
	`)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	// Set search parameter
	_, _ = pool.Exec(context.Background(), "SET hnsw.ef_search = 64")

	return nil
}

func (d *PgvectorDriver) Cleanup() error {
	_, err := d.pool.Exec(context.Background(), "DROP TABLE IF EXISTS bench_vectors")
	return err
}

func (d *PgvectorDriver) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}

func (d *PgvectorDriver) Write(key string, value []byte) error {
	vec := deterministicVector(key, benchVectorDim)
	vecStr := vectorToString(vec)

	_, err := d.pool.Exec(context.Background(),
		`INSERT INTO bench_vectors (k, embedding) VALUES ($1, $2::vector)
		 ON CONFLICT (k) DO UPDATE SET embedding = $2::vector`,
		key, vecStr)
	return err
}

func (d *PgvectorDriver) Read(key string) ([]byte, error) {
	vec := deterministicVector(key, benchVectorDim)
	vecStr := vectorToString(vec)

	var k string
	err := d.pool.QueryRow(context.Background(),
		`SELECT k FROM bench_vectors ORDER BY embedding <=> $1::vector LIMIT 1`,
		vecStr).Scan(&k)
	if err != nil {
		return nil, err
	}
	return []byte(k), nil
}

// vectorToString converts a float32 slice to pgvector string format: "[0.1,0.2,...]"
func vectorToString(vec []float32) string {
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
