# Qdrant Official Documentation Reference

Source: https://qdrant.tech/documentation/

## Overview

Qdrant is a vector similarity search engine and database, written in Rust. Designed for extended filtering, high throughput, and low latency on large-scale vector datasets. REST and gRPC APIs.

## Storage Engine: HNSW + Custom Segment Architecture

### Architecture
- **Collections**: Groups of vectors with same dimensionality
- **Segments**: Independent data partitions within a collection
- **Points**: Vectors + payload (metadata for filtering)
- **HNSW Index**: Hierarchical Navigable Small World graph per segment

### Key Features
| Feature | Description |
|---------|-------------|
| HNSW index | High-recall approximate nearest neighbor |
| Payload filtering | Filter during search (not post-filter) |
| Quantization | Scalar/Product quantization for memory reduction |
| Distributed | Raft consensus + shard replication |
| Multi-vector | Multiple vectors per point (ColBERT, late interaction) |

### Write Path
```
Upsert → Write-ahead log → In-memory segment → Background optimization → Indexed segment
```

### Search Path
```
Query vector → HNSW graph traversal → Filter by payload → Re-rank top-K → Return
```

## Performance Characteristics (Official Claims)

### Search Performance
- **Claim**: "Highest RPS and lowest latencies in almost all scenarios"
- **Benchmark**: 12K QPS at p99 <5ms on 1M 768-dim vectors (vs pgvector 4.5K QPS)
- **Why**: Rust, SIMD optimizations, concurrent segment search

### Filtering During Search
- **Claim**: "Filters are applied during search, not post-filter"
- **Why**: Custom HNSW implementation with payload index integration
- **Benefit**: Maintains recall even with strict filters

### Memory Efficiency
- **Claim**: "Scalar quantization reduces memory by 4x with minimal recall loss"
- **Typical**: 768-dim float32 → 768 bytes (from 3072 bytes) with <1% recall drop

### Scalability
- **Claim**: "Horizontally scalable via sharding and replication"
- **How**: Collections split into shards, distributed across nodes via Raft

## Best Use Cases (from qdrant.tech)

1. **RAG (Retrieval Augmented Generation)**
   - Document chunk retrieval for LLMs
   - Semantic search with metadata filtering
   - Multi-tenant RAG applications

2. **Recommendation Systems**
   - User/item embedding similarity
   - Content-based filtering
   - Real-time recommendations

3. **Image/Video Search**
   - Visual similarity (CLIP embeddings)
   - Reverse image search
   - Video frame matching

4. **Anomaly Detection**
   - Outlier detection via distance thresholds
   - Fraud pattern recognition

## Anti-Patterns (from docs)

1. **Exact match / key-value lookups**
   - Use Redis/PostgreSQL for exact queries
   - Vector search is approximate by design

2. **Tiny datasets (<1K vectors)**
   - Brute force is fast enough
   - Index overhead not justified

3. **Frequently mutating vectors**
   - Segment optimization is background process
   - High update rate creates optimization pressure

4. **Full-text search**
   - Use Elasticsearch for text search
   - Qdrant is for semantic/vector similarity

## Configuration for Benchmarks

### Collection Creation (REST API)
```json
PUT /collections/bench_vectors
{
  "vectors": {
    "size": 128,
    "distance": "Cosine"
  },
  "optimizers_config": {
    "indexing_threshold": 20000
  },
  "hnsw_config": {
    "m": 16,
    "ef_construct": 128
  }
}
```

### Search Parameters
```json
POST /collections/bench_vectors/points/search
{
  "vector": [0.1, 0.2, ...],
  "limit": 10,
  "params": {
    "hnsw_ef": 64,
    "exact": false
  }
}
```

### Quantization (Memory Optimization)
```json
PUT /collections/bench_vectors
{
  "vectors": { "size": 128, "distance": "Cosine" },
  "quantization_config": {
    "scalar": { "type": "int8", "always_ram": true }
  }
}
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| search_latency_ms | Query latency | <5ms p99 (HNSW, 1M vectors) |
| search_throughput | Queries per second | >10K QPS |
| upload_throughput | Points per second | >50K points/sec |
| recall | Search accuracy vs brute-force | >95% |
| memory_usage | RAM for index + vectors | Monitor growth |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H78 | Highest RPS among vector DBs | Exp 14: Compare vs Milvus, pgvector |
| H79 | <5ms p99 on 1M vectors | Exp 14: Search latency at scale |
| H80 | Filtering during search (no recall loss) | Exp: Search with payload filter vs without |
| H81 | Scalar quantization <1% recall drop | Exp: Compare recall with/without quantization |
| H82 | Horizontal scaling with shards | Exp: Compare 1-node vs 3-node throughput |

## Comparison: Qdrant vs Alternatives

| Aspect | Qdrant | Milvus | pgvector |
|--------|--------|--------|----------|
| Language | Rust | Go/C++ | C (PG extension) |
| Index types | HNSW | HNSW, IVF, DiskANN | HNSW, IVFFlat |
| Filtering | During search | Post-filter | SQL WHERE + vector |
| Quantization | Scalar, Product | Scalar | None (native) |
| Clustering | Native (Raft) | Native (etcd) | PG replication |
| API | REST + gRPC | gRPC + SDK | SQL |
| Memory @ 1M 768d | ~3GB (float32) | ~4GB | ~3.5GB |
| QPS @ 1M | ~12K | ~8K | ~4.5K |

## References

- Documentation: https://qdrant.tech/documentation/
- Benchmarks: https://qdrant.tech/benchmarks/
- Distributed: https://qdrant.tech/documentation/guides/distributed_deployment/
- GitHub: https://github.com/qdrant/qdrant
