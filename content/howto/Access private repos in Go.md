---
date: '2024-08-26T16:30:00+02:00'
title: 'How to access private repos in Go'
slug: 'access-private-repos-go'
description: 'Configure Go to access private repositories.'
tags:
  - 'Go'
  - 'VSCode'
  - 'GoLand'
---

When downloading packages, Go checks the public [Go Checksum Database](https://sum.golang.org/) to verify the package checksum. Private Git packages, such as those on a private GitLab instance, aren’t in that database, so downloading them will fail.

This can be fixed by setting `GOPRIVATE=private.example.com`. This instructs Go not to check the Checksum Database for packages in the `private.example.com` domain.

For GitLab, first set up access to repositories:

1. Create a GitLab Access Token with the `read_api` scope.
2. Update your `${HOME}/.netrc` file, and add:

```
machine private.example.com
    login your.email@example.com
    password gitlab_access_token
```

More information can be found in the [GitLab documentation](https://docs.gitlab.com/ee/user/project/use_project_as_go_package.html).

Then, set up the IDE to work with private Git packages.

## GoLand

Go to Settings → Go → Go Modules and add `GOPRIVATE=private.example.com`.

## VS Code

Add the following to your settings:

```json
"go.toolsEnvVars": {
    "GOPRIVATE": "private.example.com"
}
```
