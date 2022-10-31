package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const IdleTimeout = 5 * time.Minute

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	ticker := time.NewTicker(IdleTimeout)
	defer ticker.Stop()

	spm := NewSearchProcessManager(
		"/home/shashank/working/dev/scrap/meilisearch",
		"/home/shashank/working/dev/scrap/meili_data",
		"127.0.0.1:7700",
		"development",
	)
	if err := spm.Start(); err != nil {
		log.Println("could not start search:", err)
		os.Exit(1)
	}
	log.Println("started search server")
	mux := http.NewServeMux()
	index := MakeMeilisearchIndex(SearchUrl, "")
	indexer := NewMeilisearchIndexer(index)
	scraper := NewCollyScraper(indexer)

	MakeRoutes(scraper, mux)

	srv := http.Server{Addr: ":8080", Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// log request
		log.Printf("url: %s, method: %s", request.URL, request.Method)

		ticker.Reset(IdleTimeout)

		mux.ServeHTTP(writer, request)
	})}

	// start http server
	go func() {
		log.Println("Starting HTTP server")

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server Shutdown: %s\n", err)
			os.Exit(1)
		}
	}()

	select {
	case s := <-sigChan:
		log.Printf("signal recieved: %s", s)
	case <-ticker.C:
		log.Println("idle timeout")
	}
	// shutdown the server
	log.Println("shutting down server")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server Shutdown: %s\n", err)
	}
	log.Println("server shutdown")

	// wait to finish scraping
	log.Println("waiting for scraper to finish")
	scraper.C.Wait()
	log.Println("scraper finished")

	// stop the index
	log.Println("sending stop signal to search server")
	if err := spm.Stop(); err != nil {
		log.Printf("search server shutdown: %s\n", err)
	}
	// wait on the process
	log.Println("waiting on search server process")
	if err := spm.Wait(); err != nil {
		log.Printf("search server shutdown: %s\n", err)
	}
	log.Println("search server shutdown")
}
