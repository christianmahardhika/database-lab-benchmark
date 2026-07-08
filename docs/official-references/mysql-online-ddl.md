# MySQL InnoDB Online DDL - Official Documentation

Source: https://dev.mysql.com/doc/refman/8.0/en/innodb-online-ddl-operations.html
Version: MySQL 8.0

## Overview

The online DDL feature provides support for **instant and in-place table alterations** and **concurrent DML**. Benefits include:

- Improved responsiveness and availability in busy production environments
- Ability to adjust balance between performance and concurrency using the LOCK clause
- Less disk space usage and I/O overhead than the table-copy method

**Note:** ALGORITHM=INSTANT support is available for ADD COLUMN and other operations in MySQL 8.0.12.

## ALGORITHM and LOCK Clauses

```sql
ALTER TABLE tbl_name ADD PRIMARY KEY (column), ALGORITHM=INPLACE;
ALTER TABLE tbl_name ADD COLUMN col INT, ALGORITHM=INSTANT;
```

### ALGORITHM Options
- **INSTANT** - Only modifies metadata in data dictionary. No exclusive lock during preparation/execution. Instant and low impact.
- **INPLACE** - Avoids table copy but may rebuild table in-place. May take shared lock briefly during preparation/execution.
- **COPY** - Full table copy. Permits concurrent reads but not writes. High resource usage.

### LOCK Options
- **LOCK=NONE** - Permit reads and writes
- **LOCK=SHARED** - Permit reads only
- **LOCK=EXCLUSIVE** - No concurrent access
- **LOCK=DEFAULT** - Maximum concurrency allowed for the operation

---

## Index Operations

| Operation | Instant | In Place | Rebuilds Table | Permits Concurrent DML | Only Modifies Metadata |
|-----------|---------|----------|----------------|------------------------|------------------------|
| Creating/adding secondary index | No | Yes | No | Yes | No |
| Dropping an index | No | Yes | No | Yes | Yes |
| Renaming an index | No | Yes | No | Yes | Yes |
| Adding a FULLTEXT index | No | Yes* | No* | No | No |
| Adding a SPATIAL index | No | Yes | No | No | No |
| Changing index type | Yes | Yes | No | Yes | Yes |

### Notes:
- **Creating secondary index**: Table remains available for read and write while index is created
- **Adding first FULLTEXT index**: Rebuilds table if no user-defined FTS_DOC_ID column exists

---

## Primary Key Operations

| Operation | Instant | In Place | Rebuilds Table | Permits Concurrent DML | Only Modifies Metadata |
|-----------|---------|----------|----------------|------------------------|------------------------|
| Adding a primary key | No | Yes* | Yes* | Yes | No |
| Dropping a primary key | No | No | Yes | No | No |
| Dropping and adding primary key | No | Yes | Yes | Yes | No |

### Critical Notes:

**Adding a primary key:**
```sql
ALTER TABLE tbl_name ADD PRIMARY KEY (column), ALGORITHM=INPLACE, LOCK=NONE;
```
- Rebuilds the table in place. **Data is reorganized substantially, making it an expensive operation.**
- Best to define PRIMARY KEY when creating table, not via ALTER TABLE later
- ALGORITHM=INPLACE is only permitted when SQL_MODE includes strict_trans_tables or strict_all_tables

**Why primary key changes are expensive:**
- Rows are stored in a clustered index organized based on the primary key ("index-organized table")
- Table structure is closely tied to primary key, so redefining requires copying data

**ALGORITHM=INPLACE advantages over COPY:**
- No undo logging or associated redo logging required
- Secondary index entries are pre-sorted, can be loaded in order
- Change buffer is not used (no random-access inserts into secondary indexes)

**Dropping a primary key:**
```sql
ALTER TABLE tbl_name DROP PRIMARY KEY, ALGORITHM=COPY;
```
Only ALGORITHM=COPY supports dropping a primary key without adding a new one.

---

## Column Operations

| Operation | Instant | In Place | Rebuilds Table | Permits Concurrent DML | Only Modifies Metadata |
|-----------|---------|----------|----------------|------------------------|------------------------|
| Adding a column | Yes* | Yes | No* | Yes* | Yes |
| Dropping a column | Yes* | Yes | Yes | Yes | Yes |
| Renaming a column | Yes* | Yes | No | Yes* | Yes |
| Reordering columns | No | Yes | Yes | Yes | No |
| Setting column default value | Yes | Yes | No | Yes | Yes |
| Changing column data type | No | No | Yes | No | No |
| Extending VARCHAR column size | No | Yes | No | Yes | Yes |
| Dropping column default value | Yes | Yes | No | Yes | Yes |
| Changing auto-increment value | No | Yes | No | Yes | No* |
| Making a column NULL | No | Yes | Yes* | Yes | No |
| Making a column NOT NULL | No | Yes* | Yes* | Yes | No |
| Modifying ENUM/SET definition | Yes | Yes | No | Yes | Yes |

### Adding a column (INSTANT - MySQL 8.0.12+)

```sql
ALTER TABLE tbl_name ADD COLUMN column_name column_definition, ALGORITHM=INSTANT;
```

**INSTANT limitations:**
- Cannot combine with ALTER TABLE actions that don't support INSTANT
- Cannot add to: ROW_FORMAT=COMPRESSED tables, tables with FULLTEXT index, temporary tables
- MySQL checks row size and throws error if limit exceeded
- **Maximum 64 row versions** - after that, must use INPLACE/COPY to rebuild table

**Concurrent DML is NOT permitted when adding auto-increment column** (expensive operation, requires at minimum ALGORITHM=INPLACE, LOCK=SHARED)

### Dropping a column (INSTANT - MySQL 8.0.29+)

```sql
ALTER TABLE tbl_name DROP COLUMN column_name, ALGORITHM=INSTANT;
```

Same limitations as adding columns. Each instant add/drop creates a new "row version" (max 64).

### Renaming a column

```sql
ALTER TABLE tbl CHANGE old_col_name new_col_name data_type, ALGORITHM=INSTANT;
```

ALGORITHM=INSTANT support added in MySQL 8.0.28. Keep same data type for online operation.

### Changing column data type

```sql
ALTER TABLE tbl_name CHANGE c1 c1 BIGINT, ALGORITHM=COPY;
```

**Only ALGORITHM=COPY is supported** - always requires table rebuild.

### Extending VARCHAR column size

```sql
ALTER TABLE tbl_name CHANGE COLUMN c1 c1 VARCHAR(255), ALGORITHM=INPLACE, LOCK=NONE;
```

**Critical limitation:** Length bytes must stay the same:
- VARCHAR 0-255 bytes: 1 length byte
- VARCHAR 256+ bytes: 2 length bytes

**In-place only supports:**
- 0-255 → 0-255 (within 1-byte range)
- 256+ → larger (within 2-byte range)

**NOT supported in-place:**
- 255 → 256 (crosses byte boundary, requires ALGORITHM=COPY)
- Decreasing VARCHAR size (always requires COPY)

### Making a column NOT NULL

```sql
ALTER TABLE tbl_name MODIFY COLUMN column_name data_type NOT NULL, ALGORITHM=INPLACE, LOCK=NONE;
```

- Rebuilds table in place
- Requires STRICT_ALL_TABLES or STRICT_TRANS_TABLES SQL_MODE
- **Operation fails if column contains NULL values**

---

## Generated Column Operations

| Operation | Instant | In Place | Rebuilds Table | Permits Concurrent DML | Only Modifies Metadata |
|-----------|---------|----------|----------------|------------------------|------------------------|
| Adding a STORED column | No | No | Yes | No | No |
| Modifying STORED column order | No | No | Yes | No | No |
| Dropping a STORED column | No | Yes | Yes | Yes | No |
| Adding a VIRTUAL column | Yes | Yes | No | Yes | Yes |
| Modifying VIRTUAL column order | No | No | Yes | No | No |
| Dropping a VIRTUAL column | Yes | Yes | No | Yes | Yes |

**Key insight:** VIRTUAL columns are instant, STORED columns require table rebuild.

---

## Foreign Key Operations

| Operation | Instant | In Place | Rebuilds Table | Permits Concurrent DML | Only Modifies Metadata |
|-----------|---------|----------|----------------|------------------------|------------------------|
| Adding foreign key constraint | No | Yes* | No | Yes | Yes |
| Dropping foreign key constraint | No | Yes | No | Yes | Yes |

```sql
-- Adding FK (requires foreign_key_checks disabled for INPLACE)
SET foreign_key_checks = 0;
ALTER TABLE tbl1 ADD CONSTRAINT fk_name FOREIGN KEY index (col1)
  REFERENCES tbl2(col2), ALGORITHM=INPLACE;
SET foreign_key_checks = 1;

-- Dropping FK (works with foreign_key_checks enabled or disabled)
ALTER TABLE tbl DROP FOREIGN KEY fk_name;
```

**Important:** Tables with foreign keys (child tables) may wait for parent table changes due to CASCADE or SET NULL actions.

---

## Table Operations

| Operation | Instant | In Place | Rebuilds Table | Permits Concurrent DML | Only Modifies Metadata |
|-----------|---------|----------|----------------|------------------------|------------------------|
| Changing ROW_FORMAT | No | Yes | Yes | Yes | No |
| Changing KEY_BLOCK_SIZE | No | Yes | Yes | Yes | No |
| Setting persistent table stats | No | Yes | No | Yes | Yes |
| Specifying character set | No | Yes | Yes* | Yes | No |
| Converting character set | No | Yes | Yes* | No | No |
| Optimizing a table | No | Yes* | Yes | Yes | No |
| Rebuilding with FORCE | No | Yes* | Yes | Yes | No |
| Null rebuild (ENGINE=InnoDB) | No | Yes* | Yes | Yes | No |
| Renaming a table | Yes | Yes | No | Yes | Yes |

### Renaming a table (INSTANT)

```sql
ALTER TABLE old_tbl_name RENAME TO new_tbl_name, ALGORITHM=INSTANT;
```

MySQL renames files without making a copy.

### Optimizing/Rebuilding a table

```sql
OPTIMIZE TABLE tbl_name;
ALTER TABLE tbl_name FORCE, ALGORITHM=INPLACE, LOCK=NONE;
ALTER TABLE tbl_name ENGINE=InnoDB, ALGORITHM=INPLACE, LOCK=NONE;
```

**Note:** ALGORITHM=INPLACE not supported for tables with FULLTEXT indexes.

---

## Partition Operations

| Partition Clause | Instant | In Place | Permits DML | Notes |
|------------------|---------|----------|-------------|-------|
| PARTITION BY | No | No | No | ALGORITHM=COPY only |
| ADD PARTITION | No | Yes* | Yes* | INPLACE for RANGE/LIST, copies data for HASH/KEY |
| DROP PARTITION | No | Yes* | Yes* | INPLACE doesn't copy data for RANGE/LIST |
| TRUNCATE PARTITION | No | Yes | Yes | Just deletes rows, no table alteration |
| COALESCE PARTITION | No | Yes* | No | ALGORITHM=INPLACE, LOCK={SHARED\|EXCLUSIVE} |
| REORGANIZE PARTITION | No | Yes* | No | ALGORITHM=INPLACE, LOCK={SHARED\|EXCLUSIVE} |
| EXCHANGE PARTITION | No | Yes | Yes | |
| REMOVE PARTITIONING | No | No | No | ALGORITHM=COPY only |

---

## Row Version Tracking (MySQL 8.0.29+)

When using ALGORITHM=INSTANT for ADD/DROP COLUMN, each operation creates a new "row version":

```sql
SELECT NAME, TOTAL_ROW_VERSIONS 
FROM INFORMATION_SCHEMA.INNODB_TABLES 
WHERE NAME LIKE 'test/t1';
```

**Maximum 64 row versions.** After reaching limit:
```
ERROR 4092 (HY000): Maximum row versions reached for table test/t1. 
No more columns can be added or dropped instantly. 
Please use COPY/INPLACE.
```

Rebuild table (OPTIMIZE TABLE or ALTER TABLE ... FORCE) to reset counter to 0.

---

## Summary: What Requires Table Rebuild?

**Always requires COPY (most expensive):**
- Changing column data type
- Dropping primary key alone
- File-per-table tablespace encryption changes

**Requires In-Place Rebuild (expensive but allows DML):**
- Adding/dropping primary key
- Reordering columns
- Making column NULL/NOT NULL
- Adding STORED generated column
- ROW_FORMAT/KEY_BLOCK_SIZE changes
- Character set conversion

**Instant/Metadata-only (fastest):**
- Adding/dropping columns (MySQL 8.0.12+/8.0.29+)
- Renaming columns/tables
- Setting/dropping default values
- Adding/dropping VIRTUAL columns
- Modifying ENUM/SET (appending values only)
- Index type changes

---

## Online DDL Performance and Concurrency

### The Three Phases of Online DDL

**Phase 1: Initialization**
- Server determines how much concurrency is permitted
- Takes into account storage engine capabilities, operations specified, and ALGORITHM/LOCK options
- A **shared upgradeable metadata lock** is taken to protect current table definition

**Phase 2: Execution**
- Statement is prepared and executed
- Whether metadata lock is upgraded to exclusive depends on factors from Phase 1
- If exclusive metadata lock required, it is **only taken briefly** during statement preparation

**Phase 3: Commit Table Definition**
- Metadata lock is **upgraded to exclusive** to evict old table definition and commit new one
- Once granted, duration of exclusive metadata lock is **brief**

### Metadata Lock Blocking Behavior

**Critical insight:** An online DDL operation may have to wait for concurrent transactions that hold metadata locks on the table to commit or rollback.

Example scenario:
```
Session 1: START TRANSACTION; SELECT * FROM t1;  -- holds shared metadata lock
Session 2: ALTER TABLE t1 ADD COLUMN x INT;      -- waits for exclusive lock
Session 3: SELECT * FROM t1;                      -- blocked by Session 2's pending lock!
```

**Diagnosis:**
```sql
SHOW FULL PROCESSLIST\G
-- Look for "Waiting for table metadata lock"

-- Or via Performance Schema:
SELECT * FROM performance_schema.metadata_locks;
```

### Performance Tips

1. **Check "rows affected"** after DDL to understand what happened:
   - `0 rows affected` = metadata-only or in-place (fast)
   - `N rows affected` = table was rebuilt (slow)

2. **Test on clone first** for large tables:
   ```sql
   -- Clone structure
   CREATE TABLE t1_test LIKE t1;
   -- Add some data
   INSERT INTO t1_test SELECT * FROM t1 LIMIT 1000;
   -- Test DDL
   ALTER TABLE t1_test ADD COLUMN x INT;
   -- Check rows affected
   ```

3. **Monitor progress** via Performance Schema stage events

4. **Trade-off awareness:** Online DDL may take longer overall than table-copy because it records concurrent DML changes, but provides better application responsiveness

---

## ALTER TABLE Syntax Reference

Source: https://dev.mysql.com/doc/refman/8.0/en/alter-table.html

### Basic Syntax

```sql
ALTER TABLE tbl_name
    [alter_option [, alter_option] ...]
    [partition_options]

-- Key alter_options:
ADD [COLUMN] col_name column_definition [FIRST | AFTER col_name]
ADD {INDEX | KEY} [index_name] (key_part,...)
ADD [CONSTRAINT] PRIMARY KEY (key_part,...)
ADD [CONSTRAINT] UNIQUE [INDEX] [index_name] (key_part,...)
ADD [CONSTRAINT] FOREIGN KEY [index_name] (col_name,...) reference_definition
ADD [CONSTRAINT] CHECK (expr) [[NOT] ENFORCED]

DROP [COLUMN] col_name
DROP {INDEX | KEY} index_name
DROP PRIMARY KEY
DROP FOREIGN KEY fk_symbol
DROP {CHECK | CONSTRAINT} symbol

CHANGE [COLUMN] old_col_name new_col_name column_definition [FIRST | AFTER]
MODIFY [COLUMN] col_name column_definition [FIRST | AFTER]
RENAME COLUMN old_col_name TO new_col_name
RENAME {INDEX | KEY} old_index_name TO new_index_name
RENAME [TO | AS] new_tbl_name

ALTER [COLUMN] col_name {SET DEFAULT value | DROP DEFAULT}
ALTER INDEX index_name {VISIBLE | INVISIBLE}

ALGORITHM [=] {DEFAULT | INSTANT | INPLACE | COPY}
LOCK [=] {DEFAULT | NONE | SHARED | EXCLUSIVE}

FORCE
{DISABLE | ENABLE} KEYS
{DISCARD | IMPORT} TABLESPACE
```

### Multiple Operations in Single Statement

MySQL extension allows combining multiple operations:

```sql
-- Drop multiple columns
ALTER TABLE t2 DROP COLUMN c, DROP COLUMN d;

-- Add column and index together
ALTER TABLE t1 
  ADD COLUMN status VARCHAR(20) DEFAULT 'active',
  ADD INDEX idx_status (status);

-- Rename and modify in one statement
ALTER TABLE t1 
  RENAME COLUMN old_name TO new_name,
  MODIFY COLUMN other_col BIGINT;
```

### CHANGE vs MODIFY vs RENAME COLUMN

| Operation | Can Rename | Can Change Definition | Syntax |
|-----------|------------|----------------------|--------|
| CHANGE | Yes | Yes | `CHANGE old_name new_name definition` |
| MODIFY | No | Yes | `MODIFY col_name definition` |
| RENAME COLUMN | Yes | No | `RENAME COLUMN old TO new` |
| ALTER | No | Default only | `ALTER col SET DEFAULT value` |

**Important:** CHANGE and MODIFY require full column definition - attributes not specified are dropped!

```sql
-- Original: col1 INT UNSIGNED DEFAULT 1 COMMENT 'my column'

-- WRONG - loses UNSIGNED, DEFAULT, COMMENT:
ALTER TABLE t1 MODIFY col1 BIGINT;

-- CORRECT - preserves attributes:
ALTER TABLE t1 MODIFY col1 BIGINT UNSIGNED DEFAULT 1 COMMENT 'my column';
```

### Column Reordering

```sql
-- Move column to first position
ALTER TABLE t1 MODIFY col_name INT FIRST;

-- Move column after another
ALTER TABLE t1 MODIFY col_name INT AFTER other_col;

-- Works with CHANGE too
ALTER TABLE t1 CHANGE old_name new_name VARCHAR(100) AFTER target_col;
```

### Privileges Required

- **ALTER TABLE**: Requires ALTER, CREATE, and INSERT privileges
- **RENAME TABLE**: Requires ALTER and DROP on old table; ALTER, CREATE, INSERT on new table
- **DISABLE/ENABLE KEYS**: Requires INDEX privilege

### Table Options

```sql
-- Change storage engine (causes table rebuild)
ALTER TABLE t1 ENGINE = InnoDB;

-- Defragment InnoDB table (null ALTER)
ALTER TABLE t1 ENGINE = InnoDB;
ALTER TABLE t1 FORCE;

-- Change row format
ALTER TABLE t1 ROW_FORMAT = COMPRESSED;

-- Reset auto-increment
ALTER TABLE t1 AUTO_INCREMENT = 1000;

-- Change character set
ALTER TABLE t1 CHARACTER SET = utf8mb4;

-- Enable encryption
ALTER TABLE t1 ENCRYPTION = 'Y';

-- Move to tablespace (causes full rebuild)
ALTER TABLE t1 TABLESPACE = my_tablespace;
```

### Index Operations

```sql
-- Add index
ALTER TABLE t1 ADD INDEX idx_name (col1, col2);
ALTER TABLE t1 ADD UNIQUE INDEX idx_unique (col1);
ALTER TABLE t1 ADD FULLTEXT INDEX idx_ft (text_col);

-- Drop index
ALTER TABLE t1 DROP INDEX idx_name;

-- Rename index
ALTER TABLE t1 RENAME INDEX old_idx TO new_idx;

-- Make index invisible (optimizer ignores it)
ALTER TABLE t1 ALTER INDEX idx_name INVISIBLE;
ALTER TABLE t1 ALTER INDEX idx_name VISIBLE;

-- MyISAM: disable/enable non-unique indexes for bulk load
ALTER TABLE t1 DISABLE KEYS;
-- ... bulk inserts ...
ALTER TABLE t1 ENABLE KEYS;
```

### Foreign Key Operations

```sql
-- Add foreign key
ALTER TABLE child_table 
  ADD CONSTRAINT fk_parent 
  FOREIGN KEY (parent_id) REFERENCES parent_table(id)
  ON DELETE CASCADE ON UPDATE CASCADE;

-- Drop foreign key
ALTER TABLE child_table DROP FOREIGN KEY fk_parent;

-- Note: Foreign key name != index name
-- To find FK names:
SHOW CREATE TABLE child_table\G
```

### CHECK Constraints (MySQL 8.0.16+)

```sql
-- Add check constraint
ALTER TABLE t1 ADD CONSTRAINT chk_positive CHECK (amount > 0);

-- Add non-enforced check (for documentation)
ALTER TABLE t1 ADD CONSTRAINT chk_info CHECK (status IN ('A','B','C')) NOT ENFORCED;

-- Drop check constraint
ALTER TABLE t1 DROP CHECK chk_positive;

-- Enable/disable enforcement
ALTER TABLE t1 ALTER CHECK chk_info ENFORCED;
ALTER TABLE t1 ALTER CHECK chk_info NOT ENFORCED;
```

### Validation Control

```sql
-- Skip validation for generated columns (faster but unsafe)
ALTER TABLE t1 MODIFY gen_col INT AS (col1 + col2) WITHOUT VALIDATION;

-- Force validation
ALTER TABLE t1 MODIFY gen_col INT AS (col1 + col2) WITH VALIDATION;
```
