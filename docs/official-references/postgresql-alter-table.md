# PostgreSQL ALTER TABLE - Official Documentation

Source: https://www.postgresql.org/docs/current/sql-altertable.html
Version: PostgreSQL 18

## Synopsis

```sql
ALTER TABLE [ IF EXISTS ] [ ONLY ] name [ * ]
    action [, ... ]
ALTER TABLE [ IF EXISTS ] [ ONLY ] name [ * ]
    RENAME [ COLUMN ] column_name TO new_column_name
ALTER TABLE [ IF EXISTS ] [ ONLY ] name [ * ]
    RENAME CONSTRAINT constraint_name TO new_constraint_name
ALTER TABLE [ IF EXISTS ] name
    RENAME TO new_name
ALTER TABLE [ IF EXISTS ] name
    SET SCHEMA new_schema
ALTER TABLE ALL IN TABLESPACE name [ OWNED BY role_name [, ... ] ]
    SET TABLESPACE new_tablespace [ NOWAIT ]
ALTER TABLE [ IF EXISTS ] name
    ATTACH PARTITION partition_name { FOR VALUES partition_bound_spec | DEFAULT }
ALTER TABLE [ IF EXISTS ] name
    DETACH PARTITION partition_name [ CONCURRENTLY | FINALIZE ]
```

## Actions

- `ADD [ COLUMN ] [ IF NOT EXISTS ] column_name data_type [ COLLATE collation ] [ column_constraint [ ... ] ]`
- `DROP [ COLUMN ] [ IF EXISTS ] column_name [ RESTRICT | CASCADE ]`
- `ALTER [ COLUMN ] column_name [ SET DATA ] TYPE data_type [ COLLATE collation ] [ USING expression ]`
- `ALTER [ COLUMN ] column_name SET DEFAULT expression`
- `ALTER [ COLUMN ] column_name DROP DEFAULT`
- `ALTER [ COLUMN ] column_name { SET | DROP } NOT NULL`
- `ALTER [ COLUMN ] column_name SET EXPRESSION AS ( expression )`
- `ALTER [ COLUMN ] column_name DROP EXPRESSION [ IF EXISTS ]`
- `ADD table_constraint [ NOT VALID ]`
- `DROP CONSTRAINT [ IF EXISTS ] constraint_name [ RESTRICT | CASCADE ]`
- `VALIDATE CONSTRAINT constraint_name`
- And many more...

## Description

ALTER TABLE changes the definition of an existing table. There are several subforms described below. Note that **the lock level required may differ for each subform**. An **ACCESS EXCLUSIVE lock is acquired unless explicitly noted**. When multiple subcommands are given, the lock acquired will be the strictest one required by any subcommand.

### ADD COLUMN

This form adds a new column to the table, using the same syntax as CREATE TABLE. If IF NOT EXISTS is specified and a column already exists with this name, no error is thrown.

### DROP COLUMN

This form drops a column from a table. Indexes and table constraints involving the column will be automatically dropped as well. You will need to say CASCADE if anything outside the table depends on the column.

### SET DATA TYPE

This form changes the type of a column of a table. Indexes and simple table constraints involving the column will be automatically converted to use the new column type by reparsing the originally supplied expression.

When this form is used, **the column's statistics are removed**, so running ANALYZE on the table afterwards is recommended.

### SET/DROP DEFAULT

These forms set or remove the default value for a column. The new default value will only apply in subsequent INSERT or UPDATE commands; **it does not cause rows already in the table to change**.

### SET/DROP NOT NULL

SET NOT NULL may only be applied to a column provided none of the records in the table contain a NULL value for the column. Ordinarily this is checked during the ALTER TABLE by **scanning the entire table**, unless NOT VALID is specified.

### ADD table_constraint [ NOT VALID ]

Normally, this form will cause a **scan of the table** to verify that all existing rows in the table satisfy the new constraint. But if the NOT VALID option is used, this potentially-lengthy scan is skipped.

**Lock levels:**
- Most forms of ADD table_constraint require an **ACCESS EXCLUSIVE lock**
- ADD FOREIGN KEY requires only a **SHARE ROW EXCLUSIVE lock**

### VALIDATE CONSTRAINT

This form validates a foreign key, check, or not-null constraint that was previously created as NOT VALID, by scanning the table to ensure there are no rows for which the constraint is not satisfied.

This command acquires a **SHARE UPDATE EXCLUSIVE lock**.

### DROP CONSTRAINT

This form drops the specified constraint on a table, along with any index underlying the constraint.

### SET STATISTICS

This form sets the per-column statistics-gathering target for subsequent ANALYZE operations.

SET STATISTICS acquires a **SHARE UPDATE EXCLUSIVE lock**.

### SET STORAGE { PLAIN | EXTERNAL | EXTENDED | MAIN | DEFAULT }

This form sets the storage mode for a column. This controls whether this column is held inline or in a secondary TOAST table, and whether the data should be compressed or not.

- **PLAIN** - fixed-length values such as integer, inline, uncompressed
- **MAIN** - inline, compressible data
- **EXTERNAL** - external, uncompressed data
- **EXTENDED** - external, compressed data (default for most types)

Note that ALTER TABLE ... SET STORAGE **doesn't itself change anything in the table**; it just sets the strategy to be pursued during future table updates.

### SET TABLESPACE

This form changes the table's tablespace to the specified tablespace and **moves the data file(s)** associated with the table to the new tablespace. Indexes on the table, if any, **are not moved**; but they can be moved separately.

### SET { LOGGED | UNLOGGED }

This form changes the table from unlogged to logged or vice-versa. It cannot be applied to a temporary table.

### ATTACH PARTITION

This form attaches an existing table as a partition of the target table.

A partition using FOR VALUES uses same syntax for partition_bound_spec as CREATE TABLE. If the new partition is a regular table, **a full table scan is performed** to check that existing rows in the table do not violate the partition constraint.

Attaching a partition acquires a **SHARE UPDATE EXCLUSIVE lock** on the parent table, in addition to the **ACCESS EXCLUSIVE locks** on the table being attached and on the default partition.

### DETACH PARTITION

This form detaches the specified partition of the target table.

If **CONCURRENTLY** is specified, it runs using a reduced lock level to avoid blocking other sessions. In this mode, two transactions are used internally.

## Notes - Critical Performance Information

### ADD COLUMN with DEFAULT (Fast Path)

When a column is added with ADD COLUMN and a **non-volatile DEFAULT** is specified, the default value is evaluated at the time of the statement and **the result stored in the table's metadata**, where it will be returned when any existing rows are accessed. The value will be only applied when the table is rewritten, making **the ALTER TABLE very fast even on large tables**.

If no column constraints are specified, NULL is used as the DEFAULT. **In neither case is a rewrite of the table required.**

### Operations That Require Table Rewrite

Adding a column with:
- A **volatile DEFAULT** (e.g., clock_timestamp())
- A **stored generated column**
- An **identity column**
- A column with a **domain data type that has constraints**

...will cause **the entire table and its indexes to be rewritten**.

Adding a **virtual generated column never requires a rewrite**.

### Changing Column Type

Changing the type of an existing column will **normally cause the entire table and its indexes to be rewritten**.

**Exception:** When changing the type, if the USING clause does not change the column contents and the old type is either **binary coercible** to the new type or an unconstrained domain over the new type, **a table rewrite is not needed**. However, indexes will still be rebuilt unless the system can verify that the new index would be logically equivalent.

Example: A column can be changed from **text to varchar** (or vice versa) **without rebuilding the indexes** because these data types sort identically.

### Table Rebuilds

Table and/or index rebuilds may take a **significant amount of time for a large table**, and will temporarily require as much as **double the disk space**.

### Adding Constraints

Adding a CHECK or NOT NULL constraint requires **scanning the table** to verify that existing rows meet the constraint, but **does not require a table rewrite**.

### NOT VALID Pattern for Minimal Downtime

Scanning a large table to verify new foreign-key, check, or not-null constraints can take a long time, and **other updates to the table are locked out** until the ALTER TABLE ADD CONSTRAINT command is committed.

The main purpose of the **NOT VALID** constraint option is to reduce the impact of adding a constraint on concurrent updates:

1. `ALTER TABLE ADD CONSTRAINT ... NOT VALID` - does not scan the table, can be committed immediately
2. `VALIDATE CONSTRAINT` - verifies existing rows, **does not lock out concurrent updates** (only SHARE UPDATE EXCLUSIVE lock)

### DROP COLUMN Internals

The DROP COLUMN form **does not physically remove the column**, but simply makes it invisible to SQL operations. Subsequent insert and update operations will store a null value for the column.

Thus, dropping a column is **quick** but it will **not immediately reduce the on-disk size** of your table, as the space occupied by the dropped column is not reclaimed. The space will be reclaimed over time as existing rows are updated.

To force immediate reclamation of space occupied by a dropped column, execute one of the forms of ALTER TABLE that performs a **rewrite of the whole table**.

### MVCC Warning

The rewriting forms of ALTER TABLE are **not MVCC-safe**. After a table rewrite, the table will **appear empty to concurrent transactions**, if they are using a snapshot taken before the rewrite occurred.

## Lock Levels Summary

| Operation | Lock Level |
|-----------|------------|
| Most ALTER TABLE | ACCESS EXCLUSIVE |
| ADD FOREIGN KEY | SHARE ROW EXCLUSIVE |
| VALIDATE CONSTRAINT | SHARE UPDATE EXCLUSIVE |
| SET STATISTICS | SHARE UPDATE EXCLUSIVE |
| Changing per-attribute options | SHARE UPDATE EXCLUSIVE |
| Changing cluster options | SHARE UPDATE EXCLUSIVE |
| ATTACH PARTITION (on parent) | SHARE UPDATE EXCLUSIVE |
| DISABLE/ENABLE TRIGGER | SHARE ROW EXCLUSIVE |

## Examples

### Add column (fast, no rewrite):
```sql
ALTER TABLE distributors ADD COLUMN address varchar(30);
```

### Add column with non-null default (fast in PG11+):
```sql
ALTER TABLE measurements
  ADD COLUMN mtime timestamp with time zone DEFAULT now();
```

### Change column type with USING:
```sql
ALTER TABLE foo
    ALTER COLUMN foo_timestamp SET DATA TYPE timestamp with time zone
    USING
        timestamp with time zone 'epoch' + foo_timestamp * interval '1 second';
```

### Add foreign key with minimal locking:
```sql
ALTER TABLE distributors ADD CONSTRAINT distfk 
    FOREIGN KEY (address) REFERENCES addresses (address) NOT VALID;
ALTER TABLE distributors VALIDATE CONSTRAINT distfk;
```

### Recreate primary key without blocking updates:
```sql
CREATE UNIQUE INDEX CONCURRENTLY dist_id_temp_idx ON distributors (dist_id);
ALTER TABLE distributors DROP CONSTRAINT distributors_pkey,
    ADD CONSTRAINT distributors_pkey PRIMARY KEY USING INDEX dist_id_temp_idx;
```

### Attach partition:
```sql
ALTER TABLE measurement
    ATTACH PARTITION measurement_y2016m07 FOR VALUES FROM ('2016-07-01') TO ('2016-08-01');
```

### Detach partition concurrently:
```sql
ALTER TABLE measurement
    DETACH PARTITION measurement_y2015m12 CONCURRENTLY;
```
