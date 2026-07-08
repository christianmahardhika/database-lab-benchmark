#!/usr/bin/env python3
"""
B-Tree vs LSM-Tree Benchmark
PostgreSQL (B-Tree) vs ScyllaDB (LSM-Tree)
"""
import time
import uuid
import random
import string
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime

# Results storage
results = {}

def random_string(length=20):
    return ''.join(random.choices(string.ascii_letters, k=length))

# ============ PostgreSQL (B-Tree) ============
def benchmark_postgres():
    import psycopg2
    from psycopg2.extras import execute_batch
    
    conn = psycopg2.connect(
        host='localhost', port=5437, 
        user='bench', password='bench', 
        database='bench'
    )
    cur = conn.cursor()
    
    # Setup
    cur.execute("DROP TABLE IF EXISTS test_data")
    cur.execute("""
        CREATE TABLE test_data (
            id UUID PRIMARY KEY,
            name VARCHAR(100),
            email VARCHAR(255),
            value INT,
            created_at TIMESTAMP
        )
    """)
    conn.commit()
    
    # 1. Write throughput (100k rows)
    print("\n[PostgreSQL] Write throughput test...")
    data = [(str(uuid.uuid4()), random_string(), f"{random_string()}@test.com", 
             random.randint(1, 10000), datetime.now()) for _ in range(100000)]
    
    start = time.time()
    execute_batch(cur, 
        "INSERT INTO test_data (id, name, email, value, created_at) VALUES (%s, %s, %s, %s, %s)",
        data, page_size=1000)
    conn.commit()
    write_time = time.time() - start
    write_tps = 100000 / write_time
    print(f"  Write: {write_time:.2f}s ({write_tps:.0f} rows/sec)")
    results['pg_write_time'] = write_time
    results['pg_write_tps'] = write_tps
    
    # Create index for fair read comparison
    cur.execute("CREATE INDEX idx_value ON test_data(value)")
    conn.commit()
    
    # 2. Point query (1000 random lookups)
    print("[PostgreSQL] Point query test...")
    sample_ids = [d[0] for d in random.sample(data, 1000)]
    
    start = time.time()
    for sid in sample_ids:
        cur.execute("SELECT * FROM test_data WHERE id = %s", (sid,))
        cur.fetchone()
    point_time = time.time() - start
    point_avg = (point_time / 1000) * 1000  # ms
    print(f"  Point query: {point_avg:.3f}ms avg")
    results['pg_point_ms'] = point_avg
    
    # 3. Range scan (value BETWEEN)
    print("[PostgreSQL] Range scan test...")
    start = time.time()
    for _ in range(100):
        v = random.randint(1, 9000)
        cur.execute("SELECT * FROM test_data WHERE value BETWEEN %s AND %s", (v, v+1000))
        cur.fetchall()
    range_time = time.time() - start
    range_avg = (range_time / 100) * 1000
    print(f"  Range scan: {range_avg:.2f}ms avg")
    results['pg_range_ms'] = range_avg
    
    # 4. Mixed workload (80% read, 20% write)
    print("[PostgreSQL] Mixed workload (80/20)...")
    start = time.time()
    for i in range(1000):
        if random.random() < 0.8:
            cur.execute("SELECT * FROM test_data WHERE id = %s", (random.choice(sample_ids),))
            cur.fetchone()
        else:
            cur.execute("INSERT INTO test_data VALUES (%s, %s, %s, %s, %s)",
                       (str(uuid.uuid4()), random_string(), f"{random_string()}@test.com",
                        random.randint(1,10000), datetime.now()))
    conn.commit()
    mixed_time = time.time() - start
    mixed_tps = 1000 / mixed_time
    print(f"  Mixed: {mixed_time:.2f}s ({mixed_tps:.0f} ops/sec)")
    results['pg_mixed_tps'] = mixed_tps
    
    cur.close()
    conn.close()

# ============ ScyllaDB (LSM-Tree) ============
def benchmark_scylla():
    from cassandra.cluster import Cluster
    from cassandra.query import BatchStatement, SimpleStatement
    from cassandra import ConsistencyLevel
    
    cluster = Cluster(['localhost'], port=9043)
    session = cluster.connect('benchmark')
    
    # Truncate for clean test
    session.execute("TRUNCATE test_data")
    
    # 1. Write throughput (100k rows)
    print("\n[ScyllaDB] Write throughput test...")
    
    insert_stmt = session.prepare(
        "INSERT INTO test_data (id, name, email, value, created_at) VALUES (?, ?, ?, ?, ?)"
    )
    
    data = [(uuid.uuid4(), random_string(), f"{random_string()}@test.com",
             random.randint(1, 10000), datetime.now()) for _ in range(100000)]
    
    start = time.time()
    # Batch insert in chunks
    batch_size = 50
    for i in range(0, len(data), batch_size):
        batch = BatchStatement(consistency_level=ConsistencyLevel.ONE)
        for row in data[i:i+batch_size]:
            batch.add(insert_stmt, row)
        session.execute(batch)
    write_time = time.time() - start
    write_tps = 100000 / write_time
    print(f"  Write: {write_time:.2f}s ({write_tps:.0f} rows/sec)")
    results['scylla_write_time'] = write_time
    results['scylla_write_tps'] = write_tps
    
    # Create index
    try:
        session.execute("CREATE INDEX ON test_data(value)")
        time.sleep(2)  # Wait for index build
    except:
        pass
    
    # 2. Point query (1000 random lookups)
    print("[ScyllaDB] Point query test...")
    sample_ids = [d[0] for d in random.sample(data, 1000)]
    select_stmt = session.prepare("SELECT * FROM test_data WHERE id = ?")
    
    start = time.time()
    for sid in sample_ids:
        session.execute(select_stmt, (sid,))
    point_time = time.time() - start
    point_avg = (point_time / 1000) * 1000
    print(f"  Point query: {point_avg:.3f}ms avg")
    results['scylla_point_ms'] = point_avg
    
    # 3. Range scan (need ALLOW FILTERING for non-partition key)
    print("[ScyllaDB] Range scan test...")
    start = time.time()
    for _ in range(100):
        v = random.randint(1, 9000)
        session.execute(f"SELECT * FROM test_data WHERE value >= {v} AND value <= {v+1000} ALLOW FILTERING")
    range_time = time.time() - start
    range_avg = (range_time / 100) * 1000
    print(f"  Range scan: {range_avg:.2f}ms avg (with ALLOW FILTERING)")
    results['scylla_range_ms'] = range_avg
    
    # 4. Mixed workload
    print("[ScyllaDB] Mixed workload (80/20)...")
    start = time.time()
    for i in range(1000):
        if random.random() < 0.8:
            session.execute(select_stmt, (random.choice(sample_ids),))
        else:
            session.execute(insert_stmt, (uuid.uuid4(), random_string(), 
                           f"{random_string()}@test.com", random.randint(1,10000), datetime.now()))
    mixed_time = time.time() - start
    mixed_tps = 1000 / mixed_time
    print(f"  Mixed: {mixed_time:.2f}s ({mixed_tps:.0f} ops/sec)")
    results['scylla_mixed_tps'] = mixed_tps
    
    cluster.shutdown()

def print_comparison():
    print("\n" + "="*60)
    print("BENCHMARK RESULTS: B-Tree (PostgreSQL) vs LSM-Tree (ScyllaDB)")
    print("="*60)
    print(f"\n{'Metric':<25} {'PostgreSQL':<15} {'ScyllaDB':<15} {'Winner':<15}")
    print("-"*70)
    
    # Write throughput
    pg_w = results.get('pg_write_tps', 0)
    sc_w = results.get('scylla_write_tps', 0)
    winner = "ScyllaDB ✓" if sc_w > pg_w else "PostgreSQL ✓"
    print(f"{'Write (rows/sec)':<25} {pg_w:<15.0f} {sc_w:<15.0f} {winner}")
    
    # Point query
    pg_p = results.get('pg_point_ms', 0)
    sc_p = results.get('scylla_point_ms', 0)
    winner = "PostgreSQL ✓" if pg_p < sc_p else "ScyllaDB ✓"
    print(f"{'Point Query (ms)':<25} {pg_p:<15.3f} {sc_p:<15.3f} {winner}")
    
    # Range scan
    pg_r = results.get('pg_range_ms', 0)
    sc_r = results.get('scylla_range_ms', 0)
    winner = "PostgreSQL ✓" if pg_r < sc_r else "ScyllaDB ✓"
    print(f"{'Range Scan (ms)':<25} {pg_r:<15.2f} {sc_r:<15.2f} {winner}")
    
    # Mixed
    pg_m = results.get('pg_mixed_tps', 0)
    sc_m = results.get('scylla_mixed_tps', 0)
    winner = "ScyllaDB ✓" if sc_m > pg_m else "PostgreSQL ✓"
    print(f"{'Mixed 80/20 (ops/sec)':<25} {pg_m:<15.0f} {sc_m:<15.0f} {winner}")
    
    print("\n" + "="*60)
    print("KEY INSIGHTS:")
    print("-"*60)
    if sc_w > pg_w:
        ratio = sc_w / pg_w if pg_w > 0 else 0
        print(f"• LSM-Tree write: {ratio:.1f}x faster (sequential WAL + memtable)")
    if pg_p < sc_p:
        print(f"• B-Tree point query: faster (single seek vs multi-level bloom)")
    if pg_r < sc_r:
        ratio = sc_r / pg_r if pg_r > 0 else 0
        print(f"• B-Tree range scan: {ratio:.1f}x faster (sorted pages vs scattered SSTables)")

if __name__ == '__main__':
    print("B-Tree vs LSM-Tree Benchmark")
    print("PostgreSQL 16 vs ScyllaDB 5.4")
    print("100,000 rows test data\n")
    
    try:
        benchmark_postgres()
    except Exception as e:
        print(f"PostgreSQL error: {e}")
    
    try:
        benchmark_scylla()
    except Exception as e:
        print(f"ScyllaDB error: {e}")
    
    print_comparison()
