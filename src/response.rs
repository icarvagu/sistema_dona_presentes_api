use axum::{http::StatusCode, response::IntoResponse, Json};
use serde::Serialize;

pub fn ok_response<T: Serialize>(data: T) -> impl IntoResponse {
    (StatusCode::OK, Json(data))
}

pub fn created_response<T: Serialize>(data: T) -> impl IntoResponse {
    (StatusCode::CREATED, Json(data))
}

pub fn no_content() -> impl IntoResponse {
    StatusCode::NO_CONTENT
}
