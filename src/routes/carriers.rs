use axum::{
    routing::get,
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/carriers", get(handlers::carrier::list).post(handlers::carrier::create))
        .route(
            "/carriers/{id}",
            get(handlers::carrier::get_by_id)
                .put(handlers::carrier::update)
                .delete(handlers::carrier::delete),
        )
}
