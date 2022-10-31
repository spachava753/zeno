package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func MakeRoutes(s Scraper, mux *http.ServeMux) {
	searchUrl, _ := url.Parse(SearchUrl)
	rp := httputil.NewSingleHostReverseProxy(searchUrl)
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

		if visitErr := s.Scrape(parsedUrl.String()); visitErr != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(visitErr.Error()))
			return
		}

		http.Redirect(writer, request, "/", http.StatusFound)
	})

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "./static/index.html")
	})

	// this function will proxy search requests so that
	mux.HandleFunc(fmt.Sprintf("/indexes/%s/search", IndexName), func(writer http.ResponseWriter, request *http.Request) {
		rp.ServeHTTP(writer, request)
	})
}
