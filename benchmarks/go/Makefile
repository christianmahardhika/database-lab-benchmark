.PHONY: build run-write run-read run-all tidy list

BINARY=dbench

build:
	go build -o $(BINARY) .

tidy:
	go mod tidy

list:
	go run . list

# Quick benchmarks (100k rows, 10 goroutines)
run-write:
	go run . write

run-read:
	go run . read

run-mixed:
	go run . mixed

run-all:
	go run . all

# Heavy benchmarks (1M rows, 50 goroutines)
run-heavy:
	BENCH_ROWS=1000000 BENCH_CONCURRENCY=50 go run . all

# Compare Redis variants
run-redis-battle:
	BENCH_DBS=redis,valkey,dragonfly BENCH_ROWS=500000 go run . all

# Compare SQL databases
run-sql-battle:
	BENCH_DBS=postgres,mysql,cockroachdb,sqlite go run . all

# Compare time-series
run-timeseries-battle:
	BENCH_DBS=timescaledb,influxdb,clickhouse,scylladb go run . write

# Compare LSM engines
run-lsm-battle:
	BENCH_DBS=scylladb,cassandra,cockroachdb go run . all
