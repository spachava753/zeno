use time::Instant;
use typed_builder::TypedBuilder;
use url::Url;

struct DocTitle(String);

struct DocDescription(String);

enum DocType {
    Pdf,
    Html,
}

#[derive(TypedBuilder)]
struct Document {
    url: Url,
    title: DocTitle,
    description: Option<DocDescription>,
    doc_type: DocType,
    parsed_date: Instant,
}

impl Document {}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn doc_builder_test() -> color_eyre::Result<()> {
        color_eyre::install()?;
        let d = Document::builder()
            .url(Url::parse("https://sirupsen.com/index-merges")?)
            .title(DocTitle("test".to_string()))
            .description(None)
            .doc_type(DocType::Html)
            .parsed_date(Instant::now())
            .build();
        Ok(())
    }
}
