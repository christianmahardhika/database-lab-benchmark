---
name: compare-databases
description: Help choose the right database for a specific production use case by mapping requirements to database characteristics and citing lab benchmark data. Use when comparing databases or making technology selection decisions.
---

## Workflow

### Step 1: Gather Requirements
Ask the user about their workload:
- Read/write ratio (e.g., 90/10 read-heavy, 50/50 mixed, 10/90 write-heavy)
- Data model (relational with JOINs, document/nested, key-value, time-series, wide-column)
- Consistency requirements (strong ACID, eventual, tunable)
- Scale (data size, QPS, concurrent connections)
- Latency requirements (sub-ms, <10ms, <100ms acceptable)
- Operational constraints (team expertise, cloud vs self-hosted, budget)

### Step 2: Map to Database Characteristics
Reference `#database-characteristics` steering file to match requirements against:
- Storage engine strengths/weaknesses
- Best-fit use cases
- Production sizing guidance

### Step 3: Cite Lab Evidence
Reference actual benchmark results from `results/` directory:
- Write throughput numbers (Exp 1)
- Read latency percentiles (Exp 2)
- Scaling characteristics (Exp 5)
- Any relevant hypothesis verdicts

### Step 4: Provide Recommendation
Structure as:
1. **Primary recommendation** with rationale
2. **Runner-up** with tradeoff explanation
3. **Avoid** with reasons why other options don't fit
4. **Migration path** if they outgrow the choice

### Step 5: Production Readiness Checklist
Based on `#production-operations` steering file:
- HA configuration needed
- Backup strategy
- Monitoring setup
- Connection pooling
- Capacity planning formula

## Decision Framework

```
Is your data relational (JOINs needed)?
├── YES → Do you need complex analytics?
│         ├── YES → PostgreSQL (+ ClickHouse for OLAP offload)
│         └── NO → Is it simple OLTP?
│                   ├── YES → MySQL (faster for simple patterns)
│                   └── NO → PostgreSQL (more capable)
└── NO → Is it time-series?
          ├── YES → Do you need SQL JOINs with metadata?
          │         ├── YES → TimescaleDB
          │         └── NO → Is write volume extreme (>100K/s)?
          │                   ├── YES → ScyllaDB (TWCS)
          │                   └── NO → TimescaleDB or ClickHouse
          └── NO → Is it key-value with sub-ms requirement?
                    ├── YES → Redis (with size caveat)
                    └── NO → Is schema highly variable?
                              ├── YES → MongoDB
                              └── NO → Is it write-heavy (>80% writes)?
                                        ├── YES → ScyllaDB
                                        └── NO → PostgreSQL (default safe choice)
```

## Anti-Patterns to Flag
- Using MongoDB just because "it's flexible" when you actually need JOINs
- Using PostgreSQL for >1M writes/sec when ScyllaDB would be simpler
- Using Redis as primary datastore (data loss risk, RAM cost)
- Using ClickHouse for OLTP workloads (it's OLAP-optimized)
- Running all workloads on one database when polyglot persistence is appropriate
