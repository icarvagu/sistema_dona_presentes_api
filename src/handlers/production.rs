use axum::{
    extract::{Path, State},
    Extension, Json,
};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::{
    ProductionAssignmentInput, ProductionEventInput, ProductionFiscalInput,
    ProductionOccurrenceInput, ProductionOccurrenceResolutionInput,
    ProductionReceiptInput, ProductionShipmentInput, ProductionSupplyInput,
    ProductionSupplyMovementInput, ProductionTransitionInput, ProductionVolumeInput,
};
use crate::repositories;
use crate::response::{created_response, ok_response};
use crate::AppState;

pub async fn dashboard(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let data = repositories::production::dashboard(&state.db).await?;
    Ok(ok_response(data))
}

pub async fn list(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let orders = repositories::production::get_all(&state.db).await?;
    Ok(ok_response(orders))
}

pub async fn get_by_id(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let order = repositories::production::get_by_id(&state.db, id).await?;
    Ok(ok_response(order))
}

pub async fn add_receipt(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ProductionReceiptInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let result = repositories::production::add_receipt(&state.db, id, &input, user.user_id).await?;
    Ok(created_response(result))
}

pub async fn add_occurrence(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ProductionOccurrenceInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let result = repositories::production::add_occurrence(&state.db, id, &input, user.user_id).await?;
    Ok(created_response(result))
}

pub async fn resolve_occurrence(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path((id, occurrence_id)): Path<(i64, i64)>,
    Json(input): Json<ProductionOccurrenceResolutionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::production::resolve_occurrence(&state.db, id, occurrence_id, &input).await?;
    let order = repositories::production::get_by_id(&state.db, id).await?;
    Ok(ok_response(order))
}

pub async fn transition(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ProductionTransitionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let order = repositories::production::transition(&state.db, id, &input, user.user_id).await?;
    Ok(ok_response(order))
}

pub async fn assign(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ProductionAssignmentInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let order = repositories::production::assign(&state.db, id, input.owner_id, input.priority).await?;
    Ok(ok_response(order))
}

pub async fn add_engraving_event(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ProductionEventInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let result = repositories::production::add_engraving_event(&state.db, id, &input, user.user_id).await?;
    Ok(created_response(result))
}

pub async fn add_volume(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ProductionVolumeInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::production::add_volume(&state.db, id, &input).await?;
    let order = repositories::production::get_by_id(&state.db, id).await?;
    Ok(ok_response(order))
}

pub async fn add_fiscal(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ProductionFiscalInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::production::add_fiscal(&state.db, id, &input).await?;
    let order = repositories::production::get_by_id(&state.db, id).await?;
    Ok(ok_response(order))
}

pub async fn upsert_shipment(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ProductionShipmentInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::production::upsert_shipment(&state.db, id, &input).await?;
    let order = repositories::production::get_by_id(&state.db, id).await?;
    Ok(ok_response(order))
}

pub async fn list_supplies(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let supplies = repositories::production::list_supplies(&state.db).await?;
    Ok(ok_response(supplies))
}

pub async fn create_supply(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<ProductionSupplyInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let supply = repositories::production::create_supply(&state.db, &input).await?;
    Ok(created_response(supply))
}

pub async fn move_supply(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<ProductionSupplyMovementInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::production::move_supply(&state.db, &input).await?;
    Ok(ok_response(serde_json::json!({"message": "Movimentação registrada"})))
}
