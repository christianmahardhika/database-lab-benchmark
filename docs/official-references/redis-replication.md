# Redis Replication - Official Documentation

Source: https://redis.io/docs/latest/operate/oss_and_stack/management/replication/
Version: Redis Open Source

## Overview

At the base of Redis replication there is a **leader follower** (master-replica) replication that is simple to use and configure. It allows replica Redis instances to be exact copies of master instances.

## Three Main Mechanisms

1. **Streaming replication** — When master and replica are well-connected, master sends a stream of commands to replicate effects on the dataset (client writes, key expiries/evictions, any changes)

2. **Partial resynchronization** — When link breaks (network issues, timeout), replica reconnects and attempts partial resync: obtains only the commands missed during disconnection

3. **Full resynchronization** — When partial sync not possible, replica asks for full sync: master creates snapshot, sends to replica, then continues streaming commands

## Important Facts

- Redis uses **asynchronous replication** by default (low latency, high performance)
- A master can have **multiple replicas**
- Replicas can connect to other replicas (cascading structure)
- Replication is **non-blocking on master side** — master continues handling queries during sync
- Replication is **largely non-blocking on replica side** — can serve old data during initial sync
- **WAIT command** allows optional synchronous replication (but doesn't guarantee CP/strong consistency)

## How Redis Replication Works

Every Redis master has:
- **Replication ID** — large pseudo random string marking dataset history
- **Offset** — increments for every byte of replication stream produced

The pair `(Replication ID, offset)` identifies an **exact version** of the dataset.

### Full Synchronization Process

1. Master starts **background saving process** to produce RDB file
2. Master buffers all new write commands received
3. When background save complete, master **transfers RDB file to replica**
4. Replica saves to disk, then loads into memory
5. Master sends all buffered commands to replica as a stream

### Replication ID Explained

- New replication ID generated when instance restarts as master or replica promoted
- Replicas inherit master's replication ID after handshake
- **Two replication IDs**: main ID and secondary ID
- Secondary ID remembers former master's ID after failover (enables partial resync with new master)

## Diskless Replication

- Available since Redis 2.8.18
- Child process sends RDB **directly over the wire** to replicas
- **No disk as intermediate storage**
- Useful for slow disks

Configuration:
```
repl-diskless-sync yes
repl-diskless-sync-delay 5
```

## Configuration

Basic replication setup:
```
replicaof 192.168.1.1 6379
```

Or via command:
```
REPLICAOF 192.168.1.1 6379
```

Authentication:
```
masterauth <password>
```

## Read-only Replica

- Enabled by default since Redis 2.6
- Controlled by `replica-read-only` option
- Rejects all write commands

**Warning**: Writable replicas can cause inconsistency and are NOT recommended.

## Write Quorum (min-replicas)

Configure master to accept writes only if N replicas connected with lag < M seconds:

```
min-replicas-to-write 3
min-replicas-max-lag 10
```

If conditions not met, master replies with error.

## Key Expiry Handling

1. **Replicas don't expire keys** — wait for master to send DEL command
2. Replicas use logical clock to report expired keys as non-existent for reads
3. During Lua scripts, no key expiries performed (time frozen)

## Partial Sync After Restarts/Failovers

Since Redis 4.0:
- Promoted replica can do partial resync with old master's replicas
- Replicas store resync info in RDB file when shut down gracefully (use SHUTDOWN command)

## Maxmemory on Replicas

- By default, replicas **ignore maxmemory** (eviction handled by master)
- Replica may use more memory than maxmemory setting
- To change: `replica-ignore-maxmemory no`

## Docker/NAT Configuration

Force replica to announce specific IP/port:
```
replica-announce-ip 5.5.5.5
replica-announce-port 1234
```

## Safety Warning: Master Without Persistence

**DANGEROUS**: Master with persistence off + auto-restart can cause data loss:
1. Master A crashes, restarts with empty dataset
2. Replicas B, C sync from empty A
3. All data destroyed

**Always disable auto-restart** if master has persistence turned off.
