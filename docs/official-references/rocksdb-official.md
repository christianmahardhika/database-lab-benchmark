# RocksDB Official Documentation Reference

Source: https://rocksdb.org/docs/

## Overview

RocksDB is an embedded persistent key-value store optimized for fast storage (SSD). Forked from LevelDB by Facebook, now widely used as storage engine (MySQL MyRocks, TiKV, CockroachDB).

## Storage Engine: LSM-Tree

### Architecture
- **MemTable**: In-memory write buffer (default 64MB)
- **WAL**: Write-ahead log for durability
- **SST Files**: Sorted String Tables, immutable
- **Levels**: L0 (unsorted) → L1 → L2 → ... (sorted, increasing size)

### Write Path
```
Put() → WAL → MemTable → (flush) → L0 SST → (compaction) → L1+ SST
```

### Read Path
```
Get() → MemTable → Block Cache → Bloom Filter → SST Index → SST Block
```

### Compaction Strategies
| Strategy | Use Case | Write Amp | Space Amp |
|----------|----------|-----------|-----------|
| Leveled | General purpose | High | Low |
| Universal | Write-heavy | Low | High |
| FIFO | TTL data | Lowest | Highest |

## Performance Characteristics (Official Claims)

### Write Performance
- **Claim**: "Optimized for fast, low-latency storage"
- **Benchmark**: 500K+ ops/sec single-threaded (SSD)
- **Why**: Sequential writes, batching, async WAL

### Read Performance
- **Claim**: "Optimized for SSD with block cache"
- **Point lookup**: <1ms with data in cache
- **Range scan**: Sequential read from SST

### Write Amplification
- **Leveled**: 10-30x typical
- **Universal**: 5-10x typical
- **Tunable**: Trade-off with space amplification

## Best Use Cases (from rocksdb.org)

1. **Embedded Storage**
   - Application-embedded database
   - No network overhead
   - Single-process access

2. **Storage Engine Backend**
   - MySQL (MyRocks)
   - TiKV (TiDB)
   - CockroachDB

3. **Write-Heavy Workloads**
   - Time-series ingestion
   - Log/event collection
   - Blockchain state

4. **SSD-Optimized Workloads**
   - Sequential write pattern
   - Block-aligned I/O

## Anti-Patterns (from docs)

1. **Multi-process access**
   - Single process only
   - Use DB server for shared access

2. **Large values**
   - Optimized for small values (<100KB)
   - BlobDB for large values

3. **Heavy random reads**
   - LSM not optimal for point lookups
   - B-Tree better if read-heavy

## Configuration for Benchmarks

### Basic Options
```cpp
Options options;
options.create_if_missing = true;

// Write buffer
options.write_buffer_size = 64 * 1024 * 1024;  // 64MB
options.max_write_buffer_number = 3;

// Compaction
options.level_compaction_dynamic_level_bytes = true;
options.max_bytes_for_level_base = 256 * 1024 * 1024;  // 256MB

// Block cache
BlockBasedTableOptions table_options;
table_options.block_cache = NewLRUCache(512 * 1024 * 1024);  // 512MB
table_options.filter_policy.reset(NewBloomFilterPolicy(10));

options.table_factory.reset(NewBlockBasedTableFactory(table_options));
```

### Write-Optimized Config
```cpp
// Universal compaction for less write amp
options.compaction_style = kCompactionStyleUniversal;
options.universal_compaction_options.size_ratio = 10;

// Faster WAL
options.wal_dir = "/fast-ssd/wal";
options.max_total_wal_size = 1024 * 1024 * 1024;  // 1GB
```

### Read-Optimized Config
```cpp
// Larger block cache
table_options.block_cache = NewLRUCache(2 * 1024 * 1024 * 1024LL);  // 2GB

// Bloom filter for point lookups
table_options.filter_policy.reset(NewBloomFilterPolicy(10, false));

// Pin L0 + L1 in cache
options.pin_l0_filter_and_index_blocks_in_cache = true;
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| rocksdb.block.cache.hit | Cache hit rate | >90% |
| rocksdb.compaction.times.micros | Compaction latency | Monitor |
| rocksdb.stall.micros | Write stalls | <1% of time |
| rocksdb.bytes.written | Write throughput | Measure |
| rocksdb.bytes.read | Read throughput | Measure |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H17 | Embedded single-process | N/A (architecture) |
| H18 | High write throughput | Exp 1: Write benchmark |
| H19 | SSD-optimized | Exp 6: Write amplification |
| H20 | Blockchain/ledger storage | N/A (use case) |

## Comparison: RocksDB vs LevelDB

| Aspect | RocksDB | LevelDB |
|--------|---------|---------|
| Maintainer | Meta (Facebook) | Google |
| Compaction | Multiple strategies | Leveled only |
| Compression | LZ4, Snappy, ZSTD | Snappy |
| Concurrency | Multi-threaded | Single-threaded |
| Features | Column families, TTL, transactions | Basic KV |
| Performance | Faster, more tunable | Simpler |

## References

- Getting Started: https://rocksdb.org/docs/getting-started.html
- Tuning Guide: https://github.com/facebook/rocksdb/wiki/RocksDB-Tuning-Guide
- Write Amplification: https://github.com/facebook/rocksdb/wiki/Write-Amplification
- Benchmarks: https://github.com/facebook/rocksdb/wiki/Performance-Benchmarks
