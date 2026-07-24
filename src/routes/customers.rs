use axum::{
    routing::get,
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/customers", get(handlers::customer::list).post(handlers::customer::create))
        .route(
            "/customers/{id}",
            get(handlers::customer::get_by_id)
                .put(handlers::customer::update)
                .delete(handlers::customer::delete),
        )
}
