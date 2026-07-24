use axum::{http::StatusCode, response::IntoResponse, Json};
use serde::Serialize;

#[derive(Serialize)]
pub struct ApiResponse<T: Serialize> {
    pub data: T,
}

impl<T: Serialize> ApiResponse<T> {
    pub fn ok(data: T) -> (StatusCode, Json<Self>) {
        (StatusCode::OK, Json(Self { data }))
    }

    pub fn created(data: T) -> (StatusCode, Json<Self>) {
        (StatusCode::CREATED, Json(Self { data }))
    }
}

pub fn ok_response<T: Serialize>(data: T) -> impl IntoResponse {
    ApiResponse::ok(data)
}

pub fn created_response<T: Serialize>(data: T) -> impl IntoResponse {
    ApiResponse::created(data)
}

pub fn no_content() -> impl IntoResponse {
    StatusCode::NO_CONTENT
}
