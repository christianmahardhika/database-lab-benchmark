# OpenSearch Official References

> OpenSearch is an AWS-maintained fork of Elasticsearch 7.10, API-compatible but diverged after the license change.

Source: https://opensearch.org/docs/latest/

## Key Claims (Hypotheses)

### H44: Near Real-Time Search
- **Claim**: OpenSearch provides near real-time search capabilities with <1s index-to-search latency
- **Source**: https://opensearch.org/docs/latest/opensearch/index-data/
- **Test**: Exp 15 - Index documents and measure time until searchable

### H45: Horizontal Scaling
- **Claim**: Linear scaling with additional nodes for both indexing and search
- **Source**: https://opensearch.org/docs/latest/tuning-your-cluster/
- **Test**: Exp 5 - Measure throughput at 1, 2, 3 nodes

### H46: Full-Text Search Performance
- **Claim**: Sub-100ms search latency for complex queries on large datasets
- **Source**: https://opensearch.org/docs/latest/search-plugins/
- **Test**: Exp 15 - Full-text search benchmarks

### H47: Inverted Index Efficiency
- **Claim**: Inverted index provides O(1) term lookup
- **Source**: https://opensearch.org/docs/latest/im-plugin/index/
- **Test**: Exp 15 - Compare with PostgreSQL tsvector

### H48: Aggregation Performance
- **Claim**: Fast aggregations on high-cardinality fields
- **Source**: https://opensearch.org/docs/latest/aggregations/
- **Test**: Exp 15 - Aggregation benchmarks

## Configuration

### Recommended Production Settings
```yaml
# opensearch.yml
cluster.name: production
node.name: node-1
network.host: 0.0.0.0
discovery.seed_hosts: ["node-1", "node-2", "node-3"]
cluster.initial_cluster_manager_nodes: ["node-1", "node-2", "node-3"]

# JVM heap (50% of RAM, max 32GB)
# jvm.options: -Xms16g -Xmx16g
```

### Index Settings for Benchmarks
```json
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "refresh_interval": "30s",
    "index.translog.durability": "async"
  }
}
```

## Official Resources
- Documentation: https://opensearch.org/docs/latest/
- Performance Tuning: https://opensearch.org/docs/latest/tuning-your-cluster/
- Sizing Guide: https://opensearch.org/docs/latest/install-and-configure/
- Benchmarks: https://opensearch.org/benchmarks/
- GitHub: https://github.com/opensearch-project/OpenSearch

## Differences from Elasticsearch
- Forked from Elasticsearch 7.10.2
- Apache 2.0 license (vs Elastic License)
- Security plugin included by default (can be disabled)
- Compatible with Elasticsearch 7.x clients
- Diverged features: ML, Observability plugins
