package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
)

const indexName = "warehouse"

func main() {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		log.Fatalf("error creating the client: %s", err)
	}

	dropIndexIfExists(client)
	item := initStock(client)
	log.Printf("initial item state: %+v", item)

	// signals to coordinate the two processes
	p1 := signals{searched: make(chan struct{}), updated: make(chan struct{})}
	p2 := signals{searched: make(chan struct{}), updated: make(chan struct{})}

	// launch two processes that will try to update the same document
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		process1(client, p1, p2)
	}()

	go func() {
		defer wg.Done()

		process2(client, p1, p2)
	}()

	wg.Wait()

	searchResult := findItem(client)
	log.Printf("final item state: %s", searchResult.String())
}

func dropIndexIfExists(client *elasticsearch.Client) {
	res, err := client.Indices.Delete([]string{indexName}, client.Indices.Delete.WithAllowNoIndices(true))
	if err != nil {
		log.Fatalf("error deleting index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("error deleting index: %s", res.String())
	}
}

func initStock(client *elasticsearch.Client) item {
	item := item{
		ID:    123,
		Stock: 0,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(item); err != nil {
		log.Fatalf("error encoding item: %s", err)
	}

	response, err := client.Index(
		indexName,
		&buf,
		client.Index.WithRefresh("true"),
	)
	if err != nil {
		log.Fatalf("error indexing item: %s", err)
	}
	defer response.Body.Close()

	if response.IsError() {
		log.Fatalf("error indexing item: %s", response.String())
	}

	return item
}

type signals struct {
	searched chan struct{}
	updated  chan struct{}
}

func process1(client *elasticsearch.Client, p1, p2 signals) {
	response := findItem(client)

	log.Printf("process1: item found: %s", response.String())
	close(p1.searched) // signal to process2 that process1 has searched
	<-p2.searched      // wait for process2 to finish search

	// add 10 to stock
	response.Hits.Hits[0].Source.Stock += 10
	updateResponse := updateItem(client, response)
	defer updateResponse.Body.Close()

	if updateResponse.IsError() {
		panic(updateResponse.String())
	}

	log.Printf("process1: item updated: %s", response.String())
	close(p1.updated) // signal to process2 that process1 has updated
}

func process2(client *elasticsearch.Client, p1, p2 signals) {
	<-p1.searched // wait for process1 to finish searching

	response := findItem(client)

	log.Printf("process2: item found: %s", response.String())
	close(p2.searched) // signal to process1 that process2 has searched
	<-p1.updated       // wait for process1 to finish updating

	// add 15 to stock
	response.Hits.Hits[0].Source.Stock += 15
	updateResponse := updateItem(client, response)
	defer updateResponse.Body.Close()

	if updateResponse.StatusCode != 409 {
		log.Fatalf("process2: expected a version conflict, got: %s", updateResponse.String())
	}

	log.Printf("process2: version conflict occurred, fetching the latest document version")

	response = findItem(client)
	log.Printf("process2: item found: %s", response.String())

	// add 15 to stock again
	response.Hits.Hits[0].Source.Stock += 15
	updateResponse = updateItem(client, response)
	defer updateResponse.Body.Close()

	if updateResponse.IsError() {
		panic(updateResponse.String())
	}

	log.Printf("process2: item updated: %s", response.String())
}

func findItem(client *elasticsearch.Client) warehouseSearchResponse {
	response, err := client.Search(
		client.Search.WithIndex(indexName),
		client.Search.WithBody(strings.NewReader(`{
			"query": {
				"term": {
					"id": 123
				}
			}
		}`),
		),
		client.Search.WithSeqNoPrimaryTerm(true),
	)
	if err != nil {
		log.Fatalf("error searching for item: %s", err)
	}
	defer response.Body.Close()

	if response.IsError() {
		log.Fatalf("error searching for item: %s", response.String())
	}

	var responseBody warehouseSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		log.Fatalf("error decoding response body: %s", err)
	}

	if len(responseBody.Hits.Hits) == 0 {
		log.Fatalf("item not found")
	}
	if len(responseBody.Hits.Hits) != 1 {
		log.Fatalf("unexpected number of items found: %d", len(responseBody.Hits.Hits))
	}

	return responseBody
}

func updateItem(client *elasticsearch.Client, search warehouseSearchResponse) *esapi.Response {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(search.Hits.Hits[0].Source); err != nil {
		log.Fatalf("error encoding item: %s", err)
	}

	updateResponse, err := client.Index(
		indexName,
		&buf,
		client.Index.WithDocumentID(search.Hits.Hits[0].ID),
		client.Index.WithIfSeqNo(search.Hits.Hits[0].SeqNo),
		client.Index.WithIfPrimaryTerm(search.Hits.Hits[0].PrimaryTerm),
		client.Index.WithRefresh("true"),
	)
	if err != nil {
		log.Fatalf("error updating item: %s", err)
	}

	return updateResponse
}

type warehouseSearchResponse struct {
	Hits struct {
		Hits []struct {
			ID          string `json:"_id"`
			SeqNo       int    `json:"_seq_no"`
			PrimaryTerm int    `json:"_primary_term"`
			Source      item   `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (r *warehouseSearchResponse) String() string {
	hit := r.Hits.Hits[0]

	return fmt.Sprintf("item: id=%d, stock=%d, seq_no=%d, primary_term=%d", hit.Source.ID, hit.Source.Stock, hit.SeqNo, hit.PrimaryTerm)
}

type item struct {
	ID    int `json:"id"`
	Stock int `json:"stock"`
}
