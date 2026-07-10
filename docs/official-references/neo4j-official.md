# Neo4j Official Documentation Reference

Source: https://neo4j.com/docs/

## Overview

Neo4j is a native graph database using property graph model. Stores nodes, relationships, and properties. Designed for connected data queries where relationship traversal is the primary access pattern.

## Storage Engine: Native Graph Storage

### Architecture
- **Nodes**: Entities with labels and properties
- **Relationships**: Directed, typed connections between nodes
- **Properties**: Key-value pairs on nodes and relationships
- **Native storage**: Index-free adjacency (pointer-based traversal)

### Index-Free Adjacency
```
Node → Fixed-size record → Pointer to first relationship
Relationship → Pointer to start/end node + next relationship (doubly-linked list)
```

### Write Path
```
Transaction → WAL (write-ahead log) → Page cache → Background flush to store files
```

### Read Path
```
Query → Plan (Cypher optimizer) → Traverse via pointer chains → Return results
```

## Performance Characteristics (Official Claims)

### Traversal Performance
- **Claim**: "Constant-time relationship traversal regardless of graph size"
- **Why**: Index-free adjacency — each node knows its neighbors via direct pointers
- **Benchmark**: Millions of traversals per second (official)

### Query Performance
- **Claim**: "1000x faster than relational for connected queries"
- **Why**: No JOINs needed — direct pointer following vs hash lookups
- **Caveat**: Only for graph-shaped queries (multi-hop traversals)

### Scalability
- **Claim**: "Billions of nodes and relationships"
- **Limitation**: Single-instance primarily; clustering for HA, not horizontal write scaling

## Best Use Cases (from neo4j.com)

1. **Fraud Detection**
   - Relationship patterns reveal fraud rings
   - Real-time traversal for transaction scoring

2. **Recommendation Engines**
   - "People who bought X also bought Y"
   - Collaborative filtering via graph traversal

3. **Knowledge Graphs**
   - Entity relationships
   - Semantic search
   - Drug discovery

4. **Network & IT Operations**
   - Dependency graphs
   - Impact analysis
   - Root cause analysis

5. **Identity & Access Management**
   - Permission hierarchies
   - Role-based access resolution

## Anti-Patterns (from docs)

1. **Simple CRUD without relationships**
   - Graph overhead not justified
   - Use relational DB instead

2. **Aggregation-heavy analytics**
   - Not optimized for GROUP BY / SUM
   - Use columnar DB (ClickHouse)

3. **High-volume time-series data**
   - Not designed for append-heavy temporal data
   - Use TimescaleDB / InfluxDB

4. **Simple key-value lookups**
   - Overhead for direct KV access
   - Use Redis

## Configuration for Benchmarks

### Memory Settings
```conf
# neo4j.conf
server.memory.heap.initial_size=1g
server.memory.heap.max_size=1g
server.memory.pagecache.size=512m
```

### Transaction Settings
```conf
db.tx_log.rotation.retention_policy=1 days
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| db.query.execution_latency_millis | Query latency | <10ms for simple |
| vm.heap.used | JVM heap usage | <80% |
| vm.page_cache.hits/misses | Cache hit ratio | >95% |
| bolt.connections_opened | Client connections | Stable |
| db.store.size | Total store size | Monitor growth |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H49 | Constant-time traversal | Exp: Vary graph size, measure 1-hop latency |
| H50 | 1000x faster than SQL for multi-hop | Exp: Compare 3-hop query vs PG recursive CTE |
| H51 | Index-free adjacency scaling | Exp: Read latency at 100K vs 1M nodes |
| H52 | Sub-ms single-hop reads | Exp 2: Point read latency (MATCH by key) |

## Cypher vs SQL for Graph Queries

```cypher
// Neo4j: Find friends-of-friends
MATCH (p:Person {name: "Alice"})-[:KNOWS*2..2]-(fof:Person)
RETURN DISTINCT fof.name

// SQL equivalent (PostgreSQL recursive CTE):
WITH RECURSIVE fof AS (
  SELECT friend_id FROM friendships WHERE person_id = (SELECT id FROM persons WHERE name = 'Alice')
  UNION
  SELECT f.friend_id FROM friendships f JOIN fof ON f.person_id = fof.friend_id
)
SELECT name FROM persons WHERE id IN (SELECT friend_id FROM fof) AND name != 'Alice';
```

## References

- Documentation: https://neo4j.com/docs/
- Operations Manual: https://neo4j.com/docs/operations-manual/current/
- Performance: https://neo4j.com/docs/operations-manual/current/performance/
- Cypher: https://neo4j.com/docs/cypher-manual/current/
