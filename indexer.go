package main

import (
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"log"
)

const IndexName = "sites"
const SearchUrl = "http://localhost:7700"

type Indexer interface {
	Index(doc ScrapedDoc) error
}

type MeilisearchIndexer struct {
	index *meilisearch.Index
}

func (m MeilisearchIndexer) Index(doc ScrapedDoc) error {
	task, err := m.index.AddDocuments(doc)
	if err != nil {
		return fmt.Errorf("could not index scraped doc: %w", err)
	}
	log.Printf("indexing %s with task UID %d\n", doc.URL, task.TaskUID)
	return nil
}

func NewMeilisearchIndexer(index *meilisearch.Index) MeilisearchIndexer {
	if index == nil {
		panic("index field cannot be nil")
	}
	return MeilisearchIndexer{
		index: index,
	}
}

func MakeMeilisearchIndex(host, apiKey string) *meilisearch.Index {
	c := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   host,
		APIKey: apiKey,
	})
	return c.Index(IndexName)
}
