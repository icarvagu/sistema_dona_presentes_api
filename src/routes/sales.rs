use axum::{
    routing::{get, put},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route(
            "/sales",
            get(handlers::sale::list).post(handlers::sale::create),
        )
        .route(
            "/sales/{id}",
            get(handlers::sale::get_by_id)
                .put(handlers::sale::update)
                .delete(handlers::sale::delete),
        )
        .route("/sales/{id}/layout", put(handlers::sale::update_layout))
        .route("/sales/{id}/pdf", get(handlers::xbz::generate_sale_pdf))
}
