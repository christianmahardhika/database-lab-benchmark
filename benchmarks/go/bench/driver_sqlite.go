package bench

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

type SQLiteDriver struct {
	db  *sql.DB
	cfg *Config
}

func (d *SQLiteDriver) Name() string { return "sqlite" }

func (d *SQLiteDriver) Setup(cfg *Config) error {
	d.cfg = cfg

	// Remove existing file for clean start
	os.Remove(cfg.SQLitePath)

	db, err := sql.Open("sqlite", cfg.SQLitePath)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	// SQLite performance tuning
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-64000", // 64MB cache
		"PRAGMA mmap_size=268435456", // 256MB mmap
		"PRAGMA temp_store=MEMORY",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("pragma: %w", err)
		}
	}

	// Set pool to 1 for SQLite (single writer)
	db.SetMaxOpenConns(1)
	d.db = db

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bench_kv (
			k TEXT PRIMARY KEY,
			v BLOB NOT NULL
		)
	`)
	return err
}

func (d *SQLiteDriver) Cleanup() error {
	_, err := d.db.Exec("DELETE FROM bench_kv")
	return err
}

func (d *SQLiteDriver) Close() error {
	if d.db != nil {
		err := d.db.Close()
		os.Remove(d.cfg.SQLitePath)
		return err
	}
	return nil
}

func (d *SQLiteDriver) Write(key string, value []byte) error {
	_, err := d.db.Exec(
		"INSERT OR REPLACE INTO bench_kv (k, v) VALUES (?, ?)",
		key, value)
	return err
}

func (d *SQLiteDriver) Read(key string) ([]byte, error) {
	var v []byte
	err := d.db.QueryRow("SELECT v FROM bench_kv WHERE k = ?", key).Scan(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
