use axum::{
    routing::{delete, get, post, put},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route(
            "/products",
            get(handlers::product::list).post(handlers::product::create),
        )
        .route(
            "/products/{id}",
            get(handlers::product::get_by_id)
                .put(handlers::product::update)
                .delete(handlers::product::delete),
        )
        .route("/products/groups", get(handlers::product::groups))
        .route("/products/pending-approval", get(handlers::product::pending_approval))
        .route("/products/{id}/approve", post(handlers::product::approve))
        .route("/products/financial-report", get(handlers::product::financial_report))
        .route("/products/{id}/last-cost", put(handlers::product::update_last_cost))
        .route("/products/search", get(handlers::product::search))
        .route("/products/{id}/items", get(handlers::product::get_items).post(handlers::product::create_item))
        .route("/products/{id}/items/{itemId}", delete(handlers::product::delete_item))
        .route("/products/by-group/{group}", get(handlers::product::by_group))
        .route("/products/bulk-approve", post(handlers::product::bulk_approve))
}
