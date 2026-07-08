# TimescaleDB Official Documentation Reference

Source: https://docs.timescale.com/

## Overview

TimescaleDB is a PostgreSQL extension for time-series data. Full SQL compatibility with automatic time-based partitioning.

## Storage Engine: Hypertables (B-Tree based)

### Architecture
- **Hypertable**: Virtual table spanning many chunks
- **Chunks**: PostgreSQL tables partitioned by time
- **Compression**: Column-oriented compression for old data

### Auto-Partitioning
```
Hypertable → Chunk (7 days default) → PostgreSQL heap + indexes
```

### Compression (Native)
```sql
-- Enable compression on chunks older than 7 days
ALTER TABLE metrics SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'device_id'
);
SELECT add_compression_policy('metrics', INTERVAL '7 days');
```

## Performance Characteristics (Official Claims)

### Query Performance
- **Claim**: "10-100x faster than PostgreSQL for time-series"
- **Why**: Chunk exclusion, specialized indexes, compression
- **Benchmark**: 20x faster for time-range queries (official)

### Write Performance
- **Claim**: "Millions of inserts per second"
- **Benchmark**: 1M rows/sec on single node (official)
- **Why**: Parallel chunk inserts, no global index updates

### Compression
- **Claim**: "Up to 90%+ compression"
- **Typical**: 10-20x for time-series data

## Best Use Cases (from docs.timescale.com)

1. **IoT and Sensor Data**
   - Device telemetry
   - Industrial monitoring
   - Smart infrastructure

2. **DevOps Monitoring**
   - Metrics storage
   - Log analytics
   - APM data

3. **Financial Data**
   - Tick data
   - Trading analytics
   - Risk calculations

4. **When you need SQL**
   - Complex analytics
   - JOINs with relational data
   - Existing PostgreSQL ecosystem

## Anti-Patterns (from docs)

1. **Non-time-series data**
   - Regular tables better for non-temporal
   - Overhead not worth it

2. **Heavy updates**
   - Time-series is append-mostly
   - Updates on compressed chunks expensive

3. **Real-time aggregation without continuous aggregates**
   - Use continuous aggregates for dashboards
   - Raw queries on large data slow

## Configuration for Benchmarks

### PostgreSQL Settings
```sql
-- /etc/postgresql/postgresql.conf
shared_buffers = 4GB
effective_cache_size = 12GB
work_mem = 64MB
maintenance_work_mem = 1GB
timescaledb.max_background_workers = 8
```

### Hypertable Creation
```sql
-- Create table
CREATE TABLE metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL,
    value DOUBLE PRECISION
);

-- Convert to hypertable
SELECT create_hypertable('metrics', 'time');

-- Add compression
ALTER TABLE metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id',
    timescaledb.compress_orderby = 'time DESC'
);

-- Auto-compress after 7 days
SELECT add_compression_policy('metrics', INTERVAL '7 days');

-- Retention policy
SELECT add_retention_policy('metrics', INTERVAL '90 days');
```

### Continuous Aggregates
```sql
CREATE MATERIALIZED VIEW metrics_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', time) AS bucket,
    device_id,
    AVG(value) as avg_value,
    MAX(value) as max_value
FROM metrics
GROUP BY bucket, device_id;

-- Refresh policy
SELECT add_continuous_aggregate_policy('metrics_hourly',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| Chunk count | Partition health | Auto-managed |
| Compression ratio | Storage efficiency | >10x |
| Continuous aggregate lag | Freshness | <1 hour |
| Insert rate | Write throughput | >100K rows/s |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H40 | Time-series on PostgreSQL | Exp 11: Compare with plain PG |
| H41 | Automatic partitioning | Exp 11: Measure chunk creation |
| H42 | Continuous aggregates | Exp 11: Dashboard queries |
| H43 | 90%+ compression | Exp 11: Measure compression ratio |

## Comparison: TimescaleDB vs ClickHouse

| Aspect | TimescaleDB | ClickHouse |
|--------|-------------|------------|
| SQL | Full PostgreSQL | ClickHouse SQL |
| JOINs | Native, optimized | Expensive |
| Updates | Supported | Not supported |
| Compression | Good (10-20x) | Excellent (20-40x) |
| Write speed | Good (100K-1M/s) | Good (batch only) |
| Query speed | Good | Excellent for OLAP |
| Ecosystem | PostgreSQL tools | ClickHouse tools |

## References

- Getting Started: https://docs.timescale.com/getting-started/
- Hypertables: https://docs.timescale.com/use-timescale/hypertables/
- Compression: https://docs.timescale.com/use-timescale/compression/
- Continuous Aggregates: https://docs.timescale.com/use-timescale/continuous-aggregates/
- Benchmarks: https://www.timescale.com/blog/timescaledb-vs-influxdb-for-time-series-data-timescale-influx-sql-702d4c329d9e/
