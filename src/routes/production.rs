use axum::{
    routing::{get, patch, post, put},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/production/dashboard", get(handlers::production::dashboard))
        .route("/production/orders", get(handlers::production::list))
        .route("/production/orders/{id}", get(handlers::production::get_by_id))
        .route("/production/orders/{id}/receipts", post(handlers::production::add_receipt))
        .route("/production/orders/{id}/occurrences", post(handlers::production::add_occurrence))
        .route("/production/orders/{id}/occurrences/{occurrenceId}/resolve", patch(handlers::production::resolve_occurrence))
        .route("/production/orders/{id}/transition", patch(handlers::production::transition))
        .route("/production/orders/{id}/assignment", patch(handlers::production::assign))
        .route("/production/orders/{id}/engraving-events", post(handlers::production::add_engraving_event))
        .route("/production/orders/{id}/volumes", post(handlers::production::add_volume))
        .route("/production/orders/{id}/fiscal", post(handlers::production::add_fiscal))
        .route("/production/orders/{id}/shipment", put(handlers::production::upsert_shipment))
        .route("/production/supplies", get(handlers::production::list_supplies).post(handlers::production::create_supply))
        .route("/production/supplies/movements", post(handlers::production::move_supply))
}
