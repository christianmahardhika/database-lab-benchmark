---
inclusion: fileMatch
fileMatchPattern: "docker/**"
---

# Docker Operations for Database Lab

## Container Lifecycle

### Starting a Database
```bash
cd docker/<database>
docker compose up -d
# Wait for healthcheck to pass before running benchmarks
docker compose ps  # Check health status
```

### Stopping & Cleaning
```bash
docker compose down       # Stop containers, keep volumes
docker compose down -v    # Stop containers, delete volumes (full reset)
```

### Checking Logs
```bash
docker compose logs -f <service>         # Stream logs
docker compose logs --tail=50 <service>  # Last 50 lines
```

## Port Mapping Reference

| Database | Default Port | Lab Port | Notes |
|----------|-------------|----------|-------|
| PostgreSQL | 5432 | 5432 | Standard |
| TimescaleDB | 5432 | 5433 | Offset to avoid PG conflict |
| MySQL | 3306 | 3306 | Standard |
| MongoDB | 27017 | 27017 | Standard |
| Redis | 6379 | 6379 | Standard |
| Valkey | 6379 | 6381 | Redis fork, offset port |
| DragonflyDB | 6379 | 6380 | Redis-compatible, offset port |
| ScyllaDB | 9042 | 9043 | Offset to avoid Cassandra conflict |
| Cassandra | 9042 | 9042 | Standard |
| ClickHouse HTTP | 8123 | 8123 | Standard |
| ClickHouse Native | 9000 | 9000 | Standard |
| CockroachDB SQL | 26257 | 26257 | PG wire-compatible |
| CockroachDB UI | 8080 | 8081 | Offset to avoid conflicts |
| Elasticsearch | 9200 | 9200 | Standard |
| Neo4j Bolt | 7687 | 7687 | Standard |
| Neo4j HTTP | 7474 | 7474 | Standard |
| InfluxDB | 8086 | 8086 | Standard |
| Milvus gRPC | 19530 | 19530 | Standard |
| SQLite | N/A | N/A | Embedded, no container |

## Resource Constraints

All benchmark containers are limited to ensure fair comparison:
- **CPU**: 2 cores (`cpus: '2'`)
- **Memory**: 1.5-2GB (`memory: 2G`)

These mirror a modest production single-node setup and prevent one database from starving another during parallel testing.

## Docker Compose Patterns Used

### Standalone (benchmarking)
Single container with tuned settings for performance measurement.
Used in: `docker/<db>/docker-compose.yml`

### Replication (HA lab)
Primary + replica(s) with replication configured.
Used in: `docker/<db>/replication.yml`

### Pooling (connection management lab)
Database + connection pooler (PgBouncer, ProxySQL).
Used in: `docker/<db>/pooling.yml`

### Cluster (scaling lab)
Multi-node cluster with sharding or distributed consensus.
Used in: `docker/redis/cluster.yml`, `docker/mongodb/sharding.yml`

## When Modifying Docker Configs

- Always include a `healthcheck` — benchmarks depend on it
- Pin image versions (no `:latest` for reproducibility, except ClickHouse which tracks stable)
- Use named volumes for data persistence between restarts
- Set `deploy.resources.limits` for fair benchmarking
- Document non-standard ports in comments
- Use `depends_on` with `condition: service_healthy` for multi-container setups

## Troubleshooting

| Issue | Cause | Fix |
|-------|-------|-----|
| Container exits immediately | Config error or port conflict | `docker compose logs <service>` |
| Healthcheck never passes | DB startup slow | Increase `interval` and `retries` |
| "Address already in use" | Port conflict with host or other container | Change host port mapping |
| OOM killed | Exceeds memory limit | Increase `memory` limit or tune DB settings |
| Slow first run | Image not pulled yet | `docker compose pull` first |
