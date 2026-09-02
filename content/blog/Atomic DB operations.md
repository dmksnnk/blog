---
date: '2026-09-02T16:00:00+02:00'
draft: false
title: 'Atomic Database Operations'
slug: 'atomic-database-operations'
showToc: true
cover:
    image: 'images/cas.svg'
summary: |
tags:
    - distributed systems
    - PostgreSQL
---

Today we will explore how to leverage atomic operations without opening a multi-statement transaction,
increasing the system's write throughput without sacrificing correctness.

To avoid conflicts with other processes changing the same data, we can lock the rows before
updating them, for example with `SELECT ... FOR UPDATE`.
This approach has drawbacks: locks consume database resources for the entire duration of the
transaction and can lead to lock contention as writers wait for one another, reducing write throughput.
It is generally preferable for transactions to be short.

Some operations can instead be expressed optimistically. Rather than explicitly locking a resource up front
so that no other process can modify it (_pessimistic concurrency control_), we assume everything
will be fine (_optimistic concurrency control_), proceed, and resolve conflicts only when they occur.
Optimistic concurrency control allows for higher throughput when the number of conflicts is _small_
because locks are held only briefly. However, it requires a conflict resolution
mechanism, such as retrying the operation.

## Commutative operations

Let's first look at operations that do not require conflict resolution.
Operations are commutative when their order does not matter, as with addition: changing the order
does not change the result. We will review several examples and see how to implement them.

### Collecting unique values

Let's say we have an auction system. We want to collect the unique users who have bid on an item.
A simple table might look like this:

```sql
CREATE TABLE bidders (
  id      BIGINT GENERATED ALWAYS AS IDENTITY,
  item_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL
);
```

Because we want to store each (`user_id`, `item_id`) pair only once, we add a unique index on it:

```sql
CREATE UNIQUE INDEX unique_item_bidder_idx ON bidders(user_id, item_id);
```

Collecting a bidder is then a simple atomic operation. If the user has already bid on this item,
we ignore the duplicate:

```sql
INSERT INTO bidders(item_id, user_id) VALUES
($1, $2)
ON CONFLICT (item_id, user_id) DO NOTHING;
```

This way, we neither need to explicitly lock anything nor care about the order of the actions:
adding values to a set is commutative.

### Collecting MAX values

Now, let's collect the maximum bid for each item. We have a table that stores the maximum bid
for an item and the user who made it (`user_id`):

```sql
CREATE TABLE max_bids (
  id         BIGINT GENERATED ALWAYS AS IDENTITY,
  item_id    BIGINT NOT NULL,
  user_id    BIGINT NOT NULL,
  bid        NUMERIC NOT NULL
);
```

There can be only one maximum bid per item, so we create a unique index on `item_id`.

```sql
CREATE UNIQUE INDEX unique_bids_idx ON max_bids(item_id);
```

When we insert a new bid, it is added if the `item_id` has no bid yet. If the `item_id` already
has a bid (`ON CONFLICT (item_id)`), we update that row only if the new bid is higher
(`WHERE EXCLUDED.bid > max_bids.bid`):

```sql
INSERT INTO max_bids (item_id, user_id, bid)
VALUES ($1, $2, $3)
ON CONFLICT (item_id) DO UPDATE
    SET bid        = EXCLUDED.bid,
        user_id    = EXCLUDED.user_id
    WHERE EXCLUDED.bid > max_bids.bid;
```
If we need to know whether the value changed (through an insert or update), we can inspect the
number of affected rows. It will be 0 if the bid was neither inserted nor updated.
Most database libraries expose this information. For example, Go provides
`sql.Result.RowsAffected()`.

### Incrementing a counter

Now, let's say we want to know how many times each item has received a bid.
For each item, we have a counter that tracks the number of bids:

```sql
CREATE TABLE bids_count (
  id         BIGINT GENERATED ALWAYS AS IDENTITY,
  item_id    BIGINT NOT NULL,
  "count"    INT NOT NULL DEFAULT 1
);
```

Again, because we want to track the count per `item_id`, we need a unique index on it.

```sql
CREATE UNIQUE INDEX unique_bids_count_idx ON bids_count(item_id);
```

Updating the count is a single atomic operation `count = count + 1`:

```sql
INSERT INTO bids_count(item_id)
VALUES ($1)
ON CONFLICT (item_id) DO UPDATE
SET "count" = bids_count."count" + 1;
```

## Conditional writes

When an operation cannot be expressed in a commutative form, we can use a conditional update to
avoid acting on stale data. Assume each user has a wallet in which we track their balance and the
funds `reserved` for active bids:

```sql
CREATE TABLE wallets (
  user_id  BIGINT PRIMARY KEY,
  balance  NUMERIC NOT NULL,
  reserved NUMERIC NOT NULL DEFAULT 0,
  version  BIGINT NOT NULL DEFAULT 0
);
```

Multiple processes can try to update the same wallet at the same time: a user might place several
bids at once, or a bidding bot might submit them in parallel. Placing a bid is a read-modify-write
operation: we read the wallet, calculate the new `balance` and `reserved` values in application
code, and write them back. If we blindly read and then update, another transaction can change the
wallet between our read and write, causing us to overwrite its result (a lost update).

One option is to do this pessimistically by locking the row:

```sql
BEGIN;

SELECT balance, reserved, version
FROM wallets
WHERE user_id = $1
FOR UPDATE;

-- updates

COMMIT;
```

`FOR UPDATE` prevents other transactions from modifying the row while the lock is held.
The problem is lock contention: any other transaction that wants to update this row must wait.

Another option is to do this optimistically with compare-and-swap (CAS), updating a record only
if it has not changed. First, read the row without an explicit lock (`FOR UPDATE`):

```sql
SELECT balance, reserved, version
FROM wallets
WHERE user_id = $1;
```

Then update it only if nothing has changed since we read it:

```sql
UPDATE wallets
SET balance = $1, reserved = $2
WHERE user_id = $3 AND balance = $4 AND reserved = $5;
```
The number of affected rows tells us whether the update succeeded: 1 means that it did, while 0
no row matched: either the values changed or the row no longer exist.

Comparing values does not detect an [ABA problem](https://en.wikipedia.org/wiki/ABA_problem):
a value can change from 100 to 50 and back to 100 before the update.
This is harmless when the operation depends only on the current compared values.
However, it may be incorrect in other use cases, such as when intermediate transitions produce
side effects or correctness depends on the state not having changed.

This is where a version comes into play. In this case, the version is a monotonically increasing
value. The rule is simple: whenever the row is updated, the writer must increment its version.
Each update therefore produces a distinct version, allowing us to detect any concurrent
modification with a single comparison.

For example, suppose the balance is 100 at version 10. One process changes the balance to 50 and
the version to 11, and then another process changes the balance back to 100 and the version to 12.
When the original process, which read version 10, tries to update the balance, its update affects
no rows, signaling a conflict.

The update is then simply:

```sql
UPDATE wallets
SET balance = $1, reserved = $2, version = version + 1
WHERE user_id = $3 AND version = $4;
```

The number of affected rows tells us whether the update occurred. If it is 0, another writer has
modified the row, and the application must retry: read the latest values, recalculate `balance` and
`reserved`, and attempt to store them again.
