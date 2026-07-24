use axum::{
    extract::State,
    Extension,
};
use tracing::{error, info, info_span, Instrument};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::response::ok_response;
use crate::services;
use crate::AppState;

pub async fn sync_products_from_xbz(
    State(state): State<AppState>,
    _admin: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let span = info_span!("xbz_sync");
    let _enter = span.enter();

    let cnpj = state.config.xbz_cnpj.as_deref().unwrap_or("");
    let token = state.config.xbz_token.as_deref().unwrap_or("");

    if cnpj.is_empty() || token.is_empty() {
        return Err(AppError::bad_request("XBZ_CNPJ and XBZ_TOKEN are required"));
    }

    let xbz_svc = services::xbz::XBZService::new(cnpj.to_string(), token.to_string());

    match services::sync::SyncService::sync_xbz_to_local(&state.db, &xbz_svc).await {
        Ok(msg) => {
            info!("XBZ sync completed: {msg}");
            Ok(ok_response(serde_json::json!({"message": msg})))
        }
        Err(e) => {
            error!(error = %e, "XBZ sync failed");
            Err(e)
        }
    }
}

pub async fn generate_sale_pdf(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    axum::extract::Path(sale_id): axum::extract::Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let span = info_span!("pdf_generate", sale_id);
    async {
        let pdf = services::pdf::generate_order_pdf(&state.db, sale_id).await?;
        Ok(axum::response::Response::builder()
            .header("Content-Type", "application/pdf")
            .header(
                "Content-Disposition",
                format!("inline; filename=\"pedido_{sale_id}.pdf\""),
            )
            .body(axum::body::Body::from(pdf))
            .unwrap())
    }
    .instrument(span)
    .await
}

pub async fn generate_quote_pdf(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    axum::extract::Path(quote_id): axum::extract::Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let span = info_span!("pdf_quote_generate", quote_id);
    async {
        let pdf = services::pdf::generate_quote_pdf(&state.db, quote_id).await?;
        Ok(axum::response::Response::builder()
            .header("Content-Type", "application/pdf")
            .header(
                "Content-Disposition",
                format!("inline; filename=\"orcamento_{quote_id}.pdf\""),
            )
            .body(axum::body::Body::from(pdf))
            .unwrap())
    }
    .instrument(span)
    .await
}
