package main

import (
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"log"
	"os"
	"os/exec"
	"syscall"
)

const IndexName = "sites"
const SearchUrl = "http://localhost:7700"
const ZenoKeyEnv = "ZENO_KEY"

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

type SearchProcessManager struct {
	cmd *exec.Cmd
}

func NewSearchProcessManager(cmdPath, dbPath, addr, apiKey string) SearchProcessManager {
	cmd := exec.Command(cmdPath,
		"--db-path", dbPath,
		"--http-addr", addr)
	if apiKey != "" {
		log.Println("using production env for search")
		cmd.Args = append(cmd.Args, "--env=production", "--master-key", apiKey)
	} else {
		log.Println("using development env for search")
		cmd.Args = append(cmd.Args, "--env=development")
	}
	return SearchProcessManager{
		cmd: cmd,
	}
}

func (s *SearchProcessManager) Start() error {
	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	return s.cmd.Start()
}

func (s *SearchProcessManager) Stop() error {
	// check if the search index is doing any task, then interrupt
	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err != nil {
		return err
	}
	return syscall.Kill(-pgid, syscall.SIGINT)
}

func (s *SearchProcessManager) Wait() error {
	return s.cmd.Wait()
}
