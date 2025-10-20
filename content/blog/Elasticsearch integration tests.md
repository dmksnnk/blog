---
date: '2025-10-20T20:10:10+02:00'
draft: false
title: 'Elasticsearch per-index integration tests'
slug: 'elasticsearch-integration-tests'
showToc: true
cover:
    image: ''
summary: "Faster Elasticsearch tests without starting containers for each test."
tags:
    - elasticsearch
    - testing
    - Go
---

One of the approaches to do integration tests with Elasticsearch is to use [testcontainers](https://testcontainers.com/).
It is recommended by [Elastic](https://www.elastic.co/search-labs/blog/tests-with-mocks-and-real-elasticsearch),
it works well and provides a high level of isolation between tests.
But, it comes with drawbacks: Elasticsearch containers are heavy and take time to start.

What if we use a different kind of isolation? The index seems pretty isolated.
Instead of starting a new container for each test, we can just start one container and
create a new index for each test case.

It is pretty simple to implement. Let's go through it.

## A Bookstore

We want to create a storage for books on top of Elasticsearch and want to index and search books by title and author.
Here are the mappings for the index:

```json
{
    "properties": {
        "title": {
            "type": "text"
        },
        "author": {
            "type": "text"
        }
    }
}
```

And the full code for the storage is in [storage.go](https://github.com/dmksnnk/blog/tree/main/examples/elasticsearch-integration/storage.go)
(it is too long to include here in full).

## Setting up Elasticsearch

Assume we have a running Elasticsearch instance in a container.
You can use the [docker-compose.yml](https://github.com/dmksnnk/blog/tree/main/examples/elasticsearch-integration/docker-compose.yaml)
for starters.

What I like to do is to have test helpers that abstract away the details of setting up infrastructure for tests.
We will do just that. We will create a `storagetest` package with a helper to create a new storage with a unique index for each test.

We need to pass the Elasticsearch address to communicate with it. The simplest way is to use an environment variable:

```go
address := os.Getenv("ELASTICSEARCH_ADDRESS")
```

The next step is to create a unique index name for each test.
There are [some rules](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-indices-create#operation-indices-create-path)
for the index name, and we will do our best to follow them: take the test name, lowercase it,
remove unsupported characters, and add a random suffix, so the index name is unique.

{{< details summary="func newIndexName(t *testing.T) string" >}}

```go
func newIndexName(t *testing.T) string {
    name := strings.ToLower(t.Name())
    if len(name) > 247 { // 247 = 255 (max 255 bytes index name) - 8 (random suffix) - 1 (underscore)
        name = name[:247]
    }

    mapper := func(r rune) rune {
        // allow only [a-z0-9_]
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
            return r
        }

        return '_'
    }
    name = strings.Map(mapper, name)

    return name + "_" + randString()
}
```

{{< /details >}}


Then, we create a new index and apply a mapping to it:

```go
response, err := client.Indices.Create(indexName)
...

response, err := client.Indices.PutMapping(
    []string{indexName},
    strings.NewReader(storage.Mappings),
)
...

```

[Here](https://github.com/dmksnnk/blog/tree/main/examples/elasticsearch-integration/embed.go)
we are using [embed](https://pkg.go.dev/embed) to embed the mapping file into the binary.
But you can read it from anywhere else, like a local file if it is in the same repo or from a remote git repository.

The most useful part comes next: **we delete the index only if the test passed**.
If it failed, we can inspect the index and see what went wrong. We'll see it in action later.

```go
t.Cleanup(func() {
    if t.Failed() {
        t.Logf("test failed, keeping index %s for inspection", indexName)
        return
    }

    deleteIndex(t, client, indexName)
})
```

And that's pretty much it. Here is the full code for the helper in
[storagetest/storage.go](https://github.com/dmksnnk/blog/tree/main/examples/elasticsearch-integration/storagetest/storage.go).


## Writing tests

The test itself looks short and sweet, all with the help of our helper.
Store a document and search for it:

```go
func TestStorage(t *testing.T) {
    store := storagetest.NewBookstore(t)

    book := storage.Book{
        Title:  "The Great Gatsby",
        Author: "F. Scott Fitzgerald",
    }

    if err := store.IndexBook(context.TODO(), book); err != nil {
        t.Fatalf("index document: %s", err)
    }

    books, err := store.Search(context.TODO(), "Gatsby")
    if err != nil {
        t.Fatalf("search: %s", err)
    }

    if len(books) != 1 {
        t.Fatalf("expected 1 book, got %d", len(books))
    }

    if books[0] != book {
        t.Fatalf("expected %+v, got %+v", book, books[0])
    }
}
```

The full code is available in [storage_test.go](https://github.com/dmksnnk/blog/tree/main/examples/elasticsearch-integration/storage_test.go).

Run the test with the Elasticsearch address set:

```sh
ELASTICSEARCH_ADDRESS=http://localhost:9200 go test ./...
```

## Inspecting failed tests

If you change the search query to something that does not match the indexed document:

```go
store.Search(context.TODO(), "Batman")
```

You will see a helpful log message with the index name to inspect:

```
--- FAIL: TestStorage (0.43s)
    storage_test.go:12: using index: teststorage_023f73bc409aa215
...
```

And you are free to explore why the test failed, for example, to see what is inside the index:

```sh
$ curl -s \
    -X POST http://localhost:9200/teststorage_023f73bc409aa215/_search \
    -H "Content-Type: application/json" \
    -d '{"query": {"match_all": {}}}' \
    | jq
{
  "took": 1,
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 1,
      "relation": "eq"
    },
    "max_score": 1.0,
    "hits": [
      {
        "_index": "teststorage_023f73bc409aa215",
        "_id": "zDLOA5oBrce7jb_Hwkex",
        "_score": 1.0,
        "_source": {
          "title": "The Great Gatsby",
          "author": "F. Scott Fitzgerald"
        }
      }
    ]
  }
}
```

The full source code is available on [GitHub](https://github.com/dmksnnk/blog/tree/main/examples/elasticsearch-integration)
