use axum::{
    extract::{Path, Query, State},
    Extension, Json,
};
use serde::Deserialize;

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::ProductInput;
use crate::repositories;
use crate::response::{created_response, no_content, ok_response};
use crate::AppState;

#[derive(Debug, Deserialize, Default)]
pub struct ProductQuery {
    pub filter: Option<String>,
    pub page: Option<i32>,
    pub limit: Option<i32>,
    pub group: Option<String>,
    pub only_new: Option<bool>,
}

pub async fn list(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Query(query): Query<ProductQuery>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let page = query.page.unwrap_or(1);
    let limit = query.limit.unwrap_or(20);
    let filter = query.filter.unwrap_or_default();
    let result = repositories::product::get_paginated(
        &state.db,
        &filter,
        query.group.as_deref(),
        query.only_new.unwrap_or(false),
        page,
        limit,
    )
    .await?;
    Ok(ok_response(result))
}

pub async fn get_by_id(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let product = repositories::product::get_by_id(&state.db, id).await?;
    Ok(ok_response(product))
}

pub async fn create(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<ProductInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let product = repositories::product::create(&state.db, &input).await?;
    Ok(created_response(product))
}

pub async fn update(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<ProductInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let product = repositories::product::update(&state.db, id, &input).await?;
    Ok(ok_response(product))
}

pub async fn delete(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::product::delete(&state.db, id).await?;
    Ok(no_content())
}

pub async fn groups(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let groups = repositories::product::get_groups(&state.db).await?;
    Ok(ok_response(groups))
}

pub async fn pending_approval(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let products = repositories::product::get_pending_approval(&state.db).await?;
    Ok(ok_response(products))
}

pub async fn approve(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::product::approve(&state.db, id).await?;
    Ok(ok_response(serde_json::json!({"message": "Produto aprovado"})))
}

pub async fn financial_report(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let report = repositories::product::get_financial_report(&state.db).await?;
    Ok(ok_response(report))
}

#[derive(Debug, Deserialize)]
pub struct LastCostInput {
    pub cost: f64,
    pub qty1: Option<i32>,
    pub qty2: Option<i32>,
    pub qty3: Option<i32>,
    pub val1: Option<f64>,
    pub val2: Option<f64>,
    pub val3: Option<f64>,
}

pub async fn update_last_cost(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<LastCostInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let now = chrono::Utc::now().naive_utc();
    repositories::product::update_last_cost(
        &state.db, id, input.cost, now,
        input.qty1.unwrap_or(0), input.qty2.unwrap_or(0), input.qty3.unwrap_or(0),
        input.val1.unwrap_or(0.0), input.val2.unwrap_or(0.0), input.val3.unwrap_or(0.0),
        &user.username,
    ).await?;
    let product = repositories::product::get_by_id(&state.db, id).await?;
    Ok(ok_response(product))
}

#[derive(Debug, Deserialize)]
pub struct SearchQuery {
    pub q: Option<String>,
    pub limit: Option<i32>,
}

pub async fn search(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Query(query): Query<SearchQuery>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let results = repositories::product::search(
        &state.db,
        &query.q.unwrap_or_default(),
        query.limit.unwrap_or(50),
    ).await?;
    Ok(ok_response(results))
}

pub async fn get_items(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(parent_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let items = repositories::product::get_items(&state.db, parent_id).await?;
    Ok(ok_response(items))
}

pub async fn create_item(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(parent_id): Path<i32>,
    Json(input): Json<crate::models::ProductItemInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let item = repositories::product::create_item(&state.db, parent_id, &input).await?;
    Ok(created_response(item))
}

pub async fn delete_item(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path((_parent_id, item_id)): Path<(i32, i32)>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::product::delete_item(&state.db, item_id).await?;
    Ok(no_content())
}

pub async fn by_group(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(group): Path<String>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let products = repositories::product::get_by_group(&state.db, &group).await?;
    Ok(ok_response(products))
}

#[derive(Debug, Deserialize)]
pub struct BulkApproveInput {
    pub ids: Vec<i32>,
}

pub async fn bulk_approve(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<BulkApproveInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let count = repositories::product::bulk_approve(&state.db, &input.ids).await?;
    Ok(ok_response(serde_json::json!({"approved": count})))
}
