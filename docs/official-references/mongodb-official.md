# MongoDB Official Documentation Reference

Source: https://www.mongodb.com/docs/

## Overview

MongoDB is a document-oriented NoSQL database. Stores data as BSON (Binary JSON) documents. Default storage engine is WiredTiger (B-Tree based).

## Storage Engine: WiredTiger (B-Tree)

### Architecture
- **Document**: BSON object (JSON-like)
- **Collection**: Group of documents (like table)
- **Database**: Group of collections
- **WiredTiger**: B-Tree storage with compression

### WiredTiger Internals
- **In-memory**: B-Tree cache
- **On-disk**: B-Tree + checkpoints
- **Journal**: Write-ahead log (100ms default sync)
- **Compression**: Snappy (default), zstd, zlib

### Write Path
```
Insert → Memory (B-Tree cache) → Journal → Checkpoint (60s) → Disk
```

### Read Path
```
Find → Check cache → Read from disk if miss → Decompress → Return
```

## Performance Characteristics (Official Claims)

### From mongodb.com
- **Developer productivity**: Flexible schema
- **Horizontal scaling**: Native sharding
- **Performance**: "Sub-millisecond latency"

### Scaling
- **Vertical**: Single node performance
- **Horizontal**: Sharding with mongos router

## Best Use Cases (from docs.mongodb.com)

1. **Flexible Schema Applications**
   - Rapidly evolving requirements
   - Polymorphic data
   - Prototyping

2. **Content Management**
   - CMS platforms
   - Product catalogs with varying attributes

3. **Real-time Analytics**
   - Event logging
   - Time-series (5.0+ native support)
   - IoT data

4. **Mobile Applications**
   - MongoDB Realm sync
   - Offline-first apps

## Anti-Patterns (from docs)

1. **Heavy JOINs**
   - $lookup is expensive
   - Denormalize instead

2. **Strong multi-document transactions**
   - Supported since 4.0
   - But performance penalty

3. **Small, simple data**
   - Document overhead
   - Relational may be simpler

## Configuration for Benchmarks

### WiredTiger Cache
```yaml
# /etc/mongod.conf
storage:
  wiredTiger:
    engineConfig:
      cacheSizeGB: 4  # 50% of RAM minus 1GB
```

### Journal
```yaml
storage:
  journal:
    enabled: true
    commitIntervalMs: 100
```

### Write Concern
```javascript
// Client-side
db.collection.insertOne(doc, { writeConcern: { w: 1, j: false } })  // Fast
db.collection.insertOne(doc, { writeConcern: { w: "majority", j: true } })  // Safe
```

### Read Preference
```javascript
// Read from secondary for scaling
db.collection.find().readPref("secondaryPreferred")
```

## Key Metrics to Monitor

| Metric | Description | Target |
|--------|-------------|--------|
| wiredTiger.cache.bytes | Cache usage | <80% of configured |
| opcounters.query/insert/update | Operation rates | Baseline |
| globalLock.currentQueue | Lock contention | Low |
| repl.lag | Replication delay | <1s |

## Hypothesis to Test

| ID | Claim | Test |
|----|-------|------|
| H26 | Flexible schema | Exp 4, 9: Schema changes |
| H27 | Nested data locality | Exp 9: Read latency |
| H28 | Horizontal scaling | Exp 5: Sharding |
| H29 | Geospatial queries | N/A (not in scope) |
| H30 | Change streams | N/A (not in scope) |
| H31 | Time-series (5.0+) | Exp 11: Time-series battle |

## Aggregation Pipeline vs SQL

```javascript
// MongoDB
db.orders.aggregate([
  { $match: { status: "completed" } },
  { $group: { _id: "$customer_id", total: { $sum: "$amount" } } },
  { $sort: { total: -1 } },
  { $limit: 10 }
])

// SQL equivalent
SELECT customer_id, SUM(amount) as total
FROM orders
WHERE status = 'completed'
GROUP BY customer_id
ORDER BY total DESC
LIMIT 10;
```

## Time-Series Collections (MongoDB 5.0+)

```javascript
db.createCollection("sensor_data", {
   timeseries: {
      timeField: "timestamp",
      metaField: "metadata",
      granularity: "seconds"
   },
   expireAfterSeconds: 86400 * 30  // 30 day TTL
})
```

## References

- Documentation: https://www.mongodb.com/docs/manual/
- WiredTiger: https://www.mongodb.com/docs/manual/core/wiredtiger/
- Performance: https://www.mongodb.com/docs/manual/administration/production-notes/
- Sharding: https://www.mongodb.com/docs/manual/sharding/
- Time Series: https://www.mongodb.com/docs/manual/core/timeseries-collections/
