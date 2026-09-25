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
        .route("/financial/overview", get(handlers::purchase::get_financial_overview))
        .route("/financial/client-notes", get(handlers::purchase::get_client_notes))
        .route("/financial/client-notes/{sale_id}/status", axum::routing::patch(handlers::purchase::update_client_note_status))
        .route("/financial/icms", get(handlers::purchase::get_icms_credit))
        .route("/financial/icms/entries", axum::routing::post(handlers::purchase::create_icms_entry))
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
