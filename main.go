package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
)

func main() {
	var searchPath, dbPath, addr, searchEnv string
	flag.StringVar(
		&searchPath,
		"cmd",
		"./meilisearch",
		"Where the search binary is located",
	)
	flag.StringVar(
		&dbPath,
		"dbpath",
		"./meili_data",
		"Where the search binary will store data",
	)
	flag.StringVar(
		&addr,
		"addr",
		"127.0.0.1:7700",
		"Where the search binary will listen on",
	)
	flag.StringVar(
		&searchEnv,
		"env",
		"development",
		"Search binary environment",
	)

	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	spm := NewSearchProcessManager(
		searchPath,
		dbPath,
		addr,
		searchEnv,
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

	searchUrl, _ := url.Parse(SearchUrl)
	rp := httputil.NewSingleHostReverseProxy(searchUrl)
	srv := http.Server{Addr: ":8080", Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// log request
		log.Printf("url: %s, method: %s", request.URL, request.Method)

		// if the request was '/' or '/scrape', serve
		if strings.HasPrefix(request.URL.RequestURI(), "/scrape") ||
			request.URL.RequestURI() == "/" {
			mux.ServeHTTP(writer, request)
		} else {
			// otherwise proxy request to search
			rp.ServeHTTP(writer, request)
		}
	})}

	// start http server
	go func() {
		log.Println("Starting HTTP server")

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server Shutdown: %s\n", err)
			os.Exit(1)
		}
	}()

	s := <-sigChan
	log.Printf("signal recieved: %s", s)

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
