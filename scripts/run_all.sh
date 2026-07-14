#!/bin/bash
# Database Lab Benchmark — Experiment-Based Orchestrator
# Usage: ./scripts/run_all.sh [experiment_number|all|list]
# Each experiment starts only the relevant DBs, runs benchmarks, then stops them.
#
# Designed for spot instances: each experiment is a self-contained chunk (~5-15 min)
# If spot gets interrupted, just re-run the experiment that was running.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_DIR/docker"
BENCH_DIR="$PROJECT_DIR/benchmarks/go"
RESULTS_DIR="$PROJECT_DIR/results"

# Benchmark config (override via env)
# LIGHT mode: ~5 min per experiment (good for spot)
# FULL mode:  ~15-30 min per experiment (accurate results)
BENCH_MODE="${BENCH_MODE:-light}"

if [ "$BENCH_MODE" = "light" ]; then
  BENCH_ROWS="${BENCH_ROWS:-10000}"
  BENCH_CONCURRENCY="${BENCH_CONCURRENCY:-10}"
  BENCH_RUNS="${BENCH_RUNS:-1}"
elif [ "$BENCH_MODE" = "full" ]; then
  BENCH_ROWS="${BENCH_ROWS:-100000}"
  BENCH_CONCURRENCY="${BENCH_CONCURRENCY:-10}"
  BENCH_RUNS="${BENCH_RUNS:-3}"
fi

# Allow manual override regardless of mode
export BENCH_ROWS="${BENCH_ROWS}"
export BENCH_CONCURRENCY="${BENCH_CONCURRENCY}"
export BENCH_RUNS="${BENCH_RUNS}"
export PATH="$PATH:/usr/local/go/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[$(date +%H:%M:%S)]${NC} $1"; }
ok()  { echo -e "${GREEN}[✓]${NC} $1"; }
err() { echo -e "${RED}[✗]${NC} $1"; }
warn(){ echo -e "${YELLOW}[!]${NC} $1"; }

# Build Go benchmark binary once
build_bench() {
  if [ -f "$PROJECT_DIR/bench-runner" ]; then
    ok "Binary exists: bench-runner (skip build)"
    return 0
  fi
  log "Building Go benchmark binary..."
  cd "$BENCH_DIR"
  go build -o "$PROJECT_DIR/bench-runner" .
  ok "Binary built: bench-runner"
}

# Start databases for an experiment
start_dbs() {
  log "Starting databases..."
  for entry in "$@"; do
    local db="${entry%%:*}"
    local file="${entry##*:}"
    log "  ▶ Starting $db ($file)..."
    cd "$DOCKER_DIR/$db"
    docker compose -f "$file" up -d 2>/dev/null
  done
  log "Waiting for healthchecks..."
  sleep 5
  local timeout=120
  local elapsed=0
  while [ $elapsed -lt $timeout ]; do
    local all_healthy=true
    for entry in "$@"; do
      local db="${entry%%:*}"
      local file="${entry##*:}"
      cd "$DOCKER_DIR/$db"
      if docker compose -f "$file" ps 2>/dev/null | grep -q "unhealthy\|starting"; then
        all_healthy=false
        break
      fi
    done
    if [ "$all_healthy" = true ]; then
      ok "All databases healthy"
      return 0
    fi
    sleep 5
    elapsed=$((elapsed + 5))
  done
  warn "Timeout waiting for healthchecks (proceeding anyway)"
}

# Stop databases for an experiment
stop_dbs() {
  log "Stopping databases..."
  for entry in "$@"; do
    local db="${entry%%:*}"
    local file="${entry##*:}"
    cd "$DOCKER_DIR/$db"
    docker compose -f "$file" down -v 2>/dev/null &
  done
  wait
  ok "All databases stopped"
}

# Run Go benchmark
run_bench() {
  local workload="$1"
  local dbs="$2"
  log "Running: $workload | DBs: $dbs | Rows: $BENCH_ROWS | Concurrency: $BENCH_CONCURRENCY | Runs: $BENCH_RUNS"
  cd "$PROJECT_DIR"
  BENCH_DBS="$dbs" \
  BENCH_OUTPUT="$RESULTS_DIR" \
  ./bench-runner "$workload"
}

# Save checkpoint (so we know what's done if spot dies)
save_checkpoint() {
  local exp_name="$1"
  echo "$exp_name completed at $(date)" >> "$RESULTS_DIR/checkpoint.log"
  ok "Checkpoint saved: $exp_name"
}

# ═══════════════════════════════════════════════════════════════
# EXPERIMENTS — Each is a self-contained chunk (~5-15 min on light)
# ═══════════════════════════════════════════════════════════════

exp8_mysql_vs_pg() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 8: MySQL vs PostgreSQL (~5 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("postgres:docker-compose.yml" "mysql:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "postgres,mysql"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp8"
}

exp12_inmemory_battle() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 12: In-Memory Battle (~5 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("redis:docker-compose.yml" "valkey:docker-compose.yml" "dragonflydb:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "redis,valkey,dragonfly"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp12"
}

exp14_vector_search() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 14: Vector Search (~8 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("qdrant:docker-compose.yml" "pgvector:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "qdrant,pgvector"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp14"
}

exp14_milvus() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 14b: Milvus Vector Search (~10 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("milvus:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "milvus"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp14b"
}

exp7_timeseries() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 7: Time-Series Ingest (~8 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("timescaledb:docker-compose.yml" "influxdb:docker-compose.yml" "prometheus:docker-compose.yml" "clickhouse:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "timescaledb,influxdb,prometheus,clickhouse"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp7"
}

exp13_graph() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 13: Graph Traversal (~5 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("neo4j:docker-compose.yml" "postgres:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "neo4j,postgres"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp13"
}

exp15_fulltext() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 15: Full-Text Search (~5 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("opensearch:docker-compose.yml" "postgres:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "opensearch,postgres"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp15"
}

exp16_distributed_sql() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 16: Distributed SQL (~8 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("cockroachdb:docker-compose.yml" "postgres:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "cockroachdb,postgres"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp16"
}

exp6_lsm_battle() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 6: LSM Battle — ScyllaDB vs Cassandra (~10 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("scylladb:docker-compose.yml" "cassandra:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "scylladb,cassandra"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp6"
}

exp9_mongo_vs_sql() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  Experiment 9: MongoDB vs SQL (~8 min)"
  echo "══════════════════════════════════════════════════════════════"
  local dbs=("mongodb:docker-compose.yml" "postgres:docker-compose.yml" "mysql:docker-compose.yml")
  start_dbs "${dbs[@]}"
  run_bench "all" "mongodb,postgres,mysql"
  stop_dbs "${dbs[@]}"
  save_checkpoint "exp9"
}

exp_sqlite() {
  echo ""
  echo "══════════════════════════════════════════════════════════════"
  echo "  SQLite Benchmark (embedded, no Docker) (~2 min)"
  echo "══════════════════════════════════════════════════════════════"
  cd "$PROJECT_DIR"
  BENCH_DBS="sqlite" BENCH_OUTPUT="$RESULTS_DIR" ./bench-runner all
  save_checkpoint "sqlite"
}

# ═══════════════════════════════════════════════════════════════
# CLI
# ═══════════════════════════════════════════════════════════════

list_experiments() {
  echo ""
  echo "Database Lab Benchmark — Spot-Friendly Orchestrator"
  echo ""
  echo "Each experiment is a self-contained chunk (~5-15 min)."
  echo "If spot dies, just re-run the interrupted experiment."
  echo ""
  echo "Available experiments:"
  echo ""
  echo "  8     MySQL vs PostgreSQL                  (~5 min)"
  echo "  12    In-Memory: Redis vs Valkey vs Dragonfly  (~5 min)"
  echo "  14    Vector: Qdrant vs pgvector           (~8 min)"
  echo "  14b   Vector: Milvus (heavy, separate)     (~10 min)"
  echo "  7     Time-Series: Timescale, InfluxDB, Prom, CH  (~8 min)"
  echo "  13    Graph: Neo4j vs PostgreSQL           (~5 min)"
  echo "  15    Full-Text: Elasticsearch vs PG       (~5 min)"
  echo "  16    Distributed SQL: CockroachDB vs PG   (~8 min)"
  echo "  6     LSM: ScyllaDB vs Cassandra           (~10 min)"
  echo "  9     MongoDB vs SQL                       (~8 min)"
  echo "  sqlite  SQLite (no Docker needed)          (~2 min)"
  echo ""
  echo "  quick   Run 8 + 12 + 14                    (~18 min)"
  echo "  full    Run all experiments                 (~75 min)"
  echo ""
  echo "Modes (set BENCH_MODE env):"
  echo "  BENCH_MODE=light  10K rows, 1 run   (fast, ~5 min/exp)  [default]"
  echo "  BENCH_MODE=full   100K rows, 3 runs (accurate, ~15 min/exp)"
  echo ""
  echo "Examples:"
  echo "  ./scripts/run_all.sh 12                    # In-Memory battle"
  echo "  BENCH_MODE=full ./scripts/run_all.sh 8     # MySQL vs PG (accurate)"
  echo "  BENCH_ROWS=50000 ./scripts/run_all.sh 14   # Custom rows"
  echo ""
  echo "Checkpoints saved to: results/checkpoint.log"
  echo ""
}

main() {
  mkdir -p "$RESULTS_DIR"
  build_bench

  local cmd="${1:-list}"

  case "$cmd" in
    8)      exp8_mysql_vs_pg ;;
    12)     exp12_inmemory_battle ;;
    14)     exp14_vector_search ;;
    14b)    exp14_milvus ;;
    7)      exp7_timeseries ;;
    13)     exp13_graph ;;
    15)     exp15_fulltext ;;
    16)     exp16_distributed_sql ;;
    6)      exp6_lsm_battle ;;
    9)      exp9_mongo_vs_sql ;;
    sqlite) exp_sqlite ;;
    quick)
      exp8_mysql_vs_pg
      exp12_inmemory_battle
      exp14_vector_search
      ;;
    full)
      exp8_mysql_vs_pg
      exp12_inmemory_battle
      exp14_vector_search
      exp7_timeseries
      exp13_graph
      exp15_fulltext
      exp16_distributed_sql
      exp6_lsm_battle
      exp9_mongo_vs_sql
      exp14_milvus
      exp_sqlite
      ;;
    list|help|-h|--help)
      list_experiments
      ;;
    *)
      err "Unknown experiment: $cmd"
      list_experiments
      exit 1
      ;;
  esac

  echo ""
  ok "Done! Results saved to: $RESULTS_DIR/"
  ok "Checkpoints: $RESULTS_DIR/checkpoint.log"
}

main "$@"
