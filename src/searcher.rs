use std::path::Path;

use color_eyre::Result;
use tantivy::collector::TopDocs;
use tantivy::query::QueryParser;
use tantivy::schema::{Field, Schema, INDEXED, STORED, TEXT};
use tantivy::{Document, Index, IndexReader, IndexWriter};

struct Searcher {
    writer: IndexWriter,
    reader: IndexReader,
    index: Index,
    url_field: Field,
    title_field: Field,
    body_field: Field,
    description_field: Field,
    parsed_time_field: Field,
}

impl Searcher {
    const URL_FIELD: &'static str = "url";
    const TITLE_FIELD: &'static str = "title";
    const BODY_FIELD: &'static str = "body";
    const DESCRIPTION_FIELD: &'static str = "description";
    const PARSED_TIME_FIELD: &'static str = "parsed";
    fn new<P: AsRef<Path>>(index_dir: P) -> Result<Self> {
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

        Ok(Searcher {
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

    fn add_doc(&mut self, doc: Document) -> Result<()> {
        self.writer.add_document(doc)?;
        self.writer.commit()?;
        Ok(())
    }

    fn search(&self, query_str: &str, limit: usize) -> Result<Vec<Document>> {
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
    use tantivy::DateTime;
    use time::OffsetDateTime;

    use super::*;

    #[test]
    fn tantivy_test() -> Result<()> {
        let dir = tempdir::TempDir::new("tantivy-test")?;
        println!("creating tantivy index in {dir:?}");
        let mut searcher = Searcher::new(dir.path())?;
        let mut doc = Document::new();
        doc.add_text(searcher.url_field, "https://sirupsen.com/index-merges");
        doc.add_text(searcher.title_field, "Neural Network From Scratch");
        doc.add_text(
            searcher.body_field,
            include_str!("../testdata/neural-net.html"),
        );
        doc.add_text(
            searcher.description_field,
            "Article about writing a neural net from scratch",
        );
        doc.add_date(
            searcher.parsed_time_field,
            DateTime::from_timestamp_millis(OffsetDateTime::now_utc().unix_timestamp()),
        );
        searcher.add_doc(doc)?;

        searcher.reader.reload()?; // force a reload
        let query = searcher.search("Neural Network From Scratch", 10)?;
        println!("{query:?}");
        dir.close()?;
        Ok(())
    }
}
