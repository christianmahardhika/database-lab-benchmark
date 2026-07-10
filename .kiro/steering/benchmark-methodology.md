---
inclusion: fileMatch
fileMatchPattern: "benchmarks/**"
---

# Benchmark Methodology Standards

## Measurement Principles

1. **Warmup before measurement**: Run 10-20% of target operations before recording
2. **Steady state**: Measure after caches are populated and compaction/vacuum stabilizes
3. **Multiple runs**: Take median of 3+ runs, report p50/p95/p99
4. **Controlled variables**: Same CPU/RAM limits, same data distribution, same network

## Benchmark Script Template

```python
#!/usr/bin/env python3
"""<Database> <Workload> Benchmark"""

import time
import json
import statistics
from datetime import datetime

# Configuration (pin at top for reproducibility)
CONFIG = {
    "num_rows": 100_000,
    "batch_size": 1000,
    "num_threads": 1,
    "warmup_rows": 10_000,
    "data_size_bytes": 100,  # per value
    "runs": 3,
}

RESULTS = {}

def setup():
    """Create schema and prepare environment."""
    pass

def warmup():
    """Pre-populate caches and trigger initial compaction."""
    pass

def benchmark_writes():
    """Measure write throughput and latency."""
    latencies = []
    start = time.perf_counter()
    
    for i in range(CONFIG["num_rows"]):
        t0 = time.perf_counter()
        # ... write operation ...
        latencies.append(time.perf_counter() - t0)
    
    elapsed = time.perf_counter() - start
    
    RESULTS["write"] = {
        "total_ops": CONFIG["num_rows"],
        "elapsed_s": round(elapsed, 3),
        "throughput_ops_s": round(CONFIG["num_rows"] / elapsed),
        "latency_ms_p50": round(statistics.median(latencies) * 1000, 3),
        "latency_ms_p95": round(sorted(latencies)[int(len(latencies) * 0.95)] * 1000, 3),
        "latency_ms_p99": round(sorted(latencies)[int(len(latencies) * 0.99)] * 1000, 3),
    }

def benchmark_reads():
    """Measure read throughput and latency."""
    pass  # Similar structure

def teardown():
    """Clean up resources."""
    pass

def save_results(output_path):
    """Save results as JSON."""
    results = {
        "metadata": {
            "database": "<db_name>",
            "version": "<db_version>",
            "workload": "<workload_type>",
            "timestamp": datetime.now().isoformat(),
            "config": CONFIG,
        },
        "results": RESULTS,
    }
    with open(output_path, "w") as f:
        json.dump(results, f, indent=2)
    print(f"Results saved to {output_path}")

if __name__ == "__main__":
    setup()
    warmup()
    benchmark_writes()
    benchmark_reads()
    teardown()
    save_results("results.json")
```

## Data Generation Rules

- Use deterministic seeds for reproducibility: `random.seed(42)`
- Value sizes should be realistic for the use case being simulated
- Key distribution: uniform random (default), zipfian (for cache tests), sequential (for range tests)
- Include variety in data types (strings, integers, timestamps, nested objects)

## What to Report

Every benchmark result must include:
- **Throughput**: Operations per second (median of N runs)
- **Latency**: p50, p95, p99 in milliseconds
- **Resource usage**: Peak CPU %, peak memory, disk I/O (if available)
- **Configuration**: Exact database settings, hardware limits, data characteristics
- **Reproducibility**: Exact commands to reproduce the result

## Common Pitfalls to Avoid

| Pitfall | Why It's Wrong | Fix |
|---------|----------------|-----|
| Measuring cold cache | First run includes cache misses | Warmup phase |
| Single run | Variance not captured | 3+ runs, report median |
| Client bottleneck | Slow driver masks DB perf | Use native tools or async driver |
| Unrealistic data | 1-byte values don't match production | Use realistic payload sizes |
| Ignoring compaction/vacuum | LSM perf degrades after compaction | Measure during steady state |
| Network overhead in latency | Includes round-trip, not DB time | Separate network from DB latency |

## Comparing Across Databases Fairly

- Same hardware limits (Docker CPU/RAM constraints)
- Same data volume and key distribution
- Same measurement window (steady state)
- Use native benchmark tools when available for "best case":
  - PostgreSQL: `pgbench`
  - MySQL: `sysbench`
  - ScyllaDB: `cassandra-stress`
  - Redis: `redis-benchmark`
  - ClickHouse: `clickhouse-benchmark`
- Use Python driver tests for "application perspective" comparison
