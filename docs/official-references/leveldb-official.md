# LevelDB Official Documentation Reference

Source: https://github.com/google/leveldb/blob/main/doc/

## Overview

LevelDB is a fast key-value storage library written at Google. Provides ordered mapping from string keys to string values. Foundation for RocksDB.

## Storage Engine: LSM-Tree

### Architecture
- **MemTable**: In-memory skip list (default 4MB)
- **Immutable MemTable**: Being flushed to disk
- **Log**: Write-ahead log for durability
- **SSTable**: Sorted String Table files
- **Levels**: L0 → L1 → L2 → ... (10x size ratio)

### File Structure
```
dbname/
├── CURRENT           # Pointer to current MANIFEST
├── LOCK              # File lock
├── LOG               # Runtime log
├── MANIFEST-000001   # Metadata
├── 000003.log        # WAL
├── 000004.ldb        # SSTable
└── ...
```

### Write Path
```
Put(key, value)
  → Append to log (WAL)
  → Insert into MemTable (skip list)
  → Return success

Background:
  → MemTable full (4MB) → Write to L0 SSTable
  → L0 SSTable count > 4 → Compact to L1
  → Level N size limit → Compact to N+1
```

### Read Path
```
Get(key)
  → Search MemTable
  → Search Immutable MemTable
  → For each level 0..N:
      → Use bloom filter to skip SSTables
      → Binary search in SSTable index
      → Read data block
  → Return value or NotFound
```

## Performance Characteristics (Official)

### From README
> LevelDB is not a SQL database. It does not have a relational data model, it does not support SQL queries, and it has no support for indexes.

### Benchmarks (from Google)
```
# Sequential writes
fillseq      :       1.765 micros/op;   62.7 MB/s

# Random writes
fillrandom   :       2.460 micros/op;   45.0 MB/s

# Sequential reads
readseq      :       0.476 micros/op;  232.3 MB/s

# Random reads
readrandom   :       9.872 micros/op;
```

### Key Characteristics
- **Ordered iteration**: Keys are sorted
- **Batch writes**: Atomic batch operations
- **Snapshots**: Consistent read views
- **Compression**: Snappy by default

## Best Use Cases

1. **Embedded Key-Value Store**
   - Browser storage (Chrome IndexedDB backend)
   - Application state persistence
   - Configuration storage

2. **Building Block**
   - Base for other databases
   - RocksDB forked from LevelDB
   - Inspiration for many LSM implementations

3. **Simple Persistent Map**
   - When you need sorted KV
   - Single-process access
   - No complex queries

## Limitations (from docs)

1. **Single process**
   - Only one process can open DB
   - Use LOCK file to prevent corruption

2. **No client-server**
   - Embedded library only
   - No network protocol

3. **No transactions**
   - WriteBatch is atomic
   - But no multi-batch transactions

4. **Limited data types**
   - Keys and values are byte strings
   - No schema, no types

## Configuration for Benchmarks

### Basic Options
```cpp
leveldb::Options options;
options.create_if_missing = true;
options.write_buffer_size = 4 * 1024 * 1024;  // 4MB MemTable
options.max_open_files = 1000;
options.block_size = 4096;  // 4KB blocks
options.compression = leveldb::kSnappyCompression;
```

### Write-Optimized
```cpp
leveldb::WriteOptions write_options;
write_options.sync = false;  // Async WAL (faster, less durable)

// Use WriteBatch for multiple writes
leveldb::WriteBatch batch;
batch.Put("key1", "value1");
batch.Put("key2", "value2");
db->Write(write_options, &batch);
```

### Read-Optimized
```cpp
leveldb::ReadOptions read_options;
read_options.verify_checksums = false;  // Skip verification
read_options.fill_cache = true;  // Populate cache

// Use iterator for range scan
leveldb::Iterator* it = db->NewIterator(read_options);
for (it->SeekToFirst(); it->Valid(); it->Next()) {
    // Process it->key(), it->value()
}
```

### Python (plyvel)
```python
import plyvel

# Open database
db = plyvel.DB('/tmp/testdb/', create_if_missing=True)

# Write
db.put(b'key', b'value')

# Batch write
with db.write_batch() as wb:
    wb.put(b'key1', b'value1')
    wb.put(b'key2', b'value2')

# Read
value = db.get(b'key')

# Range scan
for key, value in db.iterator():
    print(key, value)

# Close
db.close()
```

## Key Metrics to Monitor

| Metric | Description | How |
|--------|-------------|-----|
| Write latency | Time per write | Benchmark |
| Read latency | Time per read | Benchmark |
| Compaction | Background merge | LOG file |
| Disk usage | SSTable size | File system |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H17 | Embedded single-process | N/A (architecture) |
| H18 | High write throughput | Exp 1: Write benchmark |

## Comparison with Other KV Stores

| Aspect | LevelDB | RocksDB | LMDB |
|--------|---------|---------|------|
| Engine | LSM | LSM | B+Tree |
| Writes | Fast | Faster | Slower |
| Reads | Good | Good | Faster |
| Space | More (write amp) | More | Less |
| Features | Basic | Rich | Basic |
| Concurrency | Single writer | Multi-thread | Multi-process |

## References

- Documentation: https://github.com/google/leveldb/tree/main/doc
- Implementation Notes: https://github.com/google/leveldb/blob/main/doc/impl.md
- Table Format: https://github.com/google/leveldb/blob/main/doc/table_format.md
- Python Binding: https://plyvel.readthedocs.io/
