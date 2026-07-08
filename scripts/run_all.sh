#!/bin/bash
# Database Lab Benchmark - Run All Experiments
# Usage: ./run_all.sh [quick|full]

set -e

MODE=${1:-quick}
RESULTS_DIR="$(dirname $0)/../results/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$RESULTS_DIR"

echo "=========================================="
echo "Database Lab Benchmark"
echo "Mode: $MODE"
echo "Results: $RESULTS_DIR"
echo "=========================================="

# Configuration
if [ "$MODE" = "quick" ]; then
    ROWS=100000
    THREADS="1 10"
    DBS="postgres mysql mongodb redis"
else
    ROWS=1000000
    THREADS="1 10 50"
    DBS="postgres mysql mongodb redis scylladb cassandra clickhouse timescaledb"
fi

# Helper functions
start_db() {
    local db=$1
    echo "Starting $db..."
    docker-compose -f docker/$db/docker-compose.yml up -d
    sleep 10  # Wait for DB to be ready
}

stop_db() {
    local db=$1
    echo "Stopping $db..."
    docker-compose -f docker/$db/docker-compose.yml down -v
}

# Experiment 1: Write Throughput
run_exp1() {
    echo ""
    echo "=========================================="
    echo "Experiment 1: Write Throughput"
    echo "=========================================="
    
    for db in $DBS; do
        start_db $db
        
        for t in $THREADS; do
            echo "[$db] Testing with $t threads, $ROWS rows..."
            python3 benchmarks/write-throughput/benchmark_${db}.py \
                --rows $ROWS \
                --threads $t \
                --output "$RESULTS_DIR/exp1_${db}_t${t}.json" 2>&1 | tee -a "$RESULTS_DIR/exp1.log"
        done
        
        stop_db $db
    done
}

# Experiment 2: Read Latency
run_exp2() {
    echo ""
    echo "=========================================="
    echo "Experiment 2: Read Latency"
    echo "=========================================="
    
    for db in $DBS; do
        start_db $db
        
        echo "[$db] Loading data..."
        python3 benchmarks/read-latency/load_${db}.py --rows $ROWS
        
        echo "[$db] Running read benchmark..."
        python3 benchmarks/read-latency/benchmark_${db}.py \
            --output "$RESULTS_DIR/exp2_${db}.json" 2>&1 | tee -a "$RESULTS_DIR/exp2.log"
        
        stop_db $db
    done
}

# Experiment 8: MySQL vs PostgreSQL
run_exp8() {
    echo ""
    echo "=========================================="
    echo "Experiment 8: MySQL vs PostgreSQL"
    echo "=========================================="
    
    # Start both
    start_db postgres
    start_db mysql
    
    # Simple OLTP
    echo "[exp8a] Simple OLTP pattern..."
    python3 benchmarks/mysql-vs-postgres/oltp_simple.py \
        --rows $ROWS \
        --output "$RESULTS_DIR/exp8a.json" 2>&1 | tee -a "$RESULTS_DIR/exp8.log"
    
    # Complex queries
    echo "[exp8b] Complex queries..."
    python3 benchmarks/mysql-vs-postgres/complex_queries.py \
        --output "$RESULTS_DIR/exp8b.json" 2>&1 | tee -a "$RESULTS_DIR/exp8.log"
    
    stop_db postgres
    stop_db mysql
}

# Main
echo "Starting benchmarks at $(date)"
echo "System info:"
echo "  CPU: $(nproc) cores"
echo "  RAM: $(free -h | grep Mem | awk '{print $2}')"
echo "  Disk: $(df -h / | tail -1 | awk '{print $4}') free"
echo ""

case "$MODE" in
    quick)
        run_exp1
        run_exp8
        ;;
    full)
        run_exp1
        run_exp2
        run_exp8
        # Add more experiments here
        ;;
    exp1) run_exp1 ;;
    exp2) run_exp2 ;;
    exp8) run_exp8 ;;
    *)
        echo "Usage: $0 [quick|full|exp1|exp2|exp8]"
        exit 1
        ;;
esac

echo ""
echo "=========================================="
echo "Benchmark complete!"
echo "Results saved to: $RESULTS_DIR"
echo "=========================================="

# Generate summary
python3 scripts/generate_summary.py "$RESULTS_DIR" > "$RESULTS_DIR/SUMMARY.md"
echo "Summary: $RESULTS_DIR/SUMMARY.md"
