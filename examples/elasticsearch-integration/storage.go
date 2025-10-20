package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9"
)

// Bookstore is a books storage.
type Bookstore struct {
	client *elasticsearch.Client
	index  string
}

func NewBookstore(client *elasticsearch.Client, index string) *Bookstore {
	return &Bookstore{
		client: client,
		index:  index,
	}
}

// IndexBook indexes a new document.
func (s *Bookstore) IndexBook(ctx context.Context, book Book) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(book); err != nil {
		return fmt.Errorf("encode document: %w", err)
	}

	response, err := s.client.Index(
		s.index,
		&body,
		s.client.Index.WithContext(ctx),
		s.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("index document: %w", err)
	}
	defer response.Body.Close()

	if response.IsError() {
		return fmt.Errorf("bad response with status %s: %s", response.Status(), response.String())
	}

	return nil
}

// Search searches for documents matching title or author.
func (s *Bookstore) Search(ctx context.Context, searchParam string) ([]Book, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  searchParam,
				"fields": []string{"title", "author"},
			},
		},
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(query); err != nil {
		return nil, fmt.Errorf("encode query: %w", err)
	}

	response, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(&body),
	)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, fmt.Errorf("bad response with status %s: %s", response.Status(), response.String())
	}

	var result searchResult
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	documents := make([]Book, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		documents = append(documents, hit.Source)
	}

	return documents, nil
}

type Book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

type searchResult struct {
	Hits struct {
		Hits []struct {
			Source Book `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
