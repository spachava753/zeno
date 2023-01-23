extern crate core;

use std::net::SocketAddr;

use crate::searcher::actor::SearchEngineMsg;
use axum::extract::State;
use axum::{
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use tokio::sync::mpsc::Sender;
use tokio::sync::oneshot;
use tokio::sync::oneshot::error::{RecvError, TryRecvError};
use tower::ServiceBuilder;
use tower_http::trace::{DefaultOnRequest, DefaultOnResponse, TraceLayer};
use tracing::*;

mod doc;
mod pdf;
mod searcher;

#[derive(Debug, Clone)]
struct AppState {
    tx: Sender<SearchEngineMsg>,
}

#[tokio::main]
async fn main() {
    // initialize tracing
    tracing_subscriber::fmt()
        .with_file(true)
        .with_line_number(true)
        .with_level(true)
        .with_max_level(Level::INFO)
        .init();

    let (handle, tx) = searcher::actor::start_actor();

    // build our application with a route
    let app = Router::new()
        .route("/", get(root))
        .route("/scrape", post(index_doc))
        .route("/search", post(search_docs))
        .with_state(AppState { tx })
        .layer(
            ServiceBuilder::new().layer(
                TraceLayer::new_for_http()
                    .on_request(DefaultOnRequest::default().level(Level::INFO))
                    .on_response(DefaultOnResponse::default().level(Level::INFO)),
            ),
        );

    // run our app with hyper
    // `axum::Server` is a re-export of `hyper::Server`
    let addr = SocketAddr::from(([127, 0, 0, 1], 3000));
    info!("listening on {}", addr);
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();

    handle.await.expect("error awaiting handle");
}

// basic handler that responds with a static string
#[instrument]
async fn root() -> &'static str {
    "Hello, World!"
}

#[instrument]
async fn index_doc(
    // this argument tells axum to parse the request body
    // as JSON into a `CreateUser` type
    State(state): State<AppState>,
    Json(payload): Json<IndexDoc>,
) -> impl IntoResponse {
    info!("indexing new document");
    let (tx, rx) = oneshot::channel();
    if let Err(_) = state.tx.send(SearchEngineMsg::Index { resp: tx }).await {
        error!("receiver dropped");
    }

    match rx.await {
        Ok(s) => {
            info!("message received: {}", s);
        }
        Err(e) => {
            error!("sender dropped: {}", e);
        }
    }
    // this will be converted into a JSON response
    // with a status code of `201 Created`
    (StatusCode::OK)
}

#[instrument]
async fn search_docs(
    State(state): State<AppState>,
    Json(payload): Json<SearchQuery>,
) -> impl IntoResponse {
    info!("handling search request");
    let (tx, rx) = oneshot::channel();
    if let Err(_) = state.tx.send(SearchEngineMsg::Search { resp: tx }).await {
        error!("receiver dropped");
    }

    match rx.await {
        Ok(s) => {
            info!("message received: {}", s);
        }
        Err(e) => {
            error!("sender dropped: {}", e);
        }
    }
    (StatusCode::OK)
}

// the input to our `create_user` handler
#[derive(Deserialize, Debug)]
struct IndexDoc {
    url: String,
    title: Option<String>,
    description: Option<String>,
}

#[derive(Deserialize, Debug)]
struct SearchQuery {
    query: String,
}
