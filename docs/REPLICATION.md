# Part 2: Replication & Scaling Lab

Hands-on simulation untuk memahami replication mechanisms di berbagai database.

---

## 2.1 PostgreSQL Streaming Replication

**Source:** [PostgreSQL Docs - Streaming Replication](https://www.postgresql.org/docs/current/warm-standby.html#STREAMING-REPLICATION)

### Bagaimana Streaming Replication Bekerja

#### Analogi: Juru Tulis dan Asisten

Bayangkan:
- **Primary** = Juru tulis utama yang mencatat semua transaksi
- **WAL (Write-Ahead Log)** = Buku jurnal kronologis semua perubahan
- **Replica** = Asisten yang menyalin dari jurnal
- **WAL Sender** = Kurir yang mengirim halaman jurnal
- **WAL Receiver** = Penerima di sisi asisten

#### Flow Replication

```
[Client] → [Primary]
              │
              ├── 1. Write to WAL (jurnal)
              ├── 2. Apply to heap (data actual)
              └── 3. WAL Sender streams to replica
                        │
                        ▼
                  [Replica]
                      │
                      ├── 4. WAL Receiver terima
                      ├── 5. Write to local WAL
                      └── 6. Recovery process apply
```

### Docker Lab Setup

```yaml
# postgres/replication/docker-compose.yml
```

### Menjalankan Lab

```bash
cd ~/db-operations-lab/postgres/replication
docker compose up -d

# Tunggu 10 detik untuk startup
sleep 10

# Cek replication status
docker exec pg-primary psql -U postgres -c "SELECT * FROM pg_stat_replication;"
```

### Eksperimen 1: Write di Primary, Read di Replica

```bash
# Insert di primary
docker exec pg-primary psql -U postgres -d testdb -c "
  CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    product TEXT,
    created_at TIMESTAMP DEFAULT NOW()
  );
  INSERT INTO orders (product) VALUES ('Laptop'), ('Mouse'), ('Keyboard');
"

# Baca di replica (harus muncul!)
docker exec pg-replica psql -U postgres -d testdb -c "SELECT * FROM orders;"
```

### Eksperimen 2: Measure Replication Lag

```bash
# Di primary, cek lag
docker exec pg-primary psql -U postgres -c "
  SELECT 
    client_addr,
    state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    pg_wal_lsn_diff(sent_lsn, replay_lsn) AS lag_bytes
  FROM pg_stat_replication;
"
```

### Eksperimen 3: Failover Manual

```bash
# Promote replica menjadi primary
docker exec pg-replica psql -U postgres -c "SELECT pg_promote();"

# Sekarang replica bisa menerima writes
docker exec pg-replica psql -U postgres -d testdb -c "
  INSERT INTO orders (product) VALUES ('Monitor');
  SELECT * FROM orders;
"
```

---

## 2.2 MySQL Binary Log Replication

**Source:** [MySQL Docs - Replication](https://dev.mysql.com/doc/refman/8.0/en/replication.html)

### Bagaimana Binary Log Replication Bekerja

#### Analogi: Kantor Pusat dan Cabang

- **Source (Master)** = Kantor pusat yang membuat kebijakan
- **Binary Log** = Memo resmi semua perubahan kebijakan
- **Replica (Slave)** = Kantor cabang yang harus ikut kebijakan
- **IO Thread** = Kurir yang mengambil memo dari pusat
- **SQL Thread** = Petugas di cabang yang menjalankan memo

#### Replication Formats

| Format | Deskripsi | Pro | Con |
|--------|-----------|-----|-----|
| STATEMENT | Log SQL statements | Compact | Non-deterministic functions bisa beda hasil |
| ROW | Log actual row changes | Deterministic | Lebih besar |
| MIXED | Auto switch | Best of both | Complex |

### Docker Lab Setup

```yaml
# mysql/replication/docker-compose.yml
```

### Menjalankan Lab

```bash
cd ~/db-operations-lab/mysql/replication
docker compose up -d

# Tunggu startup
sleep 15

# Cek replication status di replica
docker exec mysql-replica mysql -uroot -prootpass -e "SHOW REPLICA STATUS\G"
```

### Eksperimen: GTID-based Replication

```bash
# Insert di source
docker exec mysql-source mysql -uroot -prootpass -e "
  CREATE DATABASE IF NOT EXISTS shop;
  USE shop;
  CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100),
    price DECIMAL(10,2)
  );
  INSERT INTO products (name, price) VALUES ('Laptop', 15000000), ('Mouse', 250000);
"

# Verify di replica
docker exec mysql-replica mysql -uroot -prootpass -e "SELECT * FROM shop.products;"

# Cek GTID executed
docker exec mysql-replica mysql -uroot -prootpass -e "
  SELECT @@GLOBAL.gtid_executed;
  SHOW REPLICA STATUS\G
"
```

---

## 2.3 MongoDB Replica Set

**Source:** [MongoDB Docs - Replication](https://www.mongodb.com/docs/manual/replication/)

### Bagaimana Replica Set Bekerja

#### Analogi: Komite dengan Voting

- **Primary** = Ketua komite yang bisa menulis keputusan
- **Secondary** = Anggota yang menyalin keputusan ketua
- **Oplog** = Notulensi rapat (operation log)
- **Election** = Voting untuk pilih ketua baru jika ketua lama down
- **Arbiter** = Penengah yang hanya voting, tidak menyimpan data

#### Write Concern & Read Preference

```
Write Concern:
  w: 1        → Ack setelah primary write
  w: majority → Ack setelah majority nodes write
  w: all      → Ack setelah semua nodes write

Read Preference:
  primary           → Selalu baca dari primary
  primaryPreferred  → Primary, fallback ke secondary
  secondary         → Selalu baca dari secondary
  secondaryPreferred→ Secondary, fallback ke primary
  nearest           → Node dengan latency terendah
```

### Docker Lab Setup

```yaml
# mongodb/replication/docker-compose.yml
```

### Menjalankan Lab

```bash
cd ~/db-operations-lab/mongodb/replication
docker compose up -d

# Tunggu startup
sleep 10

# Initialize replica set
docker exec mongo1 mongosh --eval '
  rs.initiate({
    _id: "rs0",
    members: [
      { _id: 0, host: "mongo1:27017", priority: 2 },
      { _id: 1, host: "mongo2:27017", priority: 1 },
      { _id: 2, host: "mongo3:27017", priority: 1 }
    ]
  })
'

# Tunggu election selesai
sleep 5

# Cek status
docker exec mongo1 mongosh --eval "rs.status()"
```

### Eksperimen 1: Write dan Read

```bash
# Insert di primary
docker exec mongo1 mongosh --eval '
  use shop;
  db.products.insertMany([
    { name: "Laptop", price: 15000000 },
    { name: "Mouse", price: 250000 },
    { name: "Keyboard", price: 500000 }
  ]);
  db.products.find();
'

# Read dari secondary (harus set read preference)
docker exec mongo2 mongosh --eval '
  db.getMongo().setReadPref("secondary");
  use shop;
  db.products.find();
'
```

### Eksperimen 2: Simulate Failover

```bash
# Stop primary
docker stop mongo1

# Tunggu election (5-10 detik)
sleep 10

# Cek siapa primary baru
docker exec mongo2 mongosh --eval "rs.status().members.filter(m => m.stateStr === 'PRIMARY')"

# Insert ke primary baru
docker exec mongo2 mongosh --eval '
  use shop;
  db.products.insertOne({ name: "Monitor", price: 3000000 });
  db.products.find();
'

# Bring back mongo1
docker start mongo1
sleep 5

# mongo1 sekarang jadi secondary
docker exec mongo1 mongosh --eval "rs.status().members.filter(m => m.name.includes('mongo1'))"
```

---

## 2.4 Redis Sentinel (High Availability)

**Source:** [Redis Docs - Sentinel](https://redis.io/docs/management/sentinel/)

### Bagaimana Redis Sentinel Bekerja

#### Analogi: Sistem Monitoring dengan Auto-Failover

- **Master** = Server utama yang handle writes
- **Replica** = Server backup yang sync dari master
- **Sentinel** = Monitor 24/7 yang bisa trigger failover
- **Quorum** = Jumlah sentinel yang harus setuju sebelum failover

#### Sentinel Functions

1. **Monitoring** — Cek apakah master/replica hidup
2. **Notification** — Alert via Pub/Sub atau script
3. **Automatic Failover** — Promote replica jika master down
4. **Configuration Provider** — Client tanya sentinel untuk master address

### Docker Lab Setup

```yaml
# redis/replication/docker-compose.yml
```

### Menjalankan Lab

```bash
cd ~/db-operations-lab/redis/replication
docker compose up -d

# Tunggu startup
sleep 5

# Cek replication info di master
docker exec redis-master redis-cli INFO replication

# Cek sentinel status
docker exec redis-sentinel1 redis-cli -p 26379 SENTINEL master mymaster
```

### Eksperimen 1: Write/Read Flow

```bash
# Write ke master
docker exec redis-master redis-cli SET user:1 '{"name":"John","email":"john@example.com"}'

# Read dari replica
docker exec redis-replica1 redis-cli GET user:1

# Replica adalah read-only
docker exec redis-replica1 redis-cli SET test "value"  # Error: READONLY
```

### Eksperimen 2: Automatic Failover

```bash
# Cek current master
docker exec redis-sentinel1 redis-cli -p 26379 SENTINEL get-master-addr-by-name mymaster

# Kill master
docker stop redis-master

# Tunggu failover (default 30 detik)
sleep 35

# Cek master baru
docker exec redis-sentinel1 redis-cli -p 26379 SENTINEL get-master-addr-by-name mymaster

# Salah satu replica sekarang jadi master
docker exec redis-replica1 redis-cli INFO replication | grep role

# Write ke master baru
docker exec redis-replica1 redis-cli SET user:2 '{"name":"Jane"}'

# Bring back old master (sekarang jadi replica)
docker start redis-master
sleep 5
docker exec redis-master redis-cli INFO replication | grep role  # role:slave
```

---

## Summary: Replication Comparison

| Aspect | PostgreSQL | MySQL | MongoDB | Redis |
|--------|------------|-------|---------|-------|
| Model | Primary-Standby | Source-Replica | Replica Set | Master-Replica |
| Log Type | WAL | Binary Log | Oplog | RDB/AOF + Replication Stream |
| Failover | Manual / Patroni | Orchestrator | Automatic (Election) | Sentinel (Automatic) |
| Multi-Primary | No (use BDR) | Group Replication | No (Sharded Cluster) | No |
| Sync Mode | Sync/Async | Async/Semi-sync | Write Concern | Async (WAIT for sync) |
| Lag Monitoring | pg_stat_replication | SHOW REPLICA STATUS | rs.printReplicationInfo() | INFO replication |

---

## Cleanup

```bash
# Stop all lab containers
cd ~/db-operations-lab/postgres/replication && docker compose down -v
cd ~/db-operations-lab/mysql/replication && docker compose down -v
cd ~/db-operations-lab/mongodb/replication && docker compose down -v
cd ~/db-operations-lab/redis/replication && docker compose down -v
```
