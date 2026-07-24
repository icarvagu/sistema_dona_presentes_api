use axum::{
    routing::{delete, get, patch, post, put},
    Router,
};

use crate::{handlers, AppState};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/art-final/dashboard", get(handlers::art_final::dashboard))
        .route("/art-final/tasks", get(handlers::art_final::list_tasks).post(handlers::art_final::create_task))
        .route("/art-final/tasks/{id}", put(handlers::art_final::update_task))
        .route("/art-final/stories", post(handlers::art_final::create_story))
        .route("/art-final/stories/{id}/check", patch(handlers::art_final::check_story))
        .route("/art-final/stories/{id}", delete(handlers::art_final::delete_story))
        .route("/art-final/stories/{id}/lifecycle", patch(handlers::art_final::update_story_lifecycle))
        .route("/art-final/layout/requests", get(handlers::art_final::list_layout_requests)
               .post(handlers::art_final::create_layout_request))
        .route("/art-final/layout/requests/{id}", get(handlers::art_final::get_layout_request))
        .route("/art-final/layout/requests/{id}/transition", patch(handlers::art_final::transition_layout))
        .route("/art-final/layout/requests/{id}/messages", post(handlers::art_final::add_layout_message))
        .route("/art-final/layout/items/{itemId}/versions", post(handlers::art_final::add_layout_version))
        .route("/art-final/layout/versions/{versionId}/decision", patch(handlers::art_final::decide_layout))
        .route("/art-final/layout/items/{itemId}/jobs/{kind}", put(handlers::art_final::upsert_layout_job))
        .route("/art-final/layout/items/{itemId}/product-received", patch(handlers::art_final::confirm_layout_product))
}
