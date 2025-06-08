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
