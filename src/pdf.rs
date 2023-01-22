#[cfg(test)]
mod tests {
    use color_eyre::Result;
    use pdf_extract::extract_text;

    #[test]
    fn testdata_read_pdf() -> Result<()> {
        let text = extract_text("testdata/neural-net-optimizer.pdf")?;
        println!("{text}");
        Ok(())
    }
}
