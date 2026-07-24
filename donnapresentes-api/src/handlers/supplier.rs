use axum::{
    extract::{Path, State},
    Extension, Json,
};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::SupplierInput;
use crate::repositories;
use crate::response::{created_response, no_content, ok_response};
use crate::AppState;

pub async fn list(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let suppliers = repositories::supplier::get_all(&state.db).await?;
    Ok(ok_response(suppliers))
}

pub async fn get_by_id(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let supplier = repositories::supplier::get_by_id(&state.db, id).await?;
    Ok(ok_response(supplier))
}

pub async fn create(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<SupplierInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let supplier = repositories::supplier::create(&state.db, &input).await?;
    Ok(created_response(supplier))
}

pub async fn update(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<SupplierInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let supplier = repositories::supplier::update(&state.db, id, &input).await?;
    Ok(ok_response(supplier))
}

pub async fn delete(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::supplier::delete(&state.db, id).await?;
    Ok(no_content())
}
