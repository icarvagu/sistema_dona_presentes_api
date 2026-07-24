use axum::{
    extract::{Path, State},
    Extension, Json,
};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::{QuoteFeedbackInput, QuoteInput};
use crate::repositories;
use crate::response::{created_response, no_content, ok_response};
use crate::AppState;

pub async fn list(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let quotes = repositories::quote::get_all(&state.db).await?;
    Ok(ok_response(quotes))
}

pub async fn get_by_id(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let quote = repositories::quote::get_by_id(&state.db, id).await?;
    Ok(ok_response(quote))
}

pub async fn create(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<QuoteInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let quote = repositories::quote::create(&state.db, &input).await?;
    Ok(created_response(quote))
}

pub async fn update(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<QuoteInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let quote = repositories::quote::update(&state.db, id, &input).await?;
    Ok(ok_response(quote))
}

pub async fn delete(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::quote::delete(&state.db, id).await?;
    Ok(no_content())
}

pub async fn update_feedback(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<QuoteFeedbackInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query(
        "UPDATE quotes SET feedback_datetime = $1, feedback_observation = $2, updated_at = NOW()
         WHERE id = $3",
    )
    .bind(input.feedback_datetime)
    .bind(&input.feedback_observation)
    .bind(id)
    .execute(&state.db)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    let quote = repositories::quote::get_by_id(&state.db, id).await?;
    Ok(ok_response(quote))
}
