# adding deps

FROM golang:1.24.3-alpine3.21 AS deps
WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN apk --no-cache add make=4.4.1-r2 && \
    go mod download

# building

FROM deps AS builder
ARG BUILD_DIR=/go/src/build

COPY cmd/ ./cmd/
COPY fs.go ./fs.go
COPY public/ ./public/

RUN GOOS=linux GOARCH=amd64 go build -v -o $BUILD_DIR/server ./cmd/...

# certs

FROM golang:1.24.3-alpine3.21 AS certs
RUN apk add --no-cache ca-certificates=20241121-r1

# copy to scratch

FROM scratch
ARG BUILD_DIR=/go/src/build
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder $BUILD_DIR/server /server
ENTRYPOINT [ "/server" ]
