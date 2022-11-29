package db

import (
	"context"
	"github.com/stretchr/testify/suite"
	"path/filepath"
	"testing"
	"time"
	"zeno/domain"
	"zeno/scraper"
)

type SqliteTestSuite struct {
	suite.Suite
	repo scraper.UrlRepo
}

func (s *SqliteTestSuite) TestRepoActions() {
	dsn := filepath.Join(s.T().TempDir(), "test.db")
	s.T().Log("dsn:", dsn)
	s.repo = NewGormRepo(dsn)
	ctx := context.Background()
	testDoc := domain.ScrapedDoc{
		Title:       "Test Site",
		Description: "Test Site Description",
		Content:     "Test Site Content",
		URL:         "test.example",
		Scrape:      false,
		ParsedDate:  domain.Timestamp(time.Now()),
		DocType:     domain.Html,
	}

	// test getting nonexistent document
	result, getErr := s.repo.Get(ctx, testDoc)
	s.Assert().Equal(result, domain.ScrapedDoc{}, "expected empty result")
	s.T().Log("getErr:", getErr)
	s.Assert().Error(getErr, "expected to fail getting document")

	// test saving document with no id
	saveErr := s.repo.Save(ctx, testDoc)
	s.T().Log("saveErr:", saveErr)
	s.Assert().Error(saveErr, "expected to fail saving")

	var idErr error
	testDoc.ID, idErr = scraper.IdFromUrl(testDoc.URL)
	s.Require().NoError(idErr, "cannot fail id creation")
	s.T().Log("test id:", testDoc.ID)

	// test saving document
	s.Require().NoError(s.repo.Save(ctx, testDoc), "cannot fail saving")

	// test getting document
	result, getErr = s.repo.Get(ctx, testDoc)
	expected := testDoc
	expected.Content = ""
	s.Assert().NoError(getErr, "no error getting document")
	s.Assert().Equal(expected.String(), result.String(), "expected equal result")

	// test updating document
	testDoc.Title = "Updated Title"
	s.Require().NoError(s.repo.Save(ctx, testDoc), "cannot fail saving")

	// test getting document
	result, getErr = s.repo.Get(ctx, testDoc)
	expected = testDoc
	expected.Content = ""
	s.Assert().NoError(getErr, "no error getting document")
	s.Assert().Equal(expected.String(), result.String(), "expected equal result")

	// test deleting document with no id
	temp := testDoc.ID
	testDoc.ID = ""
	deleteErr := s.repo.Delete(ctx, testDoc)
	s.T().Log("deleteErr:", deleteErr)
	s.Assert().Error(deleteErr, "expected to fail deleting")

	// test deleting document with just id
	s.Require().NoError(s.repo.Delete(ctx, domain.ScrapedDoc{ID: temp}), "cannot fail deleting")
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(SqliteTestSuite))
}
