---
date: '2021-05-07T16:30:00+02:00'
draft: true
title: 'How-to create PostgreSQL user'
slug: 'create-postgresql-user'
tags:
  - 'PostgreSQL'
---
## Roles

We can start with 3 simple roles:

- `ro` - read-only - `select`. Use it most of the time to avoid accidental data loss.
    
- `rwud` - read-write-update-delete - `select, insert, update, delete`. Use with caution when working with data.
    
- `all_priv` - all possible privileges. Use only for creation/migration.
    

## Role creation

Here is how roles are created.

Read-only (`ro`) role:

```sql
CREATE ROLE ro NOLOGIN NOCREATEDB NOSUPERUSER NOCREATEROLE NOREPLICATION;
GRANT CONNECT ON DATABASE postgres TO ro;
GRANT USAGE ON SCHEMA public TO ro;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO ro;
GRANT SELECT
    ON ALL TABLES IN SCHEMA public
    TO ro;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO ro;
```

Read-write-update-delete (`rwud`) role:

```sql
CREATE ROLE rwud NOLOGIN NOCREATEDB NOSUPERUSER NOCREATEROLE NOREPLICATION;
GRANT CONNECT ON DATABASE postgres to rwud;
GRANT USAGE ON SCHEMA public TO rwud;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO rwud;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO rwud;
GRANT SELECT, INSERT, UPDATE, DELETE
    ON ALL TABLES IN SCHEMA public
    TO rwud;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO rwud;
```

All privileges (`all_priv`):

```sql
CREATE ROLE all_priv NOINHERIT NOLOGIN NOSUPERUSER NOCREATEROLE;
GRANT CONNECT ON DATABASE postgres TO all_priv;
GRANT ALL ON DATABASE postgres TO all_priv;
GRANT ALL ON SCHEMA public TO all_priv;
-- tables
GRANT ALL ON ALL TABLES IN SCHEMA public TO all_priv;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO all_priv;
-- sequences
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO all_priv;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO rwud;
-- functions
GRANT ALL ON ALL FUNCTIONS IN SCHEMA public TO all_priv;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO all_priv;
```

### Dropping role

```SQL
REVOKE ALL ON DATABASE postgres FROM rwud;
REVOKE ALL ON SCHEMA public from rwud;
REVOKE ALL ON ALL TABLES IN SCHEMA public from rwud;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public from rwud;
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL PRIVILEGES ON TABLES FROM rwud;
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL PRIVILEGES ON SEQUENCES FROM rwud;
DROP ROLE rwud;
```

Please DON’T use the default `postgres` user at all; it is too easy to break stuff with it.

## User creation

To create a user, first, create the user itself, then grant rights from the roles above.

The username is created as `name_surname_privileges`, for example:

```sql
CREATE USER john_doe_all NOCREATEROLE NOSUPERUSER PASSWORD 'xxxxxx'; 
GRANT all_priv TO john_doe_all;
```

```sql
CREATE USER john_doe_ro NOCREATEROLE NOSUPERUSER PASSWORD 'xxxxxx';
GRANT ro TO john_doe_ro;
```

```sq;
CREATE USER john_doe_rwud NOCREATEROLE NOSUPERUSER PASSWORD 'xxxxxx';
GRANT rwud TO john_doe_rwud;
```

We have a special user, who is the owner of the tables and can grant access to them. Use it to create new users.

### Dropping user

```sql
DROP USER john_doe_all;
```


## Sources

- [https://www.postgresql.org/docs/current/sql-createuser.html](https://www.postgresql.org/docs/current/sql-createuser.html) / [https://www.postgresql.org/docs/current/sql-createrole.html](https://www.postgresql.org/docs/current/sql-createrole.html)
    
- [https://www.postgresql.org/docs/current/ddl-priv.html](https://www.postgresql.org/docs/current/ddl-priv.html)