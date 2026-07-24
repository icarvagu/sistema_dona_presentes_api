use axum::{
    routing::get,
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/users", get(handlers::user::list))
        .route(
            "/users/{id}",
            get(handlers::user::get_by_id).put(handlers::user::update),
        )
}
