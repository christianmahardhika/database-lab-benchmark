# Part 4: Connection Pooling Lab

Hands-on simulation untuk memahami connection pooling mechanisms dan use cases.

---

## 4.1 Mengapa Connection Pooling Penting?

### Problem: Database Connection Overhead

Setiap database connection memerlukan:
1. **TCP handshake** (3-way: SYN, SYN-ACK, ACK)
2. **TLS handshake** (jika encrypted, 2-4 round trips)
3. **Authentication** (username/password verification)
4. **Session initialization** (allocate memory, set parameters)
5. **Memory allocation** per connection:
   - PostgreSQL: ~10MB per connection
   - MySQL: ~256KB - 1MB per connection

### Analogi: Connection Pool sebagai Resepsionis Hotel

**Tanpa pooling:**
- Setiap tamu harus tunggu kamar disiapkan dari nol
- Check-in 30 menit per tamu
- Hotel bisa overload jika banyak tamu datang bersamaan

**Dengan pooling:**
- Kamar sudah ready, tinggal assign ke tamu
- Check-in 30 detik
- Tamu yang checkout, kamarnya langsung available untuk tamu lain

### Connection Pool Flow

```
[App Instance 1] ─┐
[App Instance 2] ─┼─► [Connection Pool] ─► [Database]
[App Instance 3] ─┘         │
                           │
                    ┌──────┴──────┐
                    │ Pool Config │
                    │ - min: 5    │
                    │ - max: 20   │
                    │ - timeout   │
                    └─────────────┘
```

---

## 4.2 PostgreSQL + PgBouncer

**Source:** [PgBouncer Docs](https://www.pgbouncer.org/config.html)

### PgBouncer Pooling Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| **Session** | 1 client = 1 server connection for entire session | Long-lived connections, prepared statements |
| **Transaction** | Connection returned after each transaction | Most web apps (recommended) |
| **Statement** | Connection returned after each statement | Simple queries only, no transactions |

### Analogi Pooling Modes

- **Session** = Kamar hotel dedicated selama menginap
- **Transaction** = Meja restoran - duduk saat makan, pergi setelah bayar
- **Statement** = Kursi bar - order, minum, pergi, repeat

### Docker Lab Setup

```yaml
# postgres/pooling/docker-compose.yml
```

### Menjalankan Lab

```bash
cd ~/db-operations-lab/postgres/pooling
docker compose up -d
sleep 5

# Create test database and user
docker exec pg-pooling-db psql -U postgres -c "
  CREATE DATABASE appdb;
  CREATE USER appuser WITH ENCRYPTED PASSWORD 'apppass';
  GRANT ALL PRIVILEGES ON DATABASE appdb TO appuser;
"

# Connect through PgBouncer (port 6432)
docker exec pg-pooling-db psql -h pgbouncer -p 6432 -U appuser -d appdb -c "SELECT 1;"
```

### Eksperimen 1: Compare Direct vs Pooled Connections

```bash
# Install pgbench for benchmarking
docker exec pg-pooling-db apk add postgresql-contrib || true

# Benchmark: Direct connection (port 5432)
echo "=== Direct Connection Benchmark ==="
docker exec pg-pooling-db pgbench -h localhost -p 5432 -U postgres -d appdb -i -s 10
docker exec pg-pooling-db pgbench -h localhost -p 5432 -U postgres -d appdb -c 50 -j 4 -t 100

# Benchmark: Through PgBouncer (port 6432)
echo "=== PgBouncer Connection Benchmark ==="
docker exec pg-pooling-db pgbench -h pgbouncer -p 6432 -U appuser -d appdb -c 50 -j 4 -t 100
```

### Eksperimen 2: Monitor Pool Statistics

```bash
# Connect to PgBouncer admin console
docker exec -it pgbouncer psql -h localhost -p 6432 -U pgbouncer pgbouncer

# Di dalam psql:
# SHOW POOLS;       -- Pool statistics
# SHOW CLIENTS;     -- Connected clients
# SHOW SERVERS;     -- Backend connections
# SHOW STATS;       -- Query statistics
# SHOW CONFIG;      -- Current configuration
```

### Eksperimen 3: Connection Limits

```bash
# Simulate connection storm (100 connections)
for i in {1..100}; do
  docker exec pg-pooling-db psql -h pgbouncer -p 6432 -U appuser -d appdb -c "SELECT pg_sleep(0.5);" &
done
wait

# Check how PgBouncer handles it
docker exec pgbouncer psql -h localhost -p 6432 -U pgbouncer pgbouncer -c "SHOW POOLS;"
```

### PgBouncer Configuration Explained

```ini
# pgbouncer.ini

[databases]
# database = host=X port=X dbname=X
appdb = host=pg-pooling-db port=5432 dbname=appdb

[pgbouncer]
# Listen address and port
listen_addr = 0.0.0.0
listen_port = 6432

# Pooling mode: session, transaction, or statement
pool_mode = transaction

# Pool size limits
default_pool_size = 20        # Connections per user/database pair
min_pool_size = 5             # Minimum connections to keep ready
max_client_conn = 1000        # Max client connections to PgBouncer
max_db_connections = 50       # Max connections to database

# Timeouts
server_idle_timeout = 600     # Close idle server connections after 10 min
client_idle_timeout = 0       # 0 = disabled
query_timeout = 0             # 0 = disabled

# Auth
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
```

---

## 4.3 MySQL + ProxySQL

**Source:** [ProxySQL Docs](https://proxysql.com/documentation/)

### ProxySQL Features

- Connection pooling and multiplexing
- Query routing (read/write splitting)
- Query caching
- Query rewriting
- Failover handling

### Docker Lab Setup

```yaml
# mysql/pooling/docker-compose.yml
```

### Menjalankan Lab

```bash
cd ~/db-operations-lab/mysql/pooling
docker compose up -d
sleep 10

# Create test user
docker exec mysql-pooling mysql -uroot -prootpass -e "
  CREATE DATABASE IF NOT EXISTS appdb;
  CREATE USER IF NOT EXISTS 'appuser'@'%' IDENTIFIED BY 'apppass';
  GRANT ALL PRIVILEGES ON appdb.* TO 'appuser'@'%';
  FLUSH PRIVILEGES;
"

# Connect through ProxySQL (port 6033)
docker exec mysql-pooling mysql -h proxysql -P 6033 -uappuser -papppass -e "SELECT 1;"
```

### Eksperimen 1: Read/Write Splitting

```bash
# Configure read/write splitting via ProxySQL admin
docker exec proxysql mysql -h127.0.0.1 -P6032 -uradmin -pradmin --prompt='ProxySQL> ' <<EOF
-- Add MySQL server to hostgroup 0 (writer)
INSERT INTO mysql_servers(hostgroup_id, hostname, port) VALUES (0, 'mysql-pooling', 3306);

-- Add query rules for read/write splitting
INSERT INTO mysql_query_rules (rule_id, active, match_pattern, destination_hostgroup, apply)
VALUES (1, 1, '^SELECT.*FOR UPDATE', 0, 1);  -- SELECT FOR UPDATE goes to writer

INSERT INTO mysql_query_rules (rule_id, active, match_pattern, destination_hostgroup, apply)
VALUES (2, 1, '^SELECT', 1, 1);  -- Regular SELECT goes to reader hostgroup

-- Add user
INSERT INTO mysql_users(username, password, default_hostgroup) VALUES ('appuser', 'apppass', 0);

-- Apply changes
LOAD MYSQL SERVERS TO RUNTIME;
LOAD MYSQL QUERY RULES TO RUNTIME;
LOAD MYSQL USERS TO RUNTIME;
SAVE MYSQL SERVERS TO DISK;
SAVE MYSQL QUERY RULES TO DISK;
SAVE MYSQL USERS TO DISK;
EOF
```

### Eksperimen 2: Monitor Connection Pool

```bash
# Connect to ProxySQL admin
docker exec -it proxysql mysql -h127.0.0.1 -P6032 -uradmin -pradmin

# Di dalam mysql:
# SELECT * FROM stats_mysql_connection_pool;
# SELECT * FROM stats_mysql_query_digest ORDER BY sum_time DESC LIMIT 10;
# SELECT * FROM stats_mysql_global;
```

---

## 4.4 Connection Pooling Strategies & Use Cases

### Strategy 1: Web Application (Short-lived Requests)

**Scenario:** 100 app instances, each handling 50 concurrent requests

**Without pooling:**
- 100 × 50 = 5000 database connections
- PostgreSQL default max_connections = 100 → FAIL

**With PgBouncer (transaction mode):**
- PgBouncer: max_client_conn = 5000
- Backend: default_pool_size = 20
- Actual DB connections: ~20-50
- ✅ Works perfectly

```ini
# pgbouncer.ini for web apps
pool_mode = transaction
default_pool_size = 20
max_client_conn = 5000
```

### Strategy 2: Background Workers (Long-running Jobs)

**Scenario:** Workers that hold connections for minutes (batch processing)

**Recommendation:** Session pooling atau smaller pool dengan longer timeout

```ini
# pgbouncer.ini for workers
pool_mode = session
default_pool_size = 10
server_idle_timeout = 3600  # 1 hour
```

### Strategy 3: Microservices Architecture

**Scenario:** 20 microservices, each needs database access

**Pattern:** Each service has its own pool, central pool aggregates

```
[Service A] → [Local Pool 5] ─┐
[Service B] → [Local Pool 5] ─┼─► [PgBouncer 50] → [PostgreSQL 100]
[Service C] → [Local Pool 5] ─┘
```

### Strategy 4: Multi-tenant Application

**Scenario:** Each tenant has separate database

```ini
# pgbouncer.ini multi-tenant
[databases]
tenant_001 = host=db1.example.com dbname=tenant_001
tenant_002 = host=db1.example.com dbname=tenant_002
tenant_003 = host=db2.example.com dbname=tenant_003
* = host=db1.example.com  # Wildcard for dynamic tenants
```

### Strategy 5: High-Availability with Failover

```ini
# pgbouncer.ini with failover
[databases]
# Primary with fallback
mydb = host=primary.db.local,replica.db.local port=5432 dbname=mydb
```

---

## 4.5 Connection Pool Sizing Formula

### PostgreSQL Pool Sizing

```
Optimal Pool Size = (core_count * 2) + effective_spindle_count

Untuk SSD:
Optimal Pool Size ≈ (CPU cores * 2) + 1
```

**Example:** 8-core server with SSD
- Pool size = (8 × 2) + 1 = **17 connections**

### Application-side Pool Sizing

```
App Pool Size = (Database Pool Size) / (Number of App Instances)

Total connections = App instances × App pool size × Safety factor
```

**Example:** 
- Database: max 50 connections
- App instances: 5
- Per-app pool: 50 / 5 = 10 connections each

---

## 4.6 Troubleshooting Connection Issues

### Problem 1: "Too many connections"

```bash
# Check current connections
SELECT count(*) FROM pg_stat_activity;

# Check who's holding connections
SELECT usename, application_name, state, count(*) 
FROM pg_stat_activity 
GROUP BY usename, application_name, state;

# Kill idle connections
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle' 
AND query_start < NOW() - INTERVAL '10 minutes';
```

### Problem 2: "Connection timeout"

```ini
# Increase timeouts
server_connect_timeout = 30
server_login_retry = 3
```

### Problem 3: "Prepared statement does not exist"

Transaction mode doesn't support prepared statements across transactions.

**Solution:**
1. Use session mode, atau
2. Disable prepared statements di app:
   ```python
   # SQLAlchemy
   engine = create_engine(url, pool_pre_ping=True, 
                          connect_args={"prepare_threshold": 0})
   ```

---

## Summary: Pooling Comparison

| Aspect | PgBouncer | ProxySQL | pgpool-II |
|--------|-----------|----------|-----------|
| Database | PostgreSQL | MySQL | PostgreSQL |
| Memory | Very low (~2KB/conn) | Low | Higher |
| Modes | Session/Transaction/Statement | Connection multiplexing | Session |
| Query Caching | ❌ | ✅ | ✅ |
| Load Balancing | ❌ | ✅ | ✅ |
| Read/Write Split | ❌ | ✅ | ✅ |
| Query Routing | ❌ | ✅ | ✅ |
| Failover | Basic | Advanced | Advanced |

---

## Cleanup

```bash
cd ~/db-operations-lab/postgres/pooling && docker compose down -v
cd ~/db-operations-lab/mysql/pooling && docker compose down -v
```
