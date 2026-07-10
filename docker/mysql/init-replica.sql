-- Configure replica to replicate from source using GTID
-- This runs after MySQL starts on the replica
CHANGE REPLICATION SOURCE TO
    SOURCE_HOST='mysql-source',
    SOURCE_PORT=3306,
    SOURCE_USER='replicator',
    SOURCE_PASSWORD='replicator_pass',
    SOURCE_AUTO_POSITION=1,
    GET_SOURCE_PUBLIC_KEY=1;

START REPLICA;
