# Prometheus TSDB Official Documentation Reference

Source: https://prometheus.io/docs/prometheus/latest/storage/

## Overview

Prometheus is an open-source monitoring and alerting toolkit with a built-in time-series database (TSDB). Pull-based model, PromQL query language, designed for reliability over absolute durability.

## Storage Engine: Custom TSDB (Block-based, Compressed)

### Architecture
- **Head Block**: In-memory, mutable, 2-hour window (WAL-backed)
- **Persistent Blocks**: Immutable, compressed, 2h chunks
- **Compaction**: Merge small blocks into larger ones (max 31 days)
- **Index**: Inverted index on label pairs → series
- **Chunks**: XOR-encoded (Gorilla compression for timestamps and values)

### Data Model
```
metric_name{label1="value1", label2="value2"} value timestamp
```

### Write Path (Scrape Model)
```
Scrape Target → Head Block (WAL) → 2h → Cut new block → Compact → Persist
```

### Write Path (Remote Write)
```
POST /api/v1/write → Head Block (WAL) → Same as above
```

### Read Path
```
PromQL → Identify series (label index) → Scan relevant blocks/chunks → Decompress → Return
```

## Performance Characteristics (Official Claims)

### Ingestion Rate
- **Claim**: "Millions of samples per second on single node"
- **Benchmark**: 2.16M samples/sec at ~8 CPUs, 7GB RAM (tested)
- **Why**: Append-only, Gorilla compression, efficient WAL

### Query Performance
- **Claim**: "Sub-second PromQL queries on millions of time series"
- **Typical**: Instant queries <100ms, range queries depend on time window
- **Why**: Inverted index, skip irrelevant blocks by time range

### Compression
- **Claim**: "1.3-2 bytes per sample on average"
- **Why**: Gorilla float encoding (XOR), delta-of-delta for timestamps
- **Comparison**: Raw float64 + int64 = 16 bytes → 1.5 bytes = ~10x compression

### Cardinality
- **Practical limit**: ~10M active time series per instance
- **Beyond**: Need Thanos/Mimir/Cortex for horizontal scaling

## Best Use Cases (from prometheus.io)

1. **Infrastructure Monitoring**
   - Kubernetes metrics
   - Node/container resource usage
   - Service health checks

2. **Application Metrics**
   - Request rate/latency/errors (RED)
   - Custom business metrics
   - SLI/SLO tracking

3. **Alerting**
   - PromQL-based alerting rules
   - AlertManager integration
   - Multi-signal alerting

4. **Short-term Operational Data**
   - Default 15-day retention
   - Fast queries for dashboards
   - Recent data analysis

## Anti-Patterns (from docs)

1. **Long-term storage (>30 days)**
   - Local storage is ephemeral
   - Use Thanos/Mimir for long-term

2. **Event logging**
   - Not for logs or traces
   - Use Loki or Elasticsearch

3. **High-cardinality labels**
   - label values shouldn't be unbounded (user IDs, request IDs)
   - Creates series explosion

4. **100% durability requirement**
   - Designed for availability over durability
   - WAL may lose last 2h on crash without graceful shutdown

5. **Push-based ingest at scale**
   - Designed as pull-based
   - remote_write works but adds complexity

## Configuration for Benchmarks

### Storage Tuning
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

storage:
  tsdb:
    retention.time: 24h
    min-block-duration: 2h
    max-block-duration: 2h
    wal-compression: true
```

### CLI Flags
```bash
prometheus \
  --storage.tsdb.path=/prometheus \
  --storage.tsdb.retention.time=24h \
  --web.enable-remote-write-receiver \
  --storage.tsdb.min-block-duration=2h \
  --storage.tsdb.max-block-duration=2h
```

### Remote Write API (for benchmarking)
```bash
# Push metrics via remote_write (protobuf + snappy)
# Use avalanche or promremotebench for load generation
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| prometheus_tsdb_head_samples_appended_total | Ingestion rate | >1M samples/sec |
| prometheus_tsdb_head_active_appenders | Write concurrency | Stable |
| prometheus_tsdb_compaction_duration_seconds | Compaction cost | <30s |
| prometheus_tsdb_blocks_loaded | Block count | Auto-managed |
| go_memstats_heap_inuse_bytes | Memory usage | <80% of available |
| prometheus_engine_query_duration_seconds | Query latency | <1s for dashboards |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H83 | 2M+ samples/sec ingestion (remote_write) | Exp 7: Push metrics via remote_write |
| H84 | ~1.5 bytes per sample compression | Exp: Measure storage vs raw data size |
| H85 | Sub-second PromQL queries | Exp: Range query on 1M series |
| H86 | Pull-based vs Push-based trade-off | Exp: Compare remote_write throughput vs InfluxDB push |

## Comparison: Prometheus vs Other TSDBs

| Aspect | Prometheus | InfluxDB | TimescaleDB |
|--------|-----------|----------|-------------|
| Data model | Metric + labels | Measurement + tags + fields | SQL table + time |
| Query language | PromQL | Flux / SQL (3.x) | Full SQL |
| Ingest model | Pull (+ remote_write) | Push (HTTP) | Push (SQL INSERT) |
| Compression | ~1.5 bytes/sample | ~3-10x | ~10-20x |
| Clustering | No (use Thanos/Mimir) | Enterprise only | PG replication |
| Retention | Typically 15-30 days | Unlimited | Unlimited |
| JOINs | No | No | Full SQL |
| Best for | Monitoring/alerting | Generic time-series | SQL-native analytics |

## Benchmark Tools

```bash
# Avalanche: High-cardinality metric generator
# https://github.com/prometheus-community/avalanche
avalanche --remote-url=http://localhost:9090/api/v1/write \
  --metric-count=1000 \
  --series-count=1000 \
  --value-interval=15

# promremotebench: Remote write load tester
# https://github.com/m3db/prometheus_remote_client_golang
```

## References

- Storage: https://prometheus.io/docs/prometheus/latest/storage/
- Remote Write: https://prometheus.io/docs/specs/remote_write_spec/
- TSDB Design: https://ganeshvernekar.com/blog/prometheus-tsdb-the-head-block/
- Benchmarks: https://www.percona.com/blog/prometheus-2-times-series-storage-performance-analyses/
