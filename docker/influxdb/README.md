# InfluxDB Lab Notes

## Topology: Standalone Only (OSS Limitation)

InfluxDB OSS 2.x **does not support clustering or replication**.

Clustering is only available in:
- **InfluxDB Enterprise** (commercial, closed-source)
- **InfluxDB Cloud** (managed service)
- **InfluxDB 3.x IOx** (new architecture, still evolving)

### What This Means for Benchmarks

- The standalone `docker-compose.yml` is the only available near-production setup for OSS
- For HA in production, users typically run InfluxDB behind a load balancer with external backup
- Compare this limitation against TimescaleDB (which inherits PG replication for free)

### Production Alternatives

| Option | Clustering | Notes |
|--------|-----------|-------|
| InfluxDB OSS 2.x | ❌ | Single node only |
| InfluxDB Enterprise | ✅ | Commercial license |
| InfluxDB Cloud | ✅ | Managed service |
| InfluxDB 3.x (IOx) | ✅ | New engine, object storage backend |

### Relevant Hypothesis

- **H53**: Write throughput is measured on single node (fair comparison since production also runs single node)
- This is a legitimate limitation to document in final benchmark report

### Alternative HA Pattern (Not True Clustering)

```
Client → HAProxy/Nginx → InfluxDB-1 (active)
                       → InfluxDB-2 (passive, restore from backup)
```

This is NOT real replication — it's active-passive failover with periodic backup restore.
