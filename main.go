package main

import (
	"bytes"
	"golang.org/x/net/html"
	"log"
	"strings"

	"github.com/gocolly/colly"
)

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

func ParseDocument(root *html.Node) string {
	if root.Type == html.ElementNode && ignoreElement(root.Data) {
		return ""
	}
	if root.Type == html.TextNode {
		if strings.TrimSpace(root.Data) == "" {
			return ""
		}
		return strings.TrimSpace(root.Data) + " "
	}
	var sb strings.Builder
	for n := root.FirstChild; n != nil; n = n.NextSibling {
		txt := ParseDocument(n)
		sb.WriteString(txt)
	}
	return sb.String()
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// MaxDepth is 1, so only the links on the scraped page
		// is visited, and no further links are followed
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Visit link found on page
		e.Request.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("Visiting", request.URL)
	})

	c.OnResponse(func(response *colly.Response) {
		rootNode, err := html.Parse(bytes.NewReader(response.Body))
		if err != nil {
			log.Println("could not parse html response")
			return
		}
		parsedText := ParseDocument(rootNode)
		log.Println("Parsed:", response.Request.URL, parsedText)
	})

	c.Visit("https://thespblog.net/a-gophers-foray-into-rust/")

	c.Wait()
}
