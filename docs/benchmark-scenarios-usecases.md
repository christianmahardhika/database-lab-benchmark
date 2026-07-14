# Benchmark Scenarios: DDIA & Database Internals → Real-World Use Cases

> **Sources:**
> - DDIA: Designing Data-Intensive Applications (Martin Kleppmann)
> - DBI: Database Internals (Alex Petrov)
> - KB: software-engineering-kb (Qdrant collection)

---

## 1. E-Commerce Platform

### Use Case: High-Traffic Product Catalog + Orders

**Business Context:** Tokopedia/Shopee-scale platform dengan:
- 10M+ products
- 100K+ concurrent users during flash sale
- Order throughput: 10K orders/minute peak

| ID | Scenario | Theory Source | Real Problem | DBs to Test |
|----|----------|---------------|--------------|-------------|
| EC-01 | Product Search (Read-Heavy) | DDIA Ch.3 B-Tree | "Kenapa product search lambat saat flash sale?" | PostgreSQL vs OpenSearch vs MongoDB |
| EC-02 | Order Creation (Write Burst) | DDIA Ch.3 LSM vs B-Tree | "Order timeout saat 11.11 sale" | PostgreSQL vs ScyllaDB vs CockroachDB |
| EC-03 | Inventory Decrement (Hot Key) | DDIA Ch.10 Hot Keys | "Stock iPhone habis tapi oversold 500 unit" | Redis Cluster vs PostgreSQL SELECT FOR UPDATE |
| EC-04 | Cart Persistence | DBI RUM Conjecture | "Cart hilang setelah user close app" | Redis (AOF) vs DragonflyDB vs PostgreSQL |
| EC-05 | Order History (Range Query) | DDIA Ch.3 B-Tree Range | "Load order history 2 tahun lambat" | PostgreSQL vs TimescaleDB vs ClickHouse |

**Decision Matrix:**
```
Flash Sale (Write Burst) → LSM-based (ScyllaDB) atau queue + PostgreSQL
Product Catalog (Read Heavy) → B-Tree (PostgreSQL) + OpenSearch
Inventory (Consistency) → PostgreSQL dengan proper locking
Session/Cart → Redis dengan persistence
```

---

## 2. Fintech / Digital Banking

### Use Case: Transaction Processing + Fraud Detection

**Business Context:** OVO/GoPay/Dana-scale dengan:
- 1M+ daily transactions
- Fraud detection < 100ms
- Zero tolerance for data loss
- Regulatory audit trail

| ID | Scenario | Theory Source | Real Problem | DBs to Test |
|----|----------|---------------|--------------|-------------|
| FT-01 | Balance Update (ACID) | DDIA Ch.7 Transactions | "Double spend saat network timeout" | PostgreSQL vs CockroachDB vs MySQL |
| FT-02 | Transaction Ledger (Append-Only) | DDIA Ch.3 LSM Append | "Audit trail harus immutable" | PostgreSQL vs ClickHouse vs Cassandra |
| FT-03 | Fraud Scoring (Low Latency) | DBI B-Tree Lookup | "Fraud check harus < 50ms" | Redis vs PostgreSQL vs ScyllaDB |
| FT-04 | Statement Generation (Batch) | DDIA Ch.10 Batch | "Generate 1M statements monthly" | PostgreSQL vs ClickHouse |
| FT-05 | Multi-Region Consistency | DBI Raft Consensus | "User transfer Jakarta-Surabaya harus consistent" | CockroachDB vs PostgreSQL + Citus |

**Decision Matrix:**
```
Core Banking (ACID) → PostgreSQL / CockroachDB
Ledger (Immutable) → PostgreSQL + partitioning atau ClickHouse
Real-time Fraud → Redis + ML model
Reporting → ClickHouse atau TimescaleDB
```

---

## 3. Social Media / Content Platform

### Use Case: Feed Generation + User Interactions

**Business Context:** Twitter/Instagram-scale dengan:
- 10M+ users
- 1B+ posts
- Feed generation < 200ms
- Like/comment real-time

| ID | Scenario | Theory Source | Real Problem | DBs to Test |
|----|----------|---------------|--------------|-------------|
| SM-01 | Feed Fan-Out (Write) | DDIA Ch.3 Write Amplification | "Celebrity post bikin sistem lag" | Cassandra vs ScyllaDB vs Redis |
| SM-02 | Feed Read (Timeline) | DDIA Ch.2 Document Locality | "Load feed lambat untuk user dengan 1000 following" | MongoDB vs PostgreSQL vs Cassandra |
| SM-03 | Like Counter (Hot Key) | DDIA Ch.10 Hot Keys | "Like count inconsistent saat viral" | Redis Cluster vs Cassandra Counter |
| SM-04 | Follower Graph (Traversal) | DDIA Ch.2 Graph Model | "Suggest friends-of-friends" | Neo4j vs PostgreSQL recursive CTE |
| SM-05 | Hashtag Trending | DDIA Ch.11 Stream Window | "Real-time trending topics" | Redis Sorted Set vs ClickHouse |

**Decision Matrix:**
```
Feed Storage → Cassandra/ScyllaDB (write-heavy, eventually consistent OK)
User Profile → PostgreSQL (relational, ACID)
Counters → Redis (fast increment, acceptable inconsistency)
Social Graph → Neo4j atau PostgreSQL dengan ltree
Trending → Redis Sorted Set + ClickHouse untuk analytics
```

---

## 4. IoT / Telemetry Platform

### Use Case: Sensor Data Ingestion + Analytics

**Business Context:** Smart factory / fleet tracking dengan:
- 100K devices
- 1M data points/minute
- Time-series queries
- Anomaly detection

| ID | Scenario | Theory Source | Real Problem | DBs to Test |
|----|----------|---------------|--------------|-------------|
| IOT-01 | High-Volume Ingestion | DDIA Ch.3 LSM Write | "Sensor data loss saat peak" | TimescaleDB vs InfluxDB vs ClickHouse |
| IOT-02 | Downsampling | DDIA Ch.11 Windowing | "Storage cost 10x lipat karena raw data" | TimescaleDB vs InfluxDB |
| IOT-03 | Device Lookup (Point) | DBI B-Tree Lookup | "Get latest reading per device" | TimescaleDB vs InfluxDB vs PostgreSQL |
| IOT-04 | Range Query (Time) | DDIA Ch.3 B-Tree Range | "Query 1 month data untuk 1 device" | TimescaleDB vs ClickHouse |
| IOT-05 | Aggregation (Real-time) | DDIA Ch.11 Stream | "Dashboard average per 5 menit" | TimescaleDB continuous aggregate vs ClickHouse materialized view |

**Decision Matrix:**
```
Ingestion → TimescaleDB atau InfluxDB (time-series optimized)
Long-term Storage → ClickHouse (columnar, compressed)
Real-time Dashboard → TimescaleDB continuous aggregates
Alerting → Redis Streams atau Kafka
```

---

## 5. AI/ML Platform

### Use Case: Vector Search + RAG

**Business Context:** Semantic search / chatbot dengan:
- 10M+ documents
- Embedding dimension: 768-1536
- Search latency < 100ms
- Hybrid search (vector + filter)

| ID | Scenario | Theory Source | Real Problem | DBs to Test |
|----|----------|---------------|--------------|-------------|
| AI-01 | Vector Indexing | HNSW papers | "Index build 10M vectors butuh 2 jam" | Milvus vs Qdrant vs pgvector |
| AI-02 | Recall vs Latency | ANN trade-offs | "Search cepat tapi hasil tidak relevan" | Milvus vs Qdrant vs pgvector |
| AI-03 | Hybrid Search | Vector + metadata | "Filter by category + semantic search" | Milvus vs Qdrant |
| AI-04 | Dimension Scaling | Memory overhead | "Upgrade dari 768 ke 1536 dim, memory 2x" | All vector DBs |
| AI-05 | Update/Delete | Vector index rebuild | "Re-embed 100K documents, index corrupt" | Milvus vs Qdrant |

**Decision Matrix:**
```
Small scale (< 1M vectors) → pgvector (simpler, PostgreSQL ecosystem)
Large scale (> 10M vectors) → Milvus atau Qdrant
Hybrid search priority → Qdrant (better filtering)
GPU available → Milvus (GPU acceleration)
```

---

## 6. Gaming / Leaderboard

### Use Case: Real-time Leaderboard + Player State

**Business Context:** Mobile game dengan:
- 1M+ active players
- Leaderboard update real-time
- Player state persistence
- Anti-cheat validation

| ID | Scenario | Theory Source | Real Problem | DBs to Test |
|----|----------|---------------|--------------|-------------|
| GM-01 | Leaderboard Update | Redis Sorted Set | "Leaderboard lag 5 detik" | Redis vs DragonflyDB vs Valkey |
| GM-02 | Player State Save | DBI RUM Trade-off | "Progress hilang setelah crash" | Redis AOF vs PostgreSQL |
| GM-03 | Match History | DDIA Ch.3 Append-Only | "Load match history timeout" | PostgreSQL vs Cassandra vs ClickHouse |
| GM-04 | Guild/Clan Data | DDIA Ch.2 Document | "Guild dengan 1000 member lambat load" | MongoDB vs PostgreSQL JSONB |
| GM-05 | Concurrent Update | DBI Locking | "Dua player claim reward yang sama" | PostgreSQL vs Redis Lua script |

**Decision Matrix:**
```
Leaderboard → Redis Sorted Set (O(log N) insert, O(log N) rank query)
Player State → Redis + periodic PostgreSQL backup
Match History → ClickHouse (analytics) atau Cassandra (write-heavy)
Guild/Social → PostgreSQL dengan proper indexing
```

---

## 7. Log & Observability

### Use Case: Centralized Logging + APM

**Business Context:** Microservices platform dengan:
- 100+ services
- 10GB logs/day
- Query latency < 5s
- Retention 30 days

| ID | Scenario | Theory Source | Real Problem | DBs to Test |
|----|----------|---------------|--------------|-------------|
| LOG-01 | Log Ingestion | DDIA Ch.3 LSM | "Log loss saat service spike" | OpenSearch vs ClickHouse vs Loki |
| LOG-02 | Full-Text Search | Inverted Index | "Search error message across services" | OpenSearch vs ClickHouse |
| LOG-03 | Aggregation | DDIA Ch.10 Batch | "Count errors per service per hour" | OpenSearch vs ClickHouse |
| LOG-04 | Retention/Compaction | LSM Compaction | "Storage 80% full, need cleanup" | OpenSearch ILM vs ClickHouse TTL |
| LOG-05 | Trace Correlation | DDIA Ch.2 Graph | "Trace request across 10 services" | Jaeger + Cassandra vs Tempo + S3 |

**Decision Matrix:**
```
Logs (search-heavy) → OpenSearch atau Loki
Logs (analytics-heavy) → ClickHouse
Metrics → Prometheus + TimescaleDB
Traces → Jaeger/Tempo
```

---

## 8. Benchmark Implementation Priority

### Phase 1: Core Storage Engine (Week 1-2)
| ID | Scenario | Business Value |
|----|----------|----------------|
| EC-01 | Product Search | Catalog performance |
| FT-01 | Balance Update | ACID correctness |
| IOT-01 | High-Volume Ingestion | Time-series baseline |

### Phase 2: Distributed & Consistency (Week 3-4)
| ID | Scenario | Business Value |
|----|----------|----------------|
| EC-03 | Inventory Hot Key | Flash sale readiness |
| SM-01 | Feed Fan-Out | Write scaling |
| FT-05 | Multi-Region | Geo-distribution |

### Phase 3: Specialized Workloads (Week 5-6)
| ID | Scenario | Business Value |
|----|----------|----------------|
| AI-01 | Vector Indexing | RAG pipeline |
| LOG-01 | Log Ingestion | Observability |
| GM-01 | Leaderboard | Real-time ranking |

---

## Summary: Scenario → DB Recommendation

| Use Case | Primary Need | Recommended DB | Fallback |
|----------|--------------|----------------|----------|
| E-commerce Catalog | Read-heavy, search | PostgreSQL + OpenSearch | MongoDB |
| E-commerce Orders | Write burst, ACID | PostgreSQL | CockroachDB |
| Fintech Ledger | Immutable, audit | PostgreSQL | ClickHouse |
| Social Feed | Write fan-out | Cassandra/ScyllaDB | MongoDB |
| Social Graph | Traversal | Neo4j | PostgreSQL recursive |
| IoT Telemetry | Time-series ingest | TimescaleDB | InfluxDB |
| IoT Analytics | Aggregation | ClickHouse | TimescaleDB |
| Vector Search | ANN search | Qdrant | pgvector |
| Leaderboard | Sorted ranking | Redis | DragonflyDB |
| Logging | Full-text search | OpenSearch | ClickHouse |

---

## References

1. **DDIA** — Kleppmann, Martin. "Designing Data-Intensive Applications." O'Reilly, 2017.
2. **Database Internals** — Petrov, Alex. "Database Internals: A Deep Dive." O'Reilly, 2019.
3. **Real-world patterns** from Tokopedia, Gojek, Shopee engineering blogs
