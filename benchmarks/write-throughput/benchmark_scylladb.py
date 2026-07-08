#!/usr/bin/env python3
"""ScyllaDB Operations Benchmark"""
from cassandra.cluster import Cluster
from cassandra.policies import DCAwareRoundRobinPolicy
import time
import json
import subprocess
from datetime import datetime

RESULTS = {}

def get_cluster():
    return Cluster(
        ['127.0.0.1'],
        port=9043,
        load_balancing_policy=DCAwareRoundRobinPolicy(local_dc='datacenter1')
    )

def benchmark_basic_ops():
    """Test basic read/write operations"""
    print("\n=== 1. Basic Operations ===")
    
    cluster = get_cluster()
    session = cluster.connect()
    
    # Setup
    session.execute("""
        CREATE KEYSPACE IF NOT EXISTS bench 
        WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
    """)
    session.execute("USE bench")
    session.execute("DROP TABLE IF EXISTS test_ops")
    session.execute("""
        CREATE TABLE test_ops (
            id int PRIMARY KEY,
            data text
        )
    """)
    
    # Write benchmark
    NUM_KEYS = 10000
    insert = session.prepare("INSERT INTO test_ops (id, data) VALUES (?, ?)")
    
    start = time.time()
    for i in range(NUM_KEYS):
        session.execute(insert, (i, f"value_{i}"))
    write_time = time.time() - start
    write_ops = NUM_KEYS / write_time
    print(f"Write: {NUM_KEYS} keys in {write_time:.2f}s = {write_ops:.0f} ops/s")
    
    # Read benchmark
    select = session.prepare("SELECT * FROM test_ops WHERE id = ?")
    start = time.time()
    for i in range(NUM_KEYS):
        session.execute(select, (i,))
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
    
    cluster.shutdown()

def benchmark_batch_write():
    """Test batch write performance"""
    print("\n=== 2. Batch Write ===")
    
    cluster = get_cluster()
    session = cluster.connect()
    session.execute("USE bench")
    session.execute("DROP TABLE IF EXISTS test_batch")
    session.execute("""
        CREATE TABLE test_batch (
            id int PRIMARY KEY,
            data text
        )
    """)
    
    from cassandra.query import BatchStatement, BatchType
    
    NUM_KEYS = 10000
    BATCH_SIZE = 100
    
    start = time.time()
    for batch_start in range(0, NUM_KEYS, BATCH_SIZE):
        batch = BatchStatement(batch_type=BatchType.UNLOGGED)
        for i in range(batch_start, min(batch_start + BATCH_SIZE, NUM_KEYS)):
            batch.add(
                session.prepare("INSERT INTO test_batch (id, data) VALUES (?, ?)"),
                (i, f"value_{i}")
            )
        session.execute(batch)
    batch_time = time.time() - start
    batch_ops = NUM_KEYS / batch_time
    print(f"Batch write: {NUM_KEYS} keys in {batch_time:.2f}s = {batch_ops:.0f} ops/s")
    
    RESULTS["batch_write"] = {
        "num_keys": NUM_KEYS,
        "batch_size": BATCH_SIZE,
        "time_s": round(batch_time, 3),
        "ops_per_s": round(batch_ops)
    }
    
    cluster.shutdown()

def benchmark_nodetool_snapshot():
    """Test nodetool snapshot backup"""
    print("\n=== 3. Nodetool Snapshot Backup ===")
    
    snapshot_name = f"bench_snap_{int(time.time())}"
    
    start = time.time()
    result = subprocess.run(
        ["docker", "exec", "scylla-1", "nodetool", "snapshot", "-t", snapshot_name, "bench"],
        capture_output=True, text=True
    )
    snapshot_time = time.time() - start
    
    if result.returncode == 0:
        print(f"Snapshot created: {snapshot_name}")
        print(f"Snapshot time: {snapshot_time*1000:.1f}ms")
        
        RESULTS["snapshot"] = {
            "snapshot_name": snapshot_name,
            "snapshot_time_ms": round(snapshot_time * 1000, 1),
            "success": True
        }
        
        subprocess.run(
            ["docker", "exec", "scylla-1", "nodetool", "clearsnapshot", "-t", snapshot_name],
            capture_output=True
        )
    else:
        print(f"Snapshot failed: {result.stderr}")
        RESULTS["snapshot"] = {"error": result.stderr, "success": False}

def benchmark_compaction_strategies():
    """Compare compaction strategies"""
    print("\n=== 4. Compaction Strategy Comparison ===")
    
    cluster = get_cluster()
    session = cluster.connect()
    session.execute("USE bench")
    
    strategies = {
        "STCS": "SizeTieredCompactionStrategy",
        "LCS": "LeveledCompactionStrategy",
    }
    
    results = {}
    
    for name, strategy in strategies.items():
        print(f"\nTesting {name}...")
        
        table_name = f"test_{name.lower()}"
        session.execute(f"DROP TABLE IF EXISTS {table_name}")
        
        try:
            session.execute(f"""
                CREATE TABLE {table_name} (
                    id int PRIMARY KEY,
                    data text
                ) WITH compaction = {{'class': '{strategy}'}}
            """)
            
            insert = session.prepare(f"INSERT INTO {table_name} (id, data) VALUES (?, ?)")
            start = time.time()
            for i in range(5000):
                session.execute(insert, (i, f"value_{i}" * 10))
            write_time = time.time() - start
            
            subprocess.run(
                ["docker", "exec", "scylla-1", "nodetool", "flush", "bench", table_name],
                capture_output=True
            )
            
            compact_start = time.time()
            subprocess.run(
                ["docker", "exec", "scylla-1", "nodetool", "compact", "bench", table_name],
                capture_output=True
            )
            compact_time = time.time() - compact_start
            
            results[name] = {
                "write_time_s": round(write_time, 3),
                "compaction_time_s": round(compact_time, 3),
                "success": True
            }
            print(f"  Write: {write_time:.2f}s, Compaction: {compact_time:.2f}s")
            
        except Exception as e:
            results[name] = {"error": str(e), "success": False}
            print(f"  Failed: {e}")
    
    RESULTS["compaction_strategies"] = results
    cluster.shutdown()

if __name__ == "__main__":
    print("=" * 50)
    print("ScyllaDB Operations Benchmark")
    print(f"Date: {datetime.now().isoformat()}")
    print("=" * 50)
    
    benchmark_basic_ops()
    benchmark_batch_write()
    benchmark_nodetool_snapshot()
    benchmark_compaction_strategies()
    
    print("\n" + "=" * 50)
    print("RESULTS SUMMARY")
    print("=" * 50)
    print(json.dumps(RESULTS, indent=2, default=str))
    
    with open("RESULTS.json", "w") as f:
        json.dump(RESULTS, f, indent=2, default=str)
    print("\nResults saved to RESULTS.json")
