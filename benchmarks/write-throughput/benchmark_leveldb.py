#!/usr/bin/env python3
"""LevelDB Operations Benchmark"""
import plyvel
import time
import os
import shutil
import json
from datetime import datetime

DB_PATH = "/tmp/leveldb_bench"
BACKUP_PATH = "/tmp/leveldb_backup"
RESULTS = {}

def cleanup():
    shutil.rmtree(DB_PATH, ignore_errors=True)
    shutil.rmtree(BACKUP_PATH, ignore_errors=True)

def benchmark_basic_ops():
    """Test basic read/write operations"""
    print("\n=== 1. Basic Operations ===")
    cleanup()
    
    db = plyvel.DB(DB_PATH, create_if_missing=True)
    
    # Write benchmark
    NUM_KEYS = 100000
    start = time.time()
    for i in range(NUM_KEYS):
        db.put(f"key_{i:08d}".encode(), f"value_{i}".encode())
    write_time = time.time() - start
    write_ops = NUM_KEYS / write_time
    print(f"Write: {NUM_KEYS} keys in {write_time:.2f}s = {write_ops:.0f} ops/s")
    
    # Read benchmark  
    start = time.time()
    for i in range(NUM_KEYS):
        db.get(f"key_{i:08d}".encode())
    read_time = time.time() - start
    read_ops = NUM_KEYS / read_time
    print(f"Read: {NUM_KEYS} keys in {read_time:.2f}s = {read_ops:.0f} ops/s")
    
    RESULTS["basic_ops"] = {
        "num_keys": NUM_KEYS,
        "write_time_s": round(write_time, 3),
        "write_ops_per_s": round(write_ops),
        "read_time_s": round(read_time, 3),
        "read_ops_per_s": round(read_ops)
    }
    
    db.close()
    return DB_PATH

def benchmark_batch_write():
    """Test batch write performance"""
    print("\n=== 2. Batch Write ===")
    cleanup()
    
    db = plyvel.DB(DB_PATH, create_if_missing=True)
    NUM_KEYS = 100000
    BATCH_SIZE = 1000
    
    start = time.time()
    for batch_start in range(0, NUM_KEYS, BATCH_SIZE):
        with db.write_batch() as wb:
            for i in range(batch_start, min(batch_start + BATCH_SIZE, NUM_KEYS)):
                wb.put(f"batch_{i:08d}".encode(), f"value_{i}".encode())
    batch_time = time.time() - start
    batch_ops = NUM_KEYS / batch_time
    print(f"Batch write: {NUM_KEYS} keys in {batch_time:.2f}s = {batch_ops:.0f} ops/s")
    
    RESULTS["batch_write"] = {
        "num_keys": NUM_KEYS,
        "batch_size": BATCH_SIZE,
        "time_s": round(batch_time, 3),
        "ops_per_s": round(batch_ops)
    }
    
    db.close()

def benchmark_backup():
    """Test backup (offline copy) timing"""
    print("\n=== 3. Backup (Offline Copy) ===")
    
    # First create a DB with data
    cleanup()
    db = plyvel.DB(DB_PATH, create_if_missing=True)
    for i in range(100000):
        db.put(f"key_{i:08d}".encode(), f"value_{i}" .encode())
    db.close()
    
    db_size = sum(os.path.getsize(os.path.join(DB_PATH, f)) for f in os.listdir(DB_PATH))
    print(f"DB size: {db_size / 1024 / 1024:.2f} MB")
    
    # Backup timing (must close DB first for LevelDB)
    start = time.time()
    shutil.copytree(DB_PATH, BACKUP_PATH)
    backup_time = time.time() - start
    print(f"Backup time: {backup_time*1000:.1f}ms")
    
    # Verify backup
    backup_db = plyvel.DB(BACKUP_PATH)
    count = sum(1 for _ in backup_db)
    backup_db.close()
    print(f"Backup verified: {count} keys")
    
    RESULTS["backup"] = {
        "db_size_mb": round(db_size / 1024 / 1024, 2),
        "backup_time_ms": round(backup_time * 1000, 1),
        "keys_verified": count
    }

def benchmark_compaction():
    """Test manual compaction impact"""
    print("\n=== 4. Compaction Impact ===")
    cleanup()
    
    db = plyvel.DB(DB_PATH, create_if_missing=True, 
                   write_buffer_size=4*1024*1024)  # 4MB buffer
    
    # Write then delete (creates tombstones)
    for i in range(50000):
        db.put(f"key_{i:08d}".encode(), f"value_{i}".encode())
    for i in range(25000):  # Delete half
        db.delete(f"key_{i:08d}".encode())
    
    size_before = sum(os.path.getsize(os.path.join(DB_PATH, f)) for f in os.listdir(DB_PATH))
    print(f"Size before compaction: {size_before / 1024 / 1024:.2f} MB")
    
    # Compact
    start = time.time()
    db.compact_range()
    compact_time = time.time() - start
    
    size_after = sum(os.path.getsize(os.path.join(DB_PATH, f)) for f in os.listdir(DB_PATH))
    print(f"Size after compaction: {size_after / 1024 / 1024:.2f} MB")
    print(f"Compaction time: {compact_time*1000:.1f}ms")
    print(f"Space reclaimed: {(size_before - size_after) / 1024 / 1024:.2f} MB ({(1 - size_after/size_before)*100:.1f}%)")
    
    RESULTS["compaction"] = {
        "size_before_mb": round(size_before / 1024 / 1024, 2),
        "size_after_mb": round(size_after / 1024 / 1024, 2),
        "compaction_time_ms": round(compact_time * 1000, 1),
        "space_reclaimed_pct": round((1 - size_after/size_before) * 100, 1)
    }
    
    db.close()

def benchmark_bloom_filter():
    """Test bloom filter impact on reads"""
    print("\n=== 5. Bloom Filter Impact ===")
    
    # Without bloom filter
    cleanup()
    db_no_bloom = plyvel.DB(DB_PATH, create_if_missing=True)
    for i in range(50000):
        db_no_bloom.put(f"key_{i:08d}".encode(), f"value_{i}".encode())
    
    start = time.time()
    for i in range(10000):
        db_no_bloom.get(f"missing_{i}".encode())  # Non-existent keys
    no_bloom_time = time.time() - start
    db_no_bloom.close()
    print(f"Without bloom filter: {no_bloom_time*1000:.1f}ms for 10k missing key lookups")
    
    # With bloom filter
    cleanup()
    db_bloom = plyvel.DB(DB_PATH, create_if_missing=True, bloom_filter_bits=10)
    for i in range(50000):
        db_bloom.put(f"key_{i:08d}".encode(), f"value_{i}".encode())
    
    start = time.time()
    for i in range(10000):
        db_bloom.get(f"missing_{i}".encode())
    bloom_time = time.time() - start
    db_bloom.close()
    print(f"With bloom filter (10 bits): {bloom_time*1000:.1f}ms for 10k missing key lookups")
    print(f"Speedup: {no_bloom_time/bloom_time:.1f}x")
    
    RESULTS["bloom_filter"] = {
        "no_bloom_ms": round(no_bloom_time * 1000, 1),
        "bloom_10bit_ms": round(bloom_time * 1000, 1),
        "speedup": round(no_bloom_time / bloom_time, 1)
    }

if __name__ == "__main__":
    print("=" * 50)
    print("LevelDB Operations Benchmark")
    print(f"Date: {datetime.now().isoformat()}")
    print("=" * 50)
    
    benchmark_basic_ops()
    benchmark_batch_write()
    benchmark_backup()
    benchmark_compaction()
    benchmark_bloom_filter()
    
    cleanup()
    
    print("\n" + "=" * 50)
    print("RESULTS SUMMARY")
    print("=" * 50)
    print(json.dumps(RESULTS, indent=2))
    
    # Save results
    with open("RESULTS.json", "w") as f:
        json.dump(RESULTS, f, indent=2)
    print("\nResults saved to RESULTS.json")
