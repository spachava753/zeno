package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type Timestamp time.Time

func (t Timestamp) String() string {
	tt := time.Time(t)
	return fmt.Sprintf("%d", tt.Unix())
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	return json.Marshal(tt.Unix())
}

type DocType string

const (
	Html = "html"
	Pdf  = "pdf"
)

type ScrapedDoc struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	URL         string    `json:"url"`
	ID          string    `json:"id"`
	Scrape      bool      `json:"scraped"`
	ParsedDate  Timestamp `json:"parsed_date"`
	DocType     DocType   `json:"doc_type"`
}

func (s ScrapedDoc) String() string {
	return fmt.Sprintf(`domain.ScrapedDoc{Title:"%s", Description:"%s", Content:"%s", URL:"%s", ID:"%s", Scrape:%v, ParsedDate:%s, DocType:"%s"}`, s.Title, s.Description[:50], s.Content[:50], s.URL, s.ID, s.Scrape, s.ParsedDate, s.DocType)
}
