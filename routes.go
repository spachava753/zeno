package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"zeno/domain"
	"zeno/scraper"
)

func MakeRoutes(s scraper.Scraper, mux *http.ServeMux) {
	mux.HandleFunc("/scrape", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		query := request.URL.Query()
		urlStr := query.Get("url")
		titleStr := query.Get("title")
		scrapeStr := query.Get("scrape")
		scrape, _ := strconv.ParseBool(scrapeStr)
		descriptionStr := query.Get("description")
		parsedUrl, parseErr := url.Parse(urlStr)
		if parseErr != nil {
			writer.WriteHeader(http.StatusBadRequest)
			if _, err := writer.Write([]byte(parseErr.Error())); err != nil {
				log.Println("found error writing response bytes:", err)
			}
			return
		}
		log.Printf("url: %s, title: %s, description: %s, scrape: %v\n", urlStr, titleStr, descriptionStr, scrape)

		doc := domain.ScrapedDoc{
			URL:         parsedUrl.String(),
			Title:       titleStr,
			Description: descriptionStr,
			Scrape:      scrape,
		}

		if visitErr := s.Scrape(doc); visitErr != nil {
			writer.WriteHeader(http.StatusBadRequest)
			if _, err := writer.Write([]byte(visitErr.Error())); err != nil {
				log.Println("found error writing response bytes:", err)
			}
			return
		}

		writer.WriteHeader(http.StatusAccepted)
	})

	mux.Handle("/", http.FileServer(http.Dir("./static")))
}
