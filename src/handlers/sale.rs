use axum::{
    extract::{Path, State},
    Extension, Json,
};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::{SaleInput, SaleLayoutInput};
use crate::repositories;
use crate::response::{created_response, no_content, ok_response};
use crate::AppState;

pub async fn list(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let sales = repositories::sale::get_all(&state.db).await?;
    Ok(ok_response(sales))
}

pub async fn get_by_id(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let sale = repositories::sale::get_by_id(&state.db, id).await?;
    Ok(ok_response(sale))
}

pub async fn create(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<SaleInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let sale = repositories::sale::create(&state.db, &input).await?;
    Ok(created_response(sale))
}

pub async fn update(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<SaleInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let sale = repositories::sale::update(&state.db, id, &input).await?;
    Ok(ok_response(sale))
}

pub async fn delete(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::sale::delete(&state.db, id).await?;
    Ok(no_content())
}

pub async fn update_layout(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<SaleLayoutInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let sale = repositories::sale::update_layout(&state.db, id, &input.layout_urls).await?;
    Ok(ok_response(sale))
}
