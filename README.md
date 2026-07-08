# Database Lab Benchmark

Comprehensive benchmark suite for comparing database performance across different storage engines and use cases.

## 🎯 Goal

Validate vendor claims (hypotheses) with actual benchmark data:
- **B-Tree**: PostgreSQL, MySQL, MongoDB (WiredTiger)
- **LSM-Tree**: ScyllaDB, Cassandra, RocksDB, LevelDB
- **Columnar**: ClickHouse
- **Time-Series**: TimescaleDB
- **In-Memory**: Redis

## 📊 Experiments

| # | Experiment | Databases | Validates |
|---|------------|-----------|-----------|
| 1 | Write Throughput | All | H1, H5, H17, H18 |
| 2 | Read Latency | All | H3, H11, H13 |
| 3 | Complex Query | PG, MySQL, Mongo | H1, H9, H10 |
| 4 | Schema Evolution | PG, MySQL, Mongo, Scylla | H9, H24 |
| 5 | Horizontal Scaling | Scylla, Mongo, Citus | H7, H8, H28 |
| 6 | Write Amplification | PG, Scylla, RocksDB | H19 |
| 7 | Time-Series | Timescale, Mongo, Scylla, ClickHouse | H6, H31, H40 |
| 8 | MySQL vs PostgreSQL | MySQL, PG | H21-H25 |
| 9 | MongoDB vs SQL | Mongo, PG, MySQL | H26-H30 |
| 10 | OLAP | ClickHouse, PG | H36-H39 |
| 11 | Time-Series Battle | Timescale, Mongo TS, Scylla, ClickHouse | H40-H43 |

## 🚀 Quick Start

### Option A: Local (4GB+ RAM available)

```bash
# Run single experiment
cd benchmarks/write-throughput
docker-compose up -d postgres
python3 benchmark_postgres.py
docker-compose down
```

### Option B: EC2 (Recommended for full suite)

```bash
# 1. Setup infrastructure
cd terraform
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your SSH key name

terraform init
terraform apply

# 2. SSH to instance
ssh -i ~/.ssh/your-key.pem ubuntu@<instance-ip>

# 3. Run benchmarks
cd database-lab-benchmark
./scripts/run_all.sh
```

## 💰 Cost Estimate

| Instance | Spec | Hourly | 4 Hours |
|----------|------|--------|---------|
| m6i.2xlarge (on-demand) | 8 vCPU, 32GB | $0.384 | ~$1.50 |
| m6i.2xlarge (spot) | 8 vCPU, 32GB | $0.12 | ~$0.50 |
| m6i.4xlarge (on-demand) | 16 vCPU, 64GB | $0.768 | ~$3.00 |

## 📁 Structure

```
database-lab-benchmark/
├── terraform/           # EC2 infrastructure
│   ├── main.tf
│   └── terraform.tfvars.example
├── docker/              # Docker Compose per database
│   ├── postgres/
│   ├── mysql/
│   ├── mongodb/
│   ├── redis/
│   ├── scylladb/
│   ├── cassandra/
│   ├── clickhouse/
│   └── timescaledb/
├── benchmarks/          # Experiment scripts
│   ├── write-throughput/
│   ├── read-latency/
│   ├── complex-query/
│   └── ...
├── scripts/             # Helper scripts
│   └── run_all.sh
├── results/             # Benchmark outputs
└── docs/                # Documentation
```

## 📋 Hypotheses (H1-H43)

Full list of vendor claims being tested. See [Notion page](https://app.notion.so/375cd5f2e4f181c2bb00c8d9fd7a277f) for details.

### Key Hypotheses

| ID | Claim | Source |
|----|-------|--------|
| H1 | PostgreSQL best for complex OLTP | postgresql.org |
| H5 | ScyllaDB handles >10K writes/sec | scylladb.com |
| H21 | MySQL best for web apps | mysql.com |
| H26 | MongoDB flexible schema | mongodb.com |
| H36 | ClickHouse 100M+ rows/sec scan | clickhouse.com |

## 🔬 Previous Results

From local benchmark (June 2026):

| Metric | PostgreSQL | ScyllaDB | Ratio |
|--------|------------|----------|-------|
| Write TPS | 1,920 | 9,751 | 5.1x |
| Write Latency | 2.08ms | 0.4ms | 5.2x |
| Read TPS | 14,877 | 11,976 | 0.8x |
| Write Amplification | 1.6x | 9.2x | 5.8x |

## 📖 References

- [DDIA Chapter 3: Storage & Retrieval](https://dataintensive.net/)
- [PostgreSQL Docs](https://postgresql.org/docs/)
- [ScyllaDB Docs](https://docs.scylladb.com/)
- [MongoDB Docs](https://docs.mongodb.com/)

## License

MIT
