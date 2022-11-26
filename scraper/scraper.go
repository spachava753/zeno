package scraper

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"zeno/domain"
	"zeno/indexer"

	"github.com/gocolly/colly"
	"golang.org/x/net/html"
)

const DocCtxKey = "doc"

type Scraper interface {
	Scrape(doc domain.ScrapedDoc) error
}

type CollyScraper struct {
	indexer indexer.Indexer
	C       *colly.Collector
}

func NewCollyScraper(indexer indexer.Indexer) CollyScraper {
	if indexer == nil {
		panic("indexer cannot be nil")
	}
	return CollyScraper{
		indexer: indexer,
		C:       MakeCollector(indexer),
	}
}

func (c CollyScraper) Scrape(doc domain.ScrapedDoc) error {
	_, err := url.Parse(doc.URL)
	if err != nil {
		return err
	}
	ctx := colly.NewContext()
	ctx.Put(DocCtxKey, doc)
	return c.C.Request(http.MethodGet, doc.URL, nil, ctx, nil)
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

// http://corpus.tools/wiki/Justext/Algorithm
func HandleHtmlDoc(response *colly.Response, parsedDoc *domain.ScrapedDoc) error {
	rootNode, err := html.Parse(bytes.NewReader(response.Body))
	if err != nil {
		return errors.New("could not parse html response")
	}
	if parsedDoc.Scrape {
		parsedDoc.Content = parseContent(rootNode)
	}
	if parsedDoc.Title == "" {
		parsedDoc.Title = parseTitle(rootNode)
	}
	parsedDoc.URL = response.Request.URL.String()
	parsedDoc.ID, err = IdFromUrl(parsedDoc.URL)
	if err != nil {
		return err
	}
	parsedDoc.DocType = domain.Html
	log.Printf("Parsed: %#v\n", parsedDoc)
	return nil
}

func IdFromUrl(url string) (string, error) {
	var sb strings.Builder
	encoder := base64.NewEncoder(base64.URLEncoding, &sb)
	if _, encodingErr := encoder.Write([]byte(url)); encodingErr != nil {
		return "", fmt.Errorf("could not set ID for parsed document: %w", encodingErr)
	}
	return sb.String(), nil
}

func HandlePdfDoc(response *colly.Response, s *domain.ScrapedDoc) error {
	fileName := fmt.Sprintf(
		"%s-%d.pdf",
		filepath.Base(response.Request.URL.String()),
		time.Now().Unix(),
	)
	fileName = filepath.Join(os.TempDir(), fileName)
	log.Println("writing to file", fileName)
	if err := os.WriteFile(fileName, response.Body, 0755); err != nil {
		return fmt.Errorf("could not write file %s: %w", fileName, err)
	}
	defer func() {
		if err := os.Remove(fileName); err != nil {
			log.Printf("could not remove file %s: %s", fileName, fileName)
		}
	}()
	if s.Scrape {
		var buffer bytes.Buffer
		cmd := exec.Command("pdftotext", fileName, "-")
		cmd.Stdout = &buffer
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not run pdftotext cmd: %w", err)
		}
		s.Content = buffer.String()
		log.Println("parsed content is", s.Content[:50])
	}
	s.DocType = domain.Pdf
	s.URL = response.Request.URL.String()
	if s.Title == "" {
		s.Title = strings.Split(strings.TrimSpace(s.Content), "\n")[0]
	}
	log.Println("parsed title is", s.Title)
	var err error
	s.ID, err = IdFromUrl(s.URL)
	if err != nil {
		return err
	}
	return nil
}

func DocTypeOf(response *colly.Response) domain.DocType {
	ext := filepath.Ext(strings.TrimPrefix(response.Request.URL.Path, "/"))
	if ext == ".pdf" {
		return domain.Pdf
	}
	if ext == "" &&
		strings.Contains(response.Headers.Get("Content-Type"), "text/html") {
		return domain.Html
	}
	return ""
}

func MakeCollector(indexer indexer.Indexer) *colly.Collector {
	// Instantiate default collector
	c := colly.NewCollector(
		// MaxDepth is 1, so only the links on the scraped page
		// is visited, and no further links are followed
		colly.MaxDepth(1),
		colly.Async(true),
		colly.AllowURLRevisit(),
	)

	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	// On every element which has href attribute call callback
	//c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	//	link := e.Attr("href")
	//	if link != "#" {
	//		// Visit link found on page
	//		e.Request.Visit(e.Request.AbsoluteURL(link))
	//	}
	//})

	c.OnResponse(func(response *colly.Response) {
		t := DocTypeOf(response)
		log.Println("parsed doc type:", t)
		s := response.Ctx.GetAny(DocCtxKey).(domain.ScrapedDoc)
		var err error
		switch t {
		case domain.Html:
			err = HandleHtmlDoc(response, &s)
		case domain.Pdf:
			err = HandlePdfDoc(response, &s)
		default:
			log.Println("unknown document type for url", response.Request.URL)
			return
		}
		if err != nil {
			log.Println("could scrape document:", err)
			return
		}
		s.ParsedDate = domain.Timestamp(time.Now())
		if indexErr := indexer.Index(s); indexErr != nil {
			log.Println("could not index:", indexErr)
		}
	})

	c.OnError(func(response *colly.Response, err error) {
		log.Printf("error on scraping url %s: %s\n", response.Request.URL, err)
	})

	return c
}
