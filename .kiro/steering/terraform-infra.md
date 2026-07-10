---
inclusion: fileMatch
fileMatchPattern: "terraform/**"
---

# Terraform Infrastructure Guide

## AWS Setup for Database Lab

### Instance Sizing Recommendations

| Use Case | Instance Type | vCPU | RAM | Cost/hr | Notes |
|----------|--------------|------|-----|---------|-------|
| Quick validation | m6i.xlarge | 4 | 16GB | $0.192 | Runs 2-3 DBs concurrently |
| Full benchmark suite | m6i.2xlarge | 8 | 32GB | $0.384 | Recommended default |
| Heavy experiments | m6i.4xlarge | 16 | 64GB | $0.768 | Horizontal scaling tests |
| Cost-optimized | Spot instance | varies | varies | 60-70% off | May be interrupted |

### Storage Configuration

- **Volume type**: gp3 (baseline 3000 IOPS, 125 MB/s throughput)
- **Size**: 200GB (enough for all databases + benchmark data)
- **IOPS**: 3000 default, increase to 6000+ for I/O-intensive benchmarks
- **Throughput**: 250 MB/s (gp3 allows up to 1000 MB/s)

### Region Selection

Default: `ap-southeast-1` (Singapore) — lowest latency from Indonesia.
Alternative: `us-east-1` for cheaper spot pricing.

### Security Considerations

The current security group opens database ports to `0.0.0.0/0`. This is acceptable for:
- Short-lived benchmark instances (hours, not days)
- Non-production data (synthetic benchmark data only)

For longer-lived instances:
- Restrict `cidr_blocks` to your IP: `["<your-ip>/32"]`
- Add VPN or bastion host
- Enable auto-shutdown after X hours idle

## Terraform Commands

```bash
# Initialize
cd terraform
terraform init

# Plan changes
terraform plan -var="key_name=my-key"

# Apply (create instance)
terraform apply -var="key_name=my-key"

# Use spot instance (cheaper)
terraform apply -var="key_name=my-key" -var="spot_instance=true"

# Destroy (IMPORTANT: do this when done to avoid charges)
terraform destroy -var="key_name=my-key"
```

## Post-Provisioning Checklist

1. Wait for user_data to complete: `ssh ubuntu@<ip> "cat ~/setup_complete.txt"`
2. Verify Docker: `docker ps`
3. Verify Python deps: `python3 -c "import psycopg2, pymongo, redis"`
4. Pull latest repo: `cd database-lab-benchmark && git pull`
5. Run quick sanity check: `./scripts/run_all.sh quick`

## Cost Control

- **Always destroy when done**: `terraform destroy`
- Spot instances save 60-70% but may be interrupted
- Set a calendar reminder to destroy after your session
- Consider a Lambda function to auto-terminate instances idle >2 hours
