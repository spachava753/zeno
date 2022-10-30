package main

import (
	"github.com/gocolly/colly"
	"net/http"
	"net/url"
)

func MakeRoutes(c *colly.Collector, mux *http.ServeMux) {
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

		if visitErr := c.Visit(parsedUrl.String()); visitErr != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(visitErr.Error()))
			return
		}

		writer.WriteHeader(http.StatusAccepted)
		http.ServeFile(writer, request, "./static/scraped.html")
	})

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "./static/index.html")
	})
}
