# MongoDB Index Builds and DDL Operations - Official Documentation

Source: https://www.mongodb.com/docs/manual/core/index-creation/
Version: MongoDB 8.3

## Overview

MongoDB's "DDL" equivalent operations (index builds, collection modifications) use an **optimized build process** that:
- Holds an **exclusive lock only at the beginning and end** of the index build
- **Yields to interleaving read and write operations** during the rest of the build
- Builds **simultaneously across all data-bearing replica set members**

## Key Differences from SQL Databases

MongoDB is **schema-flexible** by default:
- No formal ALTER TABLE equivalent — documents can have different structures
- Primary DDL-like operations are: **index builds**, **collection renames**, **schema validation changes**
- Index builds are the most impactful "DDL" operation

---

## Index Build Process (Detailed)

### Lock Stages

| Stage | Lock Type | Behavior |
|-------|-----------|----------|
| Start | Exclusive X | Blocks all read/write. Does not yield. |
| Collection Scan | Intent Exclusive IX | Periodically yields to read/write |
| Process Side Writes | Intent Exclusive IX | Periodically yields |
| Finish Side Writes | Shared S | Blocks writes only |
| Commit | Exclusive X | Blocks all read/write. Does not yield. |
| Complete | Released | Normal operations resume |

### The Build Phases

**1. Initialization (Exclusive Lock)**
- Creates initial index metadata entry
- Creates "side writes table" for concurrent writes during build
- Creates "constraint violation table" for documents with invalid keys

**2. Collection Scan (Intent Exclusive Lock - Yields)**
- Scans each document and generates keys
- Dumps keys into external sorter
- Key generation errors stored in constraint violation table
- **Concurrent reads and writes allowed**

**3. Process Side Writes (Intent Exclusive Lock - Yields)**
- Drains side write table (FIFO)
- Handles keys from documents written during build
- Uses snapshot system to limit keys to process
- **Concurrent reads and writes allowed**

**4. Vote and Wait for Commit Quorum (Replica Sets)**
- Member submits "vote" to primary
- Primary waits for commit quorum (default: all voting members)
- Secondary waits for "commitIndexBuild" or "abortIndexBuild" oplog entry
- Continues draining side writes while waiting

**5. Finish Processing (Shared Lock)**
- Drains remaining side writes
- **Blocks writes, allows reads**
- May pause replication

**6. Commit (Exclusive Lock)**
- Applies final operations
- Drops side write table
- Processes constraint violation table (primary only)
- Creates "commitIndexBuild" oplog entry
- Marks index as ready

---

## Locking Behavior Comparison

### Historical Context

| Build Type | Lock Duration | Performance | Concurrent DML |
|------------|---------------|-------------|----------------|
| Foreground (pre-4.2) | Entire build | Fast | **No** |
| Background (pre-4.2) | Yielding | Slower | Yes |
| Optimized (4.2+) | Start/End only | Fast | **Yes** |

**Current behavior (4.2+):**
- Obtains exclusive lock **only at start and end**
- Uses **yielding behavior** during build
- Produces **efficient index data structures** (like foreground)
- **Allows concurrent read-write** (like background)

---

## Constraint Violations During Build

Unlike SQL databases, MongoDB allows **temporary constraint violations** during index build:

```javascript
// Example: Creating unique index on product_sku
db.inventory.createIndex({ product_sku: 1 }, { unique: true })
```

- Build can **start successfully** even with duplicate values
- Duplicates can be **written during build**
- Violations only checked **at commit time**
- If violations exist at commit → **build fails with error**

### Mitigation Strategies
1. Validate no documents violate constraints before build
2. Stop writes from apps that can't guarantee violation-free operations
3. For sharded collections: check all shards first

---

## Replica Set / Sharded Cluster Behavior

### Simultaneous Build Process

1. Primary receives `createIndexes` → creates "startIndexBuild" oplog entry
2. Secondaries start build after replicating oplog entry
3. Each member **"votes" to commit** when finished indexing
4. Secondaries continue processing new writes while waiting
5. Primary checks for quorum of votes
6. Primary checks for constraint violations
7. On success: "commitIndexBuild" oplog entry
8. On failure: "abortIndexBuild" oplog entry

### Commit Quorum

**Default:** `"votingMembers"` (all data-bearing voting members)

```javascript
// Custom commit quorum
db.collection.createIndex(
  { field: 1 },
  { },
  { commitQuorum: "majority" }  // or a number, or "votingMembers"
)
```

**Warning:** If a voting node becomes unreachable with default `votingMembers`, index builds can **hang until that node comes back online**.

### Commit Quorum vs Write Concern

| Aspect | Commit Quorum | Write Concern |
|--------|---------------|---------------|
| Used for | Index builds | Write operations |
| Specifies | How many nodes must be **ready to commit** | How many nodes must **acknowledge write** |
| Timing | Before primary commits | After primary commits |

---

## Build Failure and Recovery

### Clean Shutdown (MongoDB 4.4+)
- Progress is **saved to disk**
- Automatically **recovers and continues** from checkpoint on restart
- Works for both primary and secondary

### Standalone Shutdown
- **All progress is lost**
- Must re-issue `createIndex()` operation

### Rollback (MongoDB 5.0+)
- Progress saved to disk
- If rollback doesn't revert build → **resumes from checkpoint**
- If rollback reverts build → must **re-create index**

---

## Performance Considerations

### Memory Usage

```javascript
// Default: 200MB per createIndexes command
// Shared among all indexes in that command

// Example: 10 indexes = 20MB each
db.collection.createIndexes([
  { field1: 1 },
  { field2: 1 },
  // ... 8 more
])  // Each gets ~20MB

// Adjust with parameter (rare):
db.adminCommand({ setParameter: 1, maxIndexBuildMemoryUsageMegabytes: 500 })
```

When memory limit is reached → uses temporary files in `--dbpath/_tmp`

### Concurrent Build Limits

```javascript
// Default: 3 concurrent index builds
// Check/modify with:
db.adminCommand({ getParameter: 1, maxNumActiveUserIndexBuilds: 1 })
db.adminCommand({ setParameter: 1, maxNumActiveUserIndexBuilds: 5 })
```

### Write-Heavy Workloads
- Can result in **reduced write performance**
- Consider **maintenance window** with reduced writes
- Index builds may take longer due to side write processing

---

## Monitoring Index Builds

```javascript
// Check current operations
db.currentOp({ "command.createIndexes": { $exists: true } })

// Look for progress in msg field:
// "Index Build: inserted 1000 keys from external sorter into index in 5 seconds"

// Check for metadata lock waits
db.currentOp({ "waitingForLock": true })
```

### Log Messages for Stopped/Resumed Builds
```
"msg":"Index build: wrote resumable state to disk"
"msg":"Found index from unfinished build"
```

---

## Terminating Index Builds

**Correct method:**
```javascript
db.collection.dropIndex("index_name")
// or
db.collection.dropIndexes()
```

**DO NOT use `killOp`** to terminate index builds in replica sets or sharded clusters!

---

## Sharded Collection Index Consistency

Inconsistent indexes can occur when:
- Unique index creation succeeds on some shards but fails on others (due to duplicates)
- Rolling index builds fail or have different specifications

**Check for inconsistencies:**
```javascript
// On config server primary
db.adminCommand({ serverStatus: 1 }).shardedIndexConsistency

// Or use the aggregation to find inconsistent indexes
// (see MongoDB docs for full script)
```

**Config server periodically checks** for index inconsistencies across shards.

---

## MongoDB 7.1+ Improvements

| Feature | MongoDB 7.1+ | Earlier Versions |
|---------|--------------|------------------|
| Error reporting | **Immediate** during collection scan | End of build (commit phase) |
| Secondary crash on error | **No** (requests primary to stop) | Yes |
| Disk space management | **Auto-stop** if below threshold | No protection |

```javascript
// Set minimum disk space for index builds (7.1+)
db.adminCommand({ 
  setParameter: 1, 
  indexBuildMinAvailableDiskSpaceMB: 10000 
})
```

---

## Summary: MongoDB vs SQL DDL

| Aspect | MongoDB | PostgreSQL/MySQL |
|--------|---------|------------------|
| Schema changes | Flexible (no ALTER TABLE needed) | Requires ALTER TABLE |
| Add column | Just insert with new field | ALTER TABLE ADD COLUMN |
| Index creation | `createIndex()` - online by default | CREATE INDEX [CONCURRENTLY] |
| Locking | Exclusive only at start/end | Varies by operation |
| Constraint checking | At commit time | During/after operation |
| Concurrent DML during index | Yes (default since 4.2) | Depends on method |
| Distributed builds | Simultaneous across replica set | Manual or tool-assisted |
