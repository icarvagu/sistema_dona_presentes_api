use axum::{
    routing::{patch, post},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/item-layouts/{entity}/{itemId}", post(handlers::sales_workflow::create_item_layout))
        .route("/item-layouts/{id}/approval", patch(handlers::sales_workflow::approve_item_layout))
}
