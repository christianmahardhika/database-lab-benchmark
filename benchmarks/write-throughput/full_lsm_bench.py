#!/usr/bin/env python3
"""Full LSM comparison: Cassandra (Java) vs ScyllaDB (C++) vs LevelDB (embedded)"""

import time
from cassandra.cluster import Cluster
import plyvel
import subprocess
import os

NUM_OPS = 50000
DATA_SIZE = 100

def bench_cassandra():
    print("\n[1] CASSANDRA (Java LSM, Distributed)")
    cluster = Cluster(['172.17.0.2'])  # cassandra-bench
    session = cluster.connect()
    
    session.execute("CREATE KEYSPACE IF NOT EXISTS bench WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}")
    session.execute("USE bench")
    session.execute("DROP TABLE IF EXISTS test")
    session.execute("CREATE TABLE test (k TEXT PRIMARY KEY, v TEXT)")
    
    insert_stmt = session.prepare("INSERT INTO test (k, v) VALUES (?, ?)")
    select_stmt = session.prepare("SELECT v FROM test WHERE k = ?")
    value = 'x' * DATA_SIZE
    
    start = time.perf_counter()
    for i in range(NUM_OPS):
        session.execute(insert_stmt, [f"key_{i}", value])
    write_time = time.perf_counter() - start
    
    start = time.perf_counter()
    for i in range(NUM_OPS):
        session.execute(select_stmt, [f"key_{i}"])
    read_time = time.perf_counter() - start
    
    cluster.shutdown()
    return NUM_OPS/write_time, NUM_OPS/read_time

def bench_scylladb():
    print("\n[2] SCYLLADB (C++ LSM, Distributed, Cassandra-compatible)")
    # ScyllaDB on port 9043
    cluster = Cluster(['127.0.0.1'], port=9043)
    session = cluster.connect()
    
    session.execute("CREATE KEYSPACE IF NOT EXISTS bench WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}")
    session.execute("USE bench")
    session.execute("DROP TABLE IF EXISTS test")
    session.execute("CREATE TABLE test (k TEXT PRIMARY KEY, v TEXT)")
    
    insert_stmt = session.prepare("INSERT INTO test (k, v) VALUES (?, ?)")
    select_stmt = session.prepare("SELECT v FROM test WHERE k = ?")
    value = 'x' * DATA_SIZE
    
    start = time.perf_counter()
    for i in range(NUM_OPS):
        session.execute(insert_stmt, [f"key_{i}", value])
    write_time = time.perf_counter() - start
    
    start = time.perf_counter()
    for i in range(NUM_OPS):
        session.execute(select_stmt, [f"key_{i}"])
    read_time = time.perf_counter() - start
    
    cluster.shutdown()
    return NUM_OPS/write_time, NUM_OPS/read_time

def bench_leveldb():
    print("\n[3] LEVELDB/ROCKSDB (C++ LSM, Embedded)")
    db_path = "/tmp/leveldb_bench"
    subprocess.run(['rm', '-rf', db_path], capture_output=True)
    
    db = plyvel.DB(db_path, create_if_missing=True)
    value = b'x' * DATA_SIZE
    
    start = time.perf_counter()
    wb = db.write_batch()
    for i in range(NUM_OPS):
        wb.put(f"key_{i}".encode(), value)
    wb.write()
    write_time = time.perf_counter() - start
    
    start = time.perf_counter()
    for i in range(NUM_OPS):
        db.get(f"key_{i}".encode())
    read_time = time.perf_counter() - start
    
    db.close()
    return NUM_OPS/write_time, NUM_OPS/read_time

if __name__ == "__main__":
    print(f"LSM-Tree Benchmark: {NUM_OPS} ops, {DATA_SIZE} bytes/value")
    print("=" * 70)
    
    cass_w, cass_r = bench_cassandra()
    scylla_w, scylla_r = bench_scylladb()
    level_w, level_r = bench_leveldb()
    
    print("\n" + "=" * 70)
    print("RESULTS: LSM-Tree Implementations Compared")
    print("=" * 70)
    print(f"\n{'Database':<35} {'Write (ops/s)':<15} {'Read (ops/s)':<15}")
    print("-" * 70)
    print(f"{'Cassandra 4.1 (Java, distributed)':<35} {cass_w:>12,.0f}   {cass_r:>12,.0f}")
    print(f"{'ScyllaDB 5.4 (C++, distributed)':<35} {scylla_w:>12,.0f}   {scylla_r:>12,.0f}")
    print(f"{'LevelDB (C++, embedded)':<35} {level_w:>12,.0f}   {level_r:>12,.0f}")
    
    print("\n" + "=" * 70)
    print("ANALYSIS")
    print("=" * 70)
    print(f"\nScyllaDB vs Cassandra (same protocol, C++ vs Java):")
    print(f"  Write: ScyllaDB {scylla_w/cass_w:.1f}x faster")
    print(f"  Read:  ScyllaDB {scylla_r/cass_r:.1f}x faster")
    
    print(f"\nEmbedded vs Distributed overhead:")
    print(f"  Write: LevelDB {level_w/scylla_w:.0f}x faster than ScyllaDB")
    print(f"  Read:  LevelDB {level_r/scylla_r:.0f}x faster than ScyllaDB")
