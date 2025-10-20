package storagetest

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	storage "example.com/elasticsearch-ingeration"
	"github.com/elastic/go-elasticsearch/v9"
)

// NewBookstore creates a new storage.Boostore instance with a unique index for testing.
func NewBookstore(t *testing.T) *storage.Bookstore {
	t.Helper()

	address := os.Getenv("ELASTICSEARCH_ADDRESS")
	if address == "" {
		t.Fatal("please provide Elasticsearch address via ELASTICSEARCH_ADDRESS environment variable")
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
	})
	if err != nil {
		t.Fatalf("create Elasticsearch client: %s", err)
	}

	indexName := newIndexName(t)

	createIndex(t, client, indexName)
	putMapping(t, client, indexName)

	t.Logf("using index: %s", indexName)

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("test failed, keeping index %s for inspection", indexName)
			return
		}

		deleteIndex(t, client, indexName)
	})

	return storage.NewBookstore(client, indexName)
}

func newIndexName(t *testing.T) string {
	name := strings.ToLower(t.Name())
	if len(name) > 247 { // 247 = 255 (max 255 bytes index name) - 8 (random suffix) - 1 (underscore)
		name = name[:247]
	}

	mapper := func(r rune) rune {
		// allow only [a-z0-9_]
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}

		return '_'
	}
	name = strings.Map(mapper, name)

	return name + "_" + randString()
}

func putMapping(t *testing.T, client *elasticsearch.Client, indexName string) {
	response, err := client.Indices.PutMapping(
		[]string{indexName},
		strings.NewReader(storage.Mappings),
	)
	if err != nil {
		t.Fatalf("put mapping: %s", err)
	}
	defer response.Body.Close()

	if response.IsError() {
		t.Fatalf("put mapping: %s", response.String())
	}
}

func createIndex(t *testing.T, client *elasticsearch.Client, indexName string) {
	response, err := client.Indices.Create(indexName)
	if err != nil {
		t.Fatalf("create index: %s", err)
	}
	defer response.Body.Close()

	if response.IsError() {
		t.Fatalf("error creating index: %s", response.String())
	}
}

func deleteIndex(t *testing.T, client *elasticsearch.Client, indexName string) {
	t.Helper()

	res, err := client.Indices.Delete([]string{indexName})
	if err != nil {
		t.Fatalf("delete index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		t.Fatalf("delete index: %s", res.String())
	}
}

func randString() string {
	s := make([]byte, 8)
	rand.Read(s)

	return hex.EncodeToString(s)
}
