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

#[cfg(test)]
mod tests {
    use super::{created_response, no_content, ok_response};
    use axum::{
        body::to_bytes,
        http::StatusCode,
        response::IntoResponse,
    };
    use serde_json::json;

    #[tokio::test]
    async fn ok_response_returns_json_with_200() {
        let response = ok_response(json!({"value": 1})).into_response();
        assert_eq!(response.status(), StatusCode::OK);
        let body = to_bytes(response.into_body(), usize::MAX).await.unwrap();
        assert_eq!(body, r#"{"value":1}"#);
    }

    #[tokio::test]
    async fn created_response_returns_201() {
        let response = created_response(json!({"created": true})).into_response();
        assert_eq!(response.status(), StatusCode::CREATED);
    }

    #[tokio::test]
    async fn no_content_returns_204() {
        let response = no_content().into_response();
        assert_eq!(response.status(), StatusCode::NO_CONTENT);
    }
}
