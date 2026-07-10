package bench

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDriver struct {
	db  *sql.DB
	cfg *Config
}

func (d *MySQLDriver) Name() string { return "mysql" }

func (d *MySQLDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	db.SetMaxOpenConns(cfg.Concurrency + 10)
	db.SetMaxIdleConns(cfg.Concurrency)
	d.db = db

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bench_kv (
			k VARCHAR(64) PRIMARY KEY,
			v BLOB NOT NULL
		) ENGINE=InnoDB
	`)
	return err
}

func (d *MySQLDriver) Cleanup() error {
	_, err := d.db.Exec("TRUNCATE TABLE bench_kv")
	return err
}

func (d *MySQLDriver) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *MySQLDriver) Write(key string, value []byte) error {
	_, err := d.db.Exec(
		"INSERT INTO bench_kv (k, v) VALUES (?, ?) ON DUPLICATE KEY UPDATE v = VALUES(v)",
		key, value)
	return err
}

func (d *MySQLDriver) Read(key string) ([]byte, error) {
	var v []byte
	err := d.db.QueryRow("SELECT v FROM bench_kv WHERE k = ?", key).Scan(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
