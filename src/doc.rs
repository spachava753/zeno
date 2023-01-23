use std::fmt::{Display, Formatter};
use std::time::SystemTime;

use color_eyre::eyre::eyre;
use color_eyre::Result;
use typed_builder::TypedBuilder;
use url::Url;

#[derive(Clone)]
pub struct DocTitle {
    title: String,
}

impl Display for DocTitle {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.title)
    }
}

impl DocTitle {
    pub fn new(title: String) -> Result<Self> {
        if title.is_empty() {
            return Err(eyre!("Title cannot be empty"));
        }
        Ok(Self { title })
    }
    pub fn title(&self) -> &str {
        &self.title
    }
}

#[derive(Clone)]
pub struct DocBody {
    body: String,
}

impl Display for DocBody {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.body)
    }
}

impl DocBody {
    pub fn new(body: String) -> Result<Self> {
        if body.is_empty() {
            return Err(eyre!("Body cannot be empty"));
        }
        Ok(Self { body })
    }

    pub fn body(&self) -> &str {
        &self.body
    }
}

#[derive(Clone)]
pub struct DocDescription {
    description: String,
}

impl Display for DocDescription {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.description)
    }
}

impl DocDescription {
    pub fn new(description: String) -> Result<Self> {
        if description.is_empty() {
            return Err(eyre!("Description cannot be empty"));
        }
        Ok(Self { description })
    }
    pub fn description(&self) -> &str {
        &self.description
    }
}

#[derive(Clone)]
pub enum DocType {
    Pdf,
    Html,
}

impl Display for DocType {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "{}",
            match self {
                DocType::Pdf => "pdf",
                DocType::Html => "html",
            }
        )
    }
}

#[derive(Clone)]
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
pub struct Document {
    url: Url,
    title: DocTitle,
    body: Option<DocBody>,
    description: Option<DocDescription>,
    doc_type: DocType,
    parsed_date: Timestamp,
}

impl Document {
    pub fn url(&self) -> &Url {
        &self.url
    }
    pub fn title(&self) -> &DocTitle {
        &self.title
    }
    pub fn body(&self) -> Option<&DocBody> {
        self.body.as_ref()
    }
    pub fn description(&self) -> Option<&DocDescription> {
        self.description.as_ref()
    }
    pub fn doc_type(&self) -> DocType {
        self.doc_type.clone()
    }
    pub fn parsed_date(&self) -> Timestamp {
        self.parsed_date.clone()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn doc_builder_test() -> Result<()> {
        color_eyre::install()?;
        let _ = Document::builder()
            .url(Url::parse("https://sirupsen.com/index-merges")?)
            .title(DocTitle::new("test".to_string())?)
            .body(None)
            .description(None)
            .doc_type(DocType::Html)
            .parsed_date(Timestamp::now()?)
            .build();
        Ok(())
    }
}
