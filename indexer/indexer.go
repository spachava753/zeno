package indexer

import (
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
	"zeno/domain"
)

const IndexName = "sites"
const SearchUrl = "http://localhost:7700"
const ZenoKeyEnv = "ZENO_KEY"

type Indexer interface {
	Index(doc domain.ScrapedDoc) error
	Delete(doc domain.ScrapedDoc) error
}

type MeilisearchIndexer struct {
	index *meilisearch.Index
}

func (m MeilisearchIndexer) Index(doc domain.ScrapedDoc) error {
	task, err := m.index.AddDocuments(doc)
	if err != nil {
		return fmt.Errorf("could not index scraped doc: %w", err)
	}
	log.Printf("indexing %s with task UID %d\n", doc.URL, task.TaskUID)
	return nil
}

func (m MeilisearchIndexer) Delete(doc domain.ScrapedDoc) error {
	task, err := m.index.DeleteDocument(doc.ID)
	if err != nil {
		return fmt.Errorf("could not delete scraped doc: %w", err)
	}
	log.Printf("delete %s with task UID %d\n", doc.ID, task.TaskUID)
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

func MakeMeilisearchIndex(host, apiKey string) (*meilisearch.Index, func() bool) {
	c := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    host,
		APIKey:  apiKey,
		Timeout: 100 * time.Millisecond,
	})
	return c.Index(IndexName), c.IsHealthy
}

type SearchProcessManager struct {
	cmd     *exec.Cmd
	check   func() bool
	stopSig chan os.Signal
}

func NewSearchProcessManager(cmdPath, dbPath, addr, apiKey string, check func() bool, sigChan chan os.Signal) SearchProcessManager {
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
		cmd:     cmd,
		check:   check,
		stopSig: sigChan,
	}
}

func (s *SearchProcessManager) Start() error {
	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	// run health check
	go func() {
		time.Sleep(5 * time.Second)
		log.Println("started search health check")
		failCount := 0
		for failCount < 3 {
			if !s.check() {
				failCount += 1
			} else {
				failCount = 0
			}
			time.Sleep(1 * time.Second)
		}

		log.Println("search health check failed")

		if err := s.Stop(); err != nil {
			log.Println("err stopping search", err)
		}

		if waitErr := s.Wait(); waitErr != nil {
			log.Println("err waiting on search", waitErr)
		}

		log.Println("sending signal to shutdown")
		s.stopSig <- os.Interrupt
	}()
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
