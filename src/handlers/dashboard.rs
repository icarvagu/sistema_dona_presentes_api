use axum::{
    extract::{Path, State},
    Extension,
};
use serde_json::json;

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::response::ok_response;
use crate::AppState;

pub async fn dashboard(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let customer_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM customers")
        .fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;
    let product_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM products")
        .fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;
    let quote_total: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM quotes")
        .fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;
    let sales_month: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM sales WHERE created_at >= date_trunc('month', NOW())"
    ).fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;
    let sales_finished: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM sales WHERE status IN ('Concluido','Cancelado')"
    ).fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;

    Ok(ok_response(json!({
        "cadastros": {
            "clientes": customer_count.0,
            "produtos": product_count.0,
        },
        "orcamentos": {
            "total": quote_total.0,
            "por_situacao": { "vencidos": 0, "proximos": 0, "sem_data": 0 },
            "alertas_feedback": [],
        },
        "vendas": {
            "total_mes": sales_month.0,
            "pedidos_finalizados": sales_finished.0,
            "comissoes": 0.0,
            "faltam_para_meta": 0.0,
            "situacao_pedidos": [],
            "alertas_intervencao": [],
        },
    })))
}

pub async fn seller_dashboard(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(seller_id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let customer_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM customers")
        .fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;
    let product_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM products")
        .fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;
    let quote_total: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM quotes WHERE seller_id = $1")
        .bind(seller_id).fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;
    let sales_month: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM sales WHERE seller_id = $1 AND created_at >= date_trunc('month', NOW())"
    ).bind(seller_id).fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;
    let sales_finished: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM sales WHERE seller_id = $1 AND status IN ('Concluido','Cancelado')"
    ).bind(seller_id).fetch_one(&state.db).await.map_err(|e| AppError::internal(e.to_string()))?;

    Ok(ok_response(json!({
        "cadastros": {
            "clientes": customer_count.0,
            "produtos": product_count.0,
        },
        "orcamentos": {
            "total": quote_total.0,
            "por_situacao": { "vencidos": 0, "proximos": 0, "sem_data": 0 },
            "alertas_feedback": [],
        },
        "vendas": {
            "total_mes": sales_month.0,
            "pedidos_finalizados": sales_finished.0,
            "comissoes": 0.0,
            "faltam_para_meta": 0.0,
            "situacao_pedidos": [],
            "alertas_intervencao": [],
        },
    })))
}
