extern crate core;

use std::net::SocketAddr;

use axum::extract::State;
use axum::{
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use serde::Deserialize;
use tokio::signal;
use tokio::sync::mpsc::Sender;
use tokio::sync::oneshot;
use tower::ServiceBuilder;
use tower_http::trace::{DefaultOnRequest, DefaultOnResponse, TraceLayer};
use tracing::*;

use crate::searcher::actor::SearchEngineMsg;

mod doc;
mod pdf;
mod searcher;

#[derive(Debug, Clone)]
struct AppState {
    tx: Sender<SearchEngineMsg>,
}

pub enum ProcessState {
    Running,
    Shutdown,
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

    // TODO: for testing
    let (search_engine_handle, tx) =
        searcher::actor::start_actor("testdata").expect("could not start actor");

    // build our application with a route
    let app = Router::new()
        .route("/", get(root))
        .route("/scrape", post(index_doc))
        .route("/search", post(search_docs))
        .with_state(AppState { tx: tx.clone() })
        .layer(
            ServiceBuilder::new().layer(
                TraceLayer::new_for_http()
                    .on_request(DefaultOnRequest::default().level(Level::INFO))
                    .on_response(DefaultOnResponse::default().level(Level::INFO)),
            ),
        );

    // run our app with hyper
    // `axum::Server` is a re-export of `hyper::Server`
    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    info!("listening on {}", addr);
    let axum_handle = tokio::spawn(
        axum::Server::bind(&addr)
            .serve(app.into_make_service())
            .with_graceful_shutdown(shutdown_signal()),
    );

    // wait for server to shutdown
    if let Err(e) = axum_handle.await {
        error!("error from server task: {}", e);
    }

    // drop transmitter, should exit the actor
    drop(tx);

    if let Err(e) = search_engine_handle.await {
        error!("error from search engine task: {}", e);
    }

    info!("app shut down")
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }

    info!("signal received, starting graceful shutdown");
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
    // let (tx, rx) = oneshot::channel();
    // if let Err(_) = state.tx.send(SearchEngineMsg::Index { resp: tx }).await {
    //     error!("receiver dropped");
    // }
    //
    // match rx.await {
    //     Ok(s) => {
    //         info!("message received: {:?}", s);
    //     }
    //     Err(e) => {
    //         error!("sender dropped: {}", e);
    //     }
    // }
    // this will be converted into a JSON response
    // with a status code of `201 Created`
    StatusCode::OK
}

#[instrument]
async fn search_docs(
    State(state): State<AppState>,
    Json(payload): Json<SearchQuery>,
) -> impl IntoResponse {
    info!("handling search request");
    let (tx, rx) = oneshot::channel();
    if let Err(_) = state
        .tx
        .send(SearchEngineMsg::Search {
            query: payload.query,
            limit: 10,
            resp: tx,
        })
        .await
    {
        error!("receiver dropped");
    }

    match rx.await {
        Ok(s) => {
            info!("message received: {:?}", s);
        }
        Err(e) => {
            error!("sender dropped: {}", e);
        }
    }
    StatusCode::OK
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
