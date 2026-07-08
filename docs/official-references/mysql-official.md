# MySQL Official Documentation Reference

Source: https://dev.mysql.com/doc/

## Overview

MySQL is the world's most popular open-source relational database. Owned by Oracle. Default storage engine is InnoDB (B-Tree based, ACID compliant).

## Storage Engine: InnoDB (B+Tree)

### Architecture
- **Buffer Pool**: In-memory cache for data and indexes
- **Change Buffer**: Cache for secondary index changes
- **Redo Log**: Write-ahead log for crash recovery
- **Undo Log**: MVCC and rollback support
- **Doublewrite Buffer**: Corruption protection

### Clustered Index
```
Primary Key → B+Tree → Data stored in leaf nodes
Secondary Index → B+Tree → Points to Primary Key
```

### Write Path
```
INSERT → Buffer Pool → Redo Log (sync) → Background flush → Tablespace
```

### Read Path
```
SELECT → Buffer Pool → Read from disk if miss → Return
```

## Performance Characteristics (Official Claims)

### From dev.mysql.com
- **Web applications**: Optimized for LAMP stack
- **Read scaling**: Mature replication
- **Simplicity**: Easy to operate

### Scaling
- **Vertical**: Good up to ~32 cores
- **Horizontal**: Read replicas, MySQL Cluster, Vitess

## Best Use Cases (from mysql.com)

1. **Web Applications**
   - WordPress, Drupal, Magento
   - LAMP/LEMP stack
   - Content management systems

2. **E-commerce**
   - Product catalogs
   - Order management
   - Inventory tracking

3. **SaaS Applications**
   - Multi-tenant databases
   - Simple CRUD operations

4. **Read-heavy workloads**
   - Mature replication
   - Read replicas for scaling

## Anti-Patterns (from docs)

1. **Complex analytics**
   - Limited window functions (improved in 8.0)
   - PostgreSQL better for OLAP

2. **Heavy JSON operations**
   - JSON support exists but PostgreSQL JSONB faster

3. **Geospatial at scale**
   - PostGIS more mature

## Configuration for Benchmarks

### Buffer Pool (most important)
```sql
-- /etc/mysql/mysql.conf.d/mysqld.cnf
innodb_buffer_pool_size = 4G    -- 70% of RAM for dedicated server
innodb_buffer_pool_instances = 4
```

### Redo Log
```sql
innodb_log_file_size = 512M
innodb_log_buffer_size = 64M
innodb_flush_log_at_trx_commit = 1  -- 1=ACID, 2=fast
```

### I/O
```sql
innodb_flush_method = O_DIRECT
innodb_io_capacity = 2000          -- SSD
innodb_io_capacity_max = 4000
```

### Connections
```sql
max_connections = 200
thread_cache_size = 16
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| Innodb_buffer_pool_read_requests / Innodb_buffer_pool_reads | Buffer pool hit ratio | >99% |
| Threads_running | Active queries | <CPU cores |
| Slow_queries | Query performance | Minimize |
| Innodb_row_lock_waits | Lock contention | Low |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H21 | Best for web apps | Exp 8: Simple OLTP |
| H22 | Simple OLTP fast | Exp 8: SELECT by PK |
| H23 | Read scaling via replication | Exp 8d: Replica lag |
| H24 | INSTANT DDL | Exp 4: Schema evolution |
| H25 | Group Replication HA | Exp 9: Failover time |

## MySQL vs PostgreSQL Quick Comparison

| Aspect | MySQL | PostgreSQL |
|--------|-------|------------|
| Query optimizer | Simpler, faster for simple | Advanced, better for complex |
| Replication | More mature async | Streaming, logical |
| JSON | JSON, limited indexing | JSONB, full indexing |
| Full-text | Basic | Advanced |
| Window functions | 8.0+, limited | Full support |
| Extensions | Limited | Rich ecosystem |
| Default isolation | REPEATABLE READ | READ COMMITTED |

## References

- Documentation: https://dev.mysql.com/doc/refman/8.0/en/
- InnoDB: https://dev.mysql.com/doc/refman/8.0/en/innodb-storage-engine.html
- Performance: https://dev.mysql.com/doc/refman/8.0/en/optimization.html
- Replication: https://dev.mysql.com/doc/refman/8.0/en/replication.html
