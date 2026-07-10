# Milvus Official Documentation Reference

Source: https://milvus.io/docs

## Overview

Milvus is an open-source vector database built for AI/ML applications. Designed for similarity search on embedding vectors (dense and sparse). Supports billion-scale vector datasets with millisecond search latency.

## Storage Engine: HNSW / IVF / DiskANN (Vector Indexes)

### Architecture
- **Proxy**: Entry point, routes requests
- **Query Node**: Handles search/query (in-memory index)
- **Data Node**: Handles insert/delete
- **Index Node**: Builds vector indexes in background
- **Meta Store**: etcd (metadata), MinIO/S3 (object storage)

### Index Types
| Index | Type | Use Case | Memory |
|-------|------|----------|--------|
| HNSW | Graph-based | High recall, low latency | High (all in RAM) |
| IVF_FLAT | Inverted file | Balance speed/recall | Medium |
| IVF_SQ8 | Quantized IVF | Memory-efficient | Low |
| DiskANN | Disk-based | Billion-scale, limited RAM | Very low |
| FLAT | Brute force | Small datasets, 100% recall | High |

### Write Path
```
Insert → Data Node → Sealed Segment → Index Node builds vector index → Query Node loads
```

### Search Path
```
Search Request → Proxy → Query Nodes (ANN search on index) → Merge results → Return top-K
```

## Performance Characteristics (Official Claims)

### Search Performance
- **Claim**: "Millisecond-level search on billion-scale vectors"
- **Benchmark**: <10ms for top-100 on 1B vectors (DiskANN)
- **HNSW**: <1ms for top-10 on million-scale

### Recall
- **Claim**: "95%+ recall with HNSW at high throughput"
- **Tradeoff**: Higher recall = more computation = higher latency

### Scalability
- **Claim**: "Scales horizontally to billions of vectors"
- **How**: Sharding across query nodes, separated storage/compute
- **Limitation**: More nodes = more coordination overhead

### Ingestion
- **Claim**: "Supports real-time insert and immediate search"
- **Caveat**: Growing segments (unsealed) have lower search performance

## Best Use Cases (from milvus.io)

1. **Semantic Search**
   - Document similarity (RAG for LLMs)
   - Image search by visual similarity
   - Audio/video fingerprinting

2. **Recommendation Systems**
   - User/item embeddings
   - Content-based filtering
   - Similar product discovery

3. **Anomaly Detection**
   - Find outlier vectors
   - Fraud detection via embedding distance
   - Network intrusion detection

4. **Drug Discovery**
   - Molecular similarity search
   - Protein structure matching

## Anti-Patterns (from docs)

1. **Traditional CRUD/OLTP**
   - Not a general-purpose database
   - No SQL, no JOINs, no transactions

2. **Exact match queries**
   - Designed for approximate nearest neighbor (ANN)
   - Use PostgreSQL/Redis for exact lookups

3. **Small datasets (<10K vectors)**
   - Brute force FLAT index is fine
   - Milvus overhead not justified

4. **Frequently updated vectors**
   - Delete + re-insert (no in-place update)
   - Compaction needed for cleanup

## Configuration for Benchmarks

### Collection Creation
```python
from pymilvus import Collection, FieldSchema, CollectionSchema, DataType

fields = [
    FieldSchema("id", DataType.VARCHAR, is_primary=True, max_length=64),
    FieldSchema("vector", DataType.FLOAT_VECTOR, dim=128),
    FieldSchema("payload", DataType.VARCHAR, max_length=1024),
]
schema = CollectionSchema(fields)
collection = Collection("bench_kv", schema)

# Create HNSW index
index_params = {
    "metric_type": "L2",
    "index_type": "HNSW",
    "params": {"M": 16, "efConstruction": 256}
}
collection.create_index("vector", index_params)
```

### Search Parameters
```python
search_params = {
    "metric_type": "L2",
    "params": {"ef": 64}  # Higher ef = better recall, slower
}
results = collection.search(vectors, "vector", search_params, limit=10)
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| search_latency_ms | Query latency | <10ms p99 (HNSW) |
| search_throughput | Queries per second | >1000 QPS |
| recall | Search accuracy | >95% |
| insert_rate | Ingestion speed | >10K vectors/sec |
| memory_usage | Index memory | Fits in RAM |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H70 | Millisecond search on 100K vectors | Exp 2: Search latency with HNSW |
| H71 | 95%+ recall with HNSW | Exp: Compare ANN results vs brute-force |
| H72 | Real-time insert + search | Exp 1: Write throughput via Go SDK |
| H73 | Better than brute-force at scale | Exp: HNSW vs FLAT latency at 100K+ vectors |

## Comparison: Milvus vs Alternatives

| Aspect | Milvus | pgvector | Pinecone | Weaviate |
|--------|--------|----------|----------|----------|
| Scale | Billions | Millions | Billions | Millions |
| Index types | Many (HNSW, IVF, DiskANN) | HNSW, IVF | Proprietary | HNSW |
| Deployment | Self-hosted / Zilliz Cloud | PG extension | Managed only | Self-hosted |
| Filtering | Scalar + vector | SQL WHERE + vector | Metadata filters | GraphQL + vector |
| SQL support | No | Full PostgreSQL | No | GraphQL |
| Distributed | Native | PG replication | Native | Optional |

## References

- Documentation: https://milvus.io/docs
- Architecture: https://milvus.io/docs/architecture_overview.md
- Performance: https://milvus.io/docs/benchmark.md
- Index Types: https://milvus.io/docs/index.md
- GitHub: https://github.com/milvus-io/milvus
