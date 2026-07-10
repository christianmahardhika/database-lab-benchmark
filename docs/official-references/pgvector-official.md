# pgvector Official Documentation Reference

Source: https://github.com/pgvector/pgvector

## Overview

pgvector is an open-source PostgreSQL extension for vector similarity search. Adds vector data types, HNSW and IVFFlat indexes, and distance operators to standard PostgreSQL. Leverages existing PG ecosystem (SQL, transactions, replication, extensions).

## Storage Engine: PostgreSQL B-Tree + HNSW/IVFFlat Index

### Architecture
- **Vector column**: Stored in regular PostgreSQL heap (TOAST for large vectors)
- **HNSW index**: Hierarchical Navigable Small World graph (in-memory during search)
- **IVFFlat index**: Inverted file with flat quantization (faster build, lower recall)
- **Full SQL**: Vector search combined with regular WHERE clauses

### Index Types
| Index | Build Time | Query Time | Recall | Memory |
|-------|-----------|------------|--------|--------|
| None (seq scan) | 0 | O(n) | 100% | None |
| IVFFlat | Fast | Medium | ~90-95% | Low |
| HNSW | Slower | Fast | ~95-99% | Higher |

### Write Path
```
INSERT (vector) → PostgreSQL heap → Background HNSW/IVF index update
```

### Search Path
```
SQL query → Index scan (HNSW/IVF) → Filter by WHERE → Return top-K
```

## Performance Characteristics (Official/Community Claims)

### Search Performance
- **HNSW**: ~8ms p95 on 500K 768-dim vectors, 2,300 QPS on 1M vectors
- **IVFFlat**: ~35ms p95 on 500K vectors (3-5x slower than HNSW)
- **150x speedup**: Year-over-year improvements (2023→2024)

### Throughput
- **Claim**: "1,800 QPS at 91% recall on 1M OpenAI embeddings" (Supabase benchmark)
- **At 98% recall**: 670 QPS (HNSW, 1M vectors)
- **Competitive**: At 50M vectors, some benchmarks show pgvector beating dedicated DBs

### Build Time
- **HNSW on 1M vectors**: ~10-15 minutes (768-dim)
- **IVFFlat on 1M vectors**: ~2-5 minutes (faster build)

### Key Advantage
- **No additional infrastructure**: Vector search inside your existing PostgreSQL
- **Full SQL**: JOINs, transactions, filtering all native
- **Replication**: Inherits PG streaming replication for free

## Best Use Cases (from pgvector community)

1. **Small-to-Medium Vector Datasets (<10M)**
   - When you already use PostgreSQL
   - Hybrid search (SQL + vector in one query)
   - No need for separate vector DB infrastructure

2. **RAG Applications**
   - Document embeddings + metadata filtering
   - Combine vector search with relational data
   - Full ACID for vector mutations

3. **Prototyping / MVP**
   - Single dependency (PostgreSQL)
   - Familiar SQL interface
   - Easy to migrate to dedicated vector DB later

4. **Multi-modal Search**
   - Combine text search (tsvector) + vector search
   - Filter by relational data + sort by similarity

## Anti-Patterns (from community)

1. **Billion-scale datasets**
   - Memory-constrained (HNSW in shared_buffers)
   - pgvector struggles >50M vectors without tuning
   - Use Milvus (DiskANN) or Qdrant for billion-scale

2. **Ultra-low latency (<1ms)**
   - PostgreSQL overhead (parsing, planning)
   - In-memory vector DBs faster for hot data

3. **Frequent index rebuilds**
   - HNSW build is expensive
   - High insert rate + low recall threshold = frequent reindexing

4. **No PostgreSQL in stack**
   - Adding PG just for vectors is overhead
   - Use standalone vector DB instead

## Configuration for Benchmarks

### Extension Setup
```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE bench_vectors (
    id SERIAL PRIMARY KEY,
    k VARCHAR(64) UNIQUE,
    embedding vector(128)
);

-- HNSW index (recommended)
CREATE INDEX ON bench_vectors
  USING hnsw (embedding vector_l2_ops)
  WITH (m = 16, ef_construction = 128);

-- Set search parameters
SET hnsw.ef_search = 64;
```

### IVFFlat Alternative
```sql
CREATE INDEX ON bench_vectors
  USING ivfflat (embedding vector_l2_ops)
  WITH (lists = 100);

SET ivfflat.probes = 10;
```

### Performance Tuning
```sql
-- Increase shared_buffers for index caching
ALTER SYSTEM SET shared_buffers = '4GB';
ALTER SYSTEM SET effective_cache_size = '12GB';
ALTER SYSTEM SET work_mem = '256MB';
ALTER SYSTEM SET maintenance_work_mem = '2GB';  -- For index builds
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| query_latency_ms | Search latency | <10ms p95 (HNSW, 500K) |
| queries_per_second | Throughput | >1,800 QPS (1M vectors) |
| recall | Accuracy vs sequential scan | >95% |
| index_build_time | Time to build HNSW | Minutes (depends on size) |
| shared_buffers_hit_ratio | Index in memory | >99% |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H87 | 2,300 QPS with HNSW on 1M vectors | Exp 14: Throughput benchmark |
| H88 | <10ms p95 for HNSW search | Exp 14: Latency comparison |
| H89 | Full SQL + vector in same query | Exp: Hybrid search (WHERE + ORDER BY distance) |
| H90 | Competitive with dedicated vector DBs | Exp 14: Head-to-head vs Qdrant/Milvus |
| H91 | 150x improvement claim (HNSW vs seq scan) | Exp: Compare with/without index |

## Comparison: pgvector vs Dedicated Vector DBs

| Aspect | pgvector | Qdrant | Milvus |
|--------|----------|--------|--------|
| Deployment | PG extension | Standalone | Distributed |
| Language | C (PG ext) | Rust | Go/C++ |
| Query | SQL | REST/gRPC | SDK/gRPC |
| Max practical scale | ~10-50M vectors | Billions | Billions |
| Filtering | SQL WHERE (native) | Payload filter (during search) | Scalar filter |
| Quantization | None (planned) | Scalar, Product | Scalar, IVF variants |
| Index types | HNSW, IVFFlat | HNSW | HNSW, IVF, DiskANN |
| Transactions | Full ACID | None | None |
| Replication | PG streaming | Raft sharding | Pulsar-based |

## References

- GitHub: https://github.com/pgvector/pgvector
- Performance: https://jkatz05.com/post/postgres/pgvector-performance-150x-speedup/
- Supabase Benchmark: https://supabase.com/blog/pgvector-performance
- HNSW Paper: https://arxiv.org/abs/1603.09320
