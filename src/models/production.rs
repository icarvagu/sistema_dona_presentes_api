use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct ProductionOrder {
    pub id: i64,
    pub purchase_id: i32,
    pub sale_id: i32,
    pub status: String,
    pub has_engraving: bool,
    pub first_piece_required: bool,
    pub priority: i16,
    pub owner_id: Option<i32>,
    pub version: i32,
    pub released_at: NaiveDateTime,
    pub completed_at: Option<NaiveDateTime>,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
    pub conference_started_at: Option<NaiveDateTime>,
    pub conference_completed_at: Option<NaiveDateTime>,
    pub conference_user_id: Option<i32>,
    pub items: Option<Value>,
    pub receipts: Option<Value>,
    pub occurrences: Option<Value>,
    pub engraving_events: Option<Value>,
    pub volumes: Option<Value>,
    pub fiscal_documents: Option<Value>,
    pub shipment: Option<Value>,
    pub history: Option<Value>,
}

#[derive(Debug, Deserialize)]
pub struct ProductionReceiptInput {
    pub idempotency_key: String,
    pub invoice_number: Option<String>,
    pub invoice_key: Option<String>,
    pub invoice_url: Option<String>,
    pub notes: Option<String>,
    pub items: Vec<ProductionReceiptItemInput>,
}

#[derive(Debug, Deserialize)]
pub struct ProductionReceiptItemInput {
    pub item_id: i64,
    pub quantity: i32,
}

#[derive(Debug, Deserialize)]
pub struct ProductionOccurrenceInput {
    pub item_id: Option<i64>,
    pub kind: String,
    pub severity: String,
    pub quantity: i32,
    pub description: String,
    pub attachment_url: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ProductionTransitionInput {
    pub to_status: String,
    pub note: Option<String>,
    pub expected_version: i32,
}

#[derive(Debug, Deserialize)]
pub struct ProductionEventInput {
    pub event_type: String,
    pub status: String,
    pub tracking_code: Option<String>,
    pub file_url: Option<String>,
    pub notes: Option<String>,
    pub idempotency_key: String,
    pub quantity: i32,
    pub carrier_id: Option<i32>,
}

#[derive(Debug, Deserialize)]
pub struct ProductionVolumeInput {
    pub label: String,
    pub weight_kg: f64,
    pub length_cm: f64,
    pub width_cm: f64,
    pub height_cm: f64,
}

#[derive(Debug, Deserialize)]
pub struct ProductionFiscalInput {
    pub document_type: String,
    pub document_number: String,
    pub access_key: Option<String>,
    pub file_url: Option<String>,
    pub issued_at: Option<NaiveDateTime>,
}

#[derive(Debug, Deserialize)]
pub struct ProductionShipmentInput {
    pub method: String,
    pub carrier_id: Option<i32>,
    pub driver_id: Option<i32>,
    pub tracking_code: Option<String>,
    pub tracking_url: Option<String>,
    pub postal_service: Option<String>,
    pub proof_url: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ProductionAssignmentInput {
    pub owner_id: Option<i32>,
    pub priority: i32,
}

#[derive(Debug, Deserialize)]
pub struct ProductionOccurrenceResolutionInput {
    pub resolution: String,
}

#[derive(Debug, Deserialize)]
pub struct ProductionSupplyInput {
    pub name: String,
    pub unit: String,
    pub current_quantity: f64,
    pub minimum_quantity: f64,
}

#[derive(Debug, Deserialize)]
pub struct ProductionSupplyMovementInput {
    pub supply_id: i64,
    pub order_id: Option<i64>,
    pub movement_type: String,
    pub quantity: f64,
    pub reason: String,
    pub supplier_id: Option<i32>,
    pub invoice_number: Option<String>,
    pub invoice_url: Option<String>,
}
