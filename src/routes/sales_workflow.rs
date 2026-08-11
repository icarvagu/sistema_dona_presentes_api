use axum::{
    routing::{delete, get, patch, post, put},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/sales-workflow/dashboard", get(handlers::sales_workflow::dashboard))
        .route("/sales-workflow/alerts", get(handlers::sales_workflow::list_workflow_alerts)
               .post(handlers::sales_workflow::create_workflow_alert))
        .route("/sales-workflow/alerts/{id}", delete(handlers::sales_workflow::delete_workflow_alert))
        .route("/sales-workflow/alerts/{id}/resolve", patch(handlers::sales_workflow::resolve_workflow_alert))
        .route("/quotes/{id}/feedback-events", get(handlers::sales_workflow::get_quote_feedback_events)
               .post(handlers::sales_workflow::add_quote_feedback_event))
        .route("/quotes/{id}/convert", post(handlers::sales_workflow::convert_quote_to_sale))
        .route("/quotes/{id}/important", patch(handlers::sales_workflow::set_quote_important))
        .route("/sales/{id}/financial-analysis", put(handlers::sales_workflow::upsert_financial_analysis))
        .route("/sales/{id}/workflow", get(handlers::sales_workflow::get_sale_workflow))
        .route("/sales/{id}/receipts", post(handlers::sales_workflow::add_sale_receipt))
        .route("/sales/receipts/{id}/validation", patch(handlers::sales_workflow::validate_receipt))
        .route("/sales/{id}/pending-events", post(handlers::sales_workflow::add_pending_event))
        .route("/sales/pending-events/{id}/resolve", patch(handlers::sales_workflow::resolve_pending))
        .route("/sale-items/{itemId}/engraving-approvals", post(handlers::sales_workflow::add_engraving_approval))
        .route("/engraving-approvals/{id}/record", patch(handlers::sales_workflow::record_engraving_approval))
        .route("/sales/{id}/seller-approval", patch(handlers::sales_workflow::seller_approve))
}
