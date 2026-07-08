# Database Operations Lab: SQL vs NoSQL Deep Dive

Comprehensive hands-on lab untuk memahami cara kerja internal database operations berdasarkan official documentation.

## Lab Structure

```
db-operations-lab/
├── postgres/
│   ├── replication/      # Streaming & Logical Replication
│   ├── backup/           # pg_dump, pg_basebackup, PITR
│   └── pooling/          # PgBouncer
├── mysql/
│   ├── replication/      # Binary Log Replication, GTID
│   ├── backup/           # mysqldump, XtraBackup
│   └── pooling/          # ProxySQL
├── mongodb/
│   ├── replication/      # Replica Set
│   ├── backup/           # mongodump, oplog
│   └── pooling/          # Driver-level
├── cassandra/
│   ├── replication/      # Multi-node cluster
│   └── backup/           # nodetool snapshot
└── redis/
    ├── replication/      # Master-Replica, Sentinel
    └── backup/           # RDB, AOF
```

## Prerequisites

- Docker & Docker Compose
- 8GB+ RAM (untuk run multiple containers)
- Basic SQL/NoSQL knowledge

---

# Part 1: DDL Internals — What Actually Happens Inside the Machine

## 1.1 PostgreSQL ALTER TABLE Internals

**Source:** [PostgreSQL Docs - ALTER TABLE](https://www.postgresql.org/docs/current/sql-altertable.html), [System Catalogs](https://www.postgresql.org/docs/current/catalogs.html)

### Analogi: Database sebagai Perpustakaan

Bayangkan database PostgreSQL sebagai **perpustakaan raksasa**:
- **Table** = Rak buku dengan format tertentu
- **Row** = Buku individual
- **Column** = Informasi di setiap buku (judul, penulis, ISBN)
- **System Catalog** = Kartu katalog yang mencatat format setiap rak
- **MVCC (Multi-Version Concurrency Control)** = Sistem di mana pembaca bisa tetap baca buku lama sementara petugas menulis versi baru

### Apa yang Terjadi Saat ALTER TABLE?

#### Case 1: ADD COLUMN tanpa DEFAULT (Instant)

```sql
ALTER TABLE users ADD COLUMN phone TEXT;
```

**Yang terjadi di mesin:**
1. PostgreSQL HANYA mengupdate **pg_attribute** (system catalog)
2. Menambahkan entry baru: "table users punya column phone, type TEXT"
3. **TIDAK menyentuh data files sama sekali**
4. Existing rows akan return NULL untuk column baru

**Perumpamaan:** Seperti menambahkan field baru di kartu katalog perpustakaan. Buku-buku lama tidak perlu diubah — kalau ada yang tanya "telepon penulis?", jawabannya "tidak ada data" (NULL).

**Lock:** ACCESS EXCLUSIVE (sangat singkat, < 1ms untuk metadata update)

#### Case 2: ADD COLUMN dengan DEFAULT (PG 11+, Instant)

```sql
ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active';
```

**Yang terjadi di mesin:**
1. Update **pg_attribute** dengan column definition
2. Update **pg_attrdef** dengan default value
3. PostgreSQL 11+ menggunakan **"missing value" optimization**
4. Default value disimpan di catalog, BUKAN di setiap row
5. Saat SELECT, PostgreSQL menggabungkan: data dari disk + default dari catalog

**Perumpamaan:** Seperti membuat aturan baru di perpustakaan: "Semua buku yang tidak punya stiker warna, anggap saja stiker HIJAU". Tidak perlu menempelkan stiker ke semua buku lama — cukup ingat aturannya.

#### Case 3: ALTER COLUMN TYPE (Table Rewrite!)

```sql
ALTER TABLE users ALTER COLUMN age TYPE BIGINT;
```

**Yang terjadi di mesin:**
1. **Acquire ACCESS EXCLUSIVE lock** — semua query harus tunggu
2. Create new heap file (data file baru)
3. **Scan seluruh table**, convert setiap row ke format baru
4. Rebuild semua indexes yang involve column tersebut
5. Update pg_class untuk point ke file baru
6. Drop old heap file

**Perumpamaan:** Seperti memindahkan semua buku dari rak kayu ke rak besi. Harus tutup perpustakaan (lock), pindahkan satu per satu (scan), dan baru buka lagi setelah selesai.

**Durasi:** Proporsional dengan ukuran table. 1 juta rows ≈ detik, 1 miliar rows ≈ jam.

### PostgreSQL Lock Levels

| Lock Level | Blocks | Use Case |
|------------|--------|----------|
| ACCESS SHARE | Nothing | SELECT |
| ROW SHARE | Nothing significant | SELECT FOR UPDATE |
| ROW EXCLUSIVE | Nothing significant | INSERT, UPDATE, DELETE |
| SHARE UPDATE EXCLUSIVE | Other DDL | CREATE INDEX CONCURRENTLY |
| SHARE | Writes | CREATE INDEX (non-concurrent) |
| SHARE ROW EXCLUSIVE | Writes + some DDL | CREATE TRIGGER |
| EXCLUSIVE | Everything except SELECT | Rare |
| ACCESS EXCLUSIVE | **Everything** | ALTER TABLE (most forms) |

### ALTER TABLE Specific Lock Levels (from Official Docs)

| Operation | Lock Level | Notes |
|-----------|------------|-------|
| Most ALTER TABLE | ACCESS EXCLUSIVE | Default for DDL |
| ADD FOREIGN KEY | SHARE ROW EXCLUSIVE | Lighter than full DDL |
| VALIDATE CONSTRAINT | SHARE UPDATE EXCLUSIVE | **Allows concurrent DML!** |
| SET STATISTICS | SHARE UPDATE EXCLUSIVE | Metadata only |
| ATTACH PARTITION (parent) | SHARE UPDATE EXCLUSIVE | Plus ACCESS EXCLUSIVE on attached table |
| DISABLE/ENABLE TRIGGER | SHARE ROW EXCLUSIVE | |

**Key Pattern - NOT VALID untuk Minimal Downtime:**
```sql
-- Step 1: Add constraint tanpa scan (instant, tapi constraint belum enforced)
ALTER TABLE distributors ADD CONSTRAINT distfk 
    FOREIGN KEY (address) REFERENCES addresses (address) NOT VALID;

-- Step 2: Validate dengan lock ringan (SHARE UPDATE EXCLUSIVE, allows DML)
ALTER TABLE distributors VALIDATE CONSTRAINT distfk;
```

---

## 1.2 MySQL InnoDB Online DDL Internals

**Source:** [MySQL Docs - Online DDL](https://dev.mysql.com/doc/refman/8.0/en/innodb-online-ddl.html), [InnoDB Architecture](https://dev.mysql.com/doc/refman/8.0/en/innodb-architecture.html)

### Analogi: Database sebagai Pabrik

Bayangkan MySQL sebagai **pabrik dengan conveyor belt**:
- **Table** = Gudang dengan format penyimpanan tertentu
- **Row** = Produk di gudang
- **Buffer Pool** = Area staging sebelum masuk gudang
- **Redo Log** = Buku catatan semua perubahan
- **DDL** = Renovasi gudang

### MySQL DDL Algorithms

#### INSTANT (MySQL 8.0.12+)

```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20), ALGORITHM=INSTANT;
```

**Yang terjadi:**
1. Update **data dictionary** only
2. Menambah metadata di InnoDB internal tables
3. **Zero table rebuild**
4. Instant = milidetik regardless of table size

**Support:** ADD COLUMN (last position), DROP COLUMN, RENAME COLUMN, SET DEFAULT

**Perumpamaan:** Seperti menambah label baru di sistem inventory gudang. Produk fisik tidak disentuh, hanya database tracking-nya.

#### INPLACE

```sql
ALTER TABLE users ADD INDEX idx_email (email), ALGORITHM=INPLACE, LOCK=NONE;
```

**Yang terjadi:**
1. **Prepare phase:** Lock singkat untuk setup
2. **Build phase:** Scan table, build struktur baru sambil **DML tetap jalan**
3. MySQL mencatat semua DML yang terjadi selama build di **online log**
4. **Apply phase:** Apply online log ke struktur baru
5. **Commit phase:** Lock singkat untuk swap

**Perumpamaan:** Seperti membangun rak baru di gudang sambil operasi tetap jalan. Pekerja tetap bisa ambil/taruh barang. Setelah rak baru selesai, baru dipindahkan posisinya.

#### COPY (Legacy)

```sql
ALTER TABLE users MODIFY COLUMN name VARCHAR(500), ALGORITHM=COPY;
```

**Yang terjadi:**
1. Create temporary table dengan schema baru
2. **Lock table** (atau row-by-row copy dengan table lock di akhir)
3. Copy semua rows dari old table ke temp table
4. Swap tables (rename)
5. Drop old table

**Perumpamaan:** Seperti membangun gudang baru di sebelah, pindahkan semua barang, lalu hancurkan gudang lama. Mahal dan lambat.

### MySQL Online DDL Decision Matrix

| Operation | INSTANT | INPLACE | COPY | Concurrent DML |
|-----------|---------|---------|------|----------------|
| ADD COLUMN (last) | ✅ 8.0.12+ | ✅ | ✅ | Yes |
| ADD COLUMN (middle) | ❌ | ✅ | ✅ | Yes |
| DROP COLUMN | ✅ 8.0.29+ | ✅ | ✅ | Yes |
| CHANGE TYPE | ❌ | ❌ | ✅ | **No** |
| ADD INDEX | ❌ | ✅ | ✅ | Yes |
| ADD PRIMARY KEY | ❌ | ✅* | ✅ | Yes (rebuilds table) |
| Extend VARCHAR (same byte range) | ❌ | ✅ | ✅ | Yes |

### Critical INSTANT Limitations (from Official Docs)

1. **Max 64 row versions** — setiap INSTANT ADD/DROP membuat row version baru. Setelah 64x, harus rebuild dengan INPLACE/COPY
2. **Cannot combine** dengan operasi non-INSTANT dalam satu ALTER
3. **Not supported for:** ROW_FORMAT=COMPRESSED, tables with FULLTEXT index, temporary tables
4. **ADD auto-increment column** — tidak bisa concurrent DML (minimum LOCK=SHARED)

### VARCHAR Extension Gotcha

```sql
-- ✅ OK: 100 → 200 (both use 1 length byte)
ALTER TABLE t CHANGE c1 c1 VARCHAR(200), ALGORITHM=INPLACE;

-- ❌ FAIL: 255 → 256 (crosses byte boundary!)
-- Must use ALGORITHM=COPY
ALTER TABLE t CHANGE c1 c1 VARCHAR(256), ALGORITHM=COPY;
```

**Rule:** VARCHAR 0-255 = 1 byte header, 256+ = 2 byte header. Crossing boundary requires COPY.

---

## 1.3 MongoDB Schema Changes Internals

**Source:** [MongoDB Docs - Data Modeling](https://www.mongodb.com/docs/manual/core/data-modeling-introduction/), [WiredTiger Storage](https://www.mongodb.com/docs/manual/core/wiredtiger/)

### Analogi: Database sebagai Laci File Flexibel

Bayangkan MongoDB sebagai **filing cabinet dengan folder fleksibel**:
- **Collection** = Laci
- **Document** = Folder dengan isi bebas
- **Field** = Label di folder
- **Schema** = Tidak ada aturan ketat, setiap folder boleh beda format

### "Schema Change" di MongoDB

MongoDB adalah **schemaless**, jadi tidak ada ALTER TABLE. Tapi ada implikasi:

#### Adding a Field (Per-Document)

```javascript
db.users.updateOne(
  { _id: ObjectId("...") },
  { $set: { phone: "08123456789" } }
)
```

**Yang terjadi di mesin (WiredTiger):**
1. Read document dari disk/cache
2. Deserialize BSON
3. Add field ke in-memory document
4. Serialize ulang ke BSON (ukuran berubah!)
5. **Jika document membesar dan tidak muat di slot lama:**
   - Allocate slot baru di data file
   - Write document ke slot baru
   - Update index entry untuk point ke lokasi baru
   - Mark slot lama sebagai free space

**Perumpamaan:** Seperti menambah kertas ke folder. Jika folder jadi terlalu tebal untuk slotnya, pindahkan ke slot yang lebih besar.

#### Bulk Schema Migration

```javascript
db.users.updateMany(
  { phone: { $exists: false } },
  { $set: { phone: null } }
)
```

**Yang terjadi:**
1. Full collection scan (jika tidak ada index di field filter)
2. Setiap document yang match di-update satu per satu
3. Bisa menyebabkan **document movement** jika ukuran berubah signifikan

**Best Practice:** Gunakan **lazy migration** — update document saat di-read oleh application, bukan batch sekaligus.

#### Index Creation

```javascript
db.users.createIndex({ email: 1 })
```

**Yang terjadi (MongoDB 4.2+ Optimized Build from Official Docs):**

**1. Lock Phase (Exclusive Lock - Brief)**
- Acquire exclusive lock on collection
- Create empty index data structure
- Start intercepting writes via "side write table"
- Release exclusive lock

**2. Scan Phase (Intent Exclusive Lock - Yields)**
- Scan collection with IX lock (yields to reads/writes)
- Insert keys into external sorter
- Handle concurrent writes via side write table
- **Concurrent reads and writes allowed**

**3. Process Side Writes (Intent Exclusive Lock - Yields)**
- Drain side write table (FIFO)
- Handle keys from documents written during build
- **Concurrent reads and writes allowed**

**4. Vote and Wait for Commit Quorum (Replica Sets)**
- Member submits "vote" to primary
- Primary waits for commit quorum (default: all voting members)
- Ensures index consistency across replicas

**5. Finalize Phase (Exclusive Lock - Brief)**
- Acquire exclusive lock
- Mark index as ready
- Release exclusive lock

**Key Insight:** Index builds hold **exclusive lock only at start and end**. The bulk of work happens with IX lock that yields to concurrent operations.

---

## 1.4 Cassandra ALTER TABLE Internals

**Source:** [Cassandra Docs - ALTER TABLE](https://cassandra.apache.org/doc/latest/cassandra/cql/ddl.html), [SSTable](https://cassandra.apache.org/doc/latest/cassandra/architecture/storage-engine.html)

### Analogi: Database sebagai Cluster Gudang Tersebar

Bayangkan Cassandra sebagai **jaringan gudang di berbagai kota**:
- **Keyspace** = Negara (dengan aturan replikasi)
- **Table** = Jenis produk
- **Partition Key** = Kota tujuan (menentukan gudang mana)
- **SSTable** = Container pengiriman (immutable)
- **Memtable** = Area packing sebelum dikirim

### Cassandra DDL Operations

#### ADD COLUMN (Instant)

```sql
ALTER TABLE users ADD phone text;
```

**Yang terjadi di mesin:**
1. Update schema di **system_schema** keyspace
2. Propagate schema ke semua nodes via **gossip protocol**
3. **TIDAK menyentuh data sama sekali**
4. Existing rows akan return NULL untuk column baru

**Kenapa instant?** Cassandra menggunakan **sparse column storage**. Setiap row hanya menyimpan columns yang ada nilainya. Tidak ada concept "slot kosong" yang harus diisi.

**Perumpamaan:** Seperti menambah kategori baru di sistem tracking pengiriman. Paket-paket lama tidak perlu dilabeli ulang — mereka simply tidak punya label kategori baru.

#### DROP COLUMN (Tombstone)

```sql
ALTER TABLE users DROP phone;
```

**Yang terjadi:**
1. Update schema (remove column definition)
2. Existing data **TIDAK dihapus immediately**
3. Data di-mark sebagai **tombstone** (batu nisan)
4. Actual deletion terjadi saat **compaction**

**Perumpamaan:** Seperti mencoret kategori dari sistem, tapi barang di gudang masih ada sampai ada jadwal bersih-bersih (compaction).

#### Cannot Change Primary Key

```sql
-- TIDAK BISA
ALTER TABLE users ALTER id TYPE bigint;  -- Error!
```

**Kenapa?** Primary key menentukan:
- Partition placement (node mana)
- Row ordering (clustering)
- SSTable organization

Mengubah primary key = membangun ulang seluruh table di seluruh cluster.

---

## Summary: DDL Behavior Comparison

| Aspect | PostgreSQL | MySQL | MongoDB | Cassandra |
|--------|------------|-------|---------|-----------|
| ADD COLUMN | Instant (PG11+) | INSTANT (8.0.12+) | Per-doc update | Instant (sparse) |
| DROP COLUMN | Instant (marks invisible) | INSTANT (8.0.29+) | Per-doc $unset | Tombstone |
| CHANGE TYPE | **Table rewrite** | COPY algorithm | Per-doc update | **Not allowed** |
| ADD INDEX | CONCURRENTLY option | INPLACE, LOCK=NONE | Background build | Async across nodes |
| Lock Model | MVCC + heavy locks | Row-level + DDL locks | Document-level | Distributed (no global lock) |

---

# Part 2-4: See Individual Lab Directories

- [Part 2: Replication & Scaling](./REPLICATION.md)
- [Part 3: Backup & Disaster Recovery](./BACKUP.md)  
- [Part 4: Connection Pooling](./POOLING.md)
