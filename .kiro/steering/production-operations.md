---
inclusion: manual
---

# Production-Grade Database Operations Guide

## Maturity Model: From Dev to Production

### Level 1 — Development (this lab)
- Single-node Docker containers
- No persistence guarantees
- No monitoring
- Manual operations

### Level 2 — Staging
- Replica sets / replication configured
- Basic monitoring (Prometheus + Grafana)
- Automated backups (daily)
- Connection pooling

### Level 3 — Production
- Multi-node with HA (automatic failover)
- Full observability (metrics, logs, traces)
- PITR backup with tested restore procedures
- Connection pooling + circuit breakers
- Capacity planning and alerting
- Runbooks for common incidents

### Level 4 — Production at Scale
- Horizontal scaling (sharding / read replicas)
- Multi-region / multi-DC deployment
- Chaos engineering (failure injection)
- Automated remediation
- Cost optimization (right-sizing, tiered storage)

---

## 1. High Availability Patterns

### PostgreSQL HA

**Recommended Stack**: Patroni + etcd + HAProxy (or PgBouncer)

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   HAProxy   │     │   HAProxy   │     │   HAProxy   │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Patroni +  │     │  Patroni +  │     │  Patroni +  │
│  PG Primary │◄───►│  PG Replica │◄───►│  PG Replica │
└──────┬──────┘     └─────────────┘     └─────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────┐
│                  etcd cluster (3 nodes)              │
└─────────────────────────────────────────────────────┘
```

**Key Decisions**:
- Synchronous replication: 0 data loss but higher write latency
- Asynchronous replication: better performance but potential data loss on failover
- `synchronous_commit = remote_apply` for strongest guarantee

**Failover Time**: 10-30 seconds with Patroni

### MySQL HA

**Option A**: InnoDB Cluster (Group Replication + MySQL Router)
- Built-in, single-vendor solution
- Multi-primary or single-primary mode
- Automatic conflict detection (certification-based)

**Option B**: Orchestrator + ProxySQL + semi-sync replication
- Battle-tested at scale (GitHub, Booking.com)
- More control over failover behavior
- ProxySQL adds query routing + caching

**Failover Time**: 5-30 seconds

### MongoDB HA

**Replica Set** (built-in):
- Minimum 3 members (or 2 data + 1 arbiter)
- Automatic election when primary fails
- Write concern `majority` for durability
- Read preference `secondaryPreferred` for read scaling

**Failover Time**: 5-12 seconds (electionTimeoutMillis default 10s)

### ScyllaDB HA

**Built-in** (no external orchestrator needed):
- Multi-node cluster with consistent hashing
- Rack/DC-aware token placement
- Tunable consistency (ONE, QUORUM, ALL)
- Hinted handoff for temporary node failures

**No failover needed**: All nodes serve all roles; clients route to any node.

### Redis HA

**Redis Sentinel** (for single-master HA):
- 3+ Sentinel processes monitor master
- Automatic failover and client notification
- Quorum-based decision making

**Redis Cluster** (for scaling + HA):
- 16384 hash slots distributed across masters
- Each master has 1+ replicas
- Automatic failover within slot owners

**Failover Time**: 15-30 seconds (Sentinel), 1-2 seconds (Cluster)

---

## 2. Backup & Recovery Strategies

### RPO/RTO Matrix

| Database | Backup Method | RPO | RTO |
|----------|--------------|-----|-----|
| PostgreSQL | WAL archiving + base backup | 0 (continuous) | 5-30 min |
| MySQL | binlog + XtraBackup | 0 (continuous) | 5-30 min |
| MongoDB | Oplog + mongodump | seconds | 5-60 min |
| ScyllaDB | Snapshots + incremental | minutes | 10-60 min |
| ClickHouse | Partition backup + S3 | hours | 10-60 min |
| Redis | AOF (everysec) | 1 second | 1-5 min |

### Backup Testing Cadence

| Test Type | Frequency | What to Verify |
|-----------|-----------|----------------|
| Restore from backup | Weekly | Data integrity, restore time |
| PITR to specific time | Monthly | WAL/binlog continuity |
| Full DR simulation | Quarterly | End-to-end recovery in new region |
| Backup size tracking | Daily (automated) | Growth rate, storage costs |

### Retention Policy (GFS)

```yaml
retention:
  continuous: 24h    # WAL/binlog/oplog for PITR
  hourly: 48         # Last 48 hourly snapshots
  daily: 30          # Last 30 daily backups
  weekly: 12         # Last 12 weekly backups
  monthly: 12        # Last 12 monthly backups
```

---

## 3. Monitoring & Observability

### Essential Metrics per Database

#### PostgreSQL
```
# Saturation
pg_stat_activity (active connections / max_connections)
pg_stat_bgwriter (buffers_backend — should be low)
dead_tuples / live_tuples ratio (bloat indicator)

# Performance
tps (transactions per second)
query_duration_ms (p50, p95, p99)
cache_hit_ratio (should be >99%)
index_hit_ratio

# Replication
pg_stat_replication.replay_lag
pg_stat_replication.write_lag

# Storage
database_size_bytes
table_bloat_ratio
WAL generation rate (MB/s)
```

#### MySQL
```
# Saturation
Threads_connected / max_connections
Innodb_buffer_pool_reads / Innodb_buffer_pool_read_requests (miss ratio)
Innodb_row_lock_waits

# Performance
Questions / Uptime (QPS)
Slow_queries
Innodb_buffer_pool_hit_ratio (should be >99%)

# Replication
Seconds_Behind_Master
Slave_SQL_Running_State
```

#### MongoDB
```
# Saturation
connections.current / connections.available
globalLock.currentQueue.total
wiredTiger.cache.bytes currently in cache / maximum bytes configured

# Performance  
opcounters (insert, query, update, delete, command)
opLatencies (reads, writes, commands) — p50, p95, p99
document scan ratio (docsExamined / docsReturned — should be close to 1)

# Replication
rs.status().members[].optimeDate (lag)
repl.apply.ops
```

#### ScyllaDB
```
# Saturation
scylla_reactor_utilization (per-shard CPU, should be <90%)
scylla_memory_allocated_memory / scylla_memory_total_memory
scylla_storage_proxy_coordinator_write_latency (p99)

# Performance
scylla_transport_requests_served (CQL ops/sec)
scylla_storage_proxy_coordinator_read_latency
scylla_storage_proxy_coordinator_write_latency

# Compaction
scylla_compaction_manager_compactions (pending/completed)
scylla_sstables_count (per table, per shard)
```

#### Redis
```
# Saturation
connected_clients / maxclients
used_memory / maxmemory
instantaneous_ops_per_sec

# Performance
latency (redis-cli --latency)
keyspace_hits / (keyspace_hits + keyspace_misses) — hit ratio
evicted_keys (should be 0 for non-cache workloads)

# Persistence
rdb_last_bgsave_status
aof_last_bgrewrite_status
rdb_last_bgsave_time_sec
```

### Alerting Thresholds (starting point)

| Metric | Warning | Critical |
|--------|---------|----------|
| Connection usage | >70% | >90% |
| CPU utilization | >70% | >90% |
| Disk usage | >70% | >85% |
| Replication lag | >1s | >10s |
| Query latency p99 | >100ms | >500ms |
| Cache hit ratio | <98% | <95% |
| Error rate | >0.1% | >1% |

---

## 4. Connection Management

### Pool Sizing Formula

```
Optimal connections = (CPU cores * 2) + effective_spindle_count
```

For SSD: `connections ≈ (cores * 2) + 1`

### Architecture by Scale

**<100 req/s**: Direct connections with app-level pool
```
[App] → [DB Pool in app] → [Database]
```

**100-10K req/s**: External pooler
```
[App instances] → [PgBouncer/ProxySQL] → [Database]
```

**>10K req/s**: Tiered pooling
```
[App instances] → [Local pool] → [Central pooler] → [Database cluster]
```

### Connection Pooler Selection

| Pooler | Database | Killer Feature |
|--------|----------|----------------|
| PgBouncer | PostgreSQL | Minimal memory (2KB/conn), transaction pooling |
| Pgpool-II | PostgreSQL | Load balancing, query cache, watchdog HA |
| ProxySQL | MySQL | Query routing, caching, R/W split |
| MySQL Router | MySQL | Native InnoDB Cluster integration |
| mongos | MongoDB | Built-in shard router |

---

## 5. Schema Evolution in Production

### PostgreSQL Online DDL

| Operation | Locks | Safe? |
|-----------|-------|-------|
| ADD COLUMN (no default) | ACCESS EXCLUSIVE (brief) | ✅ |
| ADD COLUMN (with default) | ACCESS EXCLUSIVE (brief, PG 11+) | ✅ |
| DROP COLUMN | ACCESS EXCLUSIVE (brief) | ✅ (marks invisible) |
| ADD INDEX | ShareLock (blocks writes) | ⚠️ Use CONCURRENTLY |
| ADD INDEX CONCURRENTLY | No lock | ✅ |
| ALTER COLUMN TYPE | ACCESS EXCLUSIVE (rewrites table) | ❌ Use shadow table |

**Tool**: `pg_repack` for zero-downtime table rewrites

### MySQL Online DDL

| Algorithm | Locks | When to Use |
|-----------|-------|-------------|
| INSTANT | None | Add column, rename column (8.0.28+) |
| INPLACE | None/Brief | Add index, change column default |
| COPY | Table lock | Change column type, drop PK |

**Tool**: `pt-online-schema-change` or `gh-ost` for zero-downtime

### MongoDB Schema Evolution

- **No explicit migrations**: Schema is application-enforced
- **Pattern**: Version field in documents + application-level migration
- **Validation**: JSON Schema validation rules (optional enforcement)
- **Risk**: Inconsistent documents if app versions coexist during rolling deploy

### ScyllaDB Schema Changes

- ADD COLUMN: Instant (sparse storage, no rewrite)
- DROP COLUMN: Instant (marked as dropped, cleaned on compaction)
- ALTER TYPE: Not supported — create new table + migrate
- ADD INDEX: Secondary indexes supported but discouraged (prefer denormalization)

---

## 6. Capacity Planning

### Growth Estimation Template

```
Current state:
  - Data size: X GB
  - Growth rate: Y GB/month
  - Query load: Z QPS (peak)
  - Connections: N concurrent

6-month projection:
  - Data size: X + (Y * 6)
  - Query load: Z * 1.5 (assume 50% growth)
  - Connections: N * 1.5

Action thresholds:
  - Vertical scale: when CPU >70% sustained or RAM <20% free
  - Horizontal scale: when single node can't serve p99 latency SLA
  - Archive/partition: when table size >100GB (B-Tree) or >1TB (LSM)
```

### Storage Engine Capacity Characteristics

| Engine | Max Practical Table Size | Scaling Trigger |
|--------|--------------------------|-----------------|
| PostgreSQL (B-Tree) | 100GB-1TB per table | Vacuum time, index rebuild time |
| MySQL (InnoDB) | 100GB-2TB per table | ALTER TABLE time, backup time |
| MongoDB | 2TB per shard | Chunk migration hotspots |
| ScyllaDB | 1TB+ per node (linear) | Compaction pressure, repair time |
| ClickHouse | 10TB+ per shard | Query scan time, merge pressure |
| Redis | Limited by RAM | Memory cost, fork time for RDB |

---

## 7. Incident Response Runbooks

### High Connection Count
1. Check: `SELECT count(*) FROM pg_stat_activity` (or equivalent)
2. Identify: Idle connections, leaked connections, connection storms
3. Mitigate: Kill idle connections, restart connection pool, scale pooler
4. Prevent: Set `idle_in_transaction_session_timeout`, tune pool max

### Replication Lag Spike
1. Check: Replica I/O and SQL thread status
2. Identify: Long-running transaction on primary, heavy write burst, slow disk
3. Mitigate: Pause non-critical writes, increase parallel apply threads
4. Prevent: Monitor lag, alert at 1s, separate OLTP from batch workloads

### Disk Space Emergency
1. Check: Which tables/databases are largest
2. Identify: Bloat, WAL retention, temp files, large indexes
3. Mitigate: Remove old WAL, drop temp tables, vacuum (PG), optimize (MySQL)
4. Prevent: Monitor growth, set up data retention policies, archive old data

### Memory Pressure (OOM Risk)
1. Check: Process memory, buffer pool usage, OS cache
2. Identify: Query memory spills, connection count, sort buffers
3. Mitigate: Kill expensive queries, reduce work_mem, limit connections
4. Prevent: Set `statement_timeout`, limit `max_connections`, right-size instance

---

## 8. Security Baseline

### Authentication
- Never use default passwords in production
- Use certificate-based auth for replication connections
- Rotate credentials quarterly (or use short-lived tokens via Vault)

### Network
- Database ports never exposed to public internet
- Use private subnets + security groups / network policies
- TLS for all connections (even internal)
- VPN or bastion host for admin access

### Authorization
- Principle of least privilege: app user gets only needed tables/operations
- Separate users for: app, migrations, monitoring, backup
- No `SUPERUSER` / `root` for application connections

### Encryption
- At-rest: Volume encryption (EBS, GCE PD) or native TDE
- In-transit: TLS 1.2+ for all client-server and replication traffic
- Backup encryption: Encrypt before storing in object storage

### Audit
- Enable query logging for DDL and admin operations
- Log authentication failures
- Use pgAudit (PostgreSQL) or audit plugin (MySQL) for compliance
