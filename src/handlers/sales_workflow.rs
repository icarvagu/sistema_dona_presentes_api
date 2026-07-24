use axum::{
    extract::{Path, State},
    Extension, Json,
};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::{
    EngravingApprovalInput, FinancialAnalysisInput, QuoteConversionInput,
    QuoteFeedbackEvent, SalePendingInput,
};
use crate::response::{created_response, ok_response};
use crate::AppState;

// --- Quotes ---

pub async fn get_quote_feedback_events(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(quote_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let events = sqlx::query_as::<_, QuoteFeedbackEvent>(
        "SELECT id, quote_id, scheduled_at, observation, created_by, created_at
         FROM quote_feedback_events WHERE quote_id = $1 ORDER BY created_at DESC"
    )
    .bind(quote_id)
    .fetch_all(&state.db)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(events))
}

pub async fn add_quote_feedback_event(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(quote_id): Path<i32>,
    Json(input): Json<QuoteFeedbackEvent>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let event = sqlx::query_as::<_, QuoteFeedbackEvent>(
        "INSERT INTO quote_feedback_events (quote_id, scheduled_at, observation, created_by)
         VALUES ($1,$2,$3,$4)
         RETURNING id, quote_id, scheduled_at, observation, created_by, created_at"
    )
    .bind(quote_id).bind(input.scheduled_at).bind(&input.observation).bind(user.user_id)
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(event))
}

pub async fn convert_quote_to_sale(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(quote_id): Path<i32>,
    Json(input): Json<QuoteConversionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let mut tx = state.db.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let quote = sqlx::query_as::<_, crate::models::Quote>(
        "SELECT q.id, q.quote_number, q.seller_id, q.customer_id, q.responsible_name,
                q.quote_valid_until, q.production_lead_time, q.total_value,
                q.quote_date, q.care_of, q.sales_channel, q.observations,
                q.feedback_datetime, q.feedback_observation, q.payment_method,
                q.installments, q.installment_dates, q.carrier_id, q.freight_value,
                q.freight_tax_id_sender, q.freight_tax_id_origin, q.freight_tax_id_dest,
                q.freight_tax_id_payer, q.freight_tipo_transporte, q.freight_contato,
                q.freight_cidade_origem, q.freight_cidade_destino, q.freight_material,
                q.freight_tipo_frete, q.freight_produto, q.freight_tipo_embalagem,
                q.freight_quantidade, q.freight_volumes, q.freight_valor_nota,
                q.freight_peso_real, q.created_at, q.updated_at
         FROM quotes q WHERE q.id = $1"
    )
    .bind(quote_id)
    .fetch_optional(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?
    .ok_or_else(|| AppError::not_found("Orçamento"))?;

    let sale_id: i32 = sqlx::query_scalar(
        "INSERT INTO sales (seller_id, customer_id, payment_method, installments,
                first_installment_start, installment_dates, status, is_event,
                care_of, invoice_email, financial_email, total_value)
         VALUES ($1,$2,$3,$4,$5,$6,'Pendente',false,$7,'','',$8)
         RETURNING id"
    )
    .bind(quote.seller_id).bind(quote.customer_id)
    .bind(&quote.payment_method).bind(quote.installments.unwrap_or(1))
    .bind(&input.withdrawal_dates.values().next().copied().flatten())
    .bind(&quote.installment_dates)
    .bind(&quote.care_of).bind(quote.total_value)
    .fetch_one(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    let items = sqlx::query_as::<_, crate::models::QuoteItem>(
        "SELECT id, quote_id, product_id, quantity, unit_price, total_price,
                personalization_type, dn_code, description_summary, is_kit,
                base_cost_unit, labor_cost, extra_unit_cost1, extra_unit_cost2,
                engraving_cost, urgency_fee, logistics_cost, freight_cost,
                tax_percent, st_percent, loss_index_percent, imported_labor_percent,
                mgmt_commission_percent, seller_commission_percent,
                agency_commission_percent, publicity_percent, scrap_index,
                over_percent, financial_factor, financial_percent, sale_unit_value,
                transport_apart, production_cost_calc, transport_cost_calc,
                additional_costs_calc, profit_calc, margin_percent_calc,
                cost_unit_calc, engravings, has_price_formation, discount,
                created_at, updated_at
         FROM quote_items WHERE quote_id = $1"
    )
    .bind(quote_id)
    .fetch_all(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    for item in items {
        sqlx::query(
            "INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, total_price, discount, engravings)
             VALUES ($1,$2,$3,$4,$5,$6,$7)"
        )
        .bind(sale_id).bind(item.product_id).bind(item.quantity)
        .bind(item.unit_price).bind(item.total_price).bind(item.discount)
        .bind(&item.engravings)
        .execute(&mut *tx).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;

    let sale = crate::repositories::sale::get_by_id(&state.db, sale_id).await?;
    Ok(created_response(sale))
}

pub async fn set_quote_important(
    State(_state): State<AppState>,
    _user: Extension<UserContext>,
    Path(_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    Ok(ok_response(serde_json::json!({"message": "Quote importance not yet persisted"})))
}

// --- Sales ---

pub async fn dashboard(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let total: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM sales")
        .fetch_one(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let pending: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM sales WHERE status NOT IN ('Concluido','Cancelado')"
    ).fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;

    let quotes_active: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM quotes")
        .fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;

    let pending_events: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM sale_pending_events WHERE resolved_at IS NULL"
    ).fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;

    Ok(ok_response(serde_json::json!({
        "total_sales": total.0,
        "pending_sales": pending.0,
        "active_quotes": quotes_active.0,
        "pending_events": pending_events.0,
    })))
}

pub async fn get_sale_workflow(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(sale_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let sale = crate::repositories::sale::get_by_id(&state.db, sale_id).await?;
    let purchases = crate::repositories::purchase::get_all(&state.db).await?;
    let related = purchases.iter().find(|p| p.sale_id == sale_id);

    Ok(ok_response(serde_json::json!({
        "sale": sale,
        "purchase_order": related,
    })))
}

pub async fn upsert_financial_analysis(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(sale_id): Path<i32>,
    Json(input): Json<FinancialAnalysisInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let result = sqlx::query_scalar::<_, i32>(
        "SELECT id FROM sale_financial_analyses WHERE sale_id = $1"
    )
    .bind(sale_id)
    .fetch_optional(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    if result.is_some() {
        sqlx::query(
            "UPDATE sale_financial_analyses SET status=$1, tags=$2, observation=$3,
                    attachment_url=$4, updated_at=NOW() WHERE sale_id=$5"
        )
        .bind(&input.status).bind(&input.tags).bind(&input.observation)
        .bind(&input.attachment_url).bind(sale_id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    } else {
        sqlx::query(
            "INSERT INTO sale_financial_analyses (sale_id, status, tags, observation, attachment_url)
             VALUES ($1,$2,$3,$4,$5)"
        )
        .bind(sale_id).bind(&input.status).bind(&input.tags)
        .bind(&input.observation).bind(&input.attachment_url)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    }

    Ok(ok_response(serde_json::json!({"message": "Análise financeira salva"})))
}

pub async fn seller_approve(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(sale_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE sales SET status = 'Aprovado Vendedor', updated_at = NOW() WHERE id = $1")
        .bind(sale_id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Venda aprovada pelo vendedor"})))
}

pub async fn add_sale_receipt(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(sale_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO sale_receipts (sale_id, created_by) VALUES ($1,$2) RETURNING id"
    )
    .bind(sale_id).bind(user.user_id)
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"receipt_id": id})))
}

pub async fn validate_receipt(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(receipt_id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE sale_receipts SET validated = true, validated_at = NOW() WHERE id = $1")
        .bind(receipt_id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Recibo validado"})))
}

pub async fn add_pending_event(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(sale_id): Path<i32>,
    Json(input): Json<SalePendingInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO sale_pending_events (sale_id, sector, description, blocking, created_by)
         VALUES ($1,$2,$3,$4,$5) RETURNING id"
    )
    .bind(sale_id).bind(&input.sector).bind(&input.description)
    .bind(input.blocking).bind(user.user_id)
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"pending_id": id})))
}

pub async fn resolve_pending(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(event_id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query(
        "UPDATE sale_pending_events SET resolved_by = $1, resolved_at = NOW() WHERE id = $2"
    )
    .bind(user.user_id).bind(event_id)
    .execute(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Pendência resolvida"})))
}

pub async fn add_engraving_approval(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(item_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO engraving_approvals (sale_item_id, requested_by) VALUES ($1,$2) RETURNING id"
    )
    .bind(item_id).bind(user.user_id)
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"approval_id": id})))
}

pub async fn create_item_layout(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path((entity, item_id)): Path<(String, i32)>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO item_layouts (entity_type, item_id, created_by) VALUES ($1,$2,$3) RETURNING id"
    )
    .bind(&entity).bind(item_id).bind(user.user_id)
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"id": id})))
}

pub async fn approve_item_layout(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE item_layouts SET approval_status = 'approved', approved_by = $1, approved_at = NOW() WHERE id = $2")
        .bind(user.user_id).bind(id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Layout aprovado"})))
}

pub async fn record_engraving_approval(
    State(state): State<AppState>,
    user: Extension<UserContext>,
    Path(approval_id): Path<i64>,
    Json(input): Json<EngravingApprovalInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query(
        "UPDATE engraving_approvals SET response = $1, observation = $2,
                responded_by = $3, responded_at = NOW() WHERE id = $4"
    )
    .bind(&input.response).bind(&input.observation).bind(user.user_id).bind(approval_id)
    .execute(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Aprovação registrada"})))
}
