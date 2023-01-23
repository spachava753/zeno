use crate::doc::Document;
use std::path::Path;
use tantivy::collector::TopDocs;
use tantivy::query::QueryParser;
use tantivy::schema::{Field, Schema, INDEXED, STORED, TEXT};
use tantivy::{DateTime, Document as TantivyDocument, Index, IndexReader, IndexWriter};

pub struct SearchEngine {
    writer: IndexWriter,
    reader: IndexReader,
    index: Index,
    url_field: Field,
    title_field: Field,
    body_field: Field,
    description_field: Field,
    parsed_time_field: Field,
}

impl SearchEngine {
    const URL_FIELD: &'static str = "url";
    const TITLE_FIELD: &'static str = "title";
    const BODY_FIELD: &'static str = "body";
    const DESCRIPTION_FIELD: &'static str = "description";
    const PARSED_TIME_FIELD: &'static str = "parsed";
    pub fn new<P: AsRef<Path>>(index_dir: P) -> color_eyre::Result<Self> {
        // create schema
        let mut schema_builder = Schema::builder();
        let url_field = schema_builder.add_text_field(Self::URL_FIELD, TEXT | STORED);
        let title_field = schema_builder.add_text_field(Self::TITLE_FIELD, TEXT | STORED);
        let body_field = schema_builder.add_text_field(Self::BODY_FIELD, TEXT);
        let description_field = schema_builder.add_text_field(Self::DESCRIPTION_FIELD, TEXT);
        let parsed_time_field = schema_builder.add_date_field(Self::PARSED_TIME_FIELD, INDEXED);
        let schema = schema_builder.build();

        // create index
        let index = Index::builder()
            .schema(schema.clone())
            .create_in_dir(index_dir)?;
        let index_writer = index.writer(100_000_000)?;
        let reader = index.reader()?;

        Ok(SearchEngine {
            index,
            writer: index_writer,
            reader,
            url_field,
            title_field,
            body_field,
            description_field,
            parsed_time_field,
        })
    }

    pub fn add_doc(&mut self, document: Document) -> color_eyre::Result<()> {
        let mut doc = TantivyDocument::new();
        doc.add_text(self.url_field, document.url().path());
        doc.add_text(self.title_field, document.title());
        if let Some(body) = document.body() {
            doc.add_text(self.body_field, body);
        }
        if let Some(description) = document.description() {
            doc.add_text(self.description_field, description);
        }
        let timestamp: i64 = document.parsed_date().try_into()?;
        doc.add_date(
            self.parsed_time_field,
            DateTime::from_timestamp_millis(timestamp),
        );
        self.writer.add_document(doc)?;
        self.writer.commit()?;
        Ok(())
    }

    pub fn search(
        &self,
        query_str: &str,
        limit: usize,
    ) -> color_eyre::Result<Vec<TantivyDocument>> {
        let query_parser = QueryParser::for_index(
            &self.index,
            vec![
                self.url_field,
                self.title_field,
                self.body_field,
                self.description_field,
            ],
        );

        let query = query_parser.parse_query(query_str)?;

        let searcher = self.reader.searcher();
        let top_docs = searcher.search(&query, &TopDocs::with_limit(limit))?;

        let mut results = Vec::with_capacity(top_docs.len());
        for (_score, doc_address) in top_docs {
            // Retrieve the actual content of documents given its `doc_address`.
            results.push(searcher.doc(doc_address)?);
        }

        Ok(results)
    }
}

#[cfg(test)]
mod tests {
    use url::Url;

    use crate::doc::{DocBody, DocDescription, DocTitle, DocType, Document, Timestamp};

    use crate::searcher::engine::SearchEngine;

    #[test]
    fn tantivy_test() -> color_eyre::Result<()> {
        let dir = tempdir::TempDir::new("tantivy-test")?;
        println!("creating tantivy index in {dir:?}");
        let mut searcher = SearchEngine::new(dir.path())?;
        let doc = Document::builder()
            .url(Url::parse("https://sirupsen.com/index-merges")?)
            .title(DocTitle::new("Neural Network From Scratch".to_string())?)
            .body(Some(DocBody::new(
                include_str!("../../testdata/neural-net.html").to_string(),
            )?))
            .description(Some(DocDescription::new(
                "Article about writing a neural net from scratch".to_string(),
            )?))
            .doc_type(DocType::Html)
            .parsed_date(Timestamp::now()?)
            .build();
        searcher.add_doc(doc)?;

        searcher.reader.reload()?; // force a reload
        let query = searcher.search("Neural Network From Scratch", 10)?;
        println!("{query:?}");
        dir.close()?;
        Ok(())
    }
}
