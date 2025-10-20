package storage_test

import (
	"context"
	"testing"

	storage "example.com/elasticsearch-ingeration"
	"example.com/elasticsearch-ingeration/storagetest"
)

func TestStorage(t *testing.T) {
	store := storagetest.NewBookstore(t)

	book := storage.Book{
		Title:  "The Great Gatsby",
		Author: "F. Scott Fitzgerald",
	}

	if err := store.IndexBook(context.TODO(), book); err != nil {
		t.Fatalf("index document: %s", err)
	}

	books, err := store.Search(context.TODO(), "Gatsby")
	if err != nil {
		t.Fatalf("search: %s", err)
	}

	if len(books) != 1 {
		t.Fatalf("expected 1 book, got %d", len(books))
	}

	if books[0] != book {
		t.Fatalf("expected %+v, got %+v", book, books[0])
	}
}
