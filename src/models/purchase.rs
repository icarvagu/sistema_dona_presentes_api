use chrono::{NaiveDate, NaiveDateTime};
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct PurchaseOrder {
    pub id: i32,
    pub sale_id: i32,
    pub general_number: String,
    pub status: String,
    pub status_updated_at: NaiveDateTime,
    pub buyer_id: Option<i32>,
    pub material_supplier_id: Option<i32>,
    pub engraving_supplier_id: Option<i32>,
    pub is_sample: bool,
    pub sample_has_engraving: bool,
    pub has_engraving: bool,
    pub corel_required: bool,
    pub corel_requested_at: Option<NaiveDateTime>,
    pub corel_requested_by: Option<i32>,
    pub corel_attached_at: Option<NaiveDateTime>,
    pub corel_attached_by: Option<i32>,
    pub material_unit_cost: f64,
    pub material_total_cost: f64,
    pub engraving_cost: f64,
    pub freight_cost: f64,
    pub other_cost: f64,
    pub buyer_discount: f64,
    pub negotiation_contact: Option<String>,
    pub negotiation_notes: Option<String>,
    pub material_deadline: Option<NaiveDateTime>,
    pub engraving_deadline: Option<NaiveDateTime>,
    pub payment_method: Option<String>,
    pub requires_advance_payment: bool,
    pub material_accepted: bool,
    pub material_accepted_at: Option<NaiveDateTime>,
    pub engraving_accepted: bool,
    pub engraving_accepted_at: Option<NaiveDateTime>,
    pub first_piece_required: bool,
    pub first_piece_status: Option<String>,
    pub first_piece_url: Option<String>,
    pub commercial_notes: Option<String>,
    pub purchase_notes: Option<String>,
    pub production_released_at: Option<NaiveDateTime>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
    #[serde(default)]
    pub sale: Option<Value>,

    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    #[sqlx(skip)]
    pub attachments: Vec<PurchaseAttachment>,

    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    #[sqlx(skip)]
    pub emails: Vec<PurchaseEmail>,

    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    #[sqlx(skip)]
    pub payments: Vec<PurchasePayment>,

    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    #[sqlx(skip)]
    pub issues: Vec<PurchaseIssue>,

    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    #[sqlx(skip)]
    pub history: Vec<PurchaseHistory>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct PurchaseAttachment {
    pub id: i32,
    pub purchase_id: i32,
    pub category: String,
    pub file_name: String,
    pub url: String,
    pub uploaded_by: Option<i32>,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct PurchaseEmail {
    pub id: i32,
    pub purchase_id: i32,
    pub kind: String,
    pub recipient: String,
    pub subject: String,
    pub body: String,
    pub observation: String,
    pub attachments: Option<Value>,
    pub sent_by: Option<i32>,
    pub sent_at: NaiveDateTime,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct PurchasePayment {
    pub id: i32,
    pub purchase_id: i32,
    pub cost_type: String,
    pub supplier_id: Option<i32>,
    pub amount: f64,
    pub method: String,
    pub status: String,
    pub justification: String,
    pub receipt_url: Option<String>,
    pub requested_by: Option<i32>,
    pub approved_by: Option<i32>,
    pub requested_at: NaiveDateTime,
    pub approved_at: Option<NaiveDateTime>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct PurchaseIssue {
    pub id: i32,
    pub purchase_id: i32,
    pub issue_type: String,
    pub description: String,
    pub attachments: Option<Value>,
    pub supplier_id: Option<i32>,
    pub solution: Option<String>,
    pub occurrence_date: NaiveDateTime,
    pub resolution_deadline: NaiveDateTime,
    pub priority: i32,
    pub status: String,
    pub opened_by: Option<i32>,
    pub resolved_by: Option<i32>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
    pub resolved_at: Option<NaiveDateTime>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct PurchaseHistory {
    pub id: i32,
    pub purchase_id: i32,
    pub action: String,
    pub from_status: String,
    pub to_status: String,
    pub details: Option<String>,
    pub user_id: Option<i32>,
    pub user_name: String,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
pub struct PurchaseReleaseInput {
    pub is_sample: Option<bool>,
    pub sample_has_engraving: Option<bool>,
}

#[derive(Debug, Deserialize)]
pub struct PurchaseUpdateInput {
    pub buyer_id: Option<i32>,
    pub material_supplier_id: Option<i32>,
    pub engraving_supplier_id: Option<i32>,
    pub is_sample: Option<bool>,
    pub sample_has_engraving: Option<bool>,
    pub has_engraving: Option<bool>,
    pub material_unit_cost: Option<f64>,
    pub material_total_cost: Option<f64>,
    pub engraving_cost: Option<f64>,
    pub freight_cost: Option<f64>,
    pub other_cost: Option<f64>,
    pub buyer_discount: Option<f64>,
    pub negotiation_contact: Option<String>,
    pub negotiation_notes: Option<String>,
    pub material_deadline: Option<NaiveDateTime>,
    pub engraving_deadline: Option<NaiveDateTime>,
    pub payment_method: Option<String>,
    pub requires_advance_payment: Option<bool>,
    pub first_piece_required: Option<bool>,
    pub commercial_notes: Option<String>,
    pub purchase_notes: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct PurchaseActionInput {
    pub action: String,
    pub observation: Option<String>,
    pub url: Option<String>,
    pub file_name: Option<String>,
    pub recipient: Option<String>,
    pub attachments: Option<Vec<String>>,
}

#[derive(Debug, Deserialize)]
pub struct PurchaseBatchInput {
    pub purchase_ids: Vec<i32>,
}

#[derive(Debug, Serialize)]
pub struct PurchaseBatchResult {
    pub batch_id: i32,
    pub supplier_id: Option<i32>,
    pub purchase_ids: Vec<i32>,
    pub items: Value,
}

#[derive(Debug, Deserialize)]
pub struct PurchasePaymentInput {
    pub cost_type: String,
    pub supplier_id: Option<i32>,
    pub amount: f64,
    pub method: String,
    pub justification: String,
    pub receipt_url: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct PurchaseIssueInput {
    pub issue_type: String,
    pub description: String,
    pub attachments: Option<Vec<String>>,
    pub supplier_id: Option<i32>,
    pub solution: Option<String>,
    pub occurrence_date: NaiveDateTime,
    pub resolution_deadline: NaiveDateTime,
    pub priority: i32,
}

#[derive(Debug, Deserialize)]
pub struct PurchaseIssueUpdateInput {
    pub status: String,
    pub solution: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct PurchaseRequest {
    pub id: i32,
    pub request_type: String,
    pub product: String,
    pub product_link: String,
    pub supplier: String,
    pub quantity: Option<f64>,
    pub description: String,
    pub attachments: Option<Value>,
    pub status: String,
    pub requested_by: Option<i32>,
    pub requester_name: Option<String>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
    pub items: Option<Value>,
    pub total_value: Option<f64>,
    pub freight_value: Option<f64>,
    pub payment_method: String,
    pub delivery_date: Option<NaiveDate>,
    pub receiver_name: String,
    pub receiver_phone: String,
    pub buyer_message: String,
}

#[derive(Debug, Deserialize)]
pub struct PurchaseRequestInput {
    pub request_type: String,
    pub product: String,
    pub product_link: Option<String>,
    pub supplier: Option<String>,
    pub quantity: Option<f64>,
    pub description: Option<String>,
    pub attachments: Option<Vec<Value>>,
}

#[derive(Debug, Deserialize)]
pub struct PurchaseRequestUpdateInput {
    pub product: Option<String>,
    pub supplier: Option<String>,
    pub quantity: Option<f64>,
    pub description: Option<String>,
    pub items: Option<Value>,
    pub total_value: Option<f64>,
    pub freight_value: Option<f64>,
    pub payment_method: Option<String>,
    pub delivery_date: Option<String>,
    pub receiver_name: Option<String>,
    pub receiver_phone: Option<String>,
    pub status: Option<String>,
    pub buyer_message: Option<String>,
}
