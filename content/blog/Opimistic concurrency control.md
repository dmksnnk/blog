---
date: '2025-09-14T14:14:17+02:00'
draft: false
title: 'Optimistic concurrency control in Elasticsearch'
slug: 'optimistic-elasticsearch-updates'
showToc: true
cover:
    image: 'images/conflict.svg'
summary: "Update documents in Elasticsearch without losing data."
tags:
    - architecture
    - elasticsearch
    - database
    - concurrency
---

We will go through how to update documents in Elasticsearch without losing data.

## Scenario

Let's assume we have an index of items and their count in stock, and we have multiple processes trying to
create or update existing counts. Here is how a document might look:

```json
{
    "mappings": {
        "properties": {
            "id": { "type": "integer" },
            "stock": { "type": "integer" }
        }
    }
}
```

Our warehouse inventory system tells us that we don't have any vacuum cleaners in stock.
In one warehouse, we counted 10 items in stock, and in another, 15.
Now, both warehouse managers try to update the count of items in the database:

![Concurrent updates](images/concurrent-updates.svg)

The first warehouse manager (process 1) checks how many vacuum cleaners we have and gets 0.
It updates the count to 10.
At the same time, the second warehouse manager (process 2) checks how many vacuum cleaners we have and also gets 0.
It updates the count to 15.

Next, when the first manager checks the number again, it sees that the number is 15. Not good. We got ourselves into a data race.

## Optimistic concurrency control

Why is it called "optimistic"? Because we assume that everything will be OK and proceed with the update
instead of taking measures upfront. With pessimistic concurrency control, we would take a lock on the document
and update the value, allowing only one process to work with the document at a time.
Both approaches have their pros and cons.
Optimistic concurrency control should allow for higher throughput in scenarios where conflicts don't happen often.

## Atomicity

Elasticsearch does not support transactions like classic relational databases do, it
supports atomic operations on a single document only.
This means only one process can update a document at a time, and
if the operation completes successfully or fails, the document won't be left in a partial or inconsistent state.

## Versions

Using optimistic concurrency control, we will try to update a document only when the version of the document has not changed.
If it has changed, we need to fetch the new version of the document and update it.

![Concurrent updates with conflict](images/concurrent-updates-conflict.svg)

Assuming we have a document with version 1 in the database. When the first process fetches the document,
it also receives its version. When it updates the document, it tells the database to update the document
only if the version has not changed. In our scenario, the first process gets the document with
version 1, updates it, and tells the database to update if the version has not changed. The database successfully stores
the document and increments the version to 2.
The second process also received the document with version 1, but when it updates it,
the database sees that the versions do not match and returns a conflict status.
To proceed, the second process needs to handle this conflict. It should fetch the new version of
the document and try to update it again.

Elasticsearch has always increasing counters `_seq_no` and `_primary_term`.
`_seq_no` is incremented every time the document changes. `_primary_term` is incremented
every time a shard is promoted to primary. Both `_seq_no` and `_primary_term` give us a unique version or generation
of the document in space and time.

With Elasticsearch, optimistic concurrency control can be achieved by specifying the
[`if_primary_term`](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-update#operation-update-if_primary_term)
and
[`if_seq_no`](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-update#operation-update-if_seq_no)
query parameters when updating the document. `if_primary_term` tells Elasticsearch to update the document if no shard
rebalancing happened, and `if_seq_no` if the document itself was not changed.

{{< alert note >}}

For simple counter increments, using the Update API with a script is generally a better approach
in Elasticsearch, as it updates document atomically in-place.

```
POST /warehouse/_update/123

{
  "script" : {
    "source": "ctx._source.stock += params.count",
    "lang": "painless",
    "params" : {
      "count" : 10
    }
  }
}
```

However, the optimistic concurrency control approach described here is a more general example
applicable to other databases and scenarios beyond simple counter updates.

{{< /alert >}}

## Example

Let's try to simulate the above scenario in Go.
We will create document with 0 stock, and then run two goroutines that will try to update the stock concurrently.
You can find the full code with conflict resolution
[on GitHub](https://github.com/dmksnnk/blog/tree/main/examples/optimistic-elasticsearch-updates).


Simply start an Elasticsearch instance in Docker:

```bash
docker compose up -d
```

Then, run the Go program:

```bash
go run main.go
```

You should see output similar to this:

```
2025/09/14 20:49:39 initial item state: {ID:123 Stock:0}
2025/09/14 20:49:39 process1: item found: item: id=123, stock=0, seq_no=0, primary_term=1
2025/09/14 20:49:39 process2: item found: item: id=123, stock=0, seq_no=0, primary_term=1
2025/09/14 20:49:39 process1: item updated: item: id=123, stock=10, seq_no=0, primary_term=1
2025/09/14 20:49:39 process2: version conflict occurred, fetching the latest document version
2025/09/14 20:49:39 process2: item found: item: id=123, stock=10, seq_no=1, primary_term=1
2025/09/14 20:49:39 process2: item updated: item: id=123, stock=25, seq_no=1, primary_term=1
2025/09/14 20:49:39 final item state: item: id=123, stock=25, seq_no=2, primary_term=1
```

This shows how Elasticsearch updates the seq_no when the first process updates the document.
The second process receives a conflict status and must retrieve the latest document version to proceed.

---

## Links

- [Optimistic concurrency control](https://www.elastic.co/docs/reference/elasticsearch/rest-apis/optimistic-concurrency-control)
