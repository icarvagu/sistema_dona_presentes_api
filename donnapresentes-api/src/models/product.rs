use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Product {
    pub id: i32,
    pub product_name: String,
    pub internal_code: String,
    pub supplier_code: Option<String>,
    pub supplier_id: Option<i32>,
    pub product_group: Option<String>,
    pub description: Option<String>,
    #[sqlx(default)]
    pub photos: Vec<String>,
    pub ncm: Option<String>,
    pub material_origin: Option<String>,
    pub stock: i32,
    pub supplier_stock: i32,
    pub moves_stock: bool,
    pub enabled_for_invoice: bool,
    pub cost_price: f64,
    pub selling_price: f64,
    pub kit_type: String,
    pub is_composition: bool,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[sqlx(skip)]
    pub items: Vec<ProductItem>,
    pub source: Option<String>,
    pub imported_at: Option<NaiveDateTime>,
    pub last_synced_at: Option<NaiveDateTime>,
    pub color: Option<String>,
    pub origin: Option<String>,
    pub pending_approval: bool,
    pub last_cost: Option<f64>,
    pub last_cost_date: Option<NaiveDateTime>,
    pub last_cost_qty1: Option<i32>,
    pub last_cost_qty2: Option<i32>,
    pub last_cost_qty3: Option<i32>,
    pub last_cost_val1: Option<f64>,
    pub last_cost_val2: Option<f64>,
    pub last_cost_val3: Option<f64>,
    pub last_cost_user: Option<String>,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Deserialize)]
pub struct ProductInput {
    pub product_name: String,
    pub internal_code: String,
    pub supplier_code: Option<String>,
    pub supplier_id: Option<i32>,
    pub product_group: Option<String>,
    pub description: Option<String>,
    #[serde(default)]
    pub photos: Vec<String>,
    pub ncm: Option<String>,
    pub material_origin: Option<String>,
    pub stock: Option<i32>,
    pub supplier_stock: Option<i32>,
    pub moves_stock: Option<bool>,
    pub enabled_for_invoice: Option<bool>,
    pub cost_price: Option<f64>,
    pub selling_price: Option<f64>,
    pub kit_type: Option<String>,
    pub is_composition: Option<bool>,
    #[serde(default)]
    pub items: Vec<ProductItemInput>,
    pub color: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct PaginatedProductResponse {
    pub data: Vec<Product>,
    pub total: i64,
    pub page: i32,
    pub limit: i32,
    pub total_pages: i32,
}

use super::product_item::{ProductItem, ProductItemInput};
