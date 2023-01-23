use tokio::sync::mpsc;
use tokio::sync::mpsc::Sender;
use tokio::sync::oneshot;
use tokio::task::JoinHandle;
use tracing::*;

pub fn start_actor() -> (JoinHandle<()>, Sender<SearchEngineMsg>) {
    let (tx, rx) = mpsc::channel(32);
    let handle = tokio::spawn(searcher_actor(rx));
    (handle, tx)
}

#[derive(Debug)]
pub enum SearchEngineMsg {
    Index { resp: oneshot::Sender<String> },
    Search { resp: oneshot::Sender<String> },
}

async fn searcher_actor(mut rx: mpsc::Receiver<SearchEngineMsg>) {
    while let Some(msg) = rx.recv().instrument(info_span!("searcher actor")).await {
        match msg {
            SearchEngineMsg::Index { resp } => {
                info!("Index message received");
                if resp.send("hi".to_string()).is_err() {
                    error!("receiver dropped")
                }
            }
            SearchEngineMsg::Search { resp } => {
                info!("Search message received");
                if resp.send("hi".to_string()).is_err() {
                    error!("receiver dropped")
                }
            }
        }
    }
    info!("search actor finished running");
}
