use axum::{routing::get, Router};

use crate::{handlers, AppState};

pub fn public_router() -> Router<AppState> {
    Router::new()
        .route("/health", get(|| async { "OK" }))
        .route("/auth/login", axum::routing::post(handlers::auth::login))
        .route("/auth/forgot-password", axum::routing::post(handlers::auth::forgot_password))
        .route("/auth/reset-password", axum::routing::post(handlers::auth::reset_password))
}

pub fn protected_router() -> Router<AppState> {
    Router::new()
        .route("/auth/me", get(handlers::auth::get_current_user))
        .route("/auth/refresh", axum::routing::post(handlers::auth::refresh))
        .route("/auth/logout", axum::routing::post(handlers::auth::logout))
        .route("/auth/change-password", axum::routing::put(handlers::auth::change_password))
}
