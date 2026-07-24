use axum::{
    routing::{delete, post},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/users", post(handlers::user::create))
        .route("/users/{id}", delete(handlers::user::delete))
        .route("/products-xbz/sync", post(handlers::xbz::sync_products_from_xbz))
}
