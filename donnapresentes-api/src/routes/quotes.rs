use axum::{
    routing::{get, patch},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route(
            "/quotes",
            get(handlers::quote::list).post(handlers::quote::create),
        )
        .route(
            "/quotes/{id}",
            get(handlers::quote::get_by_id)
                .put(handlers::quote::update)
                .delete(handlers::quote::delete),
        )
        .route("/quotes/{id}/feedback", patch(handlers::quote::update_feedback))
        .route("/quotes/{id}/pdf", get(handlers::xbz::generate_quote_pdf))
}
