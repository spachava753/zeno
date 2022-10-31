package main

import (
	"bytes"
	"encoding/base64"
	"github.com/gocolly/colly"
	"golang.org/x/net/html"
	"log"
	"net/url"
	"strings"
)

type ScrapedDoc struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	URL         string `json:"url"`
	ID          string `json:"id"`
}

type Scraper interface {
	Scrape(urlStr string) error
}

type CollyScraper struct {
	indexer Indexer
	C       *colly.Collector
}

func NewCollyScraper(indexer Indexer) CollyScraper {
	if indexer == nil {
		panic("indexer cannot be nil")
	}
	return CollyScraper{
		indexer: indexer,
		C:       MakeCollector(indexer),
	}
}

func (c CollyScraper) Scrape(urlStr string) error {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	return c.C.Visit(parsedUrl.String())
}

const roleKey = "role"

func ignoreElement(tagName string) bool {
	switch tagName {
	case "head", "sup":
		return true
		// Ignore elements that often don't contain relevant info
	case "header", "footer", "nav":
		return true
		// form elements
	case "label", "textarea":
		return true
		// Ignore javascript/style nodes
	case "script", "noscript", "style":
		return true
	}
	return false
}

func parseContent(root *html.Node) string {
	if root.Type == html.ElementNode && ignoreElement(root.Data) {
		return ""
	}
	for _, attr := range root.Attr {
		if attr.Key == roleKey &&
			(attr.Val == "navigation" ||
				attr.Val == "contentinfo" ||
				attr.Val == "button") {
			return ""
		}
	}

	if root.Type == html.TextNode {
		if strings.TrimSpace(root.Data) == "" {
			return ""
		}
		return strings.TrimSpace(root.Data) + " "
	}
	var sb strings.Builder
	for n := root.FirstChild; n != nil; n = n.NextSibling {
		txt := parseContent(n)
		sb.WriteString(txt)
	}
	return sb.String()
}

func parseTitle(root *html.Node) string {
	if root.Type == html.ElementNode && root.Data == "title" {
		var s string
		if root.FirstChild != nil {
			s = root.FirstChild.Data
		}
		return s
	}
	for n := root.FirstChild; n != nil; n = n.NextSibling {
		title := parseTitle(n)
		if title != "" {
			return title
		}
	}
	return ""
}

func parseDescription(root *html.Node) string {
	return ""
}

func ParseDocument(root *html.Node) ScrapedDoc {
	var s ScrapedDoc
	s.Content = parseContent(root)
	s.Title = parseTitle(root)
	s.Description = parseDescription(root)
	return s
}

func MakeCollector(indexer Indexer) *colly.Collector {
	// Instantiate default collector
	c := colly.NewCollector(
		// MaxDepth is 1, so only the links on the scraped page
		// is visited, and no further links are followed
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// On every a element which has href attribute call callback
	//c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	//	link := e.Attr("href")
	//	if link != "#" {
	//		// Visit link found on page
	//		e.Request.Visit(e.Request.AbsoluteURL(link))
	//	}
	//})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL)
	})

	c.OnResponse(func(response *colly.Response) {
		rootNode, err := html.Parse(bytes.NewReader(response.Body))
		if err != nil {
			log.Println("could not parse html response")
			return
		}
		parsedDoc := ParseDocument(rootNode)
		parsedDoc.URL = response.Request.URL.String()
		var sb strings.Builder
		encoder := base64.NewEncoder(base64.StdEncoding, &sb)
		if _, encodingErr := encoder.Write([]byte(parsedDoc.URL)); encodingErr != nil {
			log.Println("could not set ID for parsed document:", encodingErr)
			return
		}
		parsedDoc.ID = sb.String()
		log.Printf("Parsed: %#v\n", parsedDoc)
		if indexErr := indexer.Index(parsedDoc); indexErr != nil {
			log.Println("could not index document:", indexErr)
			return
		}
	})

	c.OnError(func(response *colly.Response, err error) {
		log.Printf("error on scraping url %s: %s\n", response.Request.URL, err)
	})

	return c
}
