use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Debug, Clone)]
pub struct AppError {
    pub code: StatusCode,
    pub message: String,
    pub details: String,
    pub log_message: String,
}

impl fmt::Display for AppError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.message)
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let body = ErrorBody {
            error: self.message,
            details: if self.details.is_empty() {
                None
            } else {
                Some(self.details)
            },
            code: self.code.as_u16(),
        };
        (self.code, Json(body)).into_response()
    }
}

impl AppError {
    pub fn new(code: StatusCode, message: impl Into<String>) -> Self {
        Self {
            code,
            message: message.into(),
            details: String::new(),
            log_message: String::new(),
        }
    }

    pub fn with_details(mut self, details: impl Into<String>) -> Self {
        self.details = details.into();
        self
    }

    pub fn with_log(mut self, log_message: impl Into<String>) -> Self {
        self.log_message = log_message.into();
        self
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ErrorBody {
    pub error: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub details: Option<String>,
    pub code: u16,
}

impl AppError {
    pub fn bad_request(message: impl Into<String>) -> Self {
        Self::new(StatusCode::BAD_REQUEST, message)
    }

    pub fn not_found(resource: impl Into<String>) -> Self {
        let r = resource.into();
        Self::new(StatusCode::NOT_FOUND, format!("{} não encontrado", r))
    }

    pub fn unauthorized(message: impl Into<String>) -> Self {
        Self::new(StatusCode::UNAUTHORIZED, message)
    }

    pub fn forbidden(message: impl Into<String>) -> Self {
        Self::new(StatusCode::FORBIDDEN, message)
    }

    pub fn conflict(message: impl Into<String>) -> Self {
        Self::new(StatusCode::CONFLICT, message)
    }

    pub fn internal(message: impl Into<String>) -> Self {
        Self::new(StatusCode::INTERNAL_SERVER_ERROR, message)
    }

    pub fn validation(message: impl Into<String>) -> Self {
        Self::new(StatusCode::BAD_REQUEST, format!("Validação falhou para: {}", message.into()))
    }

    pub fn missing_field(fields: &[&str]) -> Self {
        Self::new(
            StatusCode::BAD_REQUEST,
            format!("Campos obrigatórios não fornecidos: {}", fields.join(", ")),
        )
    }

    pub fn invalid_field(field: &str, reason: &str) -> Self {
        Self::new(
            StatusCode::BAD_REQUEST,
            format!("Campo '{}' inválido: {}", field, reason),
        )
    }
}
