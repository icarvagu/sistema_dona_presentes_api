use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Sale {
    pub id: i32,
    pub seller_id: i32,
    pub customer_id: i32,
    pub payment_method: Option<String>,
    pub installments: Option<i32>,
    pub payment_term_days: Option<i32>,
    pub first_installment_start: Option<NaiveDateTime>,
    pub installment_dates: Option<Value>,
    pub status: String,
    pub is_event: bool,
    pub delivery_address: Option<String>,
    pub delivery_date: Option<NaiveDateTime>,
    pub departure_date: Option<NaiveDateTime>,
    pub arrival_date: Option<NaiveDateTime>,
    pub priority: Option<String>,
    pub care_of: Option<String>,
    pub invoice_email: Option<String>,
    pub financial_email: Option<String>,
    pub purchase_order: Option<String>,
    pub external_notes: Option<String>,
    pub internal_notes: Option<String>,
    #[sqlx(default)]
    pub layout_urls: Vec<String>,
    pub total_value: f64,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[sqlx(skip)]
    pub items: Vec<SaleItem>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[sqlx(skip)]
    pub carriers: Vec<super::Carrier>,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct SaleItem {
    pub id: i32,
    pub sale_id: i32,
    pub product_id: i32,
    pub quantity: i32,
    pub unit_price: f64,
    pub total_price: f64,
    pub discount: f64,
    pub price_formation: Option<Value>,
    pub engravings: Option<Value>,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Deserialize)]
pub struct SaleInput {
    pub seller_id: i32,
    pub customer_id: i32,
    pub payment_method: Option<String>,
    pub installments: Option<i32>,
    pub payment_term_days: Option<i32>,
    pub first_installment_start: Option<NaiveDateTime>,
    pub installment_dates: Option<Value>,
    pub status: Option<String>,
    pub is_event: Option<bool>,
    pub delivery_address: Option<String>,
    pub delivery_date: Option<NaiveDateTime>,
    pub departure_date: Option<NaiveDateTime>,
    pub arrival_date: Option<NaiveDateTime>,
    pub priority: Option<String>,
    pub care_of: Option<String>,
    pub invoice_email: Option<String>,
    pub financial_email: Option<String>,
    pub purchase_order: Option<String>,
    pub external_notes: Option<String>,
    pub internal_notes: Option<String>,
    #[serde(default)]
    pub layout_urls: Vec<String>,
    #[serde(default)]
    pub items: Vec<SaleItemInput>,
    #[serde(default)]
    pub carrier_ids: Vec<i32>,
}

#[derive(Debug, Deserialize)]
pub struct SaleItemInput {
    pub product_id: i32,
    pub quantity: i32,
    pub unit_price: f64,
    pub engravings: Option<Value>,
    pub discount: Option<f64>,
}

#[derive(Debug, Deserialize)]
pub struct SaleLayoutInput {
    pub layout_urls: Vec<String>,
}

#[derive(Debug, Deserialize)]
pub struct QuoteFeedbackInput {
    pub feedback_datetime: Option<NaiveDateTime>,
    pub feedback_observation: Option<String>,
}
