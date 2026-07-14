# Database Lab Benchmark - EC2 Infrastructure
# Usage: terraform init && terraform apply -var="key_name=your-key"

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile
}

# Variables
variable "aws_profile" {
  description = "AWS CLI profile name"
  default     = "terraform"
}

variable "aws_region" {
  description = "AWS region"
  default     = "ap-southeast-1" # Singapore - closest to Indonesia
}

variable "instance_type" {
  description = "EC2 instance type"
  default     = "m6i.2xlarge" # 8 vCPU, 32GB RAM - recommended for benchmarks
}

variable "key_name" {
  description = "SSH key pair name"
  type        = string
}

variable "spot_instance" {
  description = "Use spot instance for cost savings (60-70% cheaper)"
  default     = false
}

variable "volume_size" {
  description = "Root volume size in GB"
  default     = 300
}

variable "volume_iops" {
  description = "gp3 volume IOPS"
  default     = 3000
}

# Data sources
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_vpc" "default" {
  default = true
}

# Security Group
resource "aws_security_group" "benchmark" {
  name        = "database-benchmark-sg"
  description = "Security group for database benchmark lab"
  vpc_id      = data.aws_vpc.default.id

  # SSH
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH access"
  }

  # PostgreSQL / pgvector / TimescaleDB
  ingress {
    from_port   = 5432
    to_port     = 5439
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "PostgreSQL / pgvector / TimescaleDB"
  }

  # MySQL
  ingress {
    from_port   = 3306
    to_port     = 3309
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "MySQL"
  }

  # MongoDB
  ingress {
    from_port   = 27017
    to_port     = 27020
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "MongoDB (standalone + sharded router)"
  }

  # Redis / Valkey / DragonflyDB
  ingress {
    from_port   = 6379
    to_port     = 6384
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Redis / Valkey / DragonflyDB"
  }

  # Redis Cluster
  ingress {
    from_port   = 7001
    to_port     = 7006
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Redis Cluster nodes"
  }

  # Redis/Valkey Sentinel
  ingress {
    from_port   = 26379
    to_port     = 26384
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Redis/Valkey Sentinel"
  }

  # ScyllaDB/Cassandra CQL
  ingress {
    from_port   = 9042
    to_port     = 9044
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "CQL (Cassandra/ScyllaDB)"
  }

  # ClickHouse HTTP
  ingress {
    from_port   = 8123
    to_port     = 8126
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "ClickHouse HTTP (cluster ports)"
  }

  # ClickHouse Native
  ingress {
    from_port   = 9000
    to_port     = 9000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "ClickHouse Native"
  }

  # CockroachDB SQL + HTTP UI
  ingress {
    from_port   = 26257
    to_port     = 26259
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "CockroachDB SQL (cluster)"
  }

  ingress {
    from_port   = 8080
    to_port     = 8082
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "CockroachDB HTTP UI"
  }

  # Elasticsearch
  ingress {
    from_port   = 9200
    to_port     = 9202
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Elasticsearch HTTP (cluster)"
  }

  # Neo4j Bolt + HTTP
  ingress {
    from_port   = 7474
    to_port     = 7477
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Neo4j HTTP (cluster)"
  }

  ingress {
    from_port   = 7687
    to_port     = 7690
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Neo4j Bolt (cluster)"
  }

  # InfluxDB
  ingress {
    from_port   = 8086
    to_port     = 8086
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "InfluxDB HTTP"
  }

  # Prometheus
  ingress {
    from_port   = 9090
    to_port     = 9091
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Prometheus HTTP + Milvus metrics"
  }

  # Milvus gRPC
  ingress {
    from_port   = 19530
    to_port     = 19530
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Milvus gRPC"
  }

  # Qdrant REST + gRPC
  ingress {
    from_port   = 6333
    to_port     = 6337
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Qdrant REST (cluster)"
  }

  ingress {
    from_port   = 6334
    to_port     = 6334
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Qdrant gRPC"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name    = "database-benchmark-sg"
    Project = "database-lab-benchmark"
  }
}

# User data script to setup the instance
locals {
  user_data = <<-EOF
    #!/bin/bash
    set -e
    
    # Update system
    apt-get update && apt-get upgrade -y
    
    # Install Docker
    curl -fsSL https://get.docker.com | sh
    usermod -aG docker ubuntu
    
    # Install Docker Compose
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    
    # Install Go 1.22
    wget -q https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
    rm go1.22.5.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/ubuntu/.bashrc
    echo 'export GOPATH=/home/ubuntu/go' >> /home/ubuntu/.bashrc
    
    # Install Python and tools
    apt-get install -y python3-pip python3-venv git htop iotop sysstat
    
    # Install benchmark tools
    pip3 install psycopg2-binary pymongo redis cassandra-driver clickhouse-driver plyvel influxdb-client prometheus-client
    
    # Install pgbench (PostgreSQL client)
    apt-get install -y postgresql-client
    
    # Install sysbench for MySQL benchmarks
    apt-get install -y sysbench
    
    # Install redis-tools (for redis-benchmark, valkey, dragonfly)
    apt-get install -y redis-tools
    
    # Clone benchmark repo
    cd /home/ubuntu
    git clone https://github.com/christianmahardhika/database-lab-benchmark.git
    chown -R ubuntu:ubuntu database-lab-benchmark
    
    # Pre-build Go benchmark binary
    su - ubuntu -c "cd database-lab-benchmark/benchmarks/go && /usr/local/go/bin/go build -o /home/ubuntu/bench-runner ."
    
    # Pull all Docker images in parallel (19 databases)
    su - ubuntu -c "
      docker pull postgres:16-alpine &
      docker pull pgvector/pgvector:pg16 &
      docker pull mysql:8.0 &
      docker pull mongo:7.0 &
      docker pull redis:7-alpine &
      docker pull valkey/valkey:7.2-alpine &
      docker pull docker.dragonflydb.io/dragonflydb/dragonfly:latest &
      docker pull scylladb/scylla:5.4 &
      docker pull cassandra:4.1 &
      docker pull clickhouse/clickhouse-server:24.3 &
      docker pull timescale/timescaledb:latest-pg16 &
      docker pull cockroachdb/cockroach:v23.2.3 &
      docker pull docker.elastic.co/elasticsearch/elasticsearch:8.14.0 &
      docker pull neo4j:5-enterprise &
      docker pull influxdb:2.7-alpine &
      docker pull prom/prometheus:v2.53.0 &
      docker pull milvusdb/milvus:v2.4.5 &
      docker pull qdrant/qdrant:v1.9.7 &
      docker pull minio/minio:latest &
      wait
    "
    
    echo "Setup complete! 19 databases ready." > /home/ubuntu/setup_complete.txt
  EOF
}

# On-Demand Instance
resource "aws_instance" "benchmark" {
  count = var.spot_instance ? 0 : 1

  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  key_name               = var.key_name
  vpc_security_group_ids = [aws_security_group.benchmark.id]
  user_data              = local.user_data

  root_block_device {
    volume_size = var.volume_size
    volume_type = "gp3"
    iops        = var.volume_iops
    throughput  = 250
  }

  tags = {
    Name    = "database-benchmark-lab"
    Project = "database-lab-benchmark"
  }
}

# Spot Instance (cheaper)
resource "aws_spot_instance_request" "benchmark" {
  count = var.spot_instance ? 1 : 0

  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  key_name               = var.key_name
  vpc_security_group_ids = [aws_security_group.benchmark.id]
  user_data              = local.user_data

  spot_type            = "one-time"
  wait_for_fulfillment = true

  root_block_device {
    volume_size = var.volume_size
    volume_type = "gp3"
    iops        = var.volume_iops
    throughput  = 250
  }

  tags = {
    Name    = "database-benchmark-lab-spot"
    Project = "database-lab-benchmark"
  }
}

# Outputs
output "instance_public_ip" {
  description = "Public IP of the benchmark instance"
  value       = var.spot_instance ? aws_spot_instance_request.benchmark[0].public_ip : aws_instance.benchmark[0].public_ip
}

output "instance_id" {
  description = "Instance ID"
  value       = var.spot_instance ? aws_spot_instance_request.benchmark[0].spot_instance_id : aws_instance.benchmark[0].id
}

output "ssh_command" {
  description = "SSH command to connect"
  value       = "ssh -i ~/.ssh/${var.key_name}.pem ubuntu@${var.spot_instance ? aws_spot_instance_request.benchmark[0].public_ip : aws_instance.benchmark[0].public_ip}"
}

output "estimated_hourly_cost" {
  description = "Estimated hourly cost"
  value       = var.spot_instance ? "~$0.12/hour (spot)" : "~$0.384/hour (on-demand)"
}
