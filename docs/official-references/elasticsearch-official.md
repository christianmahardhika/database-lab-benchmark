# Elasticsearch Official Documentation Reference

Source: https://www.elastic.co/guide/en/elasticsearch/reference/current/

## Overview

Elasticsearch is a distributed search and analytics engine built on Apache Lucene. Designed for full-text search, log analytics, and real-time data exploration. RESTful API, schema-free JSON documents.

## Storage Engine: Inverted Index (Lucene)

### Architecture
- **Index**: Collection of documents (like a database)
- **Shard**: Lucene index partition (horizontal scaling unit)
- **Segment**: Immutable Lucene file (LSM-like append model)
- **Translog**: Write-ahead log for durability

### Data Structures
| Structure | Purpose |
|-----------|---------|
| Inverted Index | Full-text search (term → document mapping) |
| BKD Tree | Numeric/geo range queries |
| Doc Values | Columnar storage for sorting/aggregations |
| Stored Fields | Original document retrieval |

### Write Path
```
Index Request → Translog → In-memory buffer → Refresh (1s) → New Segment → Merge
```

### Read Path
```
Search → Coordinate → Scatter to shards → Per-shard Lucene search → Gather & rank → Return
```

### Segment Lifecycle
```
Buffer → Refresh (searchable, not durable) → Flush (durable) → Merge (compact)
```

## Performance Characteristics (Official Claims)

### Search Performance
- **Claim**: "Near real-time search" (documents searchable within 1 second of indexing)
- **Why**: Refresh interval creates new searchable segments every 1s
- **Benchmark**: Sub-second queries on billions of documents (official blog)

### Indexing Performance
- **Claim**: "Hundreds of thousands of documents per second per node"
- **Typical**: 10K-50K docs/sec single node (depends on mapping complexity)
- **Optimized**: 100K+ docs/sec with bulk API, disabled refresh, replicas=0

### Aggregation Performance
- **Claim**: "Real-time analytics on large datasets"
- **Why**: Doc values (columnar), caching, shard-level parallelism

### Scalability
- **Claim**: "Scales horizontally to hundreds of nodes"
- **How**: Shard distribution, replica shards for read scaling

## Best Use Cases (from elastic.co)

1. **Full-Text Search**
   - E-commerce product search
   - Document search
   - Autocomplete/suggestions

2. **Log and Event Analytics**
   - ELK/Elastic Stack (Elasticsearch + Logstash + Kibana)
   - Application performance monitoring
   - Security analytics (SIEM)

3. **Real-time Analytics**
   - Dashboards and visualizations
   - Metrics aggregation
   - Business intelligence

4. **Geospatial Search**
   - Location-based services
   - Geo-fencing
   - Distance calculations

## Anti-Patterns (from docs)

1. **Primary data store for OLTP**
   - Not ACID compliant
   - No transactions
   - Near-real-time (not real-time)

2. **Frequent updates to same document**
   - Immutable segments → delete + re-index
   - High update rate = high merge overhead

3. **Exact key-value lookups at scale**
   - Overhead vs Redis/DynamoDB for simple KV
   - Inverted index unnecessary for exact match

4. **Small datasets (<1GB)**
   - JVM overhead not justified
   - SQLite/PostgreSQL simpler

## Configuration for Benchmarks

### JVM Heap
```yaml
# ES_JAVA_OPTS
-Xms1g -Xmx1g  # 50% of available RAM, max 30GB
```

### Index Settings (Optimized for Write Throughput)
```json
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "refresh_interval": "30s",
    "translog.durability": "async",
    "translog.flush_threshold_size": "1gb"
  }
}
```

### Index Settings (Optimized for Search)
```json
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1,
    "refresh_interval": "1s"
  }
}
```

### Bulk Indexing Best Practices
```bash
# Optimal bulk size: 5-15 MB per request
# Typical: 1000-5000 documents per bulk
POST _bulk
{"index":{"_index":"bench_kv","_id":"key_00000001"}}
{"v":"base64_value_here"}
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| indexing.index_total | Total docs indexed | Baseline |
| search.query_time_in_millis | Search latency | <100ms p99 |
| refresh.total_time_in_millis | Refresh cost | Low |
| merges.total_time_in_millis | Merge overhead | Manageable |
| jvm.mem.heap_used_percent | Memory pressure | <75% |
| thread_pool.write.rejected | Write pressure | 0 |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H44 | Near real-time search (<1s) | Exp: Index doc, search after refresh |
| H45 | 100K+ docs/sec indexing (bulk) | Exp 1: Write throughput with bulk API |
| H46 | Sub-second aggregations on large data | Exp: Aggregation latency on 1M+ docs |
| H47 | Full-text search faster than LIKE queries | Exp: Compare ES search vs PG LIKE |
| H48 | Horizontal scaling with shards | Exp: Compare 1 vs 3 shards throughput |

## Comparison: Elasticsearch vs Alternatives

| Aspect | Elasticsearch | PostgreSQL FTS | ClickHouse |
|--------|---------------|----------------|------------|
| Full-text search | Native, optimized | tsvector (good) | Limited |
| Analytics | Aggregations | SQL window functions | Excellent |
| Write speed | Good (bulk) | Good | Batch only |
| Point lookups | Moderate | Excellent | Poor |
| JOINs | None | Full SQL | Limited |
| Storage efficiency | Moderate | Good | Excellent |
| Operational complexity | High | Low | Medium |

## References

- Documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/
- Performance Tuning: https://www.elastic.co/guide/en/elasticsearch/reference/current/tune-for-indexing-speed.html
- Sizing Guide: https://www.elastic.co/guide/en/elasticsearch/reference/current/size-your-shards.html
- Benchmarks: https://elasticsearch-benchmarks.elastic.co/
