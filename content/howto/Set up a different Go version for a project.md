---
date: '2024-08-26T16:00:00+02:00'
title: 'Set up a different Go version for a project'
slug: 'setup-project-go-version'
description: 'How to use a specific Go version for your project in VS Code and GoLand.'
tags:
  - 'Go'
  - 'VSCode'
  - 'GoLand'
---

Let’s say you have Go 1.20.1 installed locally, but the project you are working on requires Go 1.22.3.

First, install the additional version of the [Go SDK](https://go.dev/doc/manage-install).

## VS Code

Set `GOROOT` to point to the Go version you need. Add the following to the workspace settings:

```json
"go.goroot": "${env:HOME}/sdk/go1.22.3"
```

## GoLand

Go to Settings → Go → GOROOT and add or select the version you need there.