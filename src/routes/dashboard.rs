use axum::{routing::get, Router};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/dashboard", get(handlers::dashboard::dashboard))
}
