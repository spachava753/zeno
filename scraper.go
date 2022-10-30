package main

import (
	"bytes"
	"github.com/gocolly/colly"
	"golang.org/x/net/html"
	"log"
	"strings"
)

type ScrapedDoc struct {
	Title       string
	Description string
	Content     string
	URL         string
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

func MakeCollector() *colly.Collector {
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
		log.Printf("Parsed: %#v\n", parsedDoc)
	})

	c.OnError(func(response *colly.Response, err error) {
		log.Printf("error on scraping url %s: %s\n", response.Request.URL, err)
	})

	return c
}
