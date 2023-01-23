use crate::doc::{DocId, Document};
use crate::searcher::engine::SearchEngine;
use std::path::Path;
use tokio::sync::mpsc;
use tokio::sync::mpsc::Sender;
use tokio::sync::oneshot;
use tokio::task::JoinHandle;
use tracing::*;

pub fn start_actor<P: AsRef<Path>>(
    index_dir: P,
) -> color_eyre::Result<(JoinHandle<()>, Sender<SearchEngineMsg>)> {
    let (tx, rx) = mpsc::channel(32);
    let search_engine = SearchEngine::new(index_dir)?;
    let handle = tokio::spawn(searcher_actor(rx, search_engine));
    Ok((handle, tx))
}

#[derive(Debug)]
pub enum SearchEngineMsg {
    Index {
        doc: Document,
        resp: oneshot::Sender<color_eyre::Result<()>>,
    },
    Search {
        limit: usize,
        query: String,
        resp: oneshot::Sender<color_eyre::Result<Vec<DocId>>>,
    },
}

async fn searcher_actor(mut rx: mpsc::Receiver<SearchEngineMsg>, mut search_engine: SearchEngine) {
    while let Some(msg) = rx.recv().instrument(info_span!("searcher actor")).await {
        match msg {
            SearchEngineMsg::Index { doc, resp } => {
                info!("Index message received");
                if resp.send(search_engine.add_doc(doc)).is_err() {
                    error!("receiver dropped")
                }
            }
            SearchEngineMsg::Search { query, limit, resp } => {
                info!("Search message received");
                if resp
                    .send(search_engine.search(query.as_str(), limit))
                    .is_err()
                {
                    error!("receiver dropped")
                }
            }
        }
    }
    info!("search actor finished running");
}
