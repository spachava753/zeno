package main

import (
	"net/http"
	"net/url"
)

func MakeRoutes(s Scraper, mux *http.ServeMux) {
	mux.HandleFunc("/scrape", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		query := request.URL.Query()
		urlStr := query.Get("url")
		parsedUrl, parseErr := url.Parse(urlStr)
		if parseErr != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(parseErr.Error()))
			return
		}

		if visitErr := s.Scrape(ScrapedDoc{URL: parsedUrl.String()}); visitErr != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(visitErr.Error()))
			return
		}

		writer.WriteHeader(http.StatusAccepted)
	})

	mux.Handle("/", http.FileServer(http.Dir("./static")))
}
