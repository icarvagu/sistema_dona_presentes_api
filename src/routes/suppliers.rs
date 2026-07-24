use axum::{
    routing::get,
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/suppliers", get(handlers::supplier::list).post(handlers::supplier::create))
        .route(
            "/suppliers/{id}",
            get(handlers::supplier::get_by_id)
                .put(handlers::supplier::update)
                .delete(handlers::supplier::delete),
        )
}
