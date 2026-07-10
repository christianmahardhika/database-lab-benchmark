# InfluxDB Official Documentation Reference

Source: https://docs.influxdata.com/influxdb/v2/

## Overview

InfluxDB is a purpose-built time-series database. Optimized for high-ingest, real-time querying of timestamped data. Uses TSM (Time-Structured Merge Tree) storage engine.

## Storage Engine: TSM (Time-Structured Merge Tree)

### Architecture
- **TSM**: LSM-tree variant optimized for time-series
- **WAL**: Write-ahead log for durability
- **Cache**: In-memory write buffer
- **TSM Files**: Immutable, compressed, sorted by time
- **Compaction**: Background merge of TSM files

### Data Model
```
Measurement → Tags (indexed metadata) + Fields (values) + Timestamp
```

### Write Path
```
Write → WAL → In-Memory Cache → Snapshot → TSM File (compressed)
```

### Read Path
```
Query → Identify time range → Scan relevant TSM files → Decompress → Return
```

## Performance Characteristics (Official Claims)

### Write Performance
- **Claim**: "Millions of data points per second"
- **Benchmark**: 750K+ values/sec on single node (official)
- **Why**: Append-only, no random I/O, batch compression

### Query Performance
- **Claim**: "Sub-second queries on billions of points"
- **Why**: Time-based file organization, skip entire files by time range
- **Caveat**: Complex queries (GROUP BY many tags) can be slow

### Compression
- **Claim**: "Up to 10x compression for time-series data"
- **Typical**: 3-10x depending on data characteristics
- **Algorithms**: Gorilla (floats), Simple-8b (integers), Snappy (strings)

### Cardinality
- **Limitation**: High cardinality (many unique tag values) degrades performance
- **Guideline**: <10M unique series per database

## Best Use Cases (from docs.influxdata.com)

1. **Infrastructure Monitoring**
   - Server metrics (CPU, memory, disk)
   - Network telemetry
   - Container/K8s metrics

2. **IoT Sensor Data**
   - Device telemetry
   - Environmental monitoring
   - Industrial equipment

3. **Real-time Analytics**
   - Application metrics
   - Business KPIs
   - Alerting on thresholds

4. **DevOps Observability**
   - Telegraf + InfluxDB + Grafana stack
   - Log metrics
   - Trace metrics

## Anti-Patterns (from docs)

1. **High cardinality tags**
   - Unique IDs as tags = series explosion
   - Use fields for high-cardinality data

2. **Mutable data**
   - Not designed for updates
   - Overwrite = delete + insert

3. **Relational queries**
   - No JOINs
   - Limited cross-measurement queries

4. **Long-term raw data storage**
   - Downsampling recommended for old data
   - Retention policies for auto-delete

## Configuration for Benchmarks

### Storage Settings
```toml
# /etc/influxdb/config.toml
[storage]
  cache-max-memory-size = "1g"
  cache-snapshot-memory-size = "25m"
  compact-full-write-coldDuration = "4h"
```

### Bucket Creation
```bash
influx bucket create \
  --name benchmark \
  --retention 0 \
  --org benchmark
```

### Write Optimization
```bash
# Use line protocol with batching
# Optimal batch: 5000-10000 points per request
influx write --bucket benchmark --precision ns < data.lp
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| write_req_per_sec | Ingestion rate | Baseline |
| query_duration_ms | Query latency | <100ms for simple |
| cache_size_bytes | Memory pressure | <80% of configured |
| compactions_active | Background work | Low |
| series_cardinality | Index size | <10M |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H53 | 750K+ values/sec write | Exp 1: Write throughput (line protocol) |
| H54 | Sub-second time-range queries | Exp: Query last 1h from 1M+ points |
| H55 | Time-series compression 3-10x | Exp: Measure storage vs raw data size |
| H56 | Downsampling performance | Exp: Continuous query on 1M points |

## Flux vs SQL (InfluxDB 2.x uses Flux, 3.x returns to SQL)

```flux
// Flux query: Average CPU last hour, per host
from(bucket: "benchmark")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_percent")
  |> aggregateWindow(every: 5m, fn: mean)
  |> group(columns: ["host"])
```

## Comparison: InfluxDB vs Alternatives

| Aspect | InfluxDB | TimescaleDB | Prometheus |
|--------|----------|-------------|------------|
| Query language | Flux / SQL (3.x) | Full SQL | PromQL |
| Ingestion | Very fast | Fast | Pull-based |
| Compression | Good (3-10x) | Good (10-20x) | Excellent |
| JOINs | No | Full SQL | No |
| Cardinality limit | ~10M series | No hard limit | ~10M series |
| Clustering | Enterprise only | PG-based | Thanos/Cortex |

## References

- Documentation: https://docs.influxdata.com/influxdb/v2/
- Write Optimization: https://docs.influxdata.com/influxdb/v2/write-data/best-practices/
- Storage Engine: https://docs.influxdata.com/influxdb/v2/reference/internals/storage-engine/
- Performance: https://www.influxdata.com/blog/influxdb-performance-benchmarks/
