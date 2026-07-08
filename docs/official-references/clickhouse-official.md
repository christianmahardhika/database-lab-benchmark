# ClickHouse Official Documentation Reference

Source: https://clickhouse.com/docs/

## Overview

ClickHouse is a column-oriented OLAP database for real-time analytics. Developed by Yandex for web analytics (Yandex.Metrica - 20+ billion events/day).

## Storage Engine: MergeTree Family

### Architecture
- **Column-oriented**: Each column stored separately
- **Sparse indexing**: Primary key index every 8192 rows (granule)
- **Data compression**: LZ4 by default, ZSTD optional

### MergeTree Variants
| Engine | Use Case |
|--------|----------|
| MergeTree | Default, general purpose |
| ReplacingMergeTree | Deduplication by key |
| SummingMergeTree | Pre-aggregated sums |
| AggregatingMergeTree | Pre-aggregated any function |
| CollapsingMergeTree | Incremental state changes |

### Write Path
```
INSERT → In-memory buffer → Part (immutable) → Background merge
```

### Read Path
```
Query → Prune partitions → Prune granules → Decompress columns → Process
```

## Performance Characteristics (Official Claims)

### Query Performance
- **Claim**: "Process billions of rows in seconds"
- **Benchmark**: 100M-1B rows/sec scan rate (official)
- **Why**: Columnar + vectorized execution + compression

### Compression
- **Claim**: "10-20x compression ratio"
- **Typical**: 5-10x for general data, up to 40x for time-series

### Write Performance
- **Claim**: "50-200 MB/s per server"
- **Caveat**: Batch inserts only, not single-row OLTP

## Best Use Cases (from clickhouse.com)

1. **Web/App Analytics**
   - Clickstream data
   - User behavior analysis
   - A/B testing

2. **Time-Series Analytics**
   - Metrics aggregation
   - Log analysis
   - IoT analytics

3. **Real-time Dashboards**
   - Sub-second queries on TB+ data
   - High concurrency reads

4. **Data Warehousing**
   - ETL destination
   - Historical analysis

## Anti-Patterns (from docs)

1. **OLTP workloads**
   - No single-row updates/deletes
   - No transactions
   - High latency for point queries

2. **Frequent small inserts**
   - Batch inserts required (1000+ rows)
   - Each insert creates a part

3. **Normalized data with JOINs**
   - Denormalize for best performance
   - JOINs are expensive

## Configuration for Benchmarks

### Server Settings
```xml
<!-- /etc/clickhouse-server/config.xml -->
<max_memory_usage>10000000000</max_memory_usage>
<max_threads>8</max_threads>
<max_insert_block_size>1048576</max_insert_block_size>
```

### Table Creation (Optimized)
```sql
CREATE TABLE events (
    timestamp DateTime,
    user_id UInt64,
    event String,
    properties String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (user_id, timestamp)
SETTINGS index_granularity = 8192;
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| Query duration | Time to execute | <1s for dashboards |
| Rows read/s | Scan throughput | >100M rows/s |
| Compressed size | Storage efficiency | <20% of raw |
| Parts count | Merge health | <300 per partition |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H36 | OLAP with billions of rows | Exp 10: 100M row aggregation |
| H37 | 100M+ rows/sec scan | Exp 10: Full table scan |
| H38 | Time-series analytics | Exp 11: Time-series queries |
| H39 | 10-20x compression | Exp 10: Measure storage ratio |

## References

- Introduction: https://clickhouse.com/docs/en/intro
- MergeTree: https://clickhouse.com/docs/en/engines/table-engines/mergetree-family/mergetree
- Performance: https://clickhouse.com/docs/en/operations/optimizing-performance
- Benchmarks: https://clickhouse.com/benchmark/dbms/
