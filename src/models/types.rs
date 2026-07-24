use serde::Serialize;

#[derive(Debug, Serialize, sqlx::FromRow)]
pub struct FinancialReportItem {
    pub product_id: i32,
    pub product_name: String,
    pub internal_code: String,
    pub kit_type: String,
    pub cost_price: f64,
    pub selling_price: f64,
    pub margin_value: f64,
    pub margin_percent: f64,
    pub moves_stock: bool,
    pub enabled_for_invoice: bool,
}
