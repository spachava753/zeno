use std::time::SystemTime;

use color_eyre::Result;
use typed_builder::TypedBuilder;
use url::Url;

struct DocTitle(String);

struct DocDescription(String);

enum DocType {
    Pdf,
    Html,
}

pub struct Timestamp(u128);

impl Timestamp {
    pub fn now() -> Result<Self> {
        let timestamp = SystemTime::now()
            .duration_since(SystemTime::UNIX_EPOCH)?
            .as_millis();
        Ok(Self(timestamp))
    }
}

impl TryFrom<Timestamp> for i64 {
    type Error = <u128 as TryFrom<i64>>::Error;

    fn try_from(value: Timestamp) -> std::result::Result<Self, Self::Error> {
        value.0.try_into()
    }
}

#[derive(TypedBuilder)]
struct Document {
    url: Url,
    title: DocTitle,
    description: Option<DocDescription>,
    doc_type: DocType,
    parsed_date: Timestamp,
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
            .parsed_date(Timestamp::now()?)
            .build();
        Ok(())
    }
}
