# SQLite

SQLite is an embedded database — no Docker container needed.
It runs in-process with the benchmark binary.

The Go benchmark uses `modernc.org/sqlite` (pure Go, no CGO) or `github.com/mattn/go-sqlite3` (CGO).

SQLite file is created at `/tmp/bench_sqlite.db` during benchmarks.
