package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
)

//go:embed static/index.html
var f embed.FS

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

	httpFs, err := fs.Sub(f, "static")
	if err != nil {
		panic("cannot serve index")
	}

	mux.Handle("/", http.FileServer(http.FS(httpFs)))

	// this function will proxy search requests so that
	mux.HandleFunc(fmt.Sprintf("/indexes/%s/search", IndexName), func(writer http.ResponseWriter, request *http.Request) {
		rp.ServeHTTP(writer, request)
	})
}
