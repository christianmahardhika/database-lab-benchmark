---
inclusion: manual
---

# Skill: Production Readiness Review

## Purpose
Evaluate whether a database deployment is ready for production traffic by checking against a comprehensive maturity checklist.

## Workflow

### Step 1: Identify the Database and Deployment
- Which database engine and version?
- Deployment topology (single-node, replica set, cluster, multi-DC)
- Cloud provider or self-hosted?
- Expected traffic pattern and scale

### Step 2: Run Checklist

#### High Availability
- [ ] Replication configured (sync or async — document the choice and RPO impact)
- [ ] Automatic failover tested (Patroni, Sentinel, Replica Set election)
- [ ] Failover time measured and acceptable for SLA
- [ ] Client connection routing handles failover (DNS, pooler, driver config)
- [ ] Split-brain protection in place

#### Backup & Recovery
- [ ] Automated backup schedule (daily full + continuous WAL/binlog/oplog)
- [ ] Backup stored off-instance (S3, GCS, separate volume)
- [ ] Restore procedure documented and tested
- [ ] PITR capability verified (can restore to specific timestamp)
- [ ] Backup monitoring (alert on failure)
- [ ] Retention policy defined (GFS rotation)

#### Monitoring & Alerting
- [ ] Key metrics collected (connections, QPS, latency, disk, replication lag)
- [ ] Dashboards built (Grafana or equivalent)
- [ ] Alerts configured for: disk >80%, connections >70%, lag >5s, errors >0.1%
- [ ] Slow query logging enabled
- [ ] Log aggregation in place

#### Performance
- [ ] Connection pooling configured (PgBouncer, ProxySQL, driver-level)
- [ ] Query performance baselined (know normal p99)
- [ ] Indexes cover critical query patterns
- [ ] Resource limits appropriate for workload (CPU, RAM, IOPS)
- [ ] Load tested at 2x expected peak

#### Security
- [ ] No default passwords
- [ ] Network access restricted (private subnet, security groups)
- [ ] TLS enabled for client connections
- [ ] TLS enabled for replication traffic
- [ ] Least-privilege access (separate app, admin, monitoring users)
- [ ] Encryption at rest (volume-level or native TDE)
- [ ] Audit logging for DDL and admin operations

#### Operational Readiness
- [ ] Runbooks for common incidents (high connections, lag, disk, failover)
- [ ] On-call rotation knows database basics
- [ ] Capacity plan for next 6 months
- [ ] Schema change procedure documented (zero-downtime DDL)
- [ ] Maintenance windows defined (vacuum, compaction, upgrades)

#### Data Integrity
- [ ] Consistency level chosen and documented
- [ ] Application handles transient errors (retry logic)
- [ ] Data validation at application layer
- [ ] Checksums enabled (page-level corruption detection)

### Step 3: Score and Prioritize Gaps

| Score | Status | Action |
|-------|--------|--------|
| 0-40% | Not ready | Block launch until critical items resolved |
| 40-70% | Minimum viable | Launch with risk acceptance + timeline to close gaps |
| 70-90% | Production ready | Launch with monitoring; close remaining items in sprint |
| 90-100% | Mature | Maintain and iterate |

### Step 4: Generate Action Plan
For each gap, provide:
- Priority (P0: blocks launch, P1: first sprint, P2: next quarter)
- Effort estimate (hours/days)
- Specific implementation steps
- Reference to relevant lab exercise for testing

## Database-Specific Checks

### PostgreSQL Extras
- [ ] autovacuum tuned for workload (scale_factor, threshold)
- [ ] pg_stat_statements enabled for query analysis
- [ ] shared_preload_libraries configured
- [ ] pg_hba.conf restricts access properly

### MySQL Extras
- [ ] innodb_flush_log_at_trx_commit = 1 (or documented reason for 2)
- [ ] Binary log format = ROW
- [ ] GTID enabled for replication
- [ ] Performance Schema enabled

### MongoDB Extras
- [ ] Replica set with odd number of voting members
- [ ] Write concern = majority for critical writes
- [ ] Read preference documented per query type
- [ ] Shard key chosen carefully (if sharded)

### ScyllaDB Extras
- [ ] Compaction strategy matched to workload (STCS/LCS/TWCS)
- [ ] Consistency level per query type documented
- [ ] Repair scheduled (weekly recommended)
- [ ] Overprovisioned mode disabled in production

### Redis Extras
- [ ] maxmemory-policy set (not noeviction for caches)
- [ ] Persistence mode chosen (AOF everysec recommended)
- [ ] Key naming convention documented
- [ ] TTL policy for cache entries
