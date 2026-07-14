---
name: run-experiment
description: Guide through running a specific benchmark experiment from start to finish, including infrastructure setup, execution, result collection, and interpretation. Use when running database benchmarks or experiments.
---

## Workflow

### Step 1: Select Experiment
Ask the user which experiment to run (or infer from context):
- Exp 1: Write Throughput (all databases)
- Exp 2: Read Latency (all databases)
- Exp 3: Complex Query (PG, MySQL, Mongo)
- Exp 4: Schema Evolution (PG, MySQL, Mongo, Scylla)
- Exp 5: Horizontal Scaling (Scylla, Mongo, Citus)
- Exp 6: Write Amplification (PG, Scylla, RocksDB)
- Exp 7: Time-Series (TimescaleDB, Mongo, Scylla, ClickHouse)
- Exp 8: MySQL vs PostgreSQL
- Exp 9: MongoDB vs SQL
- Exp 10: OLAP (ClickHouse, PG)
- Exp 11: Time-Series Battle

### Step 2: Check Prerequisites
- Docker running: `docker info`
- Required images pulled: `docker images | grep <db>`
- Python dependencies available: test imports
- Sufficient disk space and RAM

### Step 3: Start Database Containers
```bash
cd docker/<db>
docker compose up -d
# Wait for healthcheck
docker compose ps
```

### Step 4: Run Benchmark Script
```bash
python3 benchmarks/<experiment>/benchmark_<db>.py \
  --rows <N> \
  --threads <T> \
  --output results/<experiment>_<db>.json
```

### Step 5: Collect Results
- Parse JSON output
- Display throughput, latency percentiles
- Compare with previous runs if available

### Step 6: Teardown
```bash
docker compose -f docker/<db>/docker-compose.yml down -v
```

### Step 7: Interpret
- Connect results to the hypothesis being tested
- Explain WHY the result makes sense given storage engine internals
- Note any anomalies or unexpected findings

## Parameters
- `experiment`: Which experiment number (1-11)
- `databases`: Which databases to include (default: all relevant)
- `rows`: Number of rows/operations (default: 100,000 for quick, 1,000,000 for full)
- `threads`: Concurrency levels to test (default: 1, 10)
- `mode`: "quick" (single run) or "full" (3 runs, all thread counts)

## Output
- JSON results in `results/` directory
- Markdown summary comparing databases
- Hypothesis verdict: confirmed / partially confirmed / refuted
