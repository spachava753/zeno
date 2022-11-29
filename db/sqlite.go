package db

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
	"zeno/domain"
)

type Document struct {
	ID          string `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string
	Description string
	URL         string
	Scrape      bool
	ParsedDate  time.Time
	DocType     string
}

func scrapedDocToDocument(doc *domain.ScrapedDoc) Document {
	return Document{
		ID:          doc.ID,
		Title:       doc.Title,
		Description: doc.Description,
		URL:         doc.URL,
		Scrape:      doc.Scrape,
		ParsedDate:  time.Time(doc.ParsedDate),
		DocType:     string(doc.DocType),
	}
}

func documentToScrapedDoc(doc *Document) domain.ScrapedDoc {
	return domain.ScrapedDoc{
		Title:       doc.Title,
		Description: doc.Description,
		URL:         doc.URL,
		ID:          doc.ID,
		Scrape:      doc.Scrape,
		ParsedDate:  domain.Timestamp(doc.ParsedDate),
		DocType:     domain.DocType(doc.DocType),
	}
}

var EmptyId = errors.New("empty id")

type GormRepo struct {
	db *gorm.DB
}

func (s GormRepo) Save(ctx context.Context, scrapedDoc domain.ScrapedDoc) error {
	rdoc := scrapedDocToDocument(&scrapedDoc)
	if rdoc.ID == "" {
		return EmptyId
	}
	if err := s.db.Save(&rdoc).Error; err != nil {
		return fmt.Errorf("cannot save document: %w", err)
	}
	return nil
}

func (s GormRepo) Get(ctx context.Context, scrapedDoc domain.ScrapedDoc) (domain.ScrapedDoc, error) {
	var rdoc Document
	if scrapedDoc.ID == "" {
		return domain.ScrapedDoc{}, EmptyId
	}
	if err := s.db.First(&rdoc, "id = ?", scrapedDoc.ID).Error; err != nil {
		return domain.ScrapedDoc{}, fmt.Errorf("cannot fetch document: %w", err)
	}
	sd := documentToScrapedDoc(&rdoc)
	return sd, nil
}

func (s GormRepo) GetAll(ctx context.Context) ([]domain.ScrapedDoc, error) {
	var rdocs []Document
	if err := s.db.Find(&rdocs).Error; err != nil {
		return nil, fmt.Errorf("cannot fetch documents: %w", err)
	}
	scrapedDocs := make([]domain.ScrapedDoc, len(rdocs))
	for i := range scrapedDocs {
		scrapedDocs[i] = documentToScrapedDoc(&rdocs[i])
	}
	return scrapedDocs, nil
}

func (s GormRepo) Delete(ctx context.Context, scrapedDoc domain.ScrapedDoc) error {
	rdoc := scrapedDocToDocument(&scrapedDoc)
	if rdoc.ID == "" {
		return EmptyId
	}
	if err := s.db.Delete(&rdoc).Error; err != nil {
		return fmt.Errorf("cannot delete document: %w", err)
	}
	return nil
}

func NewGormRepo(dsn string) GormRepo {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to db")
	}
	if migrateErr := db.AutoMigrate(&Document{}); migrateErr != nil {
		panic("failed to run migrations")
	}
	return GormRepo{
		db: db,
	}
}
