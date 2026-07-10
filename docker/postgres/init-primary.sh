#!/bin/bash
set -e

# Create replication user
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'replicator_pass';
    SELECT * FROM pg_create_physical_replication_slot('replica_slot');
EOSQL

# Allow replication connections from any host in the network
echo "host replication replicator 0.0.0.0/0 md5" >> "$PGDATA/pg_hba.conf"
