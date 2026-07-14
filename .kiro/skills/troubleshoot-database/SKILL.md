---
name: troubleshoot-database
description: Diagnose and resolve common database performance and operational issues using structured investigation methodology. Use when debugging slow queries, connection issues, replication problems, or any database incident.
---

## Workflow

### Step 1: Identify Symptoms
Classify the reported issue:
- **Slow queries**: Increased latency, timeouts
- **Connection issues**: "Too many connections", connection refused, pool exhaustion
- **Replication problems**: Lag increasing, replica falling behind, split brain
- **Disk pressure**: Out of space, high I/O wait, write stalls
- **Memory pressure**: OOM kills, swap usage, cache eviction
- **Lock contention**: Deadlocks, long-running transactions blocking others
- **Data issues**: Corruption, inconsistency, missing data

### Step 2: Gather Diagnostics

#### PostgreSQL
```sql
-- Active queries and locks
SELECT pid, now() - pg_stat_activity.query_start AS duration, query, state, wait_event_type, wait_event
FROM pg_stat_activity WHERE state != 'idle' ORDER BY duration DESC;

-- Lock tree
SELECT blocked.pid AS blocked_pid, blocked.query AS blocked_query,
       blocking.pid AS blocking_pid, blocking.query AS blocking_query
FROM pg_stat_activity blocked
JOIN pg_locks bl ON bl.pid = blocked.pid
JOIN pg_locks l ON l.locktype = bl.locktype AND l.relation = bl.relation AND l.pid != bl.pid
JOIN pg_stat_activity blocking ON blocking.pid = l.pid
WHERE NOT bl.granted;

-- Table bloat
SELECT schemaname, relname, n_live_tup, n_dead_tup,
       round(n_dead_tup::numeric / greatest(n_live_tup, 1) * 100, 1) AS dead_pct
FROM pg_stat_user_tables ORDER BY n_dead_tup DESC LIMIT 10;

-- Cache hit ratio
SELECT sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) AS ratio
FROM pg_statio_user_tables;

-- Replication status
SELECT client_addr, state, sent_lsn, replay_lsn,
       pg_wal_lsn_diff(sent_lsn, replay_lsn) AS lag_bytes
FROM pg_stat_replication;
```

#### MySQL
```sql
-- Active queries
SHOW PROCESSLIST;
SELECT * FROM information_schema.innodb_trx ORDER BY trx_started;

-- InnoDB status (locks, deadlocks, buffer pool)
SHOW ENGINE INNODB STATUS\G

-- Replication status
SHOW REPLICA STATUS\G

-- Buffer pool hit ratio
SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_read%';
-- Hit ratio = 1 - (Innodb_buffer_pool_reads / Innodb_buffer_pool_read_requests)
```

#### MongoDB
```javascript
// Current operations
db.currentOp({active: true, secs_running: {$gt: 5}})

// Server status
db.serverStatus()

// Collection stats
db.collection.stats()

// Profiler (slow queries)
db.setProfilingLevel(1, {slowms: 100})
db.system.profile.find().sort({ts: -1}).limit(10)

// Replication status
rs.status()
rs.printReplicationInfo()
```

#### ScyllaDB
```bash
# Node status
nodetool status

# Thread pool stats (detect saturation)
nodetool tpstats

# Compaction stats
nodetool compactionstats

# Table stats
nodetool tablestats <keyspace>.<table>
```

#### Redis
```bash
# Memory info
redis-cli INFO memory

# Slow log
redis-cli SLOWLOG GET 10

# Client list (connections)
redis-cli CLIENT LIST

# Big keys scan
redis-cli --bigkeys
```

### Step 3: Root Cause Analysis

| Symptom | Common Root Causes |
|---------|-------------------|
| High latency | Missing index, lock contention, disk I/O, network |
| Connection exhaustion | Connection leak, no pooling, long transactions |
| Replication lag | Heavy writes, large transactions, slow replica disk |
| Disk full | WAL retention, bloat, temp files, large indexes |
| OOM | Unbounded queries, too many connections, no work_mem limit |
| Deadlocks | Inconsistent lock ordering, long transactions |

### Step 4: Resolution Actions

#### Immediate (stop the bleeding)
- Kill long-running queries: `pg_terminate_backend(pid)` / `KILL <id>`
- Free disk: Remove old WAL, drop temp tables
- Reduce load: Enable connection limits, throttle batch jobs
- Failover: Promote replica if primary is unrecoverable

#### Short-term (fix the root cause)
- Add missing index
- Fix connection leak in application
- Tune autovacuum / compaction settings
- Increase instance size or disk IOPS
- Configure proper timeouts (statement_timeout, lock_timeout)

#### Long-term (prevent recurrence)
- Set up alerting for the symptom
- Add the check to monitoring dashboard
- Document in runbook
- Consider architectural change (read replicas, caching layer, sharding)

### Step 5: Verify Resolution
- Confirm metrics return to baseline
- Check for secondary effects (e.g., fixing one bottleneck reveals another)
- Document what happened and what was done (postmortem)

## Investigation Methodology

```
1. WHAT changed? (deploy, traffic spike, config change, data growth)
2. WHEN did it start? (correlate with events timeline)
3. WHERE is the bottleneck? (CPU, memory, disk, network, locks)
4. WHY is it happening? (root cause, not just symptom)
5. HOW do we fix it? (immediate + permanent)
```

## Common Anti-Patterns That Cause Issues

| Anti-Pattern | Consequence | Fix |
|-------------|-------------|-----|
| No connection pooling | Connection storm under load | Add PgBouncer/ProxySQL |
| SELECT * with no LIMIT | Memory exhaustion, network saturation | Explicit columns + pagination |
| Missing index on JOIN column | Full table scans, lock escalation | Add targeted index |
| Long-running transactions | Lock contention, replication lag | Break into smaller transactions |
| No statement_timeout | Runaway queries consume all resources | Set 30s default timeout |
| Autovacuum disabled/too slow | Table bloat, index bloat | Tune autovacuum aggressively |
