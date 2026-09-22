use axum::{
    routing::{get, post, put},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/purchases", get(handlers::purchase::list))
        .route("/purchase-requests", get(handlers::purchase::list_requests).post(handlers::purchase::create_request))
        .route("/purchase-requests/{id}", put(handlers::purchase::update_request))
        .route("/purchases/financial", get(handlers::purchase::get_financial))
        .route("/purchases/{id}", get(handlers::purchase::get_by_id).put(handlers::purchase::update))
        .route("/purchases/release/{saleId}", post(handlers::purchase::release_sale))
        .route("/purchases/batches", post(handlers::purchase::create_batch))
        .route("/purchases/{id}/actions", post(handlers::purchase::execute_action))
        .route("/purchases/{id}/attachments", post(handlers::purchase::add_attachment))
        .route("/purchases/{id}/payments", post(handlers::purchase::add_payment))
        .route("/purchases/payments/{paymentId}/approve", put(handlers::purchase::approve_payment))
        .route("/purchases/{id}/issues", post(handlers::purchase::add_issue))
        .route("/purchases/issues/{issueId}", put(handlers::purchase::update_issue))
        .route("/notifications", get(handlers::purchase::list_notifications))
        .route("/notifications/{id}/read", put(handlers::purchase::read_notification))
}
