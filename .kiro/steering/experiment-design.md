---
inclusion: manual
---

# Experiment Design Guide

## Hypotheses Catalog (H1–H43)

### B-Tree Family (PostgreSQL, MySQL)
| ID | Hypothesis | Experiment |
|----|-----------|------------|
| H1 | PostgreSQL best for complex OLTP with JOINs | Exp 3: Complex Query |
| H2 | PostgreSQL MVCC enables high read concurrency | Exp 2: Read Latency |
| H3 | B-Tree provides predictable read latency (3-4 page reads) | Exp 2: Read Latency |
| H9 | Online DDL enables zero-downtime schema changes | Exp 4: Schema Evolution |
| H10 | PostgreSQL query planner adapts to data distribution | Exp 3: Complex Query |
| H21 | MySQL optimized for simple web OLTP patterns | Exp 8: MySQL vs PG |
| H22 | InnoDB clustered index faster for PK range scans | Exp 8: MySQL vs PG |
| H23 | MySQL replication simpler to operate | Exp 8: MySQL vs PG |
| H24 | MySQL Online DDL covers most ALTER operations | Exp 4: Schema Evolution |
| H25 | ProxySQL R/W splitting reduces read latency | Exp 8: MySQL vs PG |

### LSM-Tree Family (ScyllaDB, Cassandra, RocksDB)
| ID | Hypothesis | Experiment |
|----|-----------|------------|
| H5 | ScyllaDB handles >10K writes/sec per core | Exp 1: Write Throughput |
| H7 | ScyllaDB scales linearly with nodes added | Exp 5: Horizontal Scaling |
| H8 | Consistent hashing enables zero-downtime scaling | Exp 5: Horizontal Scaling |
| H17 | LSM write path (WAL→memtable→SSTable) faster than B-Tree in-place updates | Exp 1 |
| H18 | Write amplification is the cost of LSM write speed | Exp 6: Write Amplification |
| H19 | Compaction strategies trade read vs write performance | Exp 6: Write Amplification |

### Document Store (MongoDB)
| ID | Hypothesis | Experiment |
|----|-----------|------------|
| H26 | MongoDB flexible schema reduces development time | Exp 9: MongoDB vs SQL |
| H27 | WiredTiger compression reduces storage 50-70% | Exp 9: MongoDB vs SQL |
| H28 | Sharding provides linear read/write scaling | Exp 5: Horizontal Scaling |
| H29 | Aggregation pipeline matches SQL for analytics | Exp 9: MongoDB vs SQL |
| H30 | Change streams enable event-driven architecture | Exp 9: MongoDB vs SQL |

### Columnar (ClickHouse)
| ID | Hypothesis | Experiment |
|----|-----------|------------|
| H36 | ClickHouse scans 100M+ rows/sec | Exp 10: OLAP |
| H37 | Columnar storage achieves 10-20x compression | Exp 10: OLAP |
| H38 | Materialized views enable real-time aggregation | Exp 10: OLAP |
| H39 | ClickHouse outperforms PG 10-100x for OLAP | Exp 10: OLAP |

### Time-Series (TimescaleDB)
| ID | Hypothesis | Experiment |
|----|-----------|------------|
| H40 | TimescaleDB 10-100x faster than vanilla PG for time-series | Exp 7, 11 |
| H41 | Hypertable chunking enables efficient time-range queries | Exp 11 |
| H42 | Compression reduces time-series storage 90-95% | Exp 11 |
| H43 | Continuous aggregates outperform manual materialized views | Exp 11 |

---

## Experiment Templates

### Experiment Structure
Each experiment folder should contain:
```
benchmarks/<experiment-name>/
├── README.md              # Hypothesis, methodology, expected outcome
├── benchmark_<db>.py      # One script per database
├── load_<db>.py           # Data loading script (if separate)
├── analyze.py             # Cross-database comparison script
└── results/               # Local results (gitignored)
```

### Workload Patterns

| Pattern | Description | Databases to Test |
|---------|-------------|-------------------|
| Write-heavy (95/5) | High insert rate, occasional reads | All |
| Read-heavy (5/95) | Mostly point reads + range scans | All |
| Mixed OLTP (70/30) | Typical web app pattern | PG, MySQL, Mongo |
| Append-only | Time-series ingestion, no updates | ScyllaDB, TimescaleDB, ClickHouse |
| Scan-heavy | Full table scans, aggregations | ClickHouse, PG, TimescaleDB |
| Update-heavy | Frequent updates to existing rows | PG, MySQL, Mongo |
| Hot key | Zipfian distribution (20% keys get 80% traffic) | Redis, PG, Mongo |

### Data Models per Experiment

**E-commerce (Exp 3, 8, 9)**:
- Users (id, name, email, created_at)
- Products (id, name, price, category, inventory)
- Orders (id, user_id, product_id, quantity, total, status, created_at)
- Reviews (id, user_id, product_id, rating, text, created_at)

**IoT/Time-series (Exp 7, 11)**:
- Devices (id, name, type, location)
- Measurements (device_id, timestamp, temperature, humidity, pressure, battery)

**Activity Feed (Exp 1, 5)**:
- Events (id, user_id, event_type, payload, timestamp)

---

## Running an Experiment End-to-End

```bash
# 1. Start the target database
cd docker/<db> && docker compose up -d && cd ../..

# 2. Wait for healthy
docker compose -f docker/<db>/docker-compose.yml ps

# 3. Run the benchmark
python3 benchmarks/<experiment>/benchmark_<db>.py \
  --rows 100000 \
  --threads 10 \
  --output results/<experiment>_<db>.json

# 4. Stop and clean
docker compose -f docker/<db>/docker-compose.yml down -v

# 5. Compare results
python3 benchmarks/<experiment>/analyze.py results/<experiment>_*.json
```

## Interpreting Results

### What "Faster" Means
- **Throughput**: Higher ops/sec = better for that workload
- **Latency**: Lower p99 = more predictable for user-facing apps
- **Resource efficiency**: Same throughput with less CPU/RAM = better TCO

### Context Matters
- ScyllaDB 5x write throughput doesn't mean "use ScyllaDB for everything"
- PostgreSQL 1.2x read advantage is significant for read-heavy OLTP
- Redis sub-ms latency is irrelevant if your app adds 50ms of business logic
- ClickHouse 100x scan speed doesn't help for point lookups

### Connecting Results to Architecture Decisions
Always ask: "Given MY workload pattern, which engine characteristic dominates?"

| Your Dominant Pattern | Choose | Because |
|----------------------|--------|---------|
| Writes >> Reads | LSM (ScyllaDB) | Sequential I/O, no page splits |
| Reads >> Writes, complex queries | B-Tree (PostgreSQL) | JOINs, sorted traversal |
| Reads >> Writes, simple PK lookups | B-Tree (MySQL) | Clustered index PK scan |
| Append-only + aggregation | Columnar (ClickHouse) | Vectorized scan, compression |
| Sub-ms cache | In-memory (Redis) | No disk I/O |
| Time-series + SQL analytics | TimescaleDB | Best of both worlds |
