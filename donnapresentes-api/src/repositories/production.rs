use serde_json::{json, Value};

use crate::error::AppError;
use crate::models::{
    ProductionEventInput, ProductionFiscalInput,
    ProductionOccurrenceInput, ProductionOccurrenceResolutionInput,
    ProductionReceiptInput, ProductionShipmentInput, ProductionSupplyInput,
    ProductionSupplyMovementInput, ProductionTransitionInput, ProductionVolumeInput,
    ProductionOrder,
};

pub async fn dashboard(pool: &sqlx::PgPool) -> Result<Value, AppError> {
    let total: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM production_orders")
        .fetch_one(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let by_status: Vec<(String, i64)> = sqlx::query_as(
        "SELECT status, COUNT(*) as cnt FROM production_orders GROUP BY status ORDER BY cnt DESC"
    )
    .fetch_all(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    Ok(json!({
        "total_orders": total,
        "by_status": by_status.iter().map(|(s, c)| json!({"status": s, "count": c})).collect::<Vec<_>>()
    }))
}

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<ProductionOrder>, AppError> {
    sqlx::query_as::<_, ProductionOrder>(PRODUCTION_SELECT_ALL)
        .fetch_all(pool).await
        .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i64) -> Result<ProductionOrder, AppError> {
    sqlx::query_as::<_, ProductionOrder>(&format!("{PRODUCTION_SELECT_ALL} WHERE po.id = $1"))
        .bind(id).fetch_optional(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::not_found("Ordem de produção"))
}

pub async fn add_receipt(
    pool: &sqlx::PgPool, order_id: i64, input: &ProductionReceiptInput, user_id: i32,
) -> Result<Value, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let receipt_id: i64 = sqlx::query_scalar(
        "INSERT INTO production_receipts (production_order_id, idempotency_key, invoice_number,
                invoice_key, invoice_url, notes, received_by)
         VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id"
    )
    .bind(order_id).bind(&input.idempotency_key)
    .bind(input.invoice_number.as_deref().unwrap_or(""))
    .bind(input.invoice_key.as_deref().unwrap_or(""))
    .bind(input.invoice_url.as_deref().unwrap_or(""))
    .bind(input.notes.as_deref().unwrap_or(""))
    .bind(user_id)
    .fetch_one(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    for item in &input.items {
        sqlx::query(
            "INSERT INTO production_receipt_items (receipt_id, production_item_id, quantity)
             VALUES ($1,$2,$3)"
        )
        .bind(receipt_id).bind(item.item_id).bind(item.quantity)
        .execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    Ok(json!({"receipt_id": receipt_id}))
}

pub async fn add_occurrence(
    pool: &sqlx::PgPool, order_id: i64, input: &ProductionOccurrenceInput, user_id: i32,
) -> Result<Value, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO production_occurrences (production_order_id, production_item_id, kind,
                severity, quantity, description, attachment_url, reported_by)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id"
    )
    .bind(order_id).bind(input.item_id).bind(&input.kind)
    .bind(&input.severity).bind(input.quantity)
    .bind(&input.description).bind(&input.attachment_url).bind(user_id)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(json!({"occurrence_id": id}))
}

pub async fn resolve_occurrence(
    pool: &sqlx::PgPool, _order_id: i64, occurrence_id: i64,
    input: &ProductionOccurrenceResolutionInput,
) -> Result<(), AppError> {
    sqlx::query(
        "UPDATE production_occurrences SET resolution = $1, resolved_at = NOW() WHERE id = $2"
    )
    .bind(&input.resolution).bind(occurrence_id)
    .execute(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

pub async fn transition(
    pool: &sqlx::PgPool, id: i64, input: &ProductionTransitionInput, user_id: i32,
) -> Result<ProductionOrder, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let current: (String, i32) = sqlx::query_as(
        "SELECT status, version FROM production_orders WHERE id = $1"
    )
    .bind(id).fetch_one(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    if current.1 != input.expected_version {
        return Err(AppError::conflict("Versão desatualizada. Recarregue e tente novamente."));
    }

    sqlx::query(
        "UPDATE production_orders SET status = $1, version = version + 1, updated_at = NOW() WHERE id = $2"
    )
    .bind(&input.to_status).bind(id)
    .execute(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    if let Some(ref note) = input.note {
        if !note.is_empty() {
            sqlx::query(
                "INSERT INTO production_history (order_id, action, note, user_id)
                 VALUES ($1, $2, $3, $4)"
            )
            .bind(id).bind(&input.to_status).bind(note).bind(user_id)
            .execute(&mut *tx).await.ok();
        }
    }

    if input.to_status == "CONCLUIDO" {
        sqlx::query("UPDATE production_orders SET completed_at = NOW() WHERE id = $1")
            .bind(id).execute(&mut *tx).await.ok();
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, id).await
}

pub async fn assign(
    pool: &sqlx::PgPool, id: i64, owner_id: Option<i32>, priority: i32,
) -> Result<ProductionOrder, AppError> {
    sqlx::query("UPDATE production_orders SET owner_id = $1, priority = $2, updated_at = NOW() WHERE id = $3")
        .bind(owner_id).bind(priority).bind(id)
        .execute(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, id).await
}

pub async fn add_engraving_event(
    pool: &sqlx::PgPool, order_id: i64, input: &ProductionEventInput, user_id: i32,
) -> Result<Value, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO production_engraving_events (order_id, event_type, status, tracking_code,
                file_url, notes, idempotency_key, quantity, carrier_id, created_by)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id"
    )
    .bind(order_id).bind(&input.event_type).bind(&input.status)
    .bind(&input.tracking_code).bind(&input.file_url).bind(&input.notes)
    .bind(&input.idempotency_key).bind(input.quantity)
    .bind(input.carrier_id).bind(user_id)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(json!({"event_id": id}))
}

pub async fn add_volume(
    pool: &sqlx::PgPool, order_id: i64, input: &ProductionVolumeInput,
) -> Result<(), AppError> {
    sqlx::query(
        "INSERT INTO production_volumes (order_id, label, weight_kg, length_cm, width_cm, height_cm)
         VALUES ($1,$2,$3,$4,$5,$6)"
    )
    .bind(order_id).bind(&input.label).bind(input.weight_kg)
    .bind(input.length_cm).bind(input.width_cm).bind(input.height_cm)
    .execute(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

pub async fn add_fiscal(
    pool: &sqlx::PgPool, order_id: i64, input: &ProductionFiscalInput,
) -> Result<(), AppError> {
    sqlx::query(
        "INSERT INTO production_fiscal (order_id, document_type, document_number, access_key,
                file_url, issued_at)
         VALUES ($1,$2,$3,$4,$5,$6)"
    )
    .bind(order_id).bind(&input.document_type).bind(&input.document_number)
    .bind(&input.access_key).bind(&input.file_url).bind(input.issued_at)
    .execute(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

pub async fn upsert_shipment(
    pool: &sqlx::PgPool, order_id: i64, input: &ProductionShipmentInput,
) -> Result<(), AppError> {
    sqlx::query(
        "INSERT INTO production_shipments (order_id, method, carrier_id, driver_id,
                tracking_code, tracking_url, postal_service, proof_url)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
         ON CONFLICT (order_id) DO UPDATE SET
                method = EXCLUDED.method, carrier_id = EXCLUDED.carrier_id,
                driver_id = EXCLUDED.driver_id, tracking_code = EXCLUDED.tracking_code,
                tracking_url = EXCLUDED.tracking_url, postal_service = EXCLUDED.postal_service,
                proof_url = EXCLUDED.proof_url"
    )
    .bind(order_id).bind(&input.method).bind(input.carrier_id)
    .bind(input.driver_id).bind(&input.tracking_code).bind(&input.tracking_url)
    .bind(&input.postal_service).bind(&input.proof_url)
    .execute(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

#[derive(Debug, serde::Serialize, sqlx::FromRow)]
pub struct ProductionSupply {
    pub id: i64,
    pub name: String,
    pub unit: String,
    pub current_quantity: f64,
    pub minimum_quantity: f64,
    pub created_at: chrono::NaiveDateTime,
    pub updated_at: chrono::NaiveDateTime,
}

pub async fn list_supplies(pool: &sqlx::PgPool) -> Result<Vec<ProductionSupply>, AppError> {
    sqlx::query_as::<_, ProductionSupply>(
        "SELECT id, name, unit, current_quantity, minimum_quantity, created_at, updated_at
         FROM production_supplies ORDER BY name"
    )
    .fetch_all(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn create_supply(
    pool: &sqlx::PgPool, input: &ProductionSupplyInput,
) -> Result<ProductionSupply, AppError> {
    sqlx::query_as::<_, ProductionSupply>(
        "INSERT INTO production_supplies (name, unit, current_quantity, minimum_quantity)
         VALUES ($1,$2,$3,$4)
         RETURNING id, name, unit, current_quantity, minimum_quantity, created_at, updated_at"
    )
    .bind(&input.name).bind(&input.unit)
    .bind(input.current_quantity).bind(input.minimum_quantity)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn move_supply(
    pool: &sqlx::PgPool, input: &ProductionSupplyMovementInput,
) -> Result<(), AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let delta = if input.movement_type == "in" { input.quantity } else { -input.quantity };

    sqlx::query(
        "UPDATE production_supplies SET current_quantity = current_quantity + $1,
                updated_at = NOW() WHERE id = $2"
    )
    .bind(delta).bind(input.supply_id)
    .execute(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    sqlx::query(
        "INSERT INTO production_supply_movements (supply_id, order_id, movement_type, quantity,
                reason, supplier_id, invoice_number, invoice_url)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8)"
    )
    .bind(input.supply_id).bind(input.order_id).bind(&input.movement_type)
    .bind(input.quantity).bind(&input.reason).bind(input.supplier_id)
    .bind(&input.invoice_number).bind(&input.invoice_url)
    .execute(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

const PRODUCTION_SELECT_ALL: &str = "
    SELECT po.id, po.purchase_id, po.sale_id, po.status, po.has_engraving,
           po.first_piece_required, po.priority, po.owner_id, po.version,
           po.released_at, po.completed_at, po.created_at, po.updated_at,
           po.conference_started_at, po.conference_completed_at, po.conference_user_id,
           po.items, po.receipts, po.occurrences, po.engraving_events,
           po.volumes, po.fiscal_documents, po.shipment, po.history
    FROM production_orders po";
