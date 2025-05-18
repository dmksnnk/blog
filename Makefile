BUILD_DIR ?= ./bin
current_dir = $(shell pwd)
version = $(shell git rev-parse --short HEAD)
image = ghcr.io/dmksnnk/blog:$(version)

.PHONY: docker-build
docker-build:
	@docker build --platform linux/amd64 -f Dockerfile --tag=$(image) .

.PHONY: hugo-build-local
hugo-build-local:
	@hugo build --buildDrafts --gc --baseURL localhost:8080