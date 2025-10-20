.PHONY: up down test

up:
	docker compose up -d

down:
	docker compose down

test: up
	ELASTICSEARCH_ADDRESS=http://localhost:9200 go test ./...
