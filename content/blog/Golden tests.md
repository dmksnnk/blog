---
date: '2025-06-04T18:27:15+02:00'
title: 'Golden Tests'
slug: 'golden-tests'
showToc: true
summary: 'Use golden files for testing APIs'
cover:
    image: 'images/golden-brick.svg'
tags:
    - Go
    - testing
---

This post will guide you through testing API responses using golden files.

## testdata

In Go, it is common practice to store test fixtures in the `testdata` folder.
This can be your JSONs for API, CSVs, or expected output files.

If you run `go help test` you'll see this line:

> The go tool will ignore a directory named "testdata", making it available
> to hold ancillary data needed by the tests.

The typical structure of tests with `testdata` looks like this:

```
.
├── mypackage.go
├── mypackage_test.go
└── testdata/
    └── fixture.txt

```

## Golden files

A golden file is a reference file that contains the expected output for a test. It's called "golden" because it represents the gold standard or the correct output that your code should produce. When you run tests, you compare the actual output of your code against the contents of the golden file to verify correctness.

To simplify this, I use a small package named `golden` with helper functions, which I often include in my projects. It's particularly useful when testing against known outputs, such as API responses or generated files.

```go
// Package golden contains test helpers for reading data from ./testdata/ subdirectory.
package golden

import (
    "bytes"
    "io"
    "os"
    "path"
    "strings"
    "testing"
)

// Open file and close on test cleanup.
func Open(t *testing.T, name string) io.ReadSeeker {
    t.Helper()

    f, err := os.Open(path.Join("testdata", name))
    if err != nil {
        t.Fatalf("open file: %s", err)
    }

    t.Cleanup(func() { f.Close() })

    return f
}

// ReadString reads file into string.
func ReadString(t *testing.T, name string) string {
    t.Helper()

    var buf strings.Builder
    _, err := io.Copy(&buf, Open(t, name))
    if err != nil {
        t.Fatalf("copy file: %s", err)
    }

    return buf.String()
}

// ReadBytes reads file into []byte.
func ReadBytes(t *testing.T, name string) []byte {
    t.Helper()

    var buf bytes.Buffer
    _, err := io.Copy(&buf, Open(t, name))
    if err != nil {
        t.Fatalf("copy file: %s", err)
    }

    return buf.Bytes()
}
```

## Testing JSON API with golden files

Now, let's demonstrate how to use golden files in tests. Say, we have a simple HTTP API that accepts a greeting request and returns a personalized message. Here's what the request-response flow looks like:

**Request:**
```json
POST /greet
Content-Type: application/json

{
  "name": "John Doe"
}
```

**Response:**
```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "message": "Hello, John Doe!"
}
```

Our API implementation is straightforward. We define request and response structs, then implement a handler that decodes the incoming JSON, processes it, and returns a formatted greeting:

```go
type GreetRequest struct {
    Name string `json:"name"`
}

type GreetResponse struct {
    Message string `json:"message"`
}

func (a API) Greet(w http.ResponseWriter, r *http.Request) {
    var req GreetRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    resp := GreetResponse{
        Message: fmt.Sprintf("Hello, %s!", req.Name),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

See full code in [api.go](https://github.com/dmksnnk/blog/blob/main/examples/golden/api.go).

Instead of embedding the JSON payloads directly in our test code, we'll create two separate files in our `testdata` directory. The folder structure looks like this:

```
.
├── api.go
├── api_test.go
└── testdata/
    ├── request.json
    └── response.json
```

The `request.json` file contains the exact JSON payload we'll send to our API, while `response.json` contains the expected response.


{{< details summary="request.json" >}}

```json
{
    "name": "John Doe"
}
```

{{< /details >}}


{{< details summary="response.json" >}}

```json
{
    "message": "Hello, John Doe!"
}
```

{{< /details >}}


Here is the test itself. Note the use of the `golden` package for reading from `testdata`:

```go
// reading test request
resp, err := client.Post(srv.URL+"/greet", "application/json", golden.Open(t, "request.json"))
if err != nil {
    t.Fatalf("make request: %s", err)
}
defer resp.Body.Close()

if resp.StatusCode != 200 {
    t.Fatalf("expected status 200, got %d", resp.StatusCode)
}

// asserting response
assertResponse(t, resp, "response.json")
```

The code for the test is available in [api_test.go](https://github.com/dmksnnk/blog/blob/main/examples/golden/api_test.go).

That's it! You can use the same approach for testing generated files, CLI outputs, or any other text-based outputs.

As a next step, consider using `testscript`.
For more details, see [How Go Tests "go test"](https://atlasgo.io/blog/2024/09/09/how-go-tests-go-test)
about testing CLI outputs with golden files and small test scripts.

Use these tools with caution and only where they make sense. Overusing them can unnecessarily
complicate your test code, which means the [tests won't be your helpers](/blog/test-smell/).

For the complete source code and example files, see the [full example on GitHub](https://github.com/dmksnnk/blog/tree/main/examples/golden/).
