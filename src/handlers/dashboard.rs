use axum::{
    extract::{Path, State},
    Extension,
};
use serde_json::json;

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::response::ok_response;
use crate::AppState;

#[derive(Debug, Clone, Copy)]
struct DashboardCounts {
    customer_count: i64,
    product_count: i64,
    quote_total: i64,
    sales_month: i64,
    sales_finished: i64,
}

fn build_dashboard_payload(counts: DashboardCounts) -> serde_json::Value {
    json!({
        "cadastros": {
            "clientes": counts.customer_count,
            "produtos": counts.product_count,
        },
        "orcamentos": {
            "total": counts.quote_total,
            "por_situacao": { "vencidos": 0, "proximos": 0, "sem_data": 0 },
            "alertas_feedback": [],
        },
        "vendas": {
            "total_mes": counts.sales_month,
            "pedidos_finalizados": counts.sales_finished,
            "comissoes": 0.0,
            "faltam_para_meta": 0.0,
            "situacao_pedidos": [],
            "alertas_intervencao": [],
        },
    })
}

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

    Ok(ok_response(build_dashboard_payload(DashboardCounts {
        customer_count: customer_count.0,
        product_count: product_count.0,
        quote_total: quote_total.0,
        sales_month: sales_month.0,
        sales_finished: sales_finished.0,
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

    Ok(ok_response(build_dashboard_payload(DashboardCounts {
        customer_count: customer_count.0,
        product_count: product_count.0,
        quote_total: quote_total.0,
        sales_month: sales_month.0,
        sales_finished: sales_finished.0,
    })))
}

#[cfg(test)]
mod tests {
    use super::{build_dashboard_payload, DashboardCounts};

    #[test]
    fn dashboard_payload_has_expected_shape() {
        let payload = build_dashboard_payload(DashboardCounts {
            customer_count: 3,
            product_count: 4,
            quote_total: 5,
            sales_month: 6,
            sales_finished: 7,
        });

        assert_eq!(payload["cadastros"]["clientes"], 3);
        assert_eq!(payload["cadastros"]["produtos"], 4);
        assert_eq!(payload["orcamentos"]["total"], 5);
        assert_eq!(payload["vendas"]["total_mes"], 6);
        assert_eq!(payload["vendas"]["pedidos_finalizados"], 7);
        assert!(payload["orcamentos"]["alertas_feedback"].is_array());
        assert!(payload["vendas"]["situacao_pedidos"].is_array());
    }
}
