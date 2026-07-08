# Part 3: Backup & Disaster Recovery Lab

Hands-on simulation untuk backup strategies, data retention, dan disaster recovery.

---

## 3.1 PostgreSQL Backup & PITR

**Source:** [PostgreSQL Docs - Backup and Restore](https://www.postgresql.org/docs/current/backup.html)

### Jenis Backup PostgreSQL

| Type | Tool | Use Case | Cons |
|------|------|----------|------|
| Logical | pg_dump | Small DB, cross-version migration | Slow untuk large DB |
| Physical | pg_basebackup | Large DB, exact copy | Same version only |
| Continuous | WAL Archiving + PITR | Point-in-time recovery | Complex setup |

### Analogi: Backup sebagai Sistem Asuransi

- **pg_dump** = Foto semua dokumen (logical copy)
- **pg_basebackup** = Kunci brankas dan buat duplikat (physical copy)
- **WAL Archiving** = Rekaman CCTV 24 jam (continuous changes)
- **PITR** = "Kembalikan ke keadaan jam 14:35 kemarin"

### Docker Lab Setup

```yaml
# postgres/backup/docker-compose.yml
```

### Eksperimen 1: Logical Backup dengan pg_dump

```bash
cd ~/db-operations-lab/postgres/backup
docker compose up -d
sleep 5

# Create sample data
docker exec pg-backup psql -U postgres -d testdb -c "
  CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    department TEXT,
    salary NUMERIC(10,2),
    created_at TIMESTAMP DEFAULT NOW()
  );
  INSERT INTO employees (name, department, salary) VALUES
    ('Alice', 'Engineering', 15000000),
    ('Bob', 'Marketing', 12000000),
    ('Charlie', 'Engineering', 14000000),
    ('Diana', 'HR', 11000000);
"

# Backup ke file SQL (plain text)
docker exec pg-backup pg_dump -U postgres -d testdb -F plain -f /backups/testdb_plain.sql

# Backup ke custom format (compressed, parallel restore)
docker exec pg-backup pg_dump -U postgres -d testdb -F custom -f /backups/testdb_custom.dump

# Backup schema only
docker exec pg-backup pg_dump -U postgres -d testdb --schema-only -f /backups/testdb_schema.sql

# Backup data only
docker exec pg-backup pg_dump -U postgres -d testdb --data-only -f /backups/testdb_data.sql

# List backup files
docker exec pg-backup ls -la /backups/
```

### Eksperimen 2: Restore dari Backup

```bash
# Create new database untuk restore test
docker exec pg-backup psql -U postgres -c "CREATE DATABASE testdb_restored;"

# Restore dari custom format
docker exec pg-backup pg_restore -U postgres -d testdb_restored /backups/testdb_custom.dump

# Verify data
docker exec pg-backup psql -U postgres -d testdb_restored -c "SELECT * FROM employees;"

# Atau restore dari plain SQL
docker exec pg-backup psql -U postgres -c "CREATE DATABASE testdb_restored2;"
docker exec pg-backup psql -U postgres -d testdb_restored2 -f /backups/testdb_plain.sql
```

### Eksperimen 3: Physical Backup dengan pg_basebackup

```bash
# Take base backup (includes WAL)
docker exec pg-backup pg_basebackup \
  -U postgres \
  -D /backups/basebackup \
  -Fp -Xs -P -v

# Check backup contents
docker exec pg-backup ls -la /backups/basebackup/
```

### Eksperimen 4: Point-in-Time Recovery (PITR)

```bash
# Step 1: Note current time
docker exec pg-backup psql -U postgres -c "SELECT NOW();"
# Output: 2026-06-04 10:00:00

# Step 2: Insert some data
docker exec pg-backup psql -U postgres -d testdb -c "
  INSERT INTO employees (name, department, salary) VALUES ('Eve', 'Finance', 13000000);
"

# Step 3: Note time before "disaster"
docker exec pg-backup psql -U postgres -c "SELECT NOW();"
# Output: 2026-06-04 10:05:00

# Step 4: Simulate disaster - delete all data!
docker exec pg-backup psql -U postgres -d testdb -c "DELETE FROM employees;"
docker exec pg-backup psql -U postgres -d testdb -c "SELECT * FROM employees;"
# (empty)

# Step 5: Force WAL switch to archive recent changes
docker exec pg-backup psql -U postgres -c "SELECT pg_switch_wal();"

# Step 6: Stop database
docker compose stop pg-backup

# Step 7: Restore from base backup + WAL to specific point in time
# (In production, you would restore to a recovery target time)
# For this lab, we demonstrate the concept

# Step 8: Create recovery.signal and set recovery target
# docker exec pg-backup bash -c "
#   touch /var/lib/postgresql/data/recovery.signal
#   echo \"recovery_target_time = '2026-06-04 10:05:00'\" >> /var/lib/postgresql/data/postgresql.auto.conf
#   echo \"restore_command = 'cp /backups/wal/%f %p'\" >> /var/lib/postgresql/data/postgresql.auto.conf
# "

# Step 9: Start and verify recovery
docker compose start pg-backup
```

---

## 3.2 MySQL Backup & Recovery

**Source:** [MySQL Docs - Backup and Recovery](https://dev.mysql.com/doc/refman/8.0/en/backup-and-recovery.html)

### Jenis Backup MySQL

| Tool | Type | Locking | Use Case |
|------|------|---------|----------|
| mysqldump | Logical | Table lock (InnoDB: --single-transaction) | Small-medium DB |
| mysqlpump | Logical (parallel) | Minimal | Medium DB |
| MySQL Enterprise Backup | Physical | Hot backup | Enterprise |
| Percona XtraBackup | Physical | Hot backup | Large DB, free |

### Docker Lab Setup

```yaml
# mysql/backup/docker-compose.yml
```

### Eksperimen 1: mysqldump Backup

```bash
cd ~/db-operations-lab/mysql/backup
docker compose up -d
sleep 10

# Create sample data
docker exec mysql-backup mysql -uroot -prootpass -e "
  CREATE DATABASE IF NOT EXISTS company;
  USE company;
  CREATE TABLE employees (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    department VARCHAR(50),
    salary DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  INSERT INTO employees (name, department, salary) VALUES
    ('Alice', 'Engineering', 15000000),
    ('Bob', 'Marketing', 12000000),
    ('Charlie', 'Engineering', 14000000);
"

# Full backup with single transaction (no locks for InnoDB)
docker exec mysql-backup mysqldump \
  -uroot -prootpass \
  --single-transaction \
  --routines \
  --triggers \
  --events \
  company > /tmp/company_backup.sql

# Copy backup from container
docker cp mysql-backup:/tmp/company_backup.sql ./backups/

# Backup all databases
docker exec mysql-backup mysqldump \
  -uroot -prootpass \
  --single-transaction \
  --all-databases > /tmp/all_databases.sql

docker cp mysql-backup:/tmp/all_databases.sql ./backups/
```

### Eksperimen 2: Restore MySQL Backup

```bash
# Create new database untuk restore test
docker exec mysql-backup mysql -uroot -prootpass -e "CREATE DATABASE company_restored;"

# Restore
docker cp ./backups/company_backup.sql mysql-backup:/tmp/
docker exec mysql-backup bash -c "mysql -uroot -prootpass company_restored < /tmp/company_backup.sql"

# Verify
docker exec mysql-backup mysql -uroot -prootpass -e "SELECT * FROM company_restored.employees;"
```

### Eksperimen 3: Binary Log untuk Point-in-Time Recovery

```bash
# Check binary log status
docker exec mysql-backup mysql -uroot -prootpass -e "SHOW MASTER STATUS;"

# List binary logs
docker exec mysql-backup mysql -uroot -prootpass -e "SHOW BINARY LOGS;"

# Show binary log events (last 20)
docker exec mysql-backup mysql -uroot -prootpass -e "
  SHOW BINLOG EVENTS IN 'mysql-bin.000001' LIMIT 20;
"

# Decode binary log to SQL
docker exec mysql-backup mysqlbinlog /var/lib/mysql/mysql-bin.000001 | head -100
```

---

## 3.3 MongoDB Backup & Recovery

**Source:** [MongoDB Docs - Backup Methods](https://www.mongodb.com/docs/manual/core/backups/)

### Jenis Backup MongoDB

| Method | Type | Impact | Use Case |
|--------|------|--------|----------|
| mongodump | Logical | Can impact performance | Small-medium |
| Filesystem Snapshot | Physical | Requires journaling | Large, fast |
| MongoDB Atlas | Managed | Zero impact | Cloud deployments |
| Ops Manager | Enterprise | Continuous | On-prem enterprise |

### Docker Lab Setup

```yaml
# mongodb/backup/docker-compose.yml
```

### Eksperimen 1: mongodump Backup

```bash
cd ~/db-operations-lab/mongodb/backup
docker compose up -d
sleep 5

# Create sample data
docker exec mongo-backup mongosh --eval '
  use company;
  db.employees.insertMany([
    { name: "Alice", department: "Engineering", salary: 15000000, createdAt: new Date() },
    { name: "Bob", department: "Marketing", salary: 12000000, createdAt: new Date() },
    { name: "Charlie", department: "Engineering", salary: 14000000, createdAt: new Date() }
  ]);
  db.employees.find();
'

# Full backup (all databases)
docker exec mongo-backup mongodump --out=/backups/full_backup

# Backup specific database
docker exec mongo-backup mongodump --db=company --out=/backups/company_backup

# Backup specific collection
docker exec mongo-backup mongodump --db=company --collection=employees --out=/backups/employees_backup

# List backups
docker exec mongo-backup ls -la /backups/
docker exec mongo-backup ls -la /backups/company_backup/company/
```

### Eksperimen 2: mongorestore

```bash
# Simulate disaster - drop collection
docker exec mongo-backup mongosh --eval '
  use company;
  db.employees.drop();
  db.employees.find();
'

# Restore from backup
docker exec mongo-backup mongorestore --db=company /backups/company_backup/company/

# Verify restoration
docker exec mongo-backup mongosh --eval '
  use company;
  db.employees.find();
'

# Restore to different database
docker exec mongo-backup mongorestore --db=company_restored /backups/company_backup/company/
docker exec mongo-backup mongosh --eval '
  use company_restored;
  db.employees.find();
'
```

### Eksperimen 3: Oplog Backup (untuk PITR)

```bash
# Untuk replica set, backup oplog
# docker exec mongo-backup mongodump --oplog --out=/backups/oplog_backup

# Restore dengan oplog replay
# docker exec mongo-backup mongorestore --oplogReplay /backups/oplog_backup
```

---

## 3.4 Redis Backup & Recovery

**Source:** [Redis Docs - Persistence](https://redis.io/docs/management/persistence/)

### Persistence Options

| Method | Description | Data Loss Risk | Performance |
|--------|-------------|----------------|-------------|
| RDB | Point-in-time snapshots | High (since last snapshot) | Best |
| AOF | Append-only log | Low (configurable) | Good |
| RDB + AOF | Hybrid | Lowest | Balanced |
| No persistence | Memory only | Total | Best |

### Analogi: RDB vs AOF

- **RDB** = Foto berkala (setiap jam, setiap 1000 writes)
  - Pro: Compact, fast recovery
  - Con: Kehilangan data sejak foto terakhir

- **AOF** = Rekaman video semua aktivitas
  - Pro: Minimal data loss
  - Con: File besar, replay lebih lambat

### Docker Lab Setup

```yaml
# redis/backup/docker-compose.yml
```

### Eksperimen 1: RDB Snapshot

```bash
cd ~/db-operations-lab/redis/backup
docker compose up -d
sleep 3

# Insert data
docker exec redis-backup redis-cli SET user:1 '{"name":"Alice","email":"alice@example.com"}'
docker exec redis-backup redis-cli SET user:2 '{"name":"Bob","email":"bob@example.com"}'
docker exec redis-backup redis-cli HSET product:1 name "Laptop" price 15000000
docker exec redis-backup redis-cli LPUSH orders "order-001" "order-002" "order-003"

# Check last save time
docker exec redis-backup redis-cli LASTSAVE

# Trigger manual RDB snapshot (background)
docker exec redis-backup redis-cli BGSAVE

# Wait and check
sleep 2
docker exec redis-backup redis-cli LASTSAVE

# Check RDB file
docker exec redis-backup ls -la /data/
```

### Eksperimen 2: AOF Persistence

```bash
# Check AOF status
docker exec redis-backup redis-cli CONFIG GET appendonly
docker exec redis-backup redis-cli CONFIG GET appendfsync

# View AOF file content
docker exec redis-backup cat /data/appendonly.aof | head -50

# Trigger AOF rewrite (compaction)
docker exec redis-backup redis-cli BGREWRITEAOF

sleep 2
docker exec redis-backup ls -la /data/
```

### Eksperimen 3: Recovery dari RDB

```bash
# Copy RDB file for backup
docker cp redis-backup:/data/dump.rdb ./backups/

# Simulate disaster - flush all
docker exec redis-backup redis-cli FLUSHALL
docker exec redis-backup redis-cli KEYS "*"
# (empty)

# Stop redis
docker compose stop redis-backup

# Restore RDB file
docker cp ./backups/dump.rdb redis-backup:/data/dump.rdb

# Start redis - will load from RDB
docker compose start redis-backup
sleep 2

# Verify data restored
docker exec redis-backup redis-cli KEYS "*"
docker exec redis-backup redis-cli GET user:1
```

---

## 3.5 Disaster Recovery Scenarios

### Scenario 1: Single Server Failure

**Strategy:** Restore from latest backup

```bash
# PostgreSQL
pg_restore -d newdb /backups/latest.dump

# MySQL
mysql newdb < /backups/latest.sql

# MongoDB
mongorestore --db=newdb /backups/latest/

# Redis
cp /backups/dump.rdb /var/lib/redis/
redis-server --dir /var/lib/redis
```

### Scenario 2: Data Corruption (need PITR)

**Strategy:** Restore to specific point in time before corruption

```bash
# PostgreSQL - restore to specific timestamp
# Set recovery_target_time in postgresql.conf
# recovery_target_time = '2026-06-04 14:30:00'

# MySQL - replay binlog to specific position
mysqlbinlog --stop-datetime="2026-06-04 14:30:00" binlog.000001 | mysql
```

### Scenario 3: Complete Cluster Failure

**Strategy:** 
1. Restore primary from backup
2. Rebuild replicas from restored primary
3. Reconfigure monitoring/sentinel/routing

### Backup Retention Policy Example

```yaml
# Grandfather-Father-Son (GFS) rotation
retention:
  hourly: 24      # Keep 24 hourly backups
  daily: 7        # Keep 7 daily backups
  weekly: 4       # Keep 4 weekly backups
  monthly: 12     # Keep 12 monthly backups
  yearly: 3       # Keep 3 yearly backups
```

---

## Summary: Backup Comparison

| Aspect | PostgreSQL | MySQL | MongoDB | Redis |
|--------|------------|-------|---------|-------|
| Logical Backup | pg_dump | mysqldump | mongodump | N/A |
| Physical Backup | pg_basebackup | XtraBackup | Filesystem snapshot | RDB |
| Continuous | WAL archiving | Binary log | Oplog | AOF |
| PITR | ✅ Native | ✅ Via binlog | ✅ Via oplog | ❌ |
| Hot Backup | ✅ pg_basebackup | ✅ XtraBackup | ✅ mongodump | ✅ BGSAVE |

---

## Cleanup

```bash
cd ~/db-operations-lab/postgres/backup && docker compose down -v
cd ~/db-operations-lab/mysql/backup && docker compose down -v
cd ~/db-operations-lab/mongodb/backup && docker compose down -v
cd ~/db-operations-lab/redis/backup && docker compose down -v
```
