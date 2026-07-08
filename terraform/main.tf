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
  region = var.aws_region
}

# Variables
variable "aws_region" {
  description = "AWS region"
  default     = "ap-southeast-1"  # Singapore - closest to Indonesia
}

variable "instance_type" {
  description = "EC2 instance type"
  default     = "m6i.2xlarge"  # 8 vCPU, 32GB RAM - recommended for benchmarks
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
  default     = 200
}

variable "volume_iops" {
  description = "gp3 volume IOPS"
  default     = 3000
}

# Data sources
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]  # Canonical

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

  # PostgreSQL
  ingress {
    from_port   = 5432
    to_port     = 5439
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "PostgreSQL"
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
    to_port     = 27019
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "MongoDB"
  }

  # Redis
  ingress {
    from_port   = 6379
    to_port     = 6381
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Redis"
  }

  # ScyllaDB/Cassandra
  ingress {
    from_port   = 9042
    to_port     = 9044
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "CQL (Cassandra/ScyllaDB)"
  }

  # ClickHouse
  ingress {
    from_port   = 8123
    to_port     = 8123
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "ClickHouse HTTP"
  }

  ingress {
    from_port   = 9000
    to_port     = 9000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "ClickHouse Native"
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
    
    # Install Python and tools
    apt-get install -y python3-pip python3-venv git htop iotop sysstat
    
    # Install benchmark tools
    pip3 install psycopg2-binary pymongo redis cassandra-driver clickhouse-driver plyvel
    
    # Install pgbench (PostgreSQL client)
    apt-get install -y postgresql-client
    
    # Install sysbench for MySQL benchmarks
    apt-get install -y sysbench
    
    # Clone benchmark repo
    cd /home/ubuntu
    git clone https://github.com/christianmahardhika/database-lab-benchmark.git
    chown -R ubuntu:ubuntu database-lab-benchmark
    
    # Pull all Docker images in parallel
    su - ubuntu -c "docker pull postgres:16-alpine &"
    su - ubuntu -c "docker pull mysql:8.0 &"
    su - ubuntu -c "docker pull mongo:7.0 &"
    su - ubuntu -c "docker pull redis:7-alpine &"
    su - ubuntu -c "docker pull scylladb/scylla:5.4 &"
    su - ubuntu -c "docker pull cassandra:4.1 &"
    su - ubuntu -c "docker pull clickhouse/clickhouse-server:latest &"
    su - ubuntu -c "docker pull timescale/timescaledb:latest-pg16 &"
    wait
    
    echo "Setup complete!" > /home/ubuntu/setup_complete.txt
  EOF
}

# On-Demand Instance
resource "aws_instance" "benchmark" {
  count = var.spot_instance ? 0 : 1

  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  key_name               = var.key_name
  vpc_security_group_ids = [aws_security_group.benchmark.id]
  user_data              = locals.user_data

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
  user_data              = locals.user_data

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
