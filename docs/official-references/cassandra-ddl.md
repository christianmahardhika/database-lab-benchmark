# Cassandra CQL Data Definition - Official Documentation

Source: https://cassandra.apache.org/doc/5.0/cassandra/developing/cql/ddl.html
Version: Apache Cassandra 5.0

## Overview

CQL stores data in **tables**, located in **keyspaces**. A keyspace defines options (especially replication strategy) that apply to all tables within it.

## ALTER TABLE

```sql
ALTER TABLE [ IF EXISTS ] table_name alter_table_instruction

alter_table_instruction ::= 
    ADD [ IF NOT EXISTS ] column_definition ( ',' column_definition)*
  | DROP [ IF EXISTS ] column_name ( ',' column_name )*
  | RENAME [ IF EXISTS ] column_name TO column_name
  | WITH options
```

### ADD Column

```sql
ALTER TABLE addamsFamily ADD gravesite varchar;
```

**Key characteristics:**
- **Constant-time operation** regardless of table size
- New column **cannot be part of primary key** (primary key can never be altered)
- Uses `IF NOT EXISTS` to avoid error if column already exists

**Why instant?** Cassandra uses **sparse column storage**. Each row only stores columns that have values. No concept of "empty slots" to fill.

### DROP Column

```sql
ALTER TABLE users DROP phone;
```

**Key characteristics:**
- **Constant-time operation** based on amount of data
- Column becomes immediately unavailable
- Actual content **removed lazily during compaction** (not immediately)
- Cannot re-add dropped column if it was non-frozen (like collections)
- Uses `IF EXISTS` to avoid error if column doesn't exist

**Warning:** DROP assumes timestamps are "real" microsecond timestamps. Using non-standard timestamps will cause incorrect execution.

### RENAME Column

```sql
ALTER TABLE users RENAME old_name TO new_name;
```

**Restrictions:**
- **Only primary key columns can be renamed**
- Non-primary key columns cannot be renamed
- Cannot rename to existing column name
- Renamed columns shouldn't have dependent secondary indexes

### Change Table Options (WITH)

```sql
ALTER TABLE addamsFamily
   WITH comment = 'A most excellent and useful table';
```

**Important:** Setting any compaction/compression sub-options will **erase ALL previous options** — must re-specify all sub-options you want to keep.

**Cannot change:** CLUSTERING ORDER (set at creation only)

## Primary Key — Cannot Be Altered

The PRIMARY KEY uniquely identifies a row and consists of:

1. **Partition key** — determines which node stores the data
2. **Clustering columns** — determines sort order within partition

**Critical rule:** Primary key **cannot ever be altered** after table creation.

Why? Because:
- Partition key determines physical data location (which nodes)
- Clustering columns determine on-disk sort order
- Changing these would require complete data redistribution across cluster

## Keyspace Operations

### CREATE KEYSPACE

```sql
CREATE KEYSPACE excelsior
   WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 3};

CREATE KEYSPACE excalibur
   WITH replication = {'class': 'NetworkTopologyStrategy', 'DC1' : 1, 'DC2' : 3}
   AND durable_writes = false;
```

**Replication strategies:**
- `SimpleStrategy` — single replication factor across entire cluster (not for production)
- `NetworkTopologyStrategy` — per-datacenter replication factors (production-ready)

### ALTER KEYSPACE

```sql
ALTER KEYSPACE excelsior
    WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 4};
```

### DROP KEYSPACE

```sql
DROP KEYSPACE excelsior;
```

**Warning:** Immediate, irreversible removal of keyspace, all tables, types, functions, and data.

## Table Options

| Option | Default | Description |
|--------|---------|-------------|
| comment | none | Human-readable comment |
| gc_grace_seconds | 864000 (10 days) | Time before garbage collecting tombstones |
| default_time_to_live | 0 | Default TTL in seconds |
| bloom_filter_fp_chance | 0.00075 | False positive probability for bloom filters |
| compaction | varies | Compaction strategy (STCS, LCS, TWCS) |
| compression | LZ4 | Compression algorithm |

## Compaction Strategies

- **SizeTieredCompactionStrategy (STCS)** — Default, good for write-heavy workloads
- **LeveledCompactionStrategy (LCS)** — Good for read-heavy workloads
- **TimeWindowCompactionStrategy (TWCS)** — Good for time-series data

## DROP TABLE vs TRUNCATE

```sql
-- Remove table and all data (irreversible)
DROP TABLE users;

-- Remove all data but keep table structure
TRUNCATE TABLE users;
```

## Key Differences from RDBMS

| Aspect | Cassandra | RDBMS |
|--------|-----------|-------|
| ADD COLUMN | Instant (sparse storage) | May require table scan |
| DROP COLUMN | Instant + lazy cleanup | May require table rewrite |
| ALTER PRIMARY KEY | **Not allowed** | Sometimes possible |
| RENAME COLUMN | Primary key only | Any column |
| Schema propagation | Via gossip protocol | Centralized |
