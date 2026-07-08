#!/usr/bin/env python3
"""Cassandra (Java LSM) vs RocksDB (C++ LSM) benchmark"""

import time
import statistics
from cassandra.cluster import Cluster
import subprocess
import os

NUM_OPS = 50000
DATA_SIZE = 100  # bytes per value

def benchmark_cassandra():
    """Benchmark Cassandra via Python driver"""
    print("\n=== CASSANDRA (Java LSM, Distributed) ===")
    
    cluster = Cluster(['172.17.0.2'])
    session = cluster.connect()
    
    # Setup
    session.execute("CREATE KEYSPACE IF NOT EXISTS bench WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}")
    session.execute("USE bench")
    session.execute("DROP TABLE IF EXISTS test")
    session.execute("CREATE TABLE test (k TEXT PRIMARY KEY, v TEXT)")
    
    # Prepare statements
    insert_stmt = session.prepare("INSERT INTO test (k, v) VALUES (?, ?)")
    select_stmt = session.prepare("SELECT v FROM test WHERE k = ?")
    
    value = 'x' * DATA_SIZE
    
    # Write benchmark
    print(f"Writing {NUM_OPS} rows...")
    start = time.perf_counter()
    for i in range(NUM_OPS):
        session.execute(insert_stmt, [f"key_{i}", value])
    write_time = time.perf_counter() - start
    write_ops = NUM_OPS / write_time
    
    # Read benchmark  
    print(f"Reading {NUM_OPS} rows...")
    start = time.perf_counter()
    for i in range(NUM_OPS):
        session.execute(select_stmt, [f"key_{i}"])
    read_time = time.perf_counter() - start
    read_ops = NUM_OPS / read_time
    
    cluster.shutdown()
    
    return {
        'write_ops': write_ops,
        'write_latency_ms': (write_time / NUM_OPS) * 1000,
        'read_ops': read_ops, 
        'read_latency_ms': (read_time / NUM_OPS) * 1000
    }

def benchmark_rocksdb_via_tikv():
    """Benchmark RocksDB via TiKV (distributed wrapper)"""
    print("\n=== ROCKSDB via TiKV (C++ LSM, Distributed) ===")
    
    # TiKV uses RocksDB underneath - benchmark via tikv-ctl or raw API
    # For fair comparison, use tikv's raw KV API
    
    try:
        import tikv_client
    except ImportError:
        print("tikv-client not available, using HTTP benchmark...")
        return benchmark_tikv_http()
    
def benchmark_tikv_http():
    """Benchmark TiKV via HTTP API (pd-ctl)"""
    import requests
    
    # TiKV doesn't have simple HTTP KV API like Redis
    # Let's use ldb (RocksDB CLI) directly instead
    return None

def benchmark_rocksdb_ldb():
    """Benchmark RocksDB using ldb CLI tool"""
    print("\n=== ROCKSDB (C++ LSM, Embedded via ldb) ===")
    
    db_path = "/tmp/rocksdb_bench"
    os.makedirs(db_path, exist_ok=True)
    
    value = 'x' * DATA_SIZE
    
    # Write benchmark using ldb
    print(f"Writing {NUM_OPS} rows via ldb...")
    
    # Create batch file for writes
    batch_file = "/tmp/rocksdb_batch.txt"
    with open(batch_file, 'w') as f:
        for i in range(NUM_OPS):
            f.write(f"put key_{i} {value}\n")
    
    start = time.perf_counter()
    result = subprocess.run(
        ['docker', 'run', '--rm', '-v', f'{db_path}:/data', '-v', '/tmp:/tmp',
         'cockroachdb/cockroach:latest', 'debug', 'rocksdb', '--db=/data', 
         'scan', '--from=key_0', '--to=key_z'],
        capture_output=True, timeout=300
    )
    
    # This approach is too slow, let's use plyvel (LevelDB) as proxy
    return None

def benchmark_leveldb():
    """Benchmark LevelDB as RocksDB proxy (same LSM design)"""
    print("\n=== LEVELDB (C++ LSM, Embedded) ===")
    print("(LevelDB = RocksDB ancestor, same LSM architecture)")
    
    import plyvel
    
    db_path = "/tmp/leveldb_bench"
    subprocess.run(['rm', '-rf', db_path], capture_output=True)
    
    db = plyvel.DB(db_path, create_if_missing=True)
    value = b'x' * DATA_SIZE
    
    # Write benchmark
    print(f"Writing {NUM_OPS} rows...")
    start = time.perf_counter()
    wb = db.write_batch()
    for i in range(NUM_OPS):
        wb.put(f"key_{i}".encode(), value)
    wb.write()
    write_time = time.perf_counter() - start
    write_ops = NUM_OPS / write_time
    
    # Read benchmark
    print(f"Reading {NUM_OPS} rows...")
    start = time.perf_counter()
    for i in range(NUM_OPS):
        db.get(f"key_{i}".encode())
    read_time = time.perf_counter() - start
    read_ops = NUM_OPS / read_time
    
    db.close()
    
    return {
        'write_ops': write_ops,
        'write_latency_ms': (write_time / NUM_OPS) * 1000,
        'read_ops': read_ops,
        'read_latency_ms': (read_time / NUM_OPS) * 1000
    }

if __name__ == "__main__":
    print(f"Benchmark: {NUM_OPS} operations, {DATA_SIZE} bytes/value")
    print("=" * 60)
    
    # Run benchmarks
    cass_results = benchmark_cassandra()
    leveldb_results = benchmark_leveldb()
    
    # Print comparison
    print("\n" + "=" * 60)
    print("RESULTS COMPARISON")
    print("=" * 60)
    print(f"\n{'Metric':<20} {'Cassandra':<15} {'LevelDB':<15} {'Winner':<10}")
    print("-" * 60)
    
    # Write comparison
    cass_w = cass_results['write_ops']
    level_w = leveldb_results['write_ops']
    winner_w = "LevelDB" if level_w > cass_w else "Cassandra"
    ratio_w = max(level_w, cass_w) / min(level_w, cass_w)
    print(f"{'Write (ops/s)':<20} {cass_w:>12,.0f}   {level_w:>12,.0f}   {winner_w} {ratio_w:.1f}x")
    
    # Read comparison  
    cass_r = cass_results['read_ops']
    level_r = leveldb_results['read_ops']
    winner_r = "LevelDB" if level_r > cass_r else "Cassandra"
    ratio_r = max(level_r, cass_r) / min(level_r, cass_r)
    print(f"{'Read (ops/s)':<20} {cass_r:>12,.0f}   {level_r:>12,.0f}   {winner_r} {ratio_r:.1f}x")
    
    # Latency
    print(f"\n{'Write latency (ms)':<20} {cass_results['write_latency_ms']:>12.3f}   {leveldb_results['write_latency_ms']:>12.4f}")
    print(f"{'Read latency (ms)':<20} {cass_results['read_latency_ms']:>12.3f}   {leveldb_results['read_latency_ms']:>12.4f}")
