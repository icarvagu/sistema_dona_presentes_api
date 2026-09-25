use axum::{
    extract::{Path, Query, State},
    Extension, Json,
};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::{
    PurchaseActionInput, PurchaseIssueInput, PurchaseIssueUpdateInput,
    PurchasePaymentInput, PurchaseReleaseInput, PurchaseUpdateInput, PurchaseRequestInput,
    PurchaseBatchInput,
    PurchaseRequestUpdateInput,
};
use crate::repositories;
use crate::response::{created_response, ok_response};
use crate::AppState;

// --- CRUD ---

pub async fn list(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let orders = repositories::purchase::get_all(&state.db).await?;
    Ok(ok_response(orders))
}

pub async fn list_requests(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    Ok(ok_response(repositories::purchase::get_requests(&state.db).await?))
}

pub async fn create_request(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Json(input): Json<PurchaseRequestInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    Ok(created_response(repositories::purchase::create_request(&state.db, &input, user.user_id).await?))
}

pub async fn update_request(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<PurchaseRequestUpdateInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    Ok(ok_response(repositories::purchase::update_request(&state.db, id, &input).await?))
}

pub async fn get_by_id(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let order = repositories::purchase::get_by_id(&state.db, id).await?;
    Ok(ok_response(order))
}

pub async fn get_financial(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let items = repositories::purchase::get_financial_summary(&state.db).await?;
    Ok(ok_response(items))
}

pub async fn get_financial_overview(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let overview = repositories::purchase::get_financial_overview(&state.db).await?;
    Ok(ok_response(overview))
}

pub async fn get_client_notes(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let data = repositories::purchase::get_client_notes(&state.db).await?;
    Ok(ok_response(data))
}

#[derive(Debug, serde::Deserialize)]
pub struct ClientNoteStatusInput {
    pub status: String,
    pub sent_email: Option<String>,
}

pub async fn update_client_note_status(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(sale_id): Path<i32>,
    Json(input): Json<ClientNoteStatusInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    if !matches!(input.status.as_str(), "financeiro" | "adm" | "enviadas") {
        return Err(AppError::validation("Status inválido para nota ao cliente."));
    }
    let data = repositories::purchase::update_client_note_status(
        &state.db, sale_id, &input.status, input.sent_email.as_deref().unwrap_or(""), user.user_id,
    ).await?;
    Ok(ok_response(data))
}

#[derive(Debug, serde::Deserialize, Default)]
pub struct IcmsCreditQuery {
    pub month: Option<String>,
}

pub async fn get_icms_credit(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Query(query): Query<IcmsCreditQuery>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let data = repositories::purchase::get_icms_credit(&state.db, query.month.as_deref()).await?;
    Ok(ok_response(data))
}

#[derive(Debug, serde::Deserialize)]
pub struct IcmsEntryInput {
    pub purchase_id: Option<i32>,
    pub production_order_id: Option<i64>,
    pub supplier_id: Option<i32>,
    pub invoice_number: String,
    #[serde(default)]
    pub invoice_key: String,
    pub issued_at: chrono::NaiveDateTime,
    pub tax_base: f64,
    pub aliquota: f64,
    pub icms_value: f64,
    #[serde(default)]
    pub xml_url: String,
}

pub async fn create_icms_entry(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Json(input): Json<IcmsEntryInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let data = repositories::purchase::create_icms_entry(&state.db, &input, user.user_id).await?;
    Ok(created_response(data))
}

pub async fn update(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<PurchaseUpdateInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let order = repositories::purchase::update(&state.db, id, &input).await?;
    Ok(ok_response(order))
}

pub async fn release_sale(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(sale_id): Path<i32>,
    Json(input): Json<PurchaseReleaseInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let order = repositories::purchase::release_sale_to_purchases(&state.db, sale_id, &input).await?;
    Ok(created_response(order))
}

pub async fn create_batch(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Json(input): Json<PurchaseBatchInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    Ok(created_response(repositories::purchase::create_purchase_batches(&state.db, &input, user.user_id).await?))
}

// --- Actions ---

pub async fn execute_action(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<PurchaseActionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let result = repositories::purchase::execute_action(
        &state.db, id, &input, user.user_id,
    ).await?;
    Ok(ok_response(result))
}

// --- Attachments ---

pub async fn add_attachment(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<PurchaseActionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let attachment = repositories::purchase::add_attachment(
        &state.db, id,
        &input.action,
        input.file_name.as_deref().unwrap_or(""),
        input.url.as_deref().unwrap_or(""),
        Some(user.user_id),
    ).await?;
    Ok(created_response(attachment))
}

// --- Payments ---

pub async fn add_payment(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<PurchasePaymentInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let payment = repositories::purchase::add_payment(
        &state.db, id, &input, user.user_id,
    ).await?;
    Ok(created_response(payment))
}

pub async fn approve_payment(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(payment_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let payment = repositories::purchase::approve_payment(
        &state.db, payment_id, user.user_id,
    ).await?;
    Ok(ok_response(payment))
}

// --- Issues ---

pub async fn add_issue(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<PurchaseIssueInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let issue = repositories::purchase::add_issue(
        &state.db, id, &input, user.user_id,
    ).await?;
    Ok(created_response(issue))
}

pub async fn update_issue(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(issue_id): Path<i32>,
    Json(input): Json<PurchaseIssueUpdateInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let issue = repositories::purchase::update_issue(
        &state.db, issue_id, &input,
    ).await?;
    Ok(ok_response(issue))
}

// --- Notifications ---

pub async fn list_notifications(
    State(state): State<AppState>,
    user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let notifications = repositories::purchase::get_notifications(
        &state.db, user.user_id,
    ).await?;
    Ok(ok_response(notifications))
}

pub async fn read_notification(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::purchase::mark_notification_read(&state.db, id).await?;
    Ok(ok_response(serde_json::json!({"message": "Notificação lida"})))
}
