package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const IdleTimeout = 30 * time.Second

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	ticker := time.NewTicker(IdleTimeout)
	defer ticker.Stop()

	mux := http.NewServeMux()
	index := MakeMeilisearchIndex("http://127.0.0.1:7700", "")
	m := NewMeilisearchIndexer(index)
	s := NewCollyScraper(m)

	MakeRoutes(s, mux)

	srv := http.Server{Addr: ":8080", Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// log request
		log.Printf("url: %s, method: %s", request.URL, request.Method)

		ticker.Reset(IdleTimeout)

		mux.ServeHTTP(writer, request)
	})}

	// shutdown function
	go func() {
		select {
		case s := <-sigChan:
			log.Printf("signal recieved: %s", s)
		case <-ticker.C:
			log.Println("idle timeout")
		}
		// shutdown the server
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v\n", err)
		}
		s.C.Wait()
	}()

	log.Println("Starting HTTP server")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server Shutdown: %v\n", err)
		os.Exit(1)
	}
}
