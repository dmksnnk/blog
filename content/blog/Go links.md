---
date: '2024-08-26T11:00:00+02:00'
title: 'Go Links'
slug: 'go-links'
description: 'A curated collection of useful Go programming links and resources.'
tags:
  - 'Go'
---

This page is a collection of useful links about Go.

## Getting started

- [Tutorial: Get started with Go](https://go.dev/doc/tutorial/getting-started) - Start here.
- [Effective Go](https://go.dev/doc/effective_go) - Continue here.

## Documentation

- [Documentation](https://go.dev/doc/) - Page with links to official documentation.
- [Managing Go installation](https://go.dev/doc/manage-install) - Installing additional Go versions.

## Go Wiki

- [Go Wiki: Home](https://go.dev/wiki/) - It is a collection of information about Go and a curated list of articles about Go.
- [Go Wiki: All Wiki Pages](https://go.dev/wiki/All) - All articles.

## Selected

- [Code Review Comments](https://go.dev/wiki/CodeReviewComments) - Common mistakes and best practices when writing code.
- [Test Comments](https://go.dev/wiki/TestComments) - Common mistakes and best practices when writing tests.

## The Go Blog

- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings) - Difference between string, []byte, rune , UTF-8 and how to iterate over them.
- [Blog Index](https://go.dev/blog/all) - All articles.

## Style guides

- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Google Go Style Guide ](https://google.github.io/styleguide/go/)

## Containers

- [Multi-Stage builds](https://docs.docker.com/build/building/multi-stage/) - Multistage builds with Go.

## Organizing code

- [Organizing a Go module](https://go.dev/doc/modules/layout) - Official one.
- [Go Project Layout](https://github.com/golang-standards/project-layout) - It is not an official one, but it can be used to gather ideas on where to put some exotic parts of the repo. 

## Basic structures internals

- [How Go Arrays Work](https://victoriametrics.com/blog/go-array/index.html)
- [Slices in Go](https://victoriametrics.com/blog/go-slice/index.html)
- [Go Maps Explained](https://victoriametrics.com/blog/go-map/index.html)
- [Golang Defer](https://victoriametrics.com/blog/defer-in-go/index.html)

## Concurrency primitives

- [Go sync.Mutex: Normal and Starvation Mode](https://victoriametrics.com/blog/go-sync-mutex/index.html)
- [Go sync.WaitGroup and The Alignment Problem](https://victoriametrics.com/blog/go-sync-waitgroup)
- [Go sync.Pool and the Mechanics Behind It](https://victoriametrics.com/blog/go-sync-pool)
- [Go sync.Cond, the Most Overlooked Sync Mechanism](https://victoriametrics.com/blog/go-sync-cond)
- [Go Singleflight Melts in Your Code, Not in Your DB](https://victoriametrics.com/blog/go-singleflight)

## Other

- [How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/) - Good ideas around building production quality Go HTTP services.
- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines) - How to write a concurrent pipeline.
    - [Pipeline with errgroup package](https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Pipeline)
    - Package [rheos](https://github.com/dmksnnk/rheos) provides generalized approach for this pattern.
- [Ultimate Visual Guide to Go Enums and iota](https://blog.learngoprogramming.com/golang-const-type-enums-iota-bc4befd096d3) - All about Enums and iota.
- [Using ldflags to Set Version Information for Go Applications](https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications)  - Injecting stuff during build time.
- v2 and Beyond - how to release a new major version:  
    - [Go Modules: v2 and Beyond](https://go.dev/blog/v2-go-modules)
    - [A pragmatic guide to Go module updates](https://carlosbecker.com/posts/pragmatic-gomod-bump/)
- [Go Error Propagation and API Contracts](https://matttproud.com/blog/posts/go-errors-and-api-contracts.html) - TLDR; Errors are the part of your public interface too.
- [Graceful Shutdown in Go: Practical Patterns](https://victoriametrics.com/blog/go-graceful-shutdown/index.html) - Handling signals for graceful shutdowns.

