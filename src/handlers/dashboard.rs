use axum::{
    extract::State,
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
    let sales_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM sales")
        .fetch_one(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let pending_sales: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM sales WHERE status NOT IN ('Concluido', 'Cancelado')"
    )
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    let quote_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM quotes")
        .fetch_one(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let product_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM products")
        .fetch_one(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let po_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM purchase_orders")
        .fetch_one(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let production_count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM production_orders")
        .fetch_one(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let pending_production: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM production_orders WHERE status NOT IN ('CONCLUIDO', 'EXPEDIDO', 'ENTREGUE')"
    )
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    let pending_products: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM products WHERE pending_approval = true"
    )
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    Ok(ok_response(json!({
        "sales": {
            "total": sales_count.0,
            "pending": pending_sales.0,
        },
        "quotes": {
            "total": quote_count.0,
        },
        "products": {
            "total": product_count.0,
            "pending_approval": pending_products.0,
        },
        "purchases": {
            "total": po_count.0,
        },
        "production": {
            "total": production_count.0,
            "pending": pending_production.0,
        },
    })))
}
