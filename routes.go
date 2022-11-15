package main

import (
	"net/http"
	"net/url"
	"strconv"
)

func MakeRoutes(s Scraper, mux *http.ServeMux) {
	mux.HandleFunc("/scrape", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		query := request.URL.Query()
		urlStr := query.Get("url")
		titleStr := query.Get("title")
		scrapeStr := query.Get("title")
		scrape, _ := strconv.ParseBool(scrapeStr)
		descriptionStr := query.Get("description")
		parsedUrl, parseErr := url.Parse(urlStr)
		if parseErr != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(parseErr.Error()))
			return
		}

		if visitErr := s.Scrape(ScrapedDoc{
			URL:         parsedUrl.String(),
			Title:       titleStr,
			Description: descriptionStr,
			Scrape:      scrape,
		}); visitErr != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(visitErr.Error()))
			return
		}

		writer.WriteHeader(http.StatusAccepted)
	})

	mux.Handle("/", http.FileServer(http.Dir("./static")))
}
