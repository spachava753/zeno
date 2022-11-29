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
	"zeno/db"
	"zeno/indexer"
	"zeno/scraper"
)

func main() {
	var searchPath, meiliDataPath, searchAddr, dsn, addr string
	var dev bool
	flag.StringVar(
		&searchPath,
		"cmd",
		"./meilisearch",
		"Where the search binary is located",
	)
	flag.StringVar(
		&meiliDataPath,
		"meili",
		"./meili_data",
		"Where the search binary will store data",
	)
	flag.StringVar(
		&searchAddr,
		"search-addr",
		"127.0.0.1:7700",
		"Where the search binary will listen on",
	)
	flag.StringVar(
		&dsn,
		"dsn",
		"zeno.db",
		"dsn for the document db",
	)
	flag.StringVar(
		&addr,
		"addr",
		":8080",
		"address to use to start server",
	)
	flag.BoolVar(
		&dev,
		"dev",
		false,
		"dev env",
	)

	flag.Parse()

	_, dev = os.LookupEnv("ZENO_DEV")

	log.Println("db path:", dsn)
	log.Println("meili data path:", meiliDataPath)
	log.Println("search address:", searchAddr)
	log.Println("search executable:", searchPath)

	if dev {
		log.Println("starting in dev mode")
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	apiKey := os.Getenv(indexer.ZenoKeyEnv)
	spm := indexer.NewSearchProcessManager(
		searchPath,
		meiliDataPath,
		searchAddr,
		apiKey,
	)
	if err := spm.Start(); err != nil {
		log.Println("could not start search:", err)
		os.Exit(1)
	}
	log.Println("started search server")
	mux := http.NewServeMux()
	index := indexer.MakeMeilisearchIndex(indexer.SearchUrl, apiKey)
	mIndexer := indexer.NewMeilisearchIndexer(index)
	repo := db.NewGormRepo(dsn)
	collyScraper := scraper.NewCollyScraper(mIndexer, repo)

	MakeRoutes(collyScraper, mux, repo)

	searchUrl, _ := url.Parse(indexer.SearchUrl)
	rp := httputil.NewSingleHostReverseProxy(searchUrl)
	srv := http.Server{Addr: addr, Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// log request
		log.Printf("url: %s, method: %s, uri: %s", request.URL, request.Method, request.URL.RequestURI())

		// if the request was '/' or '/scrape', serve
		if strings.HasPrefix(request.URL.String(), "/zeno") ||
			request.URL.RequestURI() == "/" {
			log.Println("handling request")
			mux.ServeHTTP(writer, request)
		} else {
			// otherwise proxy request to search
			rp.ServeHTTP(writer, request)
		}
	})}

	// start http server
	go func() {
		log.Println("Starting HTTP server at", addr)

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
	collyScraper.C.Wait()
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
