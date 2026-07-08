#!/usr/bin/env python3
"""RocksDB Operations Benchmark using rocksdict"""
import rocksdict
from rocksdict import Rdict, Options, WriteBatch, AccessType
import time
import os
import shutil
import json
from datetime import datetime

DB_PATH = "/tmp/rocksdb_bench"
CHECKPOINT_PATH = "/tmp/rocksdb_checkpoint"
RESULTS = {}

def cleanup():
    shutil.rmtree(DB_PATH, ignore_errors=True)
    shutil.rmtree(CHECKPOINT_PATH, ignore_errors=True)

def get_dir_size(path):
    total = 0
    for dirpath, dirnames, filenames in os.walk(path):
        for f in filenames:
            fp = os.path.join(dirpath, f)
            total += os.path.getsize(fp)
    return total

def benchmark_basic_ops():
    """Test basic read/write operations"""
    print("\n=== 1. Basic Operations ===")
    cleanup()
    
    db = Rdict(DB_PATH)
    
    # Write benchmark
    NUM_KEYS = 100000
    start = time.time()
    for i in range(NUM_KEYS):
        db[f"key_{i:08d}"] = f"value_{i}"
    write_time = time.time() - start
    write_ops = NUM_KEYS / write_time
    print(f"Write: {NUM_KEYS} keys in {write_time:.2f}s = {write_ops:.0f} ops/s")
    
    # Read benchmark  
    start = time.time()
    for i in range(NUM_KEYS):
        _ = db[f"key_{i:08d}"]
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

def benchmark_batch_write():
    """Test batch write performance"""
    print("\n=== 2. Batch Write ===")
    cleanup()
    
    db = Rdict(DB_PATH)
    NUM_KEYS = 100000
    BATCH_SIZE = 1000
    
    start = time.time()
    for batch_start in range(0, NUM_KEYS, BATCH_SIZE):
        wb = WriteBatch()
        for i in range(batch_start, min(batch_start + BATCH_SIZE, NUM_KEYS)):
            wb[f"batch_{i:08d}"] = f"value_{i}"
        db.write(wb)
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

def benchmark_checkpoint():
    """Test online checkpoint (backup without closing DB)"""
    print("\n=== 3. Checkpoint (Online Backup) ===")
    cleanup()
    
    db = Rdict(DB_PATH)
    
    # Write data
    for i in range(100000):
        db[f"key_{i:08d}"] = f"value_{i}"
    
    db_size = get_dir_size(DB_PATH)
    print(f"DB size: {db_size / 1024 / 1024:.2f} MB")
    
    # Create checkpoint while DB is OPEN (online backup!)
    start = time.time()
    try:
        # rocksdict doesn't expose checkpoint directly, use file copy as fallback
        # In real RocksDB: db.checkpoint(CHECKPOINT_PATH)
        db.flush()  # Ensure all data is on disk
        shutil.copytree(DB_PATH, CHECKPOINT_PATH)
        checkpoint_time = time.time() - start
        print(f"Checkpoint time: {checkpoint_time*1000:.1f}ms (flush + copy)")
        
        # Verify checkpoint while original DB still open
        checkpoint_db = Rdict(CHECKPOINT_PATH, access_type=AccessType.read_only())
        count = sum(1 for _ in checkpoint_db.keys())
        checkpoint_db.close()
        print(f"Checkpoint verified: {count} keys (DB remained open)")
        
        RESULTS["checkpoint"] = {
            "db_size_mb": round(db_size / 1024 / 1024, 2),
            "checkpoint_time_ms": round(checkpoint_time * 1000, 1),
            "keys_verified": count,
            "online": True
        }
    except Exception as e:
        print(f"Checkpoint failed: {e}")
        RESULTS["checkpoint"] = {"error": str(e)}
    
    db.close()

def benchmark_compaction():
    """Test compaction impact"""
    print("\n=== 4. Compaction Impact ===")
    cleanup()
    
    db = Rdict(DB_PATH)
    
    # Write then delete (creates tombstones)
    for i in range(50000):
        db[f"key_{i:08d}"] = f"value_{i}"
    for i in range(25000):  # Delete half
        del db[f"key_{i:08d}"]
    
    db.flush()
    size_before = get_dir_size(DB_PATH)
    print(f"Size before compaction: {size_before / 1024 / 1024:.2f} MB")
    
    # Compact
    start = time.time()
    db.compact_range(None, None)  # Full range compaction
    compact_time = time.time() - start
    
    size_after = get_dir_size(DB_PATH)
    print(f"Size after compaction: {size_after / 1024 / 1024:.2f} MB")
    print(f"Compaction time: {compact_time*1000:.1f}ms")
    if size_before > 0:
        print(f"Space reclaimed: {(size_before - size_after) / 1024 / 1024:.2f} MB ({(1 - size_after/size_before)*100:.1f}%)")
    
    RESULTS["compaction"] = {
        "size_before_mb": round(size_before / 1024 / 1024, 2),
        "size_after_mb": round(size_after / 1024 / 1024, 2),
        "compaction_time_ms": round(compact_time * 1000, 1),
        "space_reclaimed_pct": round((1 - size_after/size_before) * 100, 1) if size_before > 0 else 0
    }
    
    db.close()

def benchmark_column_families():
    """Test column families (like separate tables)"""
    print("\n=== 5. Column Families ===")
    cleanup()
    
    # Create DB with column families
    opts = Options()
    opts.create_if_missing(True)
    
    db = Rdict(DB_PATH, options=opts)
    
    # Write to default CF
    start = time.time()
    for i in range(10000):
        db[f"default_{i}"] = f"value_{i}"
    default_time = time.time() - start
    
    # Create additional column family
    try:
        cf_users = db.create_column_family("users")
        cf_orders = db.create_column_family("orders")
        
        # Write to users CF
        start = time.time()
        for i in range(10000):
            cf_users[f"user_{i}"] = f"user_data_{i}"
        users_time = time.time() - start
        
        # Write to orders CF
        start = time.time()
        for i in range(10000):
            cf_orders[f"order_{i}"] = f"order_data_{i}"
        orders_time = time.time() - start
        
        print(f"Default CF: 10k writes in {default_time*1000:.1f}ms")
        print(f"Users CF: 10k writes in {users_time*1000:.1f}ms")
        print(f"Orders CF: 10k writes in {orders_time*1000:.1f}ms")
        
        # Verify isolation
        default_count = sum(1 for _ in db.keys())
        users_count = sum(1 for _ in cf_users.keys())
        orders_count = sum(1 for _ in cf_orders.keys())
        print(f"Keys per CF: default={default_count}, users={users_count}, orders={orders_count}")
        
        RESULTS["column_families"] = {
            "default_write_ms": round(default_time * 1000, 1),
            "users_write_ms": round(users_time * 1000, 1),
            "orders_write_ms": round(orders_time * 1000, 1),
            "default_keys": default_count,
            "users_keys": users_count,
            "orders_keys": orders_count
        }
        
        cf_users.close()
        cf_orders.close()
    except Exception as e:
        print(f"Column families not supported in this build: {e}")
        RESULTS["column_families"] = {"error": str(e)}
    
    db.close()

if __name__ == "__main__":
    print("=" * 50)
    print("RocksDB Operations Benchmark")
    print(f"Date: {datetime.now().isoformat()}")
    print("=" * 50)
    
    benchmark_basic_ops()
    benchmark_batch_write()
    benchmark_checkpoint()
    benchmark_compaction()
    benchmark_column_families()
    
    cleanup()
    
    print("\n" + "=" * 50)
    print("RESULTS SUMMARY")
    print("=" * 50)
    print(json.dumps(RESULTS, indent=2))
    
    # Save results
    with open("RESULTS.json", "w") as f:
        json.dump(RESULTS, f, indent=2)
    print("\nResults saved to RESULTS.json")
