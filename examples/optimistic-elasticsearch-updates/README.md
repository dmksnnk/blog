# Optimistic Elasticsearch Updates Example

To run this example, you need to have a running Elasticsearch instance:

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

## Shutting down

To stop the Elasticsearch instance, run:

```bash
docker compose down
```