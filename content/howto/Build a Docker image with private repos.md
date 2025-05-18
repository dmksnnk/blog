---
date: '2025-01-19T13:00:00+02:00'
draft: true
title: 'Build a Docker image with private repos'
slug: 'build-docker-image-private-repos'
tags:
  - 'Docker'
  - 'Go'
  - 'CI/CD'
  - 'Secrets'
---

To build a Docker image that has dependencies from private repositories, we should pass the `.netrc` configuration (see more in [How to access private repos in Go](/howto/access-private-repos-go)). Ideally, this configuration should be passed as a secret:

```docker
FROM golang:1.22.2-alpine3.19 AS deps

COPY go.mod go.sum ./
# Copy secrets from /kaniko for Kaniko or from /run/secrets for Podman/Buddah
RUN --mount=type=secret,id=netrc \
    cp /kaniko/netrc "$HOME"/.netrc || cp /run/secrets/netrc "$HOME"/.netrc && \
    go mod download && \
    rm "$HOME"/.netrc
```

## GitLab CI/CD

In GitLab CI, to build the image, we need to create a `.netrc` configuration with `CI_JOB_TOKEN` and pass it to Kaniko secrets. Here is an example:

```yaml
build:
  image:
    name: gcr.io/kaniko-project/executor:v1.14.0-debug
    entrypoint: [""]
  before_script:
    - mkdir -p "${HOME}"
    - mkdir -p /kaniko/.docker
    - echo "{\"auths\":{\"${CI_REGISTRY}\":{\"auth\":\"$(printf \"%s:%s\" \"${CI_REGISTRY_USER}\" \"${CI_REGISTRY_PASSWORD}\" | base64)\"}}}" > /kaniko/.docker/config.json
    # Here we are creating the .netrc configuration, which will be passed to the image as a secret
    - echo -e "machine private.example.com\nlogin gitlab-ci-token\npassword ${CI_JOB_TOKEN}" > /kaniko/netrc
  script:
    - /kaniko/executor \
      --context "$CI_PROJECT_DIR" \
      --dockerfile "$CI_PROJECT_DIR/${DOCKERFILE:-Dockerfile}" \
      --cache=true \
      --build-arg "VERSION=$CI_COMMIT_REF_NAME" \
      --build-arg "NETRC_CONFIG=$NETRC_CONFIG" \
      --destination "${CI_REGISTRY_IMAGE}${DESTINATION}:$CI_COMMIT_REF_NAME" \
      --destination "${CI_REGISTRY_IMAGE}${DESTINATION}:latest"
```

If you pass the `.netrc` configuration as an `ARG` in the Dockerfile, Docker won't be able to cache this step because `.netrc` will be different each time the pipeline runs—`CI_JOB_TOKEN` is unique per pipeline.

## Docker Build

The Docker build command will look like this:

```sh
docker build -f Dockerfile --secret id=netrc,src=$HOME/.netrc .
```

## Docker Compose

In the Docker Compose file, add the local `.netrc` configuration as a secret and pass it into the build context:

```yaml
services:
  my-app:
    build:
      context: .
      secrets:
        - netrc
      
secrets: # Passing local .netrc to the image
  netrc:
    name: netrc
    file: ${HOME}/.netrc
```
