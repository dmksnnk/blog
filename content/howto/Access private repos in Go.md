---
date: '2024-08-26T16:30:00+02:00'
draft: true
title: 'How to access private repos in Go'
slug: 'access-private-repos-go'
tags:
  - 'Go'
  - 'VSCode'
  - 'GoLand'
---

When downloading packages, Go checks the public [Go Checksum Database](https://sum.golang.org/) to verify the package checksum. Private git packages, like the ones on private GitLab, aren’t in that database, so downloading will fail.

This can be fixed by setting `GOPRIVATE=private.example.com`, which will instruct Go not to check the Checksum Database for the packages on the `private.example.com` domain.

For GitLab, first set up access to repositories:

1. Create a GitLab Access Token with the `read_api` scope.

2. Update `${HOME}/.netrc` file, and add:

```
machine private.example.com
    login your.email@example.com
    password gitlab_access_token
```

More info can be found in the [GitLab docs](https://docs.gitlab.com/ee/user/project/use_project_as_go_package.html).

Then, set up the IDE to work with private git packages.

## GoLand

Go to Settings → Go → Go Modules and add `GOPRIVATE=private.example.com` there.

## VS Code

Add to your settings:

```json
"go.toolsEnvVars": {
    "GOPRIVATE": "private.example.com"
}
```
