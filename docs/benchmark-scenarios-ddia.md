# Benchmark Scenarios from DDIA & Database Internals

> **Sources:**
> - DDIA: Designing Data-Intensive Applications (Martin Kleppmann)
> - DBI: Database Internals (Alex Petrov)
> - KB: software-engineering-kb (Qdrant collection)

---

## 1. Storage Engine Fundamentals

### 1.1 Write Amplification Measurement
**Source:** DDIA Ch.3 Storage & Retrieval, DBI Ch.7

**Theory:**
```
B-Tree Write Amplification:
  1 INSERT = WAL write (1x) + Page read + Page write (1x) + Potential splits (Nx)
  Total: ~10-30x original data size

LSM-Tree Write Amplification:
  1 INSERT = WAL (1x) + Memtable + Flush L0 (1x) + Compact L0→L1→L2→...Ln
  Total: ~10-50x (depends on levels)
```

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| WA-01 | Write Amplification Ratio | Measure bytes written to disk vs logical bytes inserted | PostgreSQL, MySQL, ScyllaDB, Cassandra | Ratio (disk_bytes / logical_bytes) |
| WA-02 | WAL Size Growth | Track WAL/binlog size growth during sustained writes | PostgreSQL, MySQL, MongoDB | MB/s WAL growth |
| WA-03 | Compaction I/O | Measure I/O during LSM compaction | ScyllaDB, Cassandra, RocksDB | IOPS during compaction |

**Implementation:**
```go
type WriteAmplificationBench struct {
    LogicalBytesWritten int64
    DiskBytesWritten    int64  // from iostat or /proc/diskstats
    WALBytesWritten     int64
}

func (b *WriteAmplificationBench) Ratio() float64 {
    return float64(b.DiskBytesWritten) / float64(b.LogicalBytesWritten)
}
```

---

### 1.2 Read Amplification & Bloom Filter Efficiency
**Source:** DDIA Ch.3, DBI "Bloom Filters"

**Theory (from KB):**
> "A Bloom filter uses a large bit array and multiple hash functions. Hash functions are applied to keys of the records in the table to find indices in the bit array, bits for which are set to 1."

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| RA-01 | Read Amplification (exists) | I/O calls per successful point lookup | All | IOPS per read |
| RA-02 | Read Amplification (not exists) | I/O calls for non-existent key lookup | ScyllaDB, Cassandra | IOPS per negative lookup |
| RA-03 | Bloom Filter False Positive Rate | Measure actual vs theoretical false positive rate | ScyllaDB, Cassandra | Percentage |

**Implementation:**
```go
// Negative lookup test - key guaranteed not to exist
func BenchmarkNegativeLookup(b *testing.B) {
    for i := 0; i < b.N; i++ {
        key := fmt.Sprintf("nonexistent_%d_%d", time.Now().UnixNano(), i)
        _, _ = db.Get(key) // should return not found
    }
}
```

---

### 1.3 RUM Conjecture Validation
**Source:** DBI "RUM Conjecture" [ATHANASSOULIS16]

**Theory (from KB):**
> "RUM Conjecture states that reducing two of these overheads inevitably leads to change for the worse in the third one, and that optimizations can be done only at the expense of one of the three parameters."

```
R = Read overhead
U = Update overhead  
M = Memory overhead

Cannot optimize all three simultaneously.
```

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metrics |
|----|------|-------------|-----|---------|
| RUM-01 | RUM Triangle Plot | Measure R, U, M for each storage engine | All 18 DBs | 3-axis radar chart |
| RUM-02 | Memory vs Read Trade-off | Compare Redis (all memory) vs PostgreSQL (disk-backed) | Redis, PostgreSQL | Read latency vs memory usage |
| RUM-03 | Update vs Read Trade-off | Compare LSM (fast write) vs B-Tree (fast read) | ScyllaDB vs PostgreSQL | Write throughput vs read latency |

---

## 2. B-Tree Specific Benchmarks

### 2.1 Page Split Impact
**Source:** DDIA Ch.3, DBI "B-Tree Node Splits"

**Theory:**
```
Insert into full page:
  Page A: [151][155][160][170][175][179] — FULL!
  
  Split into two pages:
  Page A: [151][155][160]
  Page B: [165][170][175][179]
  
  Parent must update → may cascade splits up the tree!
```

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| BT-01 | Page Split Frequency | Count page splits during sequential insert | PostgreSQL, MySQL | Splits/1000 inserts |
| BT-02 | Page Split Latency Spike | Measure latency during page split events | PostgreSQL, MySQL | p99 latency during splits |
| BT-03 | Fillfactor Impact | Compare fillfactor=100 vs 70 on split frequency | PostgreSQL | Split reduction % |

**Implementation:**
```sql
-- PostgreSQL: Monitor page splits
CREATE EXTENSION pageinspect;

SELECT * FROM bt_metap('users_pkey');
SELECT * FROM bt_page_stats('users_pkey', 1);

-- Create index with fillfactor
CREATE INDEX idx_users_70 ON users(email) WITH (fillfactor = 70);
CREATE INDEX idx_users_100 ON users(email) WITH (fillfactor = 100);
```

---

### 2.2 B-Tree Depth & Lookup Complexity
**Source:** DBI "B-Tree Lookup Complexity"

**Theory (from KB):**
> "In terms of number of transfers, the logarithm base is N (number of keys per node). There are K times more nodes on each new level, and following a child pointer reduces the search space by the factor of N."

```
depth = log_b(n)
Where: b = branching factor (~500 for 8KB page), n = number of keys

Example: 500M rows → log₅₀₀(500M) ≈ 4 levels → 4 × 8KB = 32KB per lookup
```

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| BT-04 | Tree Depth vs Dataset Size | Measure index depth at 1M, 10M, 100M rows | PostgreSQL, MySQL | Index depth |
| BT-05 | Lookup Cost at Scale | Point lookup latency at different scales | PostgreSQL, MySQL | µs per lookup |
| BT-06 | Index Size Growth | Index size vs data size ratio | PostgreSQL, MySQL | Index/Data ratio |

---

## 3. LSM-Tree Specific Benchmarks

### 3.1 Compaction Impact
**Source:** DDIA Ch.3 "LSM-Tree mechanics"

**Theory:**
```
Read key (worst case):
  Memtable (miss) → L0 SSTable 1 (miss) → L0 SSTable 2 (miss)  
  → L1 (miss) → L2 (miss) → L3 (found!)
  = 6+ file reads + bloom filter checks
```

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| LSM-01 | Compaction Latency Spike | p99 read latency during compaction | ScyllaDB, Cassandra | p99 latency spike |
| LSM-02 | Compaction I/O Saturation | Disk I/O % used during compaction | ScyllaDB, Cassandra | IOPS % |
| LSM-03 | Level Fanout Impact | Compare different compaction strategies | ScyllaDB | Throughput stability |
| LSM-04 | Memtable Flush Frequency | Track memtable → SSTable flushes | ScyllaDB, Cassandra | Flushes/minute |

---

### 3.2 Space Amplification
**Source:** DDIA Ch.3

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| LSM-05 | Space Amplification Ratio | Actual disk usage vs logical data size | ScyllaDB, Cassandra | Ratio |
| LSM-06 | Tombstone Accumulation | Space used by tombstones before compaction | Cassandra | Tombstone % |

---

## 4. Data Model Benchmarks

### 4.1 Relational vs Document
**Source:** DDIA Ch.2 "Data Models and Query Languages"

**Theory:**
```
Document Model:
- Locality advantage: entire document in one read
- Disadvantage: must load entire document even for small portion

Relational Model:
- JOIN performance critical
- Normalization reduces duplication
```

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| DM-01 | Document Locality | Read partial vs full document | MongoDB, PostgreSQL(JSONB) | Latency |
| DM-02 | JOIN Performance at Scale | 2-table, 3-table, 5-table JOINs | PostgreSQL, MySQL, CockroachDB | Latency |
| DM-03 | Schema-on-Read vs Write | Add new field to existing records | MongoDB vs PostgreSQL | Migration time |

---

### 4.2 Graph Traversal
**Source:** DDIA Ch.2 "Graph Data Models"

**Theory:**
> "In a graph query, you may need to traverse a variable number of edges before you find the vertex you're looking for — that is, the number of joins is not fixed in advance."

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| GR-01 | Depth-1 Traversal | Find direct neighbors | Neo4j, PostgreSQL(recursive CTE) | Latency |
| GR-02 | Depth-3 Traversal | Friends of friends of friends | Neo4j, PostgreSQL | Latency |
| GR-03 | Variable Depth (Cypher *0..) | Path finding with unknown depth | Neo4j | Latency vs depth |

---

## 5. Replication & Consistency Benchmarks

### 5.1 Replication Lag
**Source:** DBI "Eventual Consistency", "Raft"

**Theory (from KB):**
> "Under eventual consistency, updates propagate through the system asynchronously. Formally, it states that if there are no additional updates performed against the data item, eventually all accesses will return the last updated value."

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| REP-01 | Replication Lag (sync) | Write → read from replica latency | PostgreSQL sync, MySQL semisync | ms |
| REP-02 | Replication Lag (async) | Write → eventual consistency time | PostgreSQL async, MongoDB | ms |
| REP-03 | Consistency Window | Time until all replicas agree | Cassandra (QUORUM), MongoDB | ms |

---

### 5.2 Consensus Overhead
**Source:** DBI "Raft", "Paxos"

**Theory (from KB):**
> "Raft was first presented in a paper titled 'In Search of an Understandable Consensus Algorithm' [ONGARO14]."

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| CON-01 | Leader Election Time | Time to elect new leader after failure | CockroachDB, etcd | ms |
| CON-02 | Consensus Round-Trip | Write latency with consensus | CockroachDB (Raft) | ms |
| CON-03 | Split-Brain Recovery | Recovery time after network partition | CockroachDB, MongoDB | seconds |

---

## 6. Batch & Stream Processing Benchmarks

### 6.1 Batch Insert Optimization
**Source:** DDIA Ch.10 "Batch Processing"

**Theory:**
```
MapReduce: Disk → Map → Disk → Reduce → Disk
Dataflow:  Memory → Operation Chain → Memory → Output
```

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| BAT-01 | Bulk Insert (COPY) | 1M rows via COPY/bulk loader | PostgreSQL, MySQL, ClickHouse | rows/sec |
| BAT-02 | Batch vs Single Insert | 1000 single vs 1 batch of 1000 | All | Throughput ratio |
| BAT-03 | Parallel Bulk Load | Multi-threaded bulk insert | ClickHouse, TimescaleDB | rows/sec scaling |

---

### 6.2 Time-Series Windowing
**Source:** DDIA Ch.11 "Stream Processing"

**Theory:**
```
Tumbling Windows: [12:00-12:05] [12:05-12:10] [12:10-12:15]
Sliding Windows:  [12:00-12:05] [12:01-12:06] [12:02-12:07]
Session Windows:  [--session 1--] gap [--session 2--]
```

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| TS-01 | Tumbling Window Aggregate | 5-min aggregates over 1M events | TimescaleDB, InfluxDB, ClickHouse | Query latency |
| TS-02 | Sliding Window | Moving average (1-min slide, 5-min window) | TimescaleDB, InfluxDB | Query latency |
| TS-03 | Downsampling | Aggregate 1s → 1min → 1hour | InfluxDB, TimescaleDB | Storage reduction |

---

## 7. Distributed Benchmarks

### 7.1 Hot Key / Skewed Data
**Source:** DDIA Ch.10 "Handling Skewed Data (Hot Keys)"

**Theory:**
> "Skewed distribution: Some keys have many more values than others. Load imbalance: Reducers handling hot keys become bottlenecks."

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| DIST-01 | Hot Key Write | 80% writes to 1% of keys | Redis Cluster, Cassandra | Hotspot node CPU |
| DIST-02 | Zipfian Distribution | Realistic skewed access pattern | All distributed DBs | Throughput degradation |
| DIST-03 | Key Splitting Effectiveness | With vs without key salting | Cassandra | Throughput improvement |

---

### 7.2 Partition Strategies
**Source:** DDIA Ch.6 "Partitioning" (planned)

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| PART-01 | Hash vs Range Partition | Query patterns on different strategies | CockroachDB, Cassandra | Query latency |
| PART-02 | Cross-Partition Query | Query spanning multiple partitions | CockroachDB, ScyllaDB | Latency overhead |
| PART-03 | Rebalancing Impact | Add node, measure rebalancing time | Cassandra, CockroachDB | Rebalance duration |

---

## 8. In-Memory vs Disk-Backed

### 8.1 Memory Efficiency
**Source:** DBI (from KB)

**Theory:**
> "In-memory storage engine may never need to read from disk if you have enough memory... they can be faster because they can avoid the overheads of encoding in-memory data structures in a form that can be written to disk."

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| MEM-01 | Cold vs Hot Cache | First read vs subsequent reads | PostgreSQL, MySQL | Latency ratio |
| MEM-02 | Memory Pressure | Performance as data exceeds RAM | Redis, DragonflyDB | Throughput degradation |
| MEM-03 | Persistence Overhead | Memory-only vs AOF/RDB persistence | Redis, Valkey | Write latency overhead |

---

## 9. Vector Database Specific

### 9.1 ANN Index Benchmarks
**Source:** Milvus/Qdrant documentation, HNSW papers

**Benchmark Scenario:**
| ID | Name | Description | DBs | Metric |
|----|------|-------------|-----|--------|
| VEC-01 | Recall vs Latency | ANN recall@10 at different latencies | Milvus, Qdrant, pgvector | Recall % |
| VEC-02 | Index Build Time | Time to build HNSW index on 1M vectors | Milvus, Qdrant, pgvector | seconds |
| VEC-03 | Dimension Scaling | Performance at 128, 768, 1536 dimensions | All vector DBs | Latency scaling |
| VEC-04 | Hybrid Search | Vector + metadata filter | Milvus, Qdrant | Latency |

---

## Summary: Priority Matrix

| Priority | Category | Scenarios | Business Value |
|----------|----------|-----------|----------------|
| **P0** | Storage Engine | WA-01, RA-01, BT-01, LSM-01 | Core performance understanding |
| **P1** | Data Models | DM-01, DM-02, GR-01 | Architecture decisions |
| **P1** | Replication | REP-01, REP-02, CON-01 | HA planning |
| **P2** | Batch/Stream | BAT-01, TS-01 | ETL optimization |
| **P2** | Distributed | DIST-01, PART-01 | Scaling decisions |
| **P3** | Vector | VEC-01, VEC-02 | AI/ML workloads |

---

## References

1. **DDIA** — Kleppmann, Martin. "Designing Data-Intensive Applications." O'Reilly, 2017.
   - Chapter 2: Data Models and Query Languages
   - Chapter 3: Storage and Retrieval
   - Chapter 10: Batch Processing
   - Chapter 11: Stream Processing

2. **Database Internals** — Petrov, Alex. "Database Internals: A Deep Dive." O'Reilly, 2019.
   - B-Tree Node Splits, Lookup Complexity
   - Bloom Filters
   - RUM Conjecture [ATHANASSOULIS16]
   - Raft Consensus [ONGARO14]

3. **Obsidian Vault Sources:**
   - `DDIA Storage and Retrieval - Comprehensive Digest.md`
   - `DDIA Data Models and Query Languages - Comprehensive Digest.md`
   - `DDIA Batch and Stream Processing - Digest.md`
   - `DDIA Knowledge Tree - Navigation Hub.md`

4. **Qdrant KB:** `software-engineering-kb` collection (46 chunks from Database Internals)
